package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Store defines all functions to execute db queries and transactions
// any TX required queries should defined on this interface
type Store interface {
	Querier

	// User TXes
	CreateUserTx(ctx context.Context, arg CreateUserTxParams) (UserTxResult, error)

	// User and relations TXes
	CreateUserAndRelationsTx(ctx context.Context, arg CreateUserAndRelationsTxParams) (UserAndRelationsTxResult, error)
}

// SQLStore provides all functions to execute SQL queries and transactions
type SQLStore struct {
	connPool *pgxpool.Pool
	*Queries
}

// NewStore creates a new store
func NewStore(connPool *pgxpool.Pool) Store {
	return &SQLStore{
		connPool: connPool,
		Queries:  New(connPool),
	}
}
