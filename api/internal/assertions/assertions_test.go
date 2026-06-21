package assertions

import "testing"

func resp() Response {
	return Response{
		StatusCode: 200,
		LatencyMs:  150,
		Headers:    map[string]string{"content-type": "application/json; charset=utf-8"},
		Body:       []byte(`{"status":"ok","count":3,"items":[{"name":"a"},{"name":"b"}]}`),
	}
}

func TestEvaluate(t *testing.T) {
	cases := []struct {
		name string
		a    Assertion
		want bool
	}{
		{"status equals", Assertion{Source: SourceStatusCode, Comparison: CmpEquals, Target: "200"}, true},
		{"status not equals fails", Assertion{Source: SourceStatusCode, Comparison: CmpEquals, Target: "404"}, false},
		{"response time under", Assertion{Source: SourceResponseTime, Comparison: CmpLessThan, Target: "500"}, true},
		{"response time over fails", Assertion{Source: SourceResponseTime, Comparison: CmpLessThan, Target: "100"}, false},
		{"header contains", Assertion{Source: SourceHeader, Property: "Content-Type", Comparison: CmpContains, Target: "application/json"}, true},
		{"header case-insensitive name", Assertion{Source: SourceHeader, Property: "content-type", Comparison: CmpContains, Target: "charset"}, true},
		{"body contains", Assertion{Source: SourceBody, Comparison: CmpContains, Target: `"status":"ok"`}, true},
		{"body matches regex", Assertion{Source: SourceBody, Comparison: CmpMatches, Target: `"count":\d+`}, true},
		{"json path string", Assertion{Source: SourceJSONPath, Property: "status", Comparison: CmpEquals, Target: "ok"}, true},
		{"json path number", Assertion{Source: SourceJSONPath, Property: "count", Comparison: CmpEquals, Target: "3"}, true},
		{"json path nested array", Assertion{Source: SourceJSONPath, Property: "items.1.name", Comparison: CmpEquals, Target: "b"}, true},
		{"json path missing fails", Assertion{Source: SourceJSONPath, Property: "nope", Comparison: CmpNotEquals, Target: ""}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.a.Evaluate(resp()).Passed; got != c.want {
				t.Errorf("Evaluate() = %v, want %v", got, c.want)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	bad := []Assertion{
		{Source: "bogus", Comparison: CmpEquals, Target: "x"},
		{Source: SourceHeader, Comparison: CmpEquals, Target: "x"}, // missing property
		{Source: SourceBody, Comparison: "bogus", Target: "x"},
		{Source: SourceBody, Comparison: CmpMatches, Target: "("}, // bad regex
	}
	for i, a := range bad {
		if err := a.Validate(); err == nil {
			t.Errorf("case %d: expected validation error, got nil", i)
		}
	}
	if err := (&Assertion{Source: SourceStatusCode, Comparison: CmpEquals, Target: "200"}).Validate(); err != nil {
		t.Errorf("valid assertion rejected: %v", err)
	}
}
