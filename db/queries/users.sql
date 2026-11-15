-- name: CreateUser :execresult
INSERT INTO users (name, dob)
VALUES (?, ?);

-- name: GetUser :one
SELECT id, name, dob
FROM users
WHERE id = ?;

-- name: ListUsers :many
SELECT id, name, dob
FROM users
ORDER BY id
LIMIT ? OFFSET ?;

-- name: CountUsers :one
SELECT COUNT(*)
FROM users;

-- name: UpdateUser :exec
UPDATE users
SET name = ?, dob = ?
WHERE id = ?;

-- name: DeleteUser :execresult
DELETE FROM users
WHERE id = ?;