"use client";

import type { Assertion, AssertionComparison, AssertionSource } from "@/lib/api";

const SOURCES: { value: AssertionSource; label: string; needsProp?: boolean; propPlaceholder?: string }[] = [
  { value: "status_code", label: "Status code" },
  { value: "response_time", label: "Response time (ms)" },
  { value: "header", label: "Header", needsProp: true, propPlaceholder: "Content-Type" },
  { value: "body", label: "Body" },
  { value: "json_path", label: "JSON path", needsProp: true, propPlaceholder: "data.status" },
];

// Which comparisons make sense for each source.
const COMPARISONS: Record<AssertionSource, { value: AssertionComparison; label: string }[]> = {
  status_code: [
    { value: "equals", label: "equals" },
    { value: "not_equals", label: "not equals" },
    { value: "less_than", label: "less than" },
    { value: "greater_than", label: "greater than" },
  ],
  response_time: [
    { value: "less_than", label: "less than" },
    { value: "greater_than", label: "greater than" },
  ],
  header: allComparisons(),
  body: [
    { value: "contains", label: "contains" },
    { value: "not_contains", label: "not contains" },
    { value: "matches", label: "matches regex" },
    { value: "equals", label: "equals" },
  ],
  json_path: allComparisons(),
};

function allComparisons(): { value: AssertionComparison; label: string }[] {
  return [
    { value: "equals", label: "equals" },
    { value: "not_equals", label: "not equals" },
    { value: "contains", label: "contains" },
    { value: "not_contains", label: "not contains" },
    { value: "greater_than", label: "greater than" },
    { value: "less_than", label: "less than" },
    { value: "matches", label: "matches regex" },
  ];
}

export function newAssertion(): Assertion {
  return { source: "status_code", property: "", comparison: "equals", target: "200" };
}

export function AssertionBuilder({
  value,
  onChange,
}: {
  value: Assertion[];
  onChange: (next: Assertion[]) => void;
}) {
  function update(i: number, patch: Partial<Assertion>) {
    const next = value.map((a, idx) => {
      if (idx !== i) return a;
      const merged = { ...a, ...patch };
      // when source changes, reset comparison to a valid one for the new source
      if (patch.source && patch.source !== a.source) {
        const allowed = COMPARISONS[patch.source];
        if (!allowed.some((c) => c.value === merged.comparison)) {
          merged.comparison = allowed[0].value;
        }
        if (!SOURCES.find((s) => s.value === patch.source)?.needsProp) merged.property = "";
      }
      return merged;
    });
    onChange(next);
  }

  function remove(i: number) {
    onChange(value.filter((_, idx) => idx !== i));
  }

  return (
    <div className="stack">
      <div className="spread">
        <label style={{ margin: 0 }}>Assertions</label>
        <button type="button" className="ghost" onClick={() => onChange([...value, newAssertion()])}>
          + Add assertion
        </button>
      </div>

      {value.length === 0 && (
        <p className="faint" style={{ fontSize: "0.82rem", margin: 0 }}>
          No extra assertions. Only the expected status code will be checked.
        </p>
      )}

      {value.map((a, i) => {
        const src = SOURCES.find((s) => s.value === a.source)!;
        return (
          <div className="assert-row" key={i}>
            <select value={a.source} onChange={(e) => update(i, { source: e.target.value as AssertionSource })}>
              {SOURCES.map((s) => (
                <option key={s.value} value={s.value}>{s.label}</option>
              ))}
            </select>

            {src.needsProp ? (
              <input
                placeholder={src.propPlaceholder}
                value={a.property}
                onChange={(e) => update(i, { property: e.target.value })}
              />
            ) : (
              <div className="faint" style={{ fontSize: "0.78rem", alignSelf: "center" }}>—</div>
            )}

            <select value={a.comparison} onChange={(e) => update(i, { comparison: e.target.value as AssertionComparison })}>
              {COMPARISONS[a.source].map((c) => (
                <option key={c.value} value={c.value}>{c.label}</option>
              ))}
            </select>

            <input
              placeholder="value"
              value={a.target}
              onChange={(e) => update(i, { target: e.target.value })}
            />

            <button type="button" className="danger" onClick={() => remove(i)} title="Remove">✕</button>
          </div>
        );
      })}
    </div>
  );
}
