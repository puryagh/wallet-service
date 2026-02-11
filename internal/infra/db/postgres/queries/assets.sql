-- name: GetAllWalletAssets :many
SELECT * FROM wallet_assets
WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;


-- name: GetWalletAssetByIdentifier :one
SELECT * FROM wallet_assets
WHERE identifier = $1
AND deleted_at IS NULL
LIMIT 1;


-- name: GetWalletAssetBySymbol :one
SELECT * FROM wallet_assets
WHERE symbol = $1
AND deleted_at IS NULL
LIMIT 1;
