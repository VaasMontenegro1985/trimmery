-- +goose Up
CREATE TABLE visits (
    id BIGSERIAL PRIMARY KEY,
    link_id BIGINT NOT NULL REFERENCES links(id) ON DELETE CASCADE,
    visited_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    ip INET NULL,
    user_agent TEXT NULL
);

CREATE INDEX visits_link_id_visited_at_idx
    ON visits (link_id, visited_at DESC);

-- +goose Down
DROP TABLE IF EXISTS visits;
