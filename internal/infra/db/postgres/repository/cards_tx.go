package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

// WalletCardTxParams is the params for CreateWalletCardTx.
type WalletCardTxParams struct {
	CreateWalletCardParams
	EventType    WalletCardEventType
	EventSuccess bool
	RemoteIP     pgtype.Text
	EventMetaData []byte
	AfterCreate  func(card WalletCard, event WalletCardEvent) error
}

// WalletCardTxResult is the result for CreateWalletCardTx.
type WalletCardTxResult struct {
	Card  WalletCard
	Event WalletCardEvent
}

// CreateWalletCardTx implements Store.CreateWalletCardTx.
func (store *SQLStore) CreateWalletCardTx(ctx context.Context, arg WalletCardTxParams) (WalletCardTxResult, error) {
	var result WalletCardTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		result.Card, err = q.CreateWalletCard(ctx, arg.CreateWalletCardParams)
		if err != nil {
			return err
		}

		result.Event, err = q.CreateWalletCardEvent(ctx, CreateWalletCardEventParams{
			WalletCardID:         result.Card.ID,
			WalletCardIdentifier: result.Card.Identifier.String(),
			UserIdentifier:       result.Card.UserIdentifier,
			EventType:            arg.EventType,
			Success:              arg.EventSuccess,
			RemoteIp:             arg.RemoteIP,
			MetaData:             arg.EventMetaData,
		})
		if err != nil {
			return err
		}

		if arg.AfterCreate != nil {
			return arg.AfterCreate(result.Card, result.Event)
		}

		return nil
	})

	return result, err
}