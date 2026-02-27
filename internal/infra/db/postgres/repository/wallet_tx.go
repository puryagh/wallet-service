package repository

import "context"

// WalletTxParams is the params for CreateWalletTx.
type WalletsTxParams struct {
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

		return arg.AfterCreate(result.Wallet)
	})

	return result, err
}
