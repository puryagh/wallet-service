package wallet

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/liveutil/go-lib/contextutil"
	meshlib "github.com/liveutil/go-lib/framework/mesh"
	"github.com/liveutil/go-lib/models"
	"github.com/liveutil/wallet-service/internal/config"
	"github.com/liveutil/wallet-service/internal/infra/db/postgres/repository"
	"github.com/liveutil/wallet-service/pkg/card_iso"
	ulid "github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/require"
)

type mockStore struct {
	repository.Store
	getUserWalletByIdentifierFn     func(ctx context.Context, arg repository.GetUserWalletByIdentifierParams) (repository.Wallet, error)
	getActiveWalletCardByWalletFn   func(ctx context.Context, arg repository.GetActiveWalletCardByWalletParams) (repository.WalletCard, error)
	ensureIssuerBinFn               func(ctx context.Context, arg repository.EnsureIssuerBinParams) (repository.IssuerBin, error)
	ensureCardProductFn             func(ctx context.Context, arg repository.EnsureCardProductParams) (repository.CardProduct, error)
	getWalletCardByPanFingerprintFn func(ctx context.Context, panFingerprint string) (repository.WalletCard, error)
	createWalletCardTxFn            func(ctx context.Context, arg repository.WalletCardTxParams) (repository.WalletCardTxResult, error)
	getWalletCardsByWalletFn        func(ctx context.Context, arg repository.GetWalletCardsByWalletParams) ([]repository.WalletCard, error)
}

func (m *mockStore) GetUserWalletByIdentifier(ctx context.Context, arg repository.GetUserWalletByIdentifierParams) (repository.Wallet, error) {
	return m.getUserWalletByIdentifierFn(ctx, arg)
}

func (m *mockStore) GetActiveWalletCardByWallet(ctx context.Context, arg repository.GetActiveWalletCardByWalletParams) (repository.WalletCard, error) {
	return m.getActiveWalletCardByWalletFn(ctx, arg)
}

func (m *mockStore) EnsureIssuerBin(ctx context.Context, arg repository.EnsureIssuerBinParams) (repository.IssuerBin, error) {
	return m.ensureIssuerBinFn(ctx, arg)
}

func (m *mockStore) EnsureCardProduct(ctx context.Context, arg repository.EnsureCardProductParams) (repository.CardProduct, error) {
	return m.ensureCardProductFn(ctx, arg)
}

func (m *mockStore) GetWalletCardByPanFingerprint(ctx context.Context, panFingerprint string) (repository.WalletCard, error) {
	return m.getWalletCardByPanFingerprintFn(ctx, panFingerprint)
}

func (m *mockStore) CreateWalletCardTx(ctx context.Context, arg repository.WalletCardTxParams) (repository.WalletCardTxResult, error) {
	return m.createWalletCardTxFn(ctx, arg)
}

func (m *mockStore) GetWalletCardsByWallet(ctx context.Context, arg repository.GetWalletCardsByWalletParams) ([]repository.WalletCard, error) {
	return m.getWalletCardsByWalletFn(ctx, arg)
}

type mockUsersServiceMeshClient struct {
	getUserByIdentifierFn func(ctx context.Context, identifier string) (*models.SafeUserModel, error)
	getUserByIDFn         func(ctx context.Context, id uint64) (*models.SafeUserModel, error)
}

func (m *mockUsersServiceMeshClient) GetUserByIdentifier(ctx context.Context, identifier string) (*models.SafeUserModel, error) {
	if m.getUserByIdentifierFn == nil {
		return nil, errors.New("unexpected GetUserByIdentifier call")
	}
	return m.getUserByIdentifierFn(ctx, identifier)
}

func (m *mockUsersServiceMeshClient) GetUserByID(ctx context.Context, id uint64) (*models.SafeUserModel, error) {
	if m.getUserByIDFn == nil {
		return nil, errors.New("unexpected GetUserByID call")
	}
	return m.getUserByIDFn(ctx, id)
}

var _ meshlib.UsersServiceMeshClient = (*mockUsersServiceMeshClient)(nil)

