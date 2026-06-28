-- +goose Up

-- Groups let a user organize endpoints into named collections.
-- Membership is optional: an endpoint with group_id = NULL is ungrouped.
CREATE TABLE groups (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name       TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX groups_user_idx ON groups(user_id);

-- Deleting a group leaves its endpoints intact, just ungrouped.
ALTER TABLE endpoints ADD COLUMN group_id UUID REFERENCES groups(id) ON DELETE SET NULL;
CREATE INDEX endpoints_group_idx ON endpoints(group_id);

-- +goose Down
ALTER TABLE endpoints DROP COLUMN group_id;
DROP TABLE groups;
