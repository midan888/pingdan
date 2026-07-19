package pinger

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/pingdan/api/internal/endpoints"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

// maxBodyRead caps how much of an HTTP response body is retained for
// assertion evaluation.
const maxBodyRead = 1 << 20 // 1 MiB

// ProbeResult is the protocol-independent result consumed by the scheduler.
// HTTP probes also populate status, headers, and body for assertion checks.
type ProbeResult struct {
	StatusCode *int
	LatencyMs  *int
	Headers    map[string]string
	Body       []byte
}

// Probe performs one network check. A nil error means that the protocol-level
// operation succeeded; the scheduler applies HTTP status/assertion rules.
type Probe interface {
	Run(context.Context, endpoints.Endpoint) (ProbeResult, error)
}

func defaultProbes() map[string]Probe {
	return map[string]Probe{
		endpoints.CheckTypeHTTP: httpProbe{client: sharedHTTPClient},
		endpoints.CheckTypeTCP:  tcpProbe{},
		endpoints.CheckTypeICMP: icmpProbe{},
	}
}

var sharedHTTPClient = &http.Client{
	// Per-request timeout is enforced via context; these settings keep the
	// shared connection pool bounded across all HTTP monitors.
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 4,
		IdleConnTimeout:     90 * time.Second,
	},
}

type httpProbe struct{ client *http.Client }

func (p httpProbe) Run(ctx context.Context, ep endpoints.Endpoint) (ProbeResult, error) {
	req, err := http.NewRequestWithContext(ctx, ep.Method, ep.URL, nil)
	if err != nil {
		return ProbeResult{}, err
	}
	req.Header.Set("User-Agent", "pingdan/1.0")

	start := time.Now()
	resp, err := p.client.Do(req)
	latency := int(time.Since(start).Milliseconds())
	result := ProbeResult{LatencyMs: &latency}
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBodyRead))
	if err != nil {
		return result, fmt.Errorf("read HTTP response: %w", err)
	}
	result.StatusCode = &resp.StatusCode
	result.Headers = flattenHeaders(resp.Header)
	result.Body = body
	return result, nil
}

type tcpProbe struct{}

func (tcpProbe) Run(ctx context.Context, ep endpoints.Endpoint) (ProbeResult, error) {
	u, err := url.Parse(ep.URL)
	if err != nil {
		return ProbeResult{}, fmt.Errorf("parse TCP target: %w", err)
	}
	address := net.JoinHostPort(u.Hostname(), u.Port())

	start := time.Now()
	conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", address)
	latency := int(time.Since(start).Milliseconds())
	result := ProbeResult{LatencyMs: &latency}
	if err != nil {
		return result, fmt.Errorf("TCP connect: %w", err)
	}
	if err := conn.Close(); err != nil {
		return result, fmt.Errorf("close TCP connection: %w", err)
	}
	return result, nil
}

type icmpProbe struct{}

var icmpSequence atomic.Uint32

func (icmpProbe) Run(ctx context.Context, ep endpoints.Endpoint) (ProbeResult, error) {
	u, err := url.Parse(ep.URL)
	if err != nil {
		return ProbeResult{}, fmt.Errorf("parse ICMP target: %w", err)
	}
	ip, err := resolvePingIP(ctx, u.Hostname())
	if err != nil {
		return ProbeResult{}, err
	}

	start := time.Now()
	err = pingIP(ctx, ip)
	latency := int(time.Since(start).Milliseconds())
	result := ProbeResult{LatencyMs: &latency}
	if err != nil {
		return result, err
	}
	return result, nil
}

func resolvePingIP(ctx context.Context, host string) (net.IP, error) {
	if ip := net.ParseIP(host); ip != nil {
		return ip, nil
	}
	addrs, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, fmt.Errorf("resolve ICMP target: %w", err)
	}
	// Prefer IPv4 when both families are available. It is more consistently
	// supported by container networking, while literal IPv6 targets still work.
	for _, addr := range addrs {
		if ip := addr.IP.To4(); ip != nil {
			return ip, nil
		}
	}
	if len(addrs) > 0 {
		return addrs[0].IP, nil
	}
	return nil, fmt.Errorf("resolve ICMP target: no addresses found")
}

func pingIP(ctx context.Context, ip net.IP) error {
	isIPv4 := ip.To4() != nil
	rawNetwork, packetNetwork, listenAddress := "ip6:ipv6-icmp", "udp6", "::"
	protocol := 58
	requestType, replyType := icmp.Type(ipv6.ICMPTypeEchoRequest), icmp.Type(ipv6.ICMPTypeEchoReply)
	var destination net.Addr = &net.IPAddr{IP: ip}
	if isIPv4 {
		rawNetwork, packetNetwork, listenAddress = "ip4:icmp", "udp4", "0.0.0.0"
		protocol = 1
		requestType, replyType = ipv4.ICMPTypeEcho, ipv4.ICMPTypeEchoReply
	}

	// Raw ICMP works in production when the container has NET_RAW. Developer
	// machines commonly deny raw sockets, so fall back to the OS's unprivileged
	// ping socket where it is enabled.
	conn, rawErr := icmp.ListenPacket(rawNetwork, listenAddress)
	if rawErr != nil {
		conn, rawErr = icmp.ListenPacket(packetNetwork, listenAddress)
		if rawErr != nil {
			return fmt.Errorf("open ICMP socket: %w", rawErr)
		}
		destination = &net.UDPAddr{IP: ip}
	}
	defer conn.Close()

	deadline := time.Now().Add(10 * time.Second)
	if contextDeadline, ok := ctx.Deadline(); ok {
		deadline = contextDeadline
	}
	if err := conn.SetDeadline(deadline); err != nil {
		return fmt.Errorf("set ICMP deadline: %w", err)
	}

	sequence := int(icmpSequence.Add(1) & 0xffff)
	payload := []byte("pingdan-icmp")
	message := icmp.Message{
		Type: requestType,
		Code: 0,
		Body: &icmp.Echo{ID: os.Getpid() & 0xffff, Seq: sequence, Data: payload},
	}
	wire, err := message.Marshal(nil)
	if err != nil {
		return fmt.Errorf("encode ICMP request: %w", err)
	}
	if _, err := conn.WriteTo(wire, destination); err != nil {
		return fmt.Errorf("send ICMP request: %w", err)
	}

	buffer := make([]byte, 1500)
	for {
		n, _, err := conn.ReadFrom(buffer)
		if err != nil {
			return fmt.Errorf("receive ICMP reply: %w", err)
		}
		reply, err := icmp.ParseMessage(protocol, buffer[:n])
		if err != nil || reply.Type != replyType {
			continue
		}
		echo, ok := reply.Body.(*icmp.Echo)
		if ok && echo.Seq == sequence && bytes.Equal(echo.Data, payload) {
			return nil
		}
	}
}

func flattenHeaders(h http.Header) map[string]string {
	out := make(map[string]string, len(h))
	for k, v := range h {
		if len(v) > 0 {
			out[strings.ToLower(k)] = v[0]
		}
	}
	return out
}
