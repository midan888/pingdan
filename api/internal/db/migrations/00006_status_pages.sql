-- +goose Up

-- A status page is a named, publicly shareable view that bundles a selection
-- of the owner's endpoints behind a unique slug (e.g. /status/{slug}).
CREATE TABLE status_pages (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    slug        TEXT NOT NULL UNIQUE,
    title       TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX status_pages_user_idx ON status_pages(user_id);

-- Junction: which endpoints appear on a page, in what order, under what public name.
-- display_name NULL falls back to the endpoint's real name at render time.
CREATE TABLE status_page_endpoints (
    page_id      UUID NOT NULL REFERENCES status_pages(id) ON DELETE CASCADE,
    endpoint_id  UUID NOT NULL REFERENCES endpoints(id) ON DELETE CASCADE,
    display_name TEXT,
    position     INT NOT NULL DEFAULT 0,
    PRIMARY KEY (page_id, endpoint_id)
);
CREATE INDEX status_page_endpoints_page_idx ON status_page_endpoints(page_id, position);

-- +goose Down
DROP TABLE status_page_endpoints;
DROP TABLE status_pages;
