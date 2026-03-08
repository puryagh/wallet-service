-- name: EnsureIssuerBin :one
INSERT INTO issuer_bins (
    active,
    bin,
    brand,
    issuer_name,
    country_code,
    meta_data,
    created_at,
    updated_at
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
)
ON CONFLICT (bin) DO UPDATE SET
    active = EXCLUDED.active,
    brand = EXCLUDED.brand,
    issuer_name = EXCLUDED.issuer_name,
    country_code = EXCLUDED.country_code,
    meta_data = EXCLUDED.meta_data,
    updated_at = CURRENT_TIMESTAMP
RETURNING *;

-- name: EnsureCardProduct :one
INSERT INTO card_products (
    issuer_bin_id,
    name,
    pan_length,
    cvv_length,
    expiry_months,
    service_code,
    allow_magnetic_stripe,
    meta_data,
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
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
)
ON CONFLICT (name) DO UPDATE SET
    issuer_bin_id = EXCLUDED.issuer_bin_id,
    pan_length = EXCLUDED.pan_length,
    cvv_length = EXCLUDED.cvv_length,
    expiry_months = EXCLUDED.expiry_months,
    service_code = EXCLUDED.service_code,
    allow_magnetic_stripe = EXCLUDED.allow_magnetic_stripe,
    meta_data = EXCLUDED.meta_data,
    updated_at = CURRENT_TIMESTAMP
RETURNING *;

-- name: GetActiveWalletCardByWallet :one
SELECT * FROM wallet_cards
WHERE wallet_identifier = $1
AND user_identifier = $2
AND status = 'ACTIVE'
AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT 1;

-- name: GetWalletCardsByWallet :many
SELECT * FROM wallet_cards
WHERE wallet_identifier = $1
AND user_identifier = $2
AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: GetWalletCardByPanFingerprint :one
SELECT * FROM wallet_cards
WHERE pan_fingerprint = $1
AND deleted_at IS NULL
LIMIT 1;

-- name: CreateWalletCard :one
INSERT INTO wallet_cards (
    wallet_id,
    wallet_identifier,
    user_identifier,
    card_product_id,
    issuer_bin,
    brand,
    pan_fingerprint,
    masked_pan,
    pan_last4,
    expiry_month,
    expiry_year,
    cardholder_name,
    service_code,
    cvv_digest,
    cvv_two_digest,
    track1_digest,
    track2_digest,
    pin_digest,
    last_authenticated_at,
    status,
    meta_data,
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
    $16,
    $17,
    $18,
    $19,
    $20,
    $21,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
)
RETURNING *;

-- name: UpdateWalletCardLastAuthenticatedAt :one
UPDATE wallet_cards
SET last_authenticated_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: CreateWalletCardEvent :one
INSERT INTO wallet_card_events (
    wallet_card_id,
    wallet_card_identifier,
    user_identifier,
    event_type,
    success,
    remote_ip,
    meta_data,
    created_at
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    CURRENT_TIMESTAMP
)
RETURNING *;