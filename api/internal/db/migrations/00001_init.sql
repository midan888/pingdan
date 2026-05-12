-- +goose Up
CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email           TEXT NOT NULL UNIQUE,
    name            TEXT,
    avatar_url      TEXT,
    provider        TEXT NOT NULL,
    provider_id     TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (provider, provider_id)
);

CREATE TABLE endpoints (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    url             TEXT NOT NULL,
    method          TEXT NOT NULL DEFAULT 'GET',
    expected_status INT NOT NULL DEFAULT 200,
    interval_sec    INT NOT NULL DEFAULT 60,
    timeout_sec     INT NOT NULL DEFAULT 10,
    failure_threshold INT NOT NULL DEFAULT 2,
    enabled         BOOLEAN NOT NULL DEFAULT TRUE,
    current_state   TEXT NOT NULL DEFAULT 'unknown', -- up | down | unknown
    consecutive_failures INT NOT NULL DEFAULT 0,
    last_checked_at TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX endpoints_user_idx ON endpoints(user_id);

CREATE TABLE checks (
    id              BIGSERIAL PRIMARY KEY,
    endpoint_id     UUID NOT NULL REFERENCES endpoints(id) ON DELETE CASCADE,
    status_code     INT,
    latency_ms      INT,
    ok              BOOLEAN NOT NULL,
    error           TEXT,
    checked_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX checks_endpoint_time_idx ON checks(endpoint_id, checked_at DESC);

CREATE TABLE alert_channels (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    kind            TEXT NOT NULL, -- email | telegram
    label           TEXT NOT NULL,
    config          JSONB NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX alert_channels_user_idx ON alert_channels(user_id);

CREATE TABLE endpoint_alert_channels (
    endpoint_id     UUID NOT NULL REFERENCES endpoints(id) ON DELETE CASCADE,
    channel_id      UUID NOT NULL REFERENCES alert_channels(id) ON DELETE CASCADE,
    PRIMARY KEY (endpoint_id, channel_id)
);

-- +goose Down
DROP TABLE endpoint_alert_channels;
DROP TABLE alert_channels;
DROP TABLE checks;
DROP TABLE endpoints;
DROP TABLE users;
