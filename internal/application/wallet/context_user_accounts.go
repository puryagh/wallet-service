package wallet

import (
	"context"

	"github.com/liveutil/wallet-service/internal/abstract/pb"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ContextUserAccounts implements [pb.WalletServiceServer].
func (s *service) ContextUserAccounts(ctx context.Context, req *emptypb.Empty) (*pb.ContextUserWalletResponse, error) {
	panic("unimplemented")
}
