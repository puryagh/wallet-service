package wallet

import (
	"context"

	"github.com/liveutil/wallet-service/internal/abstract/pb"
)

// QueryTransfers implements [pb.WalletServiceServer].
func (s *service) QueryTransfers(ctx context.Context, req *pb.QueryTransfersRequest) (*pb.QueryTransfersResponse, error) {
	panic("unimplemented")
}
