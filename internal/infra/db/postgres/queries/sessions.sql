-- name: CreateUserSession :one
INSERT INTO sessions (
        user_id,
        expires_in,
        notification,
        meta_data,
        host,
        device_id,
        device_name,
        device_user_agent,
        device_notification_token,
        created_at,
        updated_at
    )
VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        $7,
        $8,
        $9,
        (CURRENT_TIMESTAMP),
        (CURRENT_TIMESTAMP)
    )
RETURNING *;

-- name: GetSessionById :one
SELECT *
FROM sessions
WHERE id = $1
    AND user_id = $2
    AND deleted_at IS NULL
LIMIT 1;

-- name: RenewUserSession :one
UPDATE sessions
SET
    (meta_data, updated_at, expires_in) = ($3, (CURRENT_TIMESTAMP), $4)
WHERE
    sessions.user_id = $1
    AND sessions.id = $2
    AND sessions.deleted_at IS NULL
RETURNING
    *;