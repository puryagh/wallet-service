package wallet

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/liveutil/go-lib/contextutil"
	"github.com/liveutil/go-lib/models"
	"github.com/liveutil/wallet-service/internal/config"
	"github.com/liveutil/wallet-service/internal/infra/db/postgres/repository"
	ulid "github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/require"
	tb "github.com/tigerbeetle/tigerbeetle-go"
	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"
	"google.golang.org/protobuf/types/known/emptypb"
)

type contextUserAccountsStore struct {
	repository.Store
	getUserAccountFn                func(ctx context.Context, userIdentifier string) (repository.Account, error)
	createAccountTxFn               func(ctx context.Context, arg repository.AccountsTxParams) (repository.AccountsTxResult, error)
	getUserWalletsFn                func(ctx context.Context, arg repository.GetUserWalletsParams) ([]repository.Wallet, error)
	getWalletAssetBySymbolFn        func(ctx context.Context, symbol string) (repository.WalletAsset, error)
	getWalletAssetByIdentifierFn    func(ctx context.Context, identifier string) (repository.WalletAsset, error)
	batchCreateWalletTxFn           func(ctx context.Context, arg repository.BatchWalletsTxParams) (repository.BatchWalletsTxResult, error)
	getActiveWalletCardByWalletFn   func(ctx context.Context, arg repository.GetActiveWalletCardByWalletParams) (repository.WalletCard, error)
	getUserWalletByIdentifierFn     func(ctx context.Context, arg repository.GetUserWalletByIdentifierParams) (repository.Wallet, error)
	ensureIssuerBinFn               func(ctx context.Context, arg repository.EnsureIssuerBinParams) (repository.IssuerBin, error)
	ensureCardProductFn             func(ctx context.Context, arg repository.EnsureCardProductParams) (repository.CardProduct, error)
	getWalletCardByPanFingerprintFn func(ctx context.Context, panFingerprint string) (repository.WalletCard, error)
	createWalletCardTxFn            func(ctx context.Context, arg repository.WalletCardTxParams) (repository.WalletCardTxResult, error)
}

func (m *contextUserAccountsStore) GetUserAccount(ctx context.Context, userIdentifier string) (repository.Account, error) {
	return m.getUserAccountFn(ctx, userIdentifier)
}

func (m *contextUserAccountsStore) CreateAccountTx(ctx context.Context, arg repository.AccountsTxParams) (repository.AccountsTxResult, error) {
	return m.createAccountTxFn(ctx, arg)
}

func (m *contextUserAccountsStore) GetUserWallets(ctx context.Context, arg repository.GetUserWalletsParams) ([]repository.Wallet, error) {
	return m.getUserWalletsFn(ctx, arg)
}

func (m *contextUserAccountsStore) GetWalletAssetBySymbol(ctx context.Context, symbol string) (repository.WalletAsset, error) {
	return m.getWalletAssetBySymbolFn(ctx, symbol)
}

func (m *contextUserAccountsStore) GetWalletAssetByIdentifier(ctx context.Context, identifier string) (repository.WalletAsset, error) {
	return m.getWalletAssetByIdentifierFn(ctx, identifier)
}

func (m *contextUserAccountsStore) BatchCreateWalletTx(ctx context.Context, arg repository.BatchWalletsTxParams) (repository.BatchWalletsTxResult, error) {
	return m.batchCreateWalletTxFn(ctx, arg)
}

func (m *contextUserAccountsStore) GetActiveWalletCardByWallet(ctx context.Context, arg repository.GetActiveWalletCardByWalletParams) (repository.WalletCard, error) {
	return m.getActiveWalletCardByWalletFn(ctx, arg)
}

func (m *contextUserAccountsStore) GetUserWalletByIdentifier(ctx context.Context, arg repository.GetUserWalletByIdentifierParams) (repository.Wallet, error) {
	return m.getUserWalletByIdentifierFn(ctx, arg)
}

func (m *contextUserAccountsStore) EnsureIssuerBin(ctx context.Context, arg repository.EnsureIssuerBinParams) (repository.IssuerBin, error) {
	return m.ensureIssuerBinFn(ctx, arg)
}

