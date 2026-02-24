package wallet

import (
	"context"

	"github.com/liveutil/wallet-service/internal/abstract/pb"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ContextUserWalletTransactions implements [pb.WalletServiceServer].
func (s *service) ContextUserWalletTransactions(ctx context.Context, req *emptypb.Empty) (*pb.ContextUserWalletTransactionsResponse, error) {
	panic("unimplemented")
}
