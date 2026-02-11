package repository

import "context"

// AccountsTxParams is the params for CreateUserTx.
type AccountsTxParams struct {
	CreateAccountParams
	AfterCreate func(account Account) error
}

// AccountsTxResult is the result for CreateUserTx.
type AccountsTxResult struct {
	Account Account
}

// CreateAccountTx implements Store.CreateAccountTx
func (store *SQLStore) CreateAccountTx(ctx context.Context, arg AccountsTxParams) (AccountsTxResult, error) {
	var result AccountsTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		result.Account, err = q.CreateAccount(ctx, arg.CreateAccountParams)
		if err != nil {
			return err
		}

		return arg.AfterCreate(result.Account)
	})

	return result, err
}
