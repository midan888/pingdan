-- +goose Up

-- Assertions let a user verify more than the status code on each check:
-- status_code, response_time, header, body, and json_path comparisons.
CREATE TABLE assertions (
    id          BIGSERIAL PRIMARY KEY,
    endpoint_id UUID NOT NULL REFERENCES endpoints(id) ON DELETE CASCADE,
    source      TEXT NOT NULL,           -- status_code | response_time | header | body | json_path
    property    TEXT NOT NULL DEFAULT '',-- header name or JSON path; empty for status_code/response_time/body
    comparison  TEXT NOT NULL,           -- equals | not_equals | greater_than | less_than | contains | not_contains | matches
    target      TEXT NOT NULL,           -- expected value (string; numbers compared numerically where relevant)
    sort_order  INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX assertions_endpoint_idx ON assertions(endpoint_id, sort_order);

-- Record which assertions failed on a given check so the UI / alerts can explain a failure.
ALTER TABLE checks ADD COLUMN failed_assertions JSONB;

-- +goose Down
ALTER TABLE checks DROP COLUMN failed_assertions;
DROP TABLE assertions;