func TestIssueWalletCardSuccess(t *testing.T) {
	wallet := repository.Wallet{ID: 11, Identifier: ulid.Make(), WalletAssetID: 99}
	user := &models.SafeUserModel{
		ID:         77,
		Identifier: ulid.Make().String(),
		Profile: &models.SafeProfileModel{
			FirstName: "زهرا",
			LastName:  "محمدی",
		},
	}
	issuerBin := repository.IssuerBin{ID: 3, Bin: "502229", Brand: repository.CardNetworkLOCAL}
	cardProduct := repository.CardProduct{ID: 7, Name: "DEFAULT_MAGNETIC_CARD", AllowMagneticStripe: true}

	var capturedTx repository.WalletCardTxParams
	store := &mockStore{
		getUserWalletByIdentifierFn: func(ctx context.Context, arg repository.GetUserWalletByIdentifierParams) (repository.Wallet, error) {
			require.Equal(t, wallet.Identifier.String(), arg.Column1)
			require.Equal(t, user.Identifier, arg.UserIdentifier)
			return wallet, nil
		},
		getActiveWalletCardByWalletFn: func(ctx context.Context, arg repository.GetActiveWalletCardByWalletParams) (repository.WalletCard, error) {
			return repository.WalletCard{}, pgx.ErrNoRows
		},
		ensureIssuerBinFn: func(ctx context.Context, arg repository.EnsureIssuerBinParams) (repository.IssuerBin, error) {
			require.Equal(t, "502229", arg.Bin)
			return issuerBin, nil
		},
		ensureCardProductFn: func(ctx context.Context, arg repository.EnsureCardProductParams) (repository.CardProduct, error) {
			require.Equal(t, issuerBin.ID, arg.IssuerBinID)
			return cardProduct, nil
		},
		getWalletCardByPanFingerprintFn: func(ctx context.Context, panFingerprint string) (repository.WalletCard, error) {
			require.NotEmpty(t, panFingerprint)
			return repository.WalletCard{}, pgx.ErrNoRows
		},
		createWalletCardTxFn: func(ctx context.Context, arg repository.WalletCardTxParams) (repository.WalletCardTxResult, error) {
			capturedTx = arg
			return repository.WalletCardTxResult{
				Card: repository.WalletCard{
					ID:               123,
					Identifier:       ulid.Make(),
					WalletID:         wallet.ID,
					WalletIdentifier: wallet.Identifier.String(),
					UserIdentifier:   user.Identifier,
					PanFingerprint:   arg.PanFingerprint,
					MaskedPan:        arg.MaskedPan,
					PanLast4:         arg.PanLast4,
					ExpiryMonth:      arg.ExpiryMonth,
					ExpiryYear:       arg.ExpiryYear,
					CardholderName:   arg.CardholderName,
					ServiceCode:      arg.ServiceCode,
					Status:           arg.Status,
				},
				Event: repository.WalletCardEvent{
					ID:           456,
					Identifier:   ulid.Make(),
					WalletCardID: 123,
					EventType:    arg.EventType,
					Success:      arg.EventSuccess,
					RemoteIp:     arg.RemoteIP,
				},
			}, nil
		},
	}

	svc := &service{
		repository: store,
		config: &config.Configuration{
			TokenSymmetricKey: "top-secret",
		},
		applicationName: "wallet-service",
		userServiceMeshClient: &mockUsersServiceMeshClient{getUserByIDFn: func(ctx context.Context, id uint64) (*models.SafeUserModel, error) {
			require.Equal(t, uint64(77), id)
			return user, nil
		}},
	}

	ctx := context.WithValue(context.Background(), contextutil.ContextUserKey, contextutil.ContextUser{ID: 77})
	ctx = context.WithValue(ctx, contextutil.ContextRemoteIPKey, "127.0.0.1")

	result, err := svc.issueWalletCard(ctx, wallet.Identifier.String())
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, wallet.ID, result.Wallet.ID)
	require.Equal(t, repository.WalletCardStatusACTIVE, result.Card.Status)
	require.NoError(t, card_iso.ValidatePAN(result.Sensitive.PAN))
	require.Equal(t, "زهرا محمدی", result.Card.CardholderName)
	require.Equal(t, result.Sensitive.MaskedPAN, result.Card.MaskedPan)
	require.Contains(t, result.Sensitive.Track1, "زهرا محمدی")
	require.True(t, card_iso.VerifyDigest("top-secret", "pan_fingerprint", capturedTx.PanFingerprint, result.Sensitive.PAN))
	require.True(t, card_iso.VerifyDigest("top-secret", "track2", capturedTx.Track2Digest, result.Sensitive.Track2))
	require.Equal(t, repository.WalletCardEventTypeISSUED, capturedTx.EventType)
	require.True(t, capturedTx.EventSuccess)
	require.True(t, capturedTx.Track1Digest.Valid)
	require.True(t, capturedTx.RemoteIP.Valid)
	require.Equal(t, "127.0.0.1", capturedTx.RemoteIP.String)
	require.NotEmpty(t, capturedTx.MetaData)
	require.NotEmpty(t, capturedTx.EventMetaData)
}

