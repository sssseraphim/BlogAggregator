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

-- name: ListFeeds :many
SELECT f.name, f.url, u.name
FROM feeds f
LEFT JOIN users u ON f.user_id = u.id;

-- name: CreateFeedFollow :one
WITH inserted_feed_follow AS (
INSERT INTO feeds_follows (id, created_at, updated_at, user_id, feed_id)
VALUES (
		$1,
		$2,
		$3,
		$4,
		$5
)
RETURNING *
)
SELECT
		inserted_feed_follow.*,
		feeds.name AS feed_name,
		users.name AS user_name
FROM inserted_feed_follow
INNER JOIN feeds on inserted_feed_follow.feed_id = feeds.id
INNER JOIN users on inserted_feed_follow.user_id = users.id;

-- name: GetFeedByURL :one
SELECT *
FROM feeds
WHERE url = $1;

-- name: GetFeedFollowsForUser :many
SELECT users.name as name, feeds.name as feed
FROM users
INNER JOIN feeds_follows ON users.id = feeds_follows.user_id
INNER JOIN feeds ON feeds_follows.feed_id = feeds.id
WHERE users.name = $1;

-- name: DeleteFeedFollow :exec
DELETE FROM feeds_follows
WHERE user_id = $1
AND feed_id = $2;

-- name: MarkFeedFetched :exec
UPDATE feeds
SET last_fetched_at = $1,
		updated_at = $1
WHERE id = $2;

-- name: GetNextFeedToFetch :one
SELECT * from feeds
ORDER BY last_fetched_at ASC NULLS FIRST
LIMIT 1;

-- name: CreatePost :one
INSERT INTO posts (id, created_at, updated_at, title, url, description, published_at, feed_id) 
VALUES(
		$1,
		$2,
		$3,
		$4,
		$5,
		$6,
		$7,
		$8
)
RETURNING *;

-- name: GetPostsForUser :many
SELECT *
FROM posts
INNER JOIN feeds ON posts.feed_id = feeds.id
INNER JOIN feeds_follows ON feeds.id = feeds_follows.feed_id
WHERE feeds_follows.user_id = $1
ORDER BY posts.published_at DESC NULLS LAST
LIMIT $2;

