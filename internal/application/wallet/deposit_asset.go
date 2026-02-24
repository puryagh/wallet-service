package wallet

import (
	"context"

	"github.com/liveutil/wallet-service/internal/abstract/pb"
)

// DepositAsset implements [pb.WalletServiceServer].
func (s *service) DepositAsset(ctx context.Context, req *pb.DepositAssetRequest) (*pb.DepositAssetResponse, error) {
	panic("unimplemented")
}
