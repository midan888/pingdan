"use client";

import Link from "next/link";
import type { Assertion, AssertionComparison, AssertionSource } from "@/lib/api";

const SOURCES: {
  value: AssertionSource;
  label: string;
  needsProp?: boolean;
  propPlaceholder?: string;
  targetPlaceholder: string;
  numeric?: boolean;
}[] = [
  { value: "status_code", label: "Status code", targetPlaceholder: "200", numeric: true },
  { value: "response_time", label: "Response time (ms)", targetPlaceholder: "500", numeric: true },
  { value: "header", label: "Header", needsProp: true, propPlaceholder: "Content-Type", targetPlaceholder: "application/json" },
  { value: "body", label: "Body", targetPlaceholder: "ok" },
  { value: "json_path", label: "JSON path", needsProp: true, propPlaceholder: "data.status", targetPlaceholder: "ok" },
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
      // when source changes, reset the fields that no longer make sense
      if (patch.source && patch.source !== a.source) {
        const allowed = COMPARISONS[patch.source];
        if (!allowed.some((c) => c.value === merged.comparison)) {
          merged.comparison = allowed[0].value;
        }
        merged.property = "";
        merged.target = "";
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
        const isRegex = a.comparison === "matches";
        return (
          <div className="assert-row" key={i}>
            <select
              className="a-source"
              value={a.source}
              onChange={(e) => update(i, { source: e.target.value as AssertionSource })}
              aria-label="assertion source"
            >
              {SOURCES.map((s) => (
                <option key={s.value} value={s.value}>{s.label}</option>
              ))}
            </select>

            {src.needsProp && (
              <input
                className="a-prop"
                placeholder={src.propPlaceholder}
                value={a.property}
                onChange={(e) => update(i, { property: e.target.value })}
                aria-label={a.source === "header" ? "header name" : "JSON path"}
              />
            )}

            <select
              className="a-comparison"
              value={a.comparison}
              onChange={(e) => update(i, { comparison: e.target.value as AssertionComparison })}
              aria-label="comparison"
            >
              {COMPARISONS[a.source].map((c) => (
                <option key={c.value} value={c.value}>{c.label}</option>
              ))}
            </select>

            <input
              className="a-target"
              placeholder={isRegex ? "^2\\d\\d$" : src.targetPlaceholder}
              inputMode={src.numeric && !isRegex ? "numeric" : undefined}
              value={a.target}
              onChange={(e) => update(i, { target: e.target.value })}
              aria-label="expected value"
            />

            <button type="button" className="danger a-remove" onClick={() => remove(i)} title="Remove assertion" aria-label="remove assertion">✕</button>
          </div>
        );
      })}

      {value.length > 0 && (
        <div className="hint" style={{ marginTop: 0 }}>
          {value.length > 1 ? "All assertions must pass for a check to count as up. " : "The assertion must pass for a check to count as up. "}
          <Link href="/docs#assertions">Assertion docs</Link>.
        </div>
      )}
    </div>
  );
}
