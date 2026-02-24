package wallet

import (
	"context"

	"github.com/liveutil/wallet-service/internal/abstract/pb"
)

// GetAccountTransfers implements [pb.WalletServiceServer].
func (s *service) GetAccountTransfers(ctx context.Context, req *pb.GetAccountTransfersRequest) (*pb.GetAccountTransfersResponse, error) {
	panic("unimplemented")
}
