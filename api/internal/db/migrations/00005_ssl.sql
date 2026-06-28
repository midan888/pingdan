-- +goose Up

-- SSL/TLS certificate monitoring for HTTPS endpoints.
-- These columns are populated by the daily ssl checker; non-HTTPS endpoints
-- leave them NULL.
ALTER TABLE endpoints ADD COLUMN ssl_expires_at        TIMESTAMPTZ;
ALTER TABLE endpoints ADD COLUMN ssl_last_checked_at   TIMESTAMPTZ;
ALTER TABLE endpoints ADD COLUMN ssl_last_error        TEXT;

-- The last "days remaining" bucket we sent a warning for. Lets the checker
-- alert at most once per calendar day as the countdown ticks down
-- (15 left, 14 left, ...) without re-sending the same warning on every run.
-- Reset to NULL once the cert is renewed (expiry moves back above the window).
ALTER TABLE endpoints ADD COLUMN ssl_alerted_days      INT;

-- +goose Down
ALTER TABLE endpoints DROP COLUMN ssl_alerted_days;
ALTER TABLE endpoints DROP COLUMN ssl_last_error;
ALTER TABLE endpoints DROP COLUMN ssl_last_checked_at;
ALTER TABLE endpoints DROP COLUMN ssl_expires_at;
