package wallet

import (
	"context"

	"github.com/liveutil/wallet-service/internal/abstract/pb"
	"google.golang.org/protobuf/types/known/emptypb"
)

// authorizationMiddleware is the implementation of the wallet service authorization middleware
type authorizationMiddleware struct {
	pb.UnimplementedWalletServiceServer

	next pb.WalletServiceServer
}

// NewAuthorizationMiddleware creates a new wallet service authorization middleware
func NewAuthorizationMiddleware(service pb.WalletServiceServer) pb.WalletServiceServer {
	return &authorizationMiddleware{
		next: service,
	}
}

// ContextUserAccounts implements [pb.WalletServiceServer].
func (a *authorizationMiddleware) ContextUserAccounts(ctx context.Context, req *emptypb.Empty) (*pb.ContextUserWalletResponse, error) {
	return a.next.ContextUserAccounts(ctx, req)
}

// ContextUserWalletTransactions implements [pb.WalletServiceServer].
func (a *authorizationMiddleware) ContextUserWalletTransactions(ctx context.Context, req *emptypb.Empty) (*pb.ContextUserWalletTransactionsResponse, error) {
	return a.next.ContextUserWalletTransactions(ctx, req)
}

// DepositAsset implements [pb.WalletServiceServer].
func (a *authorizationMiddleware) DepositAsset(ctx context.Context, req *pb.DepositAssetRequest) (*pb.DepositAssetResponse, error) {
	return a.next.DepositAsset(ctx, req)
}

// ExchangeAsset implements [pb.WalletServiceServer].
func (a *authorizationMiddleware) ExchangeAsset(ctx context.Context, req *pb.ExchangeAssetRequest) (*pb.ExchangeAssetResponse, error) {
	return a.next.ExchangeAsset(ctx, req)
}

// GetAccountTransfers implements [pb.WalletServiceServer].
func (a *authorizationMiddleware) GetAccountTransfers(ctx context.Context, req *pb.GetAccountTransfersRequest) (*pb.GetAccountTransfersResponse, error) {
	return a.next.GetAccountTransfers(ctx, req)
}

// LookupTransfers implements [pb.WalletServiceServer].
func (a *authorizationMiddleware) LookupTransfers(ctx context.Context, req *pb.LookupTransfersRequest) (*pb.LookupTransfersResponse, error) {
	return a.next.LookupTransfers(ctx, req)
}

// QueryTransfers implements [pb.WalletServiceServer].
func (a *authorizationMiddleware) QueryTransfers(ctx context.Context, req *pb.QueryTransfersRequest) (*pb.QueryTransfersResponse, error) {
	return a.next.QueryTransfers(ctx, req)
}

// TransferAsset implements [pb.WalletServiceServer].
func (a *authorizationMiddleware) TransferAsset(ctx context.Context, req *pb.TransferAssetRequest) (*pb.TransferAssetResponse, error) {
	return a.next.TransferAsset(ctx, req)
}

// WithdrawAsset implements [pb.WalletServiceServer].
func (a *authorizationMiddleware) WithdrawAsset(ctx context.Context, req *pb.WithdrawAssetRequest) (*pb.WithdrawAssetResponse, error) {
	return a.next.WithdrawAsset(ctx, req)
}
