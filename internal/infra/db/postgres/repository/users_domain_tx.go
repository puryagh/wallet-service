package repository

import "context"

// CreateUserAndRelationsTxParams is the params for CreateUserAndRelationsTx.
type CreateUserAndRelationsTxParams struct {
	CreateUserAndRelationsParams
	AfterCreate func(user CreateUserAndRelationsRow) error
}

// UserAndRelationsTxResult is the result for CreateUserAndRelationsTx.
type UserAndRelationsTxResult struct {
	CreateUserAndRelationsRow
}

// CreateUserAndRelationsTx implements Store.CreateUserAndRelationsTx
func (store *SQLStore) CreateUserAndRelationsTx(ctx context.Context, arg CreateUserAndRelationsTxParams) (UserAndRelationsTxResult, error) {
	var result UserAndRelationsTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		result.CreateUserAndRelationsRow, err = q.CreateUserAndRelations(ctx, arg.CreateUserAndRelationsParams)
		if err != nil {
			return err
		}

		return arg.AfterCreate(result.CreateUserAndRelationsRow)
	})

	return result, err
}
