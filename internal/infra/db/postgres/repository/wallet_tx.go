package repository

import (
	"context"
	"errors"
	"fmt"

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

		ledgerAccountID := types.ToUint128(uint64(arg.LedgerAccountID + result.Wallet.ID))
		expectedAccounts := []types.Account{
			buildTigerBeetleAccount(arg.UserID, ledgerAccountID, arg.LedgerAccountCode, true),
		}

		if err := ensureTigerBeetleAccounts(arg.TigerBeetle, expectedAccounts); err != nil {
			return err
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
		for _, wallet := range arg.CreateWalletParams {
			createdWallet, err := q.CreateWallet(ctx, wallet)
			if err != nil {
				return err
			}

			result.Wallets = append(result.Wallets, createdWallet)
		}

		ledgerAccounts := []types.Account{}

		for index, wallet := range result.Wallets {
			ledgerAccountID := types.ToUint128(uint64(wallet.LedgerAccountID + wallet.ID))

			linked := false
			if len(result.Wallets) > 1 && index < len(result.Wallets)-1 {
				linked = true
			}

			ledgerAccounts = append(ledgerAccounts, buildTigerBeetleAccount(arg.UserID, ledgerAccountID, wallet.LedgerAccountCode, linked))
		}

		if err := ensureTigerBeetleAccounts(arg.TigerBeetle, ledgerAccounts); err != nil {
			return err
		}

		return arg.AfterCreate(result.Wallets)
	})

	return result, err
}

func buildTigerBeetleAccount(userID uint64, ledgerAccountID types.Uint128, ledgerAccountCode int32, linked bool) types.Account {
	userData := types.ToUint128(userID)

	return types.Account{
		ID:          ledgerAccountID,
		UserData128: userData,
		UserData64:  userID,
		UserData32:  0,
		Ledger:      1,
		Code:        uint16(ledgerAccountCode),
		Flags: types.AccountFlags{
			DebitsMustNotExceedCredits: true,
			Linked:                     linked,
			CreditsMustNotExceedDebits: false,
			History:                    true,
			Imported:                   false,
			Closed:                     false,
		}.ToUint16(),
	}
}

func ensureTigerBeetleAccounts(client tb.Client, expectedAccounts []types.Account) error {
	accountsToCreate, err := filterExistingTigerBeetleAccounts(client, expectedAccounts)
	if err != nil {
		return err
	}

	if len(accountsToCreate) == 0 {
		return nil
	}

	results, err := client.CreateAccounts(accountsToCreate)
	if err != nil {
		return err
	}

	return validateTigerBeetleCreateResults(client, accountsToCreate, results)
}

func filterExistingTigerBeetleAccounts(client tb.Client, expectedAccounts []types.Account) ([]types.Account, error) {
	if len(expectedAccounts) == 0 {
		return nil, nil
	}

	existingAccounts, err := lookupTigerBeetleAccounts(client, expectedAccounts)
	if err != nil {
		return nil, err
	}

	accountsToCreate := make([]types.Account, 0, len(expectedAccounts))
	for _, expectedAccount := range expectedAccounts {
		existingAccount, exists := existingAccounts[tigerBeetleUint128Key(expectedAccount.ID)]
		if !exists {
			accountsToCreate = append(accountsToCreate, expectedAccount)
			continue
		}

		if !tigerBeetleAccountMatchesExpected(existingAccount, expectedAccount) {
			return nil, fmt.Errorf("tigerbeetle account %s exists with unexpected attributes", tigerBeetleUint128Key(expectedAccount.ID))
		}
	}

	return accountsToCreate, nil
}

func validateTigerBeetleCreateResults(client tb.Client, expectedAccounts []types.Account, results []types.AccountEventResult) error {
	if len(results) == 0 {
		return nil
	}

	existingAccounts, err := lookupTigerBeetleAccounts(client, expectedAccounts)
	if err != nil {
		return err
	}

	for _, result := range results {
		if result.Result == types.AccountOK {
			continue
		}
		if int(result.Index) >= len(expectedAccounts) {
			return fmt.Errorf("unexpected tigerbeetle create account result index %d", result.Index)
		}

		expectedAccount := expectedAccounts[result.Index]
		existingAccount, exists := existingAccounts[tigerBeetleUint128Key(expectedAccount.ID)]
		if exists && tigerBeetleAccountMatchesExpected(existingAccount, expectedAccount) {
			continue
		}

		return errors.New(result.Result.String())
	}

	return nil
}

func lookupTigerBeetleAccounts(client tb.Client, expectedAccounts []types.Account) (map[string]types.Account, error) {
	accountIDs := make([]types.Uint128, 0, len(expectedAccounts))
	for _, expectedAccount := range expectedAccounts {
		accountIDs = append(accountIDs, expectedAccount.ID)
	}

	accounts, err := client.LookupAccounts(accountIDs)
	if err != nil {
		return nil, err
	}

	result := make(map[string]types.Account, len(accounts))
	for _, account := range accounts {
		result[tigerBeetleUint128Key(account.ID)] = account
	}

	return result, nil
}

func tigerBeetleAccountMatchesExpected(existingAccount, expectedAccount types.Account) bool {
	if tigerBeetleUint128Key(existingAccount.ID) != tigerBeetleUint128Key(expectedAccount.ID) {
		return false
	}
	if tigerBeetleUint128Key(existingAccount.UserData128) != tigerBeetleUint128Key(expectedAccount.UserData128) {
		return false
	}
	if existingAccount.UserData64 != expectedAccount.UserData64 || existingAccount.UserData32 != expectedAccount.UserData32 {
		return false
	}
	if existingAccount.Ledger != expectedAccount.Ledger || existingAccount.Code != expectedAccount.Code {
		return false
	}

	existingFlags := existingAccount.AccountFlags()
	expectedFlags := expectedAccount.AccountFlags()

	return existingFlags.DebitsMustNotExceedCredits == expectedFlags.DebitsMustNotExceedCredits &&
		existingFlags.CreditsMustNotExceedDebits == expectedFlags.CreditsMustNotExceedDebits &&
		existingFlags.History == expectedFlags.History &&
		existingFlags.Imported == expectedFlags.Imported &&
		existingFlags.Closed == expectedFlags.Closed
}

func tigerBeetleUint128Key(value types.Uint128) string {
	bigInt := value.BigInt()
	return bigInt.String()
}
