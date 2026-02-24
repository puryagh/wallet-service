package wallet

import (
	"context"

	"github.com/liveutil/wallet-service/internal/abstract/pb"
)

// ExchangeAsset implements [pb.WalletServiceServer].
func (s *service) ExchangeAsset(ctx context.Context, req *pb.ExchangeAssetRequest) (*pb.ExchangeAssetResponse, error) {
	panic("unimplemented")
}
