-- +goose Up
CREATE TABLE links (
    id BIGSERIAL PRIMARY KEY,
    code TEXT NOT NULL,
    original_url TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    clicks_count BIGINT NOT NULL DEFAULT 0,
    deleted_at TIMESTAMPTZ NULL
);

CREATE UNIQUE INDEX links_code_unique_active_idx
    ON links (code)
    WHERE deleted_at IS NULL;

CREATE INDEX links_user_id_created_at_idx
    ON links (user_id, created_at DESC);

CREATE INDEX links_user_id_deleted_at_created_at_idx
    ON links (user_id, deleted_at, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS links;
