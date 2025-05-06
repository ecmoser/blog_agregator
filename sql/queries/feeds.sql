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