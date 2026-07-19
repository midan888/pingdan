FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/server ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata libcap && adduser -D -u 10001 app
COPY --from=build /out/server /usr/local/bin/server
# The service remains non-root; only its executable may open the raw socket
# needed by ICMP monitors. TCP and HTTP monitors need no extra privilege.
RUN setcap cap_net_raw=ep /usr/local/bin/server
USER app
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/server"]
