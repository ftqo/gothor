-- name: CreateUser :one
-- Inserts a new user into the users table
INSERT INTO users (username, email, hashed_password, salt)
VALUES ('', $1, $2, $3)
RETURNING *;

-- name: GetUserByEmail :one
-- Retrieves user by email (used for login)
SELECT *
FROM users
WHERE email = $1;

-- name: CreatePost :one
-- Inserts a new post into the posts table
INSERT INTO posts (user_id, title, content, created_at, updated_at)
VALUES ($1, $2, $3, NOW(), NOW())
RETURNING *;

-- name: GetAllPosts :many
-- Retrieves all posts from the posts table
SELECT * FROM posts
ORDER BY created_at DESC;

-- name: GetPostByID :one
-- Retrieves a specific post from the posts table by ID
SELECT * FROM posts
WHERE id = $1;

-- name: UpdatePost :one
-- Updates a specific post in the posts table by ID
UPDATE posts
SET title = $2, content = $3, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeletePost :exec
-- Deletes a specific post from the posts table by ID
DELETE FROM posts
WHERE id = $1;

-- name: CreateComment :one
-- Inserts a new comment into the comments table
INSERT INTO comments (user_id, post_id, content, created_at, updated_at)
VALUES ($1, $2, $3, NOW(), NOW())
RETURNING *;

-- name: GetAllCommentsForPost :many
-- Retrieves all comments for a specific post from the comments table
SELECT * FROM comments
WHERE post_id = $1
ORDER BY created_at DESC;

-- name: UpdateComment :one
-- Updates a specific comment in the comments table by ID
UPDATE comments
SET content = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteComment :exec
-- Deletes a specific comment from the comments table by ID
DELETE FROM comments
WHERE id = $1;
