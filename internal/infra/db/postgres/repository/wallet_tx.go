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
				Flags: types.AccountFlags{
					DebitsMustNotExceedCredits: true,
					Linked:                     true,
					CreditsMustNotExceedDebits: false,
					History:                    true,
					Imported:                   false,
					Closed:                     false,
				}.ToUint16(),
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

// BatchWalletsTxParams is the params for BatchCreateWalletTx.
type BatchWalletsTxParams struct {
	TigerBeetle        tb.Client
	UserID             uint64
	CreateWalletParams []CreateWalletParams
	AfterCreate        func(wallets []Wallet) error
}

// AccountsTxResult is the result for CreateUserTx.
type BatchWalletsTxResult struct {
	Wallets []Wallet
}

// BatchCreateWalletTx implements Store.BatchCreateWalletTx
func (store *SQLStore) BatchCreateWalletTx(ctx context.Context, arg BatchWalletsTxParams) (BatchWalletsTxResult, error) {
	var result BatchWalletsTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		for _, wallet := range arg.CreateWalletParams {
			createdWallet, err := q.CreateWallet(ctx, wallet)
			if err != nil {
				return err
			}

			result.Wallets = append(result.Wallets, createdWallet)
		}

		ledgerAccounts := []types.Account{}

		for index, wallet := range result.Wallets {
			userData := types.ToUint128(arg.UserID)
			ledgerAccountId := types.ToUint128(uint64(wallet.LedgerAccountID + wallet.ID))

			linked := false
			if len(result.Wallets) > 1 && index < len(result.Wallets)-1 {
				linked = true
			}

			ledgerAccounts = append(ledgerAccounts, types.Account{
				ID:          ledgerAccountId,
				UserData128: userData,
				UserData64:  arg.UserID,
				UserData32:  0,
				Ledger:      1,
				Code:        uint16(wallet.LedgerAccountCode),
				Flags: types.AccountFlags{
					DebitsMustNotExceedCredits: true,
					Linked:                     linked,
					CreditsMustNotExceedDebits: false,
					History:                    true,
					Imported:                   false,
					Closed:                     false,
				}.ToUint16(),
			})
		}

		// create tigerbeetle account
		res, err := arg.TigerBeetle.CreateAccounts(ledgerAccounts)
		if err != nil {
			return err
		}

		for _, err := range res {
			return errors.New(err.Result.String())
		}

		return arg.AfterCreate(result.Wallets)
	})

	return result, err
}
