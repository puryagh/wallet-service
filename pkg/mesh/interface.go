package mesh

import (
	"context"

	"github.com/liveutil/wallet-service/pkg/models"
	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"
)

type TigerBittleAccount types.Account

// WalletServiceMesh is the interface for the wallet service mesh.
type WalletServiceMesh interface {
	GetUserAccounts(ctx context.Context, userId int64) ([]*models.WalletAccount, error)
	GetUserAccount(ctx context.Context, userId int64, userIdentifier string) (*models.WalletAccount, error)
	CreateUserAccount(ctx context.Context, userId int64, userIdentifier string) (*models.WalletAccount, error)
	GetTigerBeetleAccount(ctx context.Context, accountID types.Uint128) (*TigerBittleAccount, error)
}
