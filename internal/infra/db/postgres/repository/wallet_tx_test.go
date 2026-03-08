package repository

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	tb "github.com/tigerbeetle/tigerbeetle-go"
	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"
)

type mockTigerBeetleClient struct {
	lookupAccountsFn func(accountIDs []types.Uint128) ([]types.Account, error)
	createAccountsFn func(accounts []types.Account) ([]types.AccountEventResult, error)
}

func (m *mockTigerBeetleClient) CreateAccounts(accounts []types.Account) ([]types.AccountEventResult, error) {
	if m.createAccountsFn == nil {
		return nil, errors.New("unexpected CreateAccounts call")
	}
	return m.createAccountsFn(accounts)
}

func (m *mockTigerBeetleClient) CreateTransfers(transfers []types.Transfer) ([]types.TransferEventResult, error) {
	return nil, errors.New("unexpected CreateTransfers call")
}

func (m *mockTigerBeetleClient) LookupAccounts(accountIDs []types.Uint128) ([]types.Account, error) {
	if m.lookupAccountsFn == nil {
		return nil, errors.New("unexpected LookupAccounts call")
	}
	return m.lookupAccountsFn(accountIDs)
}

func (m *mockTigerBeetleClient) LookupTransfers(transferIDs []types.Uint128) ([]types.Transfer, error) {
	return nil, errors.New("unexpected LookupTransfers call")
}

func (m *mockTigerBeetleClient) GetAccountTransfers(filter types.AccountFilter) ([]types.Transfer, error) {
	return nil, errors.New("unexpected GetAccountTransfers call")
}

func (m *mockTigerBeetleClient) GetAccountBalances(filter types.AccountFilter) ([]types.AccountBalance, error) {
	return nil, errors.New("unexpected GetAccountBalances call")
}

func (m *mockTigerBeetleClient) QueryAccounts(filter types.QueryFilter) ([]types.Account, error) {
	return nil, errors.New("unexpected QueryAccounts call")
}

func (m *mockTigerBeetleClient) QueryTransfers(filter types.QueryFilter) ([]types.Transfer, error) {
	return nil, errors.New("unexpected QueryTransfers call")
}

func (m *mockTigerBeetleClient) GetChangeEvents(filter types.ChangeEventsFilter) ([]types.ChangeEvent, error) {
	return nil, errors.New("unexpected GetChangeEvents call")
}

func (m *mockTigerBeetleClient) Nop() error { return nil }

func (m *mockTigerBeetleClient) Close() {}

var _ tb.Client = (*mockTigerBeetleClient)(nil)

func TestEnsureTigerBeetleAccountsSkipsMatchingExistingAccounts(t *testing.T) {
	existing := buildTigerBeetleAccount(77, types.ToUint128(1001), 700, true)
	missing := buildTigerBeetleAccount(77, types.ToUint128(1002), 701, false)

	createCalled := false
	client := &mockTigerBeetleClient{
		lookupAccountsFn: func(accountIDs []types.Uint128) ([]types.Account, error) {
			require.Len(t, accountIDs, 2)
			return []types.Account{existing}, nil
		},
		createAccountsFn: func(accounts []types.Account) ([]types.AccountEventResult, error) {
			createCalled = true
			require.Equal(t, []types.Account{missing}, accounts)
			return nil, nil
		},
	}

	err := ensureTigerBeetleAccounts(client, []types.Account{existing, missing})
	require.NoError(t, err)
	require.True(t, createCalled)
}

func TestEnsureTigerBeetleAccountsFailsForConflictingExistingAccount(t *testing.T) {
	expected := buildTigerBeetleAccount(77, types.ToUint128(1001), 700, true)
	conflicting := expected
	conflicting.UserData64 = 99

	client := &mockTigerBeetleClient{
		lookupAccountsFn: func(accountIDs []types.Uint128) ([]types.Account, error) {
			return []types.Account{conflicting}, nil
		},
	}

	err := ensureTigerBeetleAccounts(client, []types.Account{expected})
	require.ErrorContains(t, err, "exists with unexpected attributes")
}

func TestEnsureTigerBeetleAccountsIgnoresCreateRaceWhenLookupMatches(t *testing.T) {
	expected := buildTigerBeetleAccount(77, types.ToUint128(1001), 700, true)
	lookupCalls := 0

	client := &mockTigerBeetleClient{
		lookupAccountsFn: func(accountIDs []types.Uint128) ([]types.Account, error) {
			lookupCalls++
			if lookupCalls == 1 {
				return nil, nil
			}
			return []types.Account{expected}, nil
		},
		createAccountsFn: func(accounts []types.Account) ([]types.AccountEventResult, error) {
			return []types.AccountEventResult{{Index: 0, Result: types.AccountExists}}, nil
		},
	}

	err := ensureTigerBeetleAccounts(client, []types.Account{expected})
	require.NoError(t, err)
	require.Equal(t, 2, lookupCalls)
}

func TestEnsureTigerBeetleAccountsReturnsCreateErrorWhenLookupStillMissing(t *testing.T) {
	expected := buildTigerBeetleAccount(77, types.ToUint128(1001), 700, true)
	lookupCalls := 0

	client := &mockTigerBeetleClient{
		lookupAccountsFn: func(accountIDs []types.Uint128) ([]types.Account, error) {
			lookupCalls++
			return nil, nil
		},
		createAccountsFn: func(accounts []types.Account) ([]types.AccountEventResult, error) {
			return []types.AccountEventResult{{Index: 0, Result: types.AccountExists}}, nil
		},
	}

	err := ensureTigerBeetleAccounts(client, []types.Account{expected})
	require.ErrorContains(t, err, "AccountExists")
	require.Equal(t, 2, lookupCalls)
}
