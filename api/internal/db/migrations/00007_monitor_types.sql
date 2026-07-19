-- +goose Up

-- Existing monitors are HTTP checks. New monitors can instead perform a TCP
-- connection or an ICMP echo request against the URI stored in url.
ALTER TABLE endpoints
    ADD COLUMN check_type TEXT NOT NULL DEFAULT 'http',
    ADD CONSTRAINT endpoints_check_type_valid CHECK (check_type IN ('http', 'tcp', 'icmp'));

-- +goose Down
ALTER TABLE endpoints DROP CONSTRAINT endpoints_check_type_valid;
ALTER TABLE endpoints DROP COLUMN check_type;
