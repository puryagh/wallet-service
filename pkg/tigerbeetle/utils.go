package tigerbeetle

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"

	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"
)

// GenerateLedgerAccountID generates a ledger account ID for the given user and asset.
func GenerateLedgerAccountID(
	userIdentifier string,
	symbol string,
	assetIdentifier string,
	baseAccountNumber uint64,
	userId uint64,
) (types.Uint128, error) {

	payload := fmt.Sprintf(
		"v1|user:%s|asset:%s|base:%d|user_id:%d|account:%d",
		userIdentifier,
		symbol,
		baseAccountNumber,
		userId,
		baseAccountNumber+userId,
	)

	hash := hmac.New(sha256.New, []byte(assetIdentifier)[:16])
	hash.Write([]byte(payload))
	hexValue := fmt.Sprintf("%x", hash.Sum(nil))

	return types.HexStringToUint128(hexValue)
}
