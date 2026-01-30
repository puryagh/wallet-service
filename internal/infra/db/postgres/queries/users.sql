-- name: GetUserById :one
SELECT
    *
FROM
    users
WHERE
    id = $1
    AND deleted_at IS NULL
LIMIT
    1;

-- name: GetActiveUserById :one
SELECT
    *
FROM
    users
WHERE
    id = $1
    AND banned = false
    AND expires_at > CURRENT_TIMESTAMP
    AND deleted_at IS NULL
LIMIT
    1;

-- name: GetActiveUserByIdentifier :one
SELECT
    *
FROM
    users
WHERE
    identifier = $1
    AND banned = false
    AND expires_at > CURRENT_TIMESTAMP
    AND deleted_at IS NULL
LIMIT
    1;

-- name: GetUserByIdentifier :one
SELECT
    *
FROM
    users
WHERE
    identifier = $1
    AND deleted_at IS NULL
LIMIT
    1;



-- name: GetSafeUserById :one
SELECT
    id,
    identifier,
    approved,
    banned,
    roles
FROM
    users
WHERE
    id = $1
    AND deleted_at IS NULL
LIMIT
    1;


-- name: CreateUser :one
INSERT INTO
    users (
        approved,
        banned,
        meta_data,
        roles,
        expires_at,
        created_at,
        updated_at
    )
VALUES
    (
        $1,
        $2,
        $3,
        $4,
        $5,
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP
    )
RETURNING
    *;

-- name: UpdateUser :one
UPDATE users
SET
    (
        approved,
        banned,
        meta_data,
        roles,
        expires_at,
        updated_at
    ) = (
        $1,
        $2,
        $3,
        $4,
        $5,
        CURRENT_TIMESTAMP
    )
WHERE
    id = $6
RETURNING
    *;