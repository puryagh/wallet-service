package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Store defines all functions to execute db queries and transactions
// any TX required queries should defined on this interface
type Store interface {
	Querier

	// Account TXes
	CreateAccountTx(ctx context.Context, arg AccountsTxParams) (AccountsTxResult, error)

	// Wallet TXes
	CreateWalletTx(ctx context.Context, arg WalletsTxParams) (WalletsTxResult, error)

	// Batch Wallet TXes
	BatchCreateWalletTx(ctx context.Context, arg BatchWalletsTxParams) (BatchWalletsTxResult, error)
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