func (m *contextUserAccountsStore) EnsureCardProduct(ctx context.Context, arg repository.EnsureCardProductParams) (repository.CardProduct, error) {
	return m.ensureCardProductFn(ctx, arg)
}

func (m *contextUserAccountsStore) GetWalletCardByPanFingerprint(ctx context.Context, panFingerprint string) (repository.WalletCard, error) {
	return m.getWalletCardByPanFingerprintFn(ctx, panFingerprint)
}

func (m *contextUserAccountsStore) CreateWalletCardTx(ctx context.Context, arg repository.WalletCardTxParams) (repository.WalletCardTxResult, error) {
	return m.createWalletCardTxFn(ctx, arg)
}

type mockTigerBeetleClient struct {
	lookupAccountsFn func(accountIDs []types.Uint128) ([]types.Account, error)
}

func (m *mockTigerBeetleClient) CreateAccounts(accounts []types.Account) ([]types.AccountEventResult, error) {
	return nil, errors.New("unexpected CreateAccounts call")
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

func TestContextUserAccountsCreatesCardsForInitialWallets(t *testing.T) {
	now := time.Now().UTC()
	user := &models.SafeUserModel{ID: 77, Identifier: ulid.Make().String(), Profile: &models.SafeProfileModel{FirstName: "زهرا", LastName: "محمدی"}}
	account := repository.Account{ID: 12, Identifier: ulid.Make(), Title: "Default Account", Description: pgtype.Text{String: "Default Account", Valid: true}, UserIdentifier: user.Identifier, MetaData: []byte(`{"holder_name":"زهرا محمدی"}`), CreatedAt: now, UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true}}
	assetIRR := repository.WalletAsset{ID: 1, Identifier: ulid.Make(), Symbol: "IRR", Code: "1", Title: "ریال", Description: pgtype.Text{String: "ریال ایران", Valid: true}, Unit: repository.WalletAssetUnitPIECE, UnitTitle: pgtype.Text{String: "ریال", Valid: true}, Network: pgtype.Text{String: "IRCBN", Valid: true}, IconUrl: pgtype.Text{String: "irr.svg", Valid: true}, MetaData: []byte(`{}`), LedgerCode: 1}
	assetXAG := repository.WalletAsset{ID: 2, Identifier: ulid.Make(), Symbol: "XAG", Code: "2", Title: "نقره", Description: pgtype.Text{String: "میلی‌گرم نقره", Valid: true}, Unit: repository.WalletAssetUnitMILLIGRAM, UnitTitle: pgtype.Text{String: "میلی‌گرم", Valid: true}, Network: pgtype.Text{String: "LIVEUTIL", Valid: true}, IconUrl: pgtype.Text{String: "xag.svg", Valid: true}, MetaData: []byte(`{}`), LedgerCode: 2}
	walletIRR := repository.Wallet{ID: 101, Identifier: ulid.Make(), UserIdentifier: user.Identifier, AccountIdentifier: account.Identifier.String(), AccountID: account.ID, AssetIdentifier: assetIRR.Identifier.String(), WalletAssetID: assetIRR.ID, MetaData: []byte(`{"holder_name":"زهرا محمدی"}`), LedgerAccountID: 9001, LedgerAccountCode: assetIRR.LedgerCode, Status: repository.WalletAccountStatusACTIVE, CreatedAt: now, UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true}}
	walletXAG := repository.Wallet{ID: 102, Identifier: ulid.Make(), UserIdentifier: user.Identifier, AccountIdentifier: account.Identifier.String(), AccountID: account.ID, AssetIdentifier: assetXAG.Identifier.String(), WalletAssetID: assetXAG.ID, MetaData: []byte(`{"holder_name":"زهرا محمدی"}`), LedgerAccountID: 9002, LedgerAccountCode: assetXAG.LedgerCode, Status: repository.WalletAccountStatusACTIVE, CreatedAt: now, UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true}}
	issuerBin := repository.IssuerBin{ID: 5, Bin: "502229", Brand: repository.CardNetworkLOCAL}
	cardProduct := repository.CardProduct{ID: 8, Name: "DEFAULT_MAGNETIC_CARD", AllowMagneticStripe: true}

	createWalletCardCount := 0
	store := &contextUserAccountsStore{
		getUserAccountFn: func(ctx context.Context, userIdentifier string) (repository.Account, error) {
			return account, nil
		},
		getUserWalletsFn: func(ctx context.Context, arg repository.GetUserWalletsParams) ([]repository.Wallet, error) {
			return []repository.Wallet{}, nil
		},
		getWalletAssetBySymbolFn: func(ctx context.Context, symbol string) (repository.WalletAsset, error) {
			switch symbol {
			case "IRR":
				return assetIRR, nil
			case "XAG":
				return assetXAG, nil
			default:
				return repository.WalletAsset{}, errors.New("unexpected asset symbol")
			}
		},
		batchCreateWalletTxFn: func(ctx context.Context, arg repository.BatchWalletsTxParams) (repository.BatchWalletsTxResult, error) {
			require.Len(t, arg.CreateWalletParams, 2)
			return repository.BatchWalletsTxResult{Wallets: []repository.Wallet{walletIRR, walletXAG}}, nil
		},
		getActiveWalletCardByWalletFn: func(ctx context.Context, arg repository.GetActiveWalletCardByWalletParams) (repository.WalletCard, error) {
			return repository.WalletCard{}, pgx.ErrNoRows
		},
		getUserWalletByIdentifierFn: func(ctx context.Context, arg repository.GetUserWalletByIdentifierParams) (repository.Wallet, error) {
			switch arg.Column1 {
			case walletIRR.Identifier.String():
				return walletIRR, nil
			case walletXAG.Identifier.String():
				return walletXAG, nil
			default:
				return repository.Wallet{}, pgx.ErrNoRows
			}
		},
		ensureIssuerBinFn: func(ctx context.Context, arg repository.EnsureIssuerBinParams) (repository.IssuerBin, error) {
			return issuerBin, nil
		},
		ensureCardProductFn: func(ctx context.Context, arg repository.EnsureCardProductParams) (repository.CardProduct, error) {
			return cardProduct, nil
		},
		getWalletCardByPanFingerprintFn: func(ctx context.Context, panFingerprint string) (repository.WalletCard, error) {
			return repository.WalletCard{}, pgx.ErrNoRows
		},
		createWalletCardTxFn: func(ctx context.Context, arg repository.WalletCardTxParams) (repository.WalletCardTxResult, error) {
			createWalletCardCount++
			return repository.WalletCardTxResult{Card: repository.WalletCard{ID: int64(createWalletCardCount), Identifier: ulid.Make(), WalletID: arg.WalletID, WalletIdentifier: arg.WalletIdentifier, UserIdentifier: arg.UserIdentifier, Status: arg.Status, CardholderName: arg.CardholderName}}, nil
		},
	}

	svc := &service{repository: store, config: &config.Configuration{AccountsInitialAssets: []string{"IRR", "XAG"}, TokenSymmetricKey: "top-secret", TigerbeetleReservedAccountNumber: 9000}, applicationName: "wallet-service", userServiceMeshClient: &mockUsersServiceMeshClient{getUserByIDFn: func(ctx context.Context, id uint64) (*models.SafeUserModel, error) { return user, nil }}}
	ctx := context.WithValue(context.Background(), contextutil.ContextUserKey, contextutil.ContextUser{ID: int64(user.ID)})

	response, err := svc.ContextUserAccounts(ctx, &emptypb.Empty{})
	require.NoError(t, err)
	require.Len(t, response.Wallets, 2)
	require.Equal(t, 2, createWalletCardCount)
	require.Equal(t, "context user account and related wallets fetched success", response.Message)
	for _, wallet := range response.Wallets {
		require.Equal(t, "0", wallet.Balance.Balance)
	}
}

func TestContextUserAccountsIssuesMissingCardForExistingWallet(t *testing.T) {
	now := time.Now().UTC()
	user := &models.SafeUserModel{ID: 77, Identifier: ulid.Make().String(), Profile: &models.SafeProfileModel{FirstName: "زهرا", LastName: "محمدی"}}
	account := repository.Account{ID: 12, Identifier: ulid.Make(), Title: "Default Account", Description: pgtype.Text{String: "Default Account", Valid: true}, UserIdentifier: user.Identifier, MetaData: []byte(`{}`), CreatedAt: now, UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true}}
	asset := repository.WalletAsset{ID: 1, Identifier: ulid.Make(), Symbol: "IRR", Code: "1", Title: "ریال", Description: pgtype.Text{String: "ریال ایران", Valid: true}, Unit: repository.WalletAssetUnitPIECE, UnitTitle: pgtype.Text{String: "ریال", Valid: true}, Network: pgtype.Text{String: "IRCBN", Valid: true}, IconUrl: pgtype.Text{String: "irr.svg", Valid: true}, MetaData: []byte(`{}`), LedgerCode: 1}
	wallet := repository.Wallet{ID: 101, Identifier: ulid.Make(), UserIdentifier: user.Identifier, AccountIdentifier: account.Identifier.String(), AccountID: account.ID, AssetIdentifier: asset.Identifier.String(), WalletAssetID: asset.ID, MetaData: []byte(`{}`), LedgerAccountID: 9001, LedgerAccountCode: asset.LedgerCode, Status: repository.WalletAccountStatusACTIVE, CreatedAt: now, UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true}}
	issuerBin := repository.IssuerBin{ID: 5, Bin: "502229", Brand: repository.CardNetworkLOCAL}
	cardProduct := repository.CardProduct{ID: 8, Name: "DEFAULT_MAGNETIC_CARD", AllowMagneticStripe: true}

	createWalletCardCount := 0
	store := &contextUserAccountsStore{
		getUserAccountFn: func(ctx context.Context, userIdentifier string) (repository.Account, error) { return account, nil },
		getUserWalletsFn: func(ctx context.Context, arg repository.GetUserWalletsParams) ([]repository.Wallet, error) {
			return []repository.Wallet{wallet}, nil
		},
		getWalletAssetByIdentifierFn: func(ctx context.Context, identifier string) (repository.WalletAsset, error) { return asset, nil },
		getActiveWalletCardByWalletFn: func(ctx context.Context, arg repository.GetActiveWalletCardByWalletParams) (repository.WalletCard, error) {
			return repository.WalletCard{}, pgx.ErrNoRows
		},
		getUserWalletByIdentifierFn: func(ctx context.Context, arg repository.GetUserWalletByIdentifierParams) (repository.Wallet, error) {
			return wallet, nil
		},
		ensureIssuerBinFn: func(ctx context.Context, arg repository.EnsureIssuerBinParams) (repository.IssuerBin, error) {
			return issuerBin, nil
		},
		ensureCardProductFn: func(ctx context.Context, arg repository.EnsureCardProductParams) (repository.CardProduct, error) {
			return cardProduct, nil
		},
		getWalletCardByPanFingerprintFn: func(ctx context.Context, panFingerprint string) (repository.WalletCard, error) {
			return repository.WalletCard{}, pgx.ErrNoRows
		},
		createWalletCardTxFn: func(ctx context.Context, arg repository.WalletCardTxParams) (repository.WalletCardTxResult, error) {
			createWalletCardCount++
			return repository.WalletCardTxResult{Card: repository.WalletCard{ID: 1, Identifier: ulid.Make(), WalletID: wallet.ID, WalletIdentifier: wallet.Identifier.String(), UserIdentifier: user.Identifier, Status: repository.WalletCardStatusACTIVE}}, nil
		},
	}

	tbClient := &mockTigerBeetleClient{lookupAccountsFn: func(accountIDs []types.Uint128) ([]types.Account, error) {
		return []types.Account{{CreditsPosted: types.ToUint128(100), DebitsPosted: types.ToUint128(25), CreditsPending: types.ToUint128(10), DebitsPending: types.ToUint128(2)}}, nil
	}}
	svc := &service{repository: store, config: &config.Configuration{TokenSymmetricKey: "top-secret"}, applicationName: "wallet-service", tigerbeetle: tbClient, userServiceMeshClient: &mockUsersServiceMeshClient{getUserByIDFn: func(ctx context.Context, id uint64) (*models.SafeUserModel, error) { return user, nil }}}
	ctx := context.WithValue(context.Background(), contextutil.ContextUserKey, contextutil.ContextUser{ID: int64(user.ID)})

	response, err := svc.ContextUserAccounts(ctx, &emptypb.Empty{})
	require.NoError(t, err)
	require.Len(t, response.Wallets, 1)
	require.Equal(t, 1, createWalletCardCount)
	require.Equal(t, "75", response.Wallets[0].Balance.Balance)
	require.Equal(t, "8", response.Wallets[0].Balance.Pending)
}

