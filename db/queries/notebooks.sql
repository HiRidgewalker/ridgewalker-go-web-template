-- name: CreateNotebook :one
INSERT INTO notebooks (title, content)
VALUES ($1, $2)
RETURNING id, title, content, created_at, updated_at;

-- name: GetNotebook :one
SELECT id, title, content, created_at, updated_at
FROM notebooks
WHERE id = $1;

-- name: ListNotebooks :many
SELECT id, title, content, created_at, updated_at
FROM notebooks
ORDER BY created_at DESC, id DESC
LIMIT $1 OFFSET $2;

-- name: CountNotebooks :one
SELECT count(*) FROM notebooks;

-- name: UpdateNotebook :one
UPDATE notebooks
SET title = $2,
    content = $3,
    updated_at = now()
WHERE id = $1
RETURNING id, title, content, created_at, updated_at;

-- name: DeleteNotebook :execrows
DELETE FROM notebooks
WHERE id = $1;
