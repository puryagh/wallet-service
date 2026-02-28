package repository

import (
	"context"
	"errors"

	tb "github.com/tigerbeetle/tigerbeetle-go"
	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"
)

// WalletTxParams is the params for CreateWalletTx.
type WalletsTxParams struct {
	TigerBeetle tb.Client
	UserID      uint64
	CreateWalletParams
	AfterCreate func(wallet Wallet) error
}

// AccountsTxResult is the result for CreateUserTx.
type WalletsTxResult struct {
	Wallet Wallet
}

// CreateWalletTx implements Store.CreateWalletTx
func (store *SQLStore) CreateWalletTx(ctx context.Context, arg WalletsTxParams) (WalletsTxResult, error) {
	var result WalletsTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		result.Wallet, err = q.CreateWallet(ctx, arg.CreateWalletParams)
		if err != nil {
			return err
		}

		userData := types.ToUint128(arg.UserID)
		ledgerAccountId := types.ToUint128(uint64(arg.LedgerAccountID + result.Wallet.ID))
		// create tigerbeetle account
		res, err := arg.TigerBeetle.CreateAccounts([]types.Account{
			{
				ID:          ledgerAccountId,
				UserData128: userData,
				UserData64:  arg.UserID,
				UserData32:  0,
				Ledger:      1,
				Code:        uint16(arg.LedgerAccountCode),
				Flags:       types.AccountFlags{DebitsMustNotExceedCredits: true}.ToUint16(),
			},
		})
		if err != nil {
			return err
		}

		for _, err := range res {
			return errors.New(err.Result.String())
		}

		return arg.AfterCreate(result.Wallet)
	})

	return result, err
}
