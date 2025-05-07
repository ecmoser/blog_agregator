-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
)
RETURNING *;

-- name: GetFeeds :many
SELECT * FROM feeds;

-- name: CreateFeedFollow :one
INSERT INTO feed_follows (id, created_at, updated_at, feed_id, user_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING *,
    (SELECT name FROM feeds WHERE feeds.id = feed_follows.feed_id) AS feed_name,
    (SELECT name FROM users WHERE users.id = feed_follows.user_id) AS user_name
    ;

-- name: GetFeed :one
SELECT * FROM feeds
WHERE url = $1;

-- name: DeleteFeedFollow :one
DELETE FROM feed_follows
WHERE feed_id = $1
AND user_id = $2
RETURNING *;

-- name: MarkFeedFetched :one
UPDATE feeds
SET last_fetched_at = NOW(), updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: GetNextFeedToFetch :one
SELECT * FROM feeds
ORDER BY last_fetched_at ASC NULLS FIRST
LIMIT 1;

-- name: CreatePost :one
INSERT INTO posts (id, created_at, updated_at, title, url, description, published_at, feed_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8
) RETURNING *;

-- name: GetPostsForUser :many
SELECT * FROM posts
WHERE feed_id IN (
    SELECT feed_id FROM feed_follows WHERE user_id = $1
) LIMIT $2;