package wallet

import (
	"context"
	"strconv"

	"github.com/liveutil/go-lib/contextutil"
	"github.com/liveutil/wallet-service/internal/abstract/pb"
	"github.com/liveutil/wallet-service/internal/infra/db/postgres/repository"
	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"
	"google.golang.org/protobuf/types/known/emptypb"
)

// GetAccountTransfers implements [pb.WalletServiceServer].
func (s *service) GetAccountTransfers(ctx context.Context, req *emptypb.Empty) (*pb.GetAccountTransfersResponse, error) {
	contextUser := &contextutil.ContextUser{}
	if err := contextutil.CatchUser(ctx, contextUser); err != nil {
		return nil, err
	}

	user, err := s.userServiceMeshClient.GetUserByID(ctx, uint64(contextUser.ID))
	if err != nil {
		return nil, err
	}

	wallets, err := s.repository.GetUserWallets(ctx, repository.GetUserWalletsParams{
		UserIdentifier: user.Identifier,
		Limit:          100,
		Offset:         0,
	})
	if err != nil {
		return nil, err
	}

	if len(wallets) == 0 {
		return &pb.GetAccountTransfersResponse{
			NextTimestamp: 0,
			TotalCount:    0,
			HasMore:       false,
			Transfers:     []*pb.Transfer{},
		}, nil
	}

	accountIds := []types.Uint128{}

	for _, wallet := range wallets {
		accountIds = append(accountIds, types.ToUint128(uint64(wallet.LedgerAccountID)))
	}

	transfers, err := s.tigerbeetle.LookupTransfers(accountIds)
	if err != nil {
		return nil, err
	}

	response := &pb.GetAccountTransfersResponse{
		Error:         false,
		Message:       "get account transfers success",
		NextTimestamp: 0,
		TotalCount:    uint32(len(transfers)),
		HasMore:       false,
		Transfers:     []*pb.Transfer{},
	}

	for _, transfer := range transfers {
		response.Transfers = append(response.Transfers, &pb.Transfer{
			Id:              transfer.ID.String(),
			DebitAccountId:  transfer.DebitAccountID.String(),
			CreditAccountId: transfer.CreditAccountID.String(),
			Amount:          transfer.Amount.String(),
			PendingId:       transfer.PendingID.String(),
			UserData_128:    transfer.UserData128[:],
			Timestamp:       transfer.Timestamp,
			UserData_64:     strconv.FormatUint(transfer.UserData64, 10),
			Timeout:         transfer.Timeout,
			Ledger:          strconv.FormatUint(uint64(transfer.Ledger), 10),
			Code:            strconv.FormatUint(uint64(transfer.Code), 10),
			Flags: &pb.TransferFlags{
				Linked:              transfer.TransferFlags().Linked,
				Pending:             transfer.TransferFlags().Pending,
				PostPendingTransfer: transfer.TransferFlags().PostPendingTransfer,
				VoidPendingTransfer: transfer.TransferFlags().VoidPendingTransfer,
				BalancingDebit:      transfer.TransferFlags().BalancingDebit,
				BalancingCredit:     transfer.TransferFlags().BalancingCredit,
				ClosingDebit:        transfer.TransferFlags().ClosingDebit,
				ClosingCredit:       transfer.TransferFlags().ClosingCredit,
				Imported:            transfer.TransferFlags().Imported,
			},
			UserData_32: transfer.UserData32,
		})
	}

	return response, nil
}
