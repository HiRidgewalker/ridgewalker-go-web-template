-- +goose Up
CREATE TABLE notebooks (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    title      text NOT NULL,
    content    text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_notebooks_created_at ON notebooks (created_at DESC, id DESC);

-- +goose Down
DROP TABLE notebooks;
