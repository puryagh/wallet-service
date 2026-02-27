-- name: GetUserAccount :one
SELECT * FROM accounts
WHERE accounts.user_identifier = $1
LIMIT 1;

-- name: GetAccountByIdentifier :one
SELECT * FROM accounts
WHERE identifier::varchar(32) = $1::varchar(32)
AND deleted_at IS NULL
LIMIT 1;

-- name: CreateAccount :one
INSERT INTO accounts (
  title,
  description,
  status,
  banned,
  user_identifier,
  base_asset_identifier,
  meta_data,
  expires_at,
  created_at,
  updated_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, CURRENT_TIMESTAMP, NULL
)
RETURNING *;


