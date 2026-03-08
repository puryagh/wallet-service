-- name: GetWalletByIdentifier :one
SELECT * FROM wallets
WHERE identifier::text = $1::text
AND deleted_at IS NULL
LIMIT 1;

-- name: GetUserWalletByIdentifier :one
SELECT * FROM wallets
WHERE identifier::text = $1::text
AND user_identifier = $2
AND deleted_at IS NULL
LIMIT 1;

-- name: GetUserWallets :many
SELECT * FROM wallets
WHERE user_identifier = $1
AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;


-- join accounts, wallets and wallet_assets tables to get wallet and base asset information
-- 1: select account with specified user identifier if it status that is enum is not 'PENDING' , deleted_at IS NULL and banned is FALSE and expires_at after now
-- 2 : get wallet asset by symbol if active field IS TRUE and deleted_at IS NULL
-- 3 : get wallet by user identifier and base asset identifier that match with input user identifier and base asset identifier selected in step 2
--  3-1: selected wallet asset deleted_at IS NULL
--  3-3: selected wallet deleted_at IS NULL
-- name: GetUserAssetWallet :one
SELECT wallets.* FROM wallets
JOIN wallet_assets ON wallets.wallet_asset_id = wallet_assets.id
JOIN accounts ON wallets.account_id = accounts.id
WHERE accounts.user_identifier = $1
AND accounts.status != 'PENDING'
AND accounts.deleted_at IS NULL
AND accounts.banned = FALSE
AND (accounts.expires_at IS NULL OR accounts.expires_at > NOW())
AND wallet_assets.symbol = $2
AND wallet_assets.active = TRUE
AND wallet_assets.deleted_at IS NULL
AND wallets.user_identifier = $1
AND wallets.deleted_at IS NULL
LIMIT 1;

-- name: CreateWallet :one
INSERT INTO wallets (
    user_identifier,
    account_identifier,
    account_id,
    asset_identifier,
    wallet_asset_id,
    meta_data,
    ledger_account_id,
    ledger_account_code,
    primary_account_number,
    iban,
    cvv,
    cvv_two,
    expire_date,
    pin_code,
    status,
    created_at,
    updated_at
) VALUES (
    $1, 
    $2, 
    $3, 
    $4, 
    $5, 
    $6, 
    $7,
    $8,
    $9,
    $10,
    $11,
    $12,
    $13,
    $14,
    $15,
    CURRENT_TIMESTAMP, 
    CURRENT_TIMESTAMP
)
RETURNING *;

