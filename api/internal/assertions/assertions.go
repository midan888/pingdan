package assertions

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Sources an assertion can read from.
const (
	SourceStatusCode   = "status_code"
	SourceResponseTime = "response_time"
	SourceHeader       = "header"
	SourceBody         = "body"
	SourceJSONPath     = "json_path"
)

// Comparisons supported by an assertion.
const (
	CmpEquals      = "equals"
	CmpNotEquals   = "not_equals"
	CmpGreaterThan = "greater_than"
	CmpLessThan    = "less_than"
	CmpContains    = "contains"
	CmpNotContains = "not_contains"
	CmpMatches     = "matches" // regular expression
)

type Assertion struct {
	ID         int64  `json:"id"`
	EndpointID string `json:"endpointId"`
	Source     string `json:"source"`
	Property   string `json:"property"`
	Comparison string `json:"comparison"`
	Target     string `json:"target"`
	SortOrder  int    `json:"sortOrder"`
}

// Validate checks the assertion is well-formed before it is persisted.
func (a *Assertion) Validate() error {
	switch a.Source {
	case SourceStatusCode, SourceResponseTime, SourceBody, SourceHeader, SourceJSONPath:
	default:
		return fmt.Errorf("invalid source %q", a.Source)
	}
	if (a.Source == SourceHeader || a.Source == SourceJSONPath) && strings.TrimSpace(a.Property) == "" {
		return fmt.Errorf("%s assertion requires a property", a.Source)
	}
	switch a.Comparison {
	case CmpEquals, CmpNotEquals, CmpGreaterThan, CmpLessThan, CmpContains, CmpNotContains, CmpMatches:
	default:
		return fmt.Errorf("invalid comparison %q", a.Comparison)
	}
	if a.Comparison == CmpMatches {
		if _, err := regexp.Compile(a.Target); err != nil {
			return fmt.Errorf("invalid regex: %v", err)
		}
	}
	return nil
}

// Result describes the outcome of evaluating one assertion against a response.
type Result struct {
	Source     string `json:"source"`
	Property   string `json:"property,omitempty"`
	Comparison string `json:"comparison"`
	Target     string `json:"target"`
	Actual     string `json:"actual"`
	Passed     bool   `json:"passed"`
}

// Response is the captured data an assertion is evaluated against.
type Response struct {
	StatusCode int
	LatencyMs  int
	Headers    map[string]string // canonicalised header name -> first value
	Body       []byte
}

// Evaluate runs the assertion against the response and returns the result.
func (a *Assertion) Evaluate(resp Response) Result {
	actual := a.extract(resp)
	res := Result{
		Source: a.Source, Property: a.Property, Comparison: a.Comparison,
		Target: a.Target, Actual: actual,
	}
	res.Passed = compare(actual, a.Comparison, a.Target)
	return res
}

func (a *Assertion) extract(resp Response) string {
	switch a.Source {
	case SourceStatusCode:
		return strconv.Itoa(resp.StatusCode)
	case SourceResponseTime:
		return strconv.Itoa(resp.LatencyMs)
	case SourceHeader:
		return resp.Headers[strings.ToLower(a.Property)]
	case SourceBody:
		return string(resp.Body)
	case SourceJSONPath:
		return jsonPath(resp.Body, a.Property)
	}
	return ""
}

func compare(actual, comparison, target string) bool {
	switch comparison {
	case CmpEquals:
		return actual == target
	case CmpNotEquals:
		return actual != target
	case CmpContains:
		return strings.Contains(actual, target)
	case CmpNotContains:
		return !strings.Contains(actual, target)
	case CmpMatches:
		re, err := regexp.Compile(target)
		return err == nil && re.MatchString(actual)
	case CmpGreaterThan, CmpLessThan:
		af, aerr := strconv.ParseFloat(strings.TrimSpace(actual), 64)
		tf, terr := strconv.ParseFloat(strings.TrimSpace(target), 64)
		if aerr != nil || terr != nil {
			return false
		}
		if comparison == CmpGreaterThan {
			return af > tf
		}
		return af < tf
	}
	return false
}

// jsonPath resolves a dotted path like "data.items.0.name" against a JSON body.
// Returns the value as a string (objects/arrays are re-marshalled).
func jsonPath(body []byte, path string) string {
	var v any
	if err := json.Unmarshal(body, &v); err != nil {
		return ""
	}
	for _, seg := range strings.Split(path, ".") {
		seg = strings.TrimSpace(seg)
		if seg == "" || seg == "$" {
			continue
		}
		switch cur := v.(type) {
		case map[string]any:
			next, ok := cur[seg]
			if !ok {
				return ""
			}
			v = next
		case []any:
			idx, err := strconv.Atoi(seg)
			if err != nil || idx < 0 || idx >= len(cur) {
				return ""
			}
			v = cur[idx]
		default:
			return ""
		}
	}
	switch val := v.(type) {
	case string:
		return val
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(val)
	case nil:
		return ""
	default:
		b, _ := json.Marshal(val)
		return string(b)
	}
}

// ----- Store -----

type Store struct{ Pool *pgxpool.Pool }

// ListForEndpoint returns assertions ordered for display and evaluation.
func (s *Store) ListForEndpoint(ctx context.Context, endpointID string) ([]Assertion, error) {
	rows, err := s.Pool.Query(ctx, `
		SELECT id, endpoint_id, source, property, comparison, target, sort_order
		FROM assertions WHERE endpoint_id=$1 ORDER BY sort_order, id
	`, endpointID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Assertion{}
	for rows.Next() {
		var a Assertion
		if err := rows.Scan(&a.ID, &a.EndpointID, &a.Source, &a.Property, &a.Comparison, &a.Target, &a.SortOrder); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// Replace swaps the full assertion set for an endpoint in a single transaction.
func (s *Store) Replace(ctx context.Context, endpointID string, list []Assertion) error {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM assertions WHERE endpoint_id=$1`, endpointID); err != nil {
		return err
	}
	for i, a := range list {
		if _, err := tx.Exec(ctx, `
			INSERT INTO assertions (endpoint_id, source, property, comparison, target, sort_order)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, endpointID, a.Source, a.Property, a.Comparison, a.Target, i); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
