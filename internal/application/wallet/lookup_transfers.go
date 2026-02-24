package wallet

import (
	"context"

	"github.com/liveutil/wallet-service/internal/abstract/pb"
)

// LookupTransfers implements [pb.WalletServiceServer].
func (s *service) LookupTransfers(ctx context.Context, req *pb.LookupTransfersRequest) (*pb.LookupTransfersResponse, error) {
	panic("unimplemented")
}