func TestIssueWalletCardRejectsExistingActiveCard(t *testing.T) {
	wallet := repository.Wallet{ID: 11, Identifier: ulid.Make()}
	user := &models.SafeUserModel{ID: 77, Identifier: ulid.Make().String()}

	store := &mockStore{
		getUserWalletByIdentifierFn: func(ctx context.Context, arg repository.GetUserWalletByIdentifierParams) (repository.Wallet, error) {
			return wallet, nil
		},
		getActiveWalletCardByWalletFn: func(ctx context.Context, arg repository.GetActiveWalletCardByWalletParams) (repository.WalletCard, error) {
			return repository.WalletCard{ID: 1}, nil
		},
	}

	svc := &service{
		repository: store,
		config: &config.Configuration{
			TokenSymmetricKey: "top-secret",
		},
		userServiceMeshClient: &mockUsersServiceMeshClient{getUserByIDFn: func(ctx context.Context, id uint64) (*models.SafeUserModel, error) {
			return user, nil
		}},
	}

	ctx := context.WithValue(context.Background(), contextutil.ContextUserKey, contextutil.ContextUser{ID: 77})
	result, err := svc.issueWalletCard(ctx, wallet.Identifier.String())
	require.Nil(t, result)
	require.ErrorContains(t, err, "active card already exists")
}

func TestListWalletCardsUsesOwnedWallet(t *testing.T) {
	wallet := repository.Wallet{ID: 11, Identifier: ulid.Make()}
	user := &models.SafeUserModel{ID: 77, Identifier: ulid.Make().String()}
	expectedCards := []repository.WalletCard{{ID: 1}, {ID: 2}}

	store := &mockStore{
		getUserWalletByIdentifierFn: func(ctx context.Context, arg repository.GetUserWalletByIdentifierParams) (repository.Wallet, error) {
			return wallet, nil
		},
		getWalletCardsByWalletFn: func(ctx context.Context, arg repository.GetWalletCardsByWalletParams) ([]repository.WalletCard, error) {
			require.Equal(t, wallet.Identifier.String(), arg.WalletIdentifier)
			require.Equal(t, user.Identifier, arg.UserIdentifier)
			return expectedCards, nil
		},
	}

	svc := &service{
		repository: store,
		userServiceMeshClient: &mockUsersServiceMeshClient{getUserByIDFn: func(ctx context.Context, id uint64) (*models.SafeUserModel, error) {
			return user, nil
		}},
	}

	ctx := context.WithValue(context.Background(), contextutil.ContextUserKey, contextutil.ContextUser{ID: 77})
	cards, err := svc.listWalletCards(ctx, wallet.Identifier.String())
	require.NoError(t, err)
	require.Equal(t, expectedCards, cards)
}

func TestValidateCardHelpers(t *testing.T) {
	svc := &service{}
	require.NoError(t, svc.validateCardPAN("4242 4242 4242 4242"))
	require.Error(t, svc.validateCardPAN("1234"))
	require.NoError(t, svc.validateCardExpiry(12, 2099))
	require.Error(t, svc.validateCardExpiry(1, 2000))
}

func TestCreateWalletCardTxParamsExposeRemoteIPType(t *testing.T) {
	var params repository.WalletCardTxParams
	params.RemoteIP = pgtype.Text{String: "127.0.0.1", Valid: true}
	require.True(t, params.RemoteIP.Valid)
}
