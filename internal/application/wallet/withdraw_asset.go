package wallet

import (
	"context"

	"github.com/liveutil/wallet-service/internal/abstract/pb"
)

// WithdrawAsset implements [pb.WalletServiceServer].
func (s *service) WithdrawAsset(ctx context.Context, req *pb.WithdrawAssetRequest) (*pb.WithdrawAssetResponse, error) {
	panic("unimplemented")
}
