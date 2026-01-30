package repository

import "context"

// CreateUserTxParams is the params for CreateUserTx.
type CreateUserTxParams struct {
	CreateUserParams
	AfterCreate func(user User) error
}

// UserTxResult is the result for CreateUserTx.
type UserTxResult struct {
	User User
}

// CreateUserTx implements Store.CreateUserTx
func (store *SQLStore) CreateUserTx(ctx context.Context, arg CreateUserTxParams) (UserTxResult, error) {
	var result UserTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		result.User, err = q.CreateUser(ctx, arg.CreateUserParams)
		if err != nil {
			return err
		}

		return arg.AfterCreate(result.User)
	})

	return result, err
}
