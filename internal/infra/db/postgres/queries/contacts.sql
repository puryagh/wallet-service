-- name: GetContactById :one
SELECT *
FROM contacts
WHERE id = $1
    AND deleted_at IS NULL
LIMIT 1;

-- name: GetContactByUserID :one
SELECT *
FROM contacts
WHERE user_id = $1
    AND deleted_at IS NULL
LIMIT 1;

-- name: FindActiveMobileContact :one
SELECT 
    c.*,
    u.id as user_id,
    u.identifier as user_identifier,
    u.approved as user_approved,
    u.banned as user_banned,
    u.roles as user_roles,
    p.id as profile_id,
    p.identifier as profile_identifier,
    p.profile_type as profile_type,
    p.first_name as profile_first_name,
    p.last_name as profile_last_name,
    p.national_id as profile_national_id,
    p.status as profile_status,
    p.meta_data as profile_meta_data,
    p.created_at as profile_created_at
FROM contacts c
JOIN users u ON u.id = c.user_id
LEFT JOIN profiles p ON p.user_id = u.id AND p.deleted_at IS NULL
WHERE 
    c.mobile = $1
    AND c.contact_type = 'MOBILE'
    AND c.deleted_at IS NULL
    AND u.deleted_at IS NULL
    AND u.banned = FALSE
    AND u.roles @> sqlc.arg(roles)::text[]
LIMIT 1;


-- name: FindActiveEmailContact :one
SELECT 
    c.*,
    u.id as user_id,
    u.identifier as user_identifier,
    u.approved as user_approved,
    u.banned as user_banned,
    u.roles as user_roles,
    p.id as profile_id,
    p.identifier as profile_identifier,
    p.profile_type as profile_type,
    p.first_name as profile_first_name,
    p.last_name as profile_last_name,
    p.national_id as profile_national_id,
    p.status as profile_status,
    p.meta_data as profile_meta_data,
    p.created_at as profile_created_at
FROM contacts c
JOIN users u ON u.id = c.user_id
LEFT JOIN profiles p ON p.user_id = u.id AND p.deleted_at IS NULL
WHERE 
    c.email = $1
    AND c.contact_type = 'EMAIL'
    AND c.is_email_verified = TRUE
    AND c.deleted_at IS NULL
    AND u.deleted_at IS NULL
    AND u.banned = FALSE
    AND $2 = ANY(u.roles)
LIMIT 1;

-- name: GetContactByUserId :one
SELECT *
FROM contacts
WHERE user_id = $1
    AND deleted_at IS NULL
LIMIT 1;

-- name: GetContactByMobile :one
SELECT *
FROM contacts
WHERE mobile = $1
    AND deleted_at IS NULL
LIMIT 1;

-- name: GetContactByEmail :one
SELECT *
FROM contacts
WHERE email = $1
    AND deleted_at IS NULL
LIMIT 1;

-- name: CreateContact :one
INSERT INTO contacts (
        contact_type,
        user_id,
        mobile,
        email,
        meta_data
    )
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateContact :one
UPDATE contacts
SET (
        mobile,
        email,
        meta_data,
        mobile_totp,
        mobile_totp_expires_at,
        email_totp,
        email_totp_expires_at,
        updated_at
    ) = ($1, $2, $3, $4, $5, $6, $7, $8)
WHERE id = $9
RETURNING *;

-- name: SafeDeleteContact :exec
UPDATE contacts
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: SetMobileOTP :one
UPDATE contacts
SET mobile_totp = $1, mobile_totp_expires_at = $2
WHERE id = $3
RETURNING *;

-- name: SetEmailOTP :one
UPDATE contacts
SET email_totp = $1, email_totp_expires_at = $2
WHERE id = $3
RETURNING *;