func TestContextUserAccountsSkipsExistingActiveWalletCard(t *testing.T) {
	now := time.Now().UTC()
	user := &models.SafeUserModel{ID: 77, Identifier: ulid.Make().String()}
	account := repository.Account{ID: 12, Identifier: ulid.Make(), Title: "Default Account", Description: pgtype.Text{String: "Default Account", Valid: true}, UserIdentifier: user.Identifier, MetaData: []byte(`{}`), CreatedAt: now, UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true}}
	asset := repository.WalletAsset{ID: 1, Identifier: ulid.Make(), Symbol: "IRR", Code: "1", Title: "ریال", Description: pgtype.Text{String: "ریال ایران", Valid: true}, Unit: repository.WalletAssetUnitPIECE, UnitTitle: pgtype.Text{String: "ریال", Valid: true}, Network: pgtype.Text{String: "IRCBN", Valid: true}, IconUrl: pgtype.Text{String: "irr.svg", Valid: true}, MetaData: []byte(`{}`), LedgerCode: 1}
	wallet := repository.Wallet{ID: 101, Identifier: ulid.Make(), UserIdentifier: user.Identifier, AccountIdentifier: account.Identifier.String(), AccountID: account.ID, AssetIdentifier: asset.Identifier.String(), WalletAssetID: asset.ID, MetaData: []byte(`{}`), LedgerAccountID: 9001, LedgerAccountCode: asset.LedgerCode, Status: repository.WalletAccountStatusACTIVE, CreatedAt: now, UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true}}

	store := &contextUserAccountsStore{
		getUserAccountFn: func(ctx context.Context, userIdentifier string) (repository.Account, error) { return account, nil },
		getUserWalletsFn: func(ctx context.Context, arg repository.GetUserWalletsParams) ([]repository.Wallet, error) {
			return []repository.Wallet{wallet}, nil
		},
		getWalletAssetByIdentifierFn: func(ctx context.Context, identifier string) (repository.WalletAsset, error) { return asset, nil },
		getActiveWalletCardByWalletFn: func(ctx context.Context, arg repository.GetActiveWalletCardByWalletParams) (repository.WalletCard, error) {
			return repository.WalletCard{ID: 1, WalletID: wallet.ID, WalletIdentifier: wallet.Identifier.String(), UserIdentifier: user.Identifier, Status: repository.WalletCardStatusACTIVE}, nil
		},
	}

	tbClient := &mockTigerBeetleClient{lookupAccountsFn: func(accountIDs []types.Uint128) ([]types.Account, error) {
		return []types.Account{{CreditsPosted: types.ToUint128(10), DebitsPosted: types.ToUint128(3)}}, nil
	}}
	svc := &service{repository: store, config: &config.Configuration{TokenSymmetricKey: "top-secret"}, applicationName: "wallet-service", tigerbeetle: tbClient, userServiceMeshClient: &mockUsersServiceMeshClient{getUserByIDFn: func(ctx context.Context, id uint64) (*models.SafeUserModel, error) { return user, nil }}}
	ctx := context.WithValue(context.Background(), contextutil.ContextUserKey, contextutil.ContextUser{ID: int64(user.ID)})

	response, err := svc.ContextUserAccounts(ctx, &emptypb.Empty{})
	require.NoError(t, err)
	require.Len(t, response.Wallets, 1)
	require.Equal(t, "7", response.Wallets[0].Balance.Balance)
	if store.createWalletCardTxFn != nil {
		t.Fatal("card issuing should not be invoked when an active card already exists")
	}
}
