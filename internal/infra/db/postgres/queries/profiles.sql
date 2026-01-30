-- name: GetProfileByUserID :one
SELECT *
FROM profiles
WHERE user_id = $1
    AND deleted_at IS NULL
LIMIT 1;