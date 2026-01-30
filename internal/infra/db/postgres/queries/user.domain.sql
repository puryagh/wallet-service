-- name: CreateUserAndRelations :one
WITH new_user AS (
    INSERT INTO users (
            expires_at,
            created_at,
            updated_at
            
        )
    VALUES (
            $1,
            CURRENT_TIMESTAMP,
            CURRENT_TIMESTAMP
        )
    RETURNING *
),
new_contact AS (
    INSERT INTO contacts (
            contact_type,
            user_id,
            mobile,
            mobile_totp,
            mobile_totp_expires_at,
            email,
            email_totp,
            email_totp_expires_at,
            created_at,
            updated_at
        )
    VALUES (
            $2,
            (
                SELECT id
                FROM new_user
            ),
            $3,
            $4,
            $5,
            $6,
            $7,
            $8,
            CURRENT_TIMESTAMP,
            CURRENT_TIMESTAMP
        )
    RETURNING *
),
new_profile AS (
    INSERT INTO profiles (
            user_id,
            profile_type,
            first_name,
            last_name,
            national_id,
            created_at,
            updated_at
        )
    VALUES (
            (
                SELECT id
                FROM new_user
            ),
            $9,
            $10,
            $11,
            $12,
            CURRENT_TIMESTAMP,
            CURRENT_TIMESTAMP
        )
    RETURNING *
)
SELECT u.id as user_id,
    u.identifier as user_identifier,
    u.approved as user_approved,
    u.banned as user_banned,
    u.meta_data as user_meta_data,
    u.roles as user_roles,
    u.created_at as user_created_at,
    c.id as contact_id,
    c.identifier as contact_identifier,
    c.mobile as contact_mobile,
    c.email as contact_email,
    c.meta_data as contact_meta_data,
    c.created_at as contact_created_at,
    p.id as profile_id,
    p.identifier as profile_identifier,
    p.profile_type as profile_type,
    p.first_name as profile_first_name,
    p.last_name as profile_last_name,
    p.national_id as profile_national_id,
    p.status as profile_status,
    p.meta_data as profile_meta_data,
    p.created_at as profile_created_at
FROM new_user u
    JOIN new_contact c ON c.user_id = u.id
    JOIN new_profile p ON p.user_id = u.id;