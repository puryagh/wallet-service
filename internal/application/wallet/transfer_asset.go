package wallet

import (
	"context"

	"github.com/liveutil/wallet-service/internal/abstract/pb"
)

// TransferAsset implements [pb.WalletServiceServer].
func (s *service) TransferAsset(ctx context.Context, req *pb.TransferAssetRequest) (*pb.TransferAssetResponse, error) {
	panic("unimplemented")
}
