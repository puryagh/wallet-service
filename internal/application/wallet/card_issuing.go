package wallet

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/liveutil/go-lib/contextutil"
	"github.com/liveutil/go-lib/models"
	"github.com/liveutil/wallet-service/internal/config"
	"github.com/liveutil/wallet-service/internal/infra/db/postgres/repository"
	"github.com/liveutil/wallet-service/pkg/card_iso"
)

const (
	defaultCardIssuerBIN               = "502229"
	defaultCardIssuerName              = "LIVEUTIL"
	defaultCardIssuerCountryCode       = "IR"
	defaultCardBrand                   = "LOCAL"
	defaultCardProductName             = "DEFAULT_MAGNETIC_CARD"
	defaultCardPANLength               = 16
	defaultCardCVVLength               = 3
	defaultCardExpiryMonths            = 36
	defaultCardServiceCode             = "201"
	defaultCardDiscretionaryDataLength = 7
	maxCardIssuingAttempts             = 5
)

var errActiveWalletCardAlreadyExists = errors.New("active card already exists for wallet")

type cardIssuerConfig struct {
	BIN                     string
	IssuerName              string
	CountryCode             string
	Brand                   repository.CardNetwork
	ProductName             string
	PANLength               int
	CVVLength               int
	ExpiryMonths            int
	ServiceCode             string
	DiscretionaryDataLength int
	HMACSecret              string
	AllowMagneticStripe     bool
}

// issuedWalletCard contains the stored card plus one-time sensitive issuance material.
type issuedWalletCard struct {
	Wallet    repository.Wallet
	Card      repository.WalletCard
	Event     repository.WalletCardEvent
	Sensitive card_iso.IssuedCard
}

// issueWalletCard issues one active magnetic card for a wallet owned by the context user.
func (s *service) issueWalletCard(ctx context.Context, walletIdentifier string) (*issuedWalletCard, error) {
	wallet, user, cardholderName, err := s.resolveContextUserWallet(ctx, walletIdentifier)
	if err != nil {
		return nil, err
	}

	_, err = s.repository.GetActiveWalletCardByWallet(ctx, repository.GetActiveWalletCardByWalletParams{
		WalletIdentifier: wallet.Identifier.String(),
		UserIdentifier:   user.Identifier,
	})
	if err == nil {
		return nil, errActiveWalletCardAlreadyExists
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	issuerCfg, err := loadCardIssuerConfig(s.config, s.applicationName)
	if err != nil {
		return nil, err
	}

	issuerBin, cardProduct, err := s.ensureCardProduct(ctx, issuerCfg)
	if err != nil {
		return nil, err
	}

	issued, err := s.issueUniqueWalletCard(ctx, issuerCfg, cardholderName)
	if err != nil {
		return nil, err
	}

	metadata, err := json.Marshal(map[string]any{
		"issuer_bin":      issuerBin.Bin,
		"brand":           string(issuerBin.Brand),
		"product_name":    cardProduct.Name,
		"masked_pan":      issued.MaskedPAN,
		"pan_last4":       issued.PANLast4,
		"issued_at":       time.Now().UTC().Format(time.RFC3339),
		"wallet_id":       wallet.ID,
		"wallet_asset_id": wallet.WalletAssetID,
	})
	if err != nil {
		return nil, err
	}

	remoteIP := pgtype.Text{}
	if ip := contextutil.RemoteIP(ctx); ip != "" {
		remoteIP = pgtype.Text{String: ip, Valid: true}
	}

	txResult, err := s.repository.CreateWalletCardTx(ctx, repository.WalletCardTxParams{
		CreateWalletCardParams: repository.CreateWalletCardParams{
			WalletID:            wallet.ID,
			WalletIdentifier:    wallet.Identifier.String(),
			UserIdentifier:      user.Identifier,
			CardProductID:       cardProduct.ID,
			IssuerBin:           issuerBin.Bin,
			Brand:               issuerBin.Brand,
			PanFingerprint:      issued.PANFingerprint,
			MaskedPan:           issued.MaskedPAN,
			PanLast4:            issued.PANLast4,
			ExpiryMonth:         int32(issued.ExpiryMonth),
			ExpiryYear:          int32(issued.ExpiryYear),
			CardholderName:      issued.CardholderName,
			ServiceCode:         issued.ServiceCode,
			CvvDigest:           issued.CVVDigest,
			CvvTwoDigest:        issued.CVV2Digest,
			Track1Digest:        pgtype.Text{String: issued.Track1Digest, Valid: true},
			Track2Digest:        issued.Track2Digest,
			PinDigest:           pgtype.Text{},
			LastAuthenticatedAt: pgtype.Timestamptz{},
			Status:              repository.WalletCardStatusACTIVE,
			MetaData:            metadata,
		},
		EventType:     repository.WalletCardEventTypeISSUED,
		EventSuccess:  true,
		RemoteIP:      remoteIP,
		EventMetaData: metadata,
	})
	if err != nil {
		return nil, err
	}

	return &issuedWalletCard{
		Wallet:    wallet,
		Card:      txResult.Card,
		Event:     txResult.Event,
		Sensitive: issued,
	}, nil
}

// listWalletCards returns all cards belonging to a context user's wallet.
func (s *service) listWalletCards(ctx context.Context, walletIdentifier string) ([]repository.WalletCard, error) {
	wallet, user, _, err := s.resolveContextUserWallet(ctx, walletIdentifier)
	if err != nil {
		return nil, err
	}

	return s.repository.GetWalletCardsByWallet(ctx, repository.GetWalletCardsByWalletParams{
		WalletIdentifier: wallet.Identifier.String(),
		UserIdentifier:   user.Identifier,
	})
}

// validateCardPAN validates an ISO PAN for future card flows.
func (s *service) validateCardPAN(pan string) error {
	return card_iso.ValidatePAN(pan)
}

// validateCardExpiry validates a card expiry for future card flows.
func (s *service) validateCardExpiry(month, year int) error {
	return card_iso.ValidateExpiry(month, year, time.Now().UTC())
}

func (s *service) resolveContextUserWallet(ctx context.Context, walletIdentifier string) (repository.Wallet, *models.SafeUserModel, string, error) {
	contextUser := &contextutil.ContextUser{}
	if err := contextutil.CatchUser(ctx, contextUser); err != nil {
		return repository.Wallet{}, nil, "", err
	}

	user, err := s.userServiceMeshClient.GetUserByID(ctx, uint64(contextUser.ID))
	if err != nil {
		return repository.Wallet{}, nil, "", err
	}

	wallet, err := s.repository.GetUserWalletByIdentifier(ctx, repository.GetUserWalletByIdentifierParams{
		Column1:        walletIdentifier,
		UserIdentifier: user.Identifier,
	})
	if err != nil {
		return repository.Wallet{}, nil, "", err
	}

	cardholderName, err := deriveCardholderName(user)
	if err != nil {
		return repository.Wallet{}, nil, "", err
	}

	return wallet, user, cardholderName, nil
}

func (s *service) ensureCardProduct(ctx context.Context, issuerCfg cardIssuerConfig) (repository.IssuerBin, repository.CardProduct, error) {
	metaData, err := json.Marshal(map[string]any{
		"managed_by": "wallet-service",
	})
	if err != nil {
		return repository.IssuerBin{}, repository.CardProduct{}, err
	}

	issuerBin, err := s.repository.EnsureIssuerBin(ctx, repository.EnsureIssuerBinParams{
		Active:      true,
		Bin:         issuerCfg.BIN,
		Brand:       issuerCfg.Brand,
		IssuerName:  issuerCfg.IssuerName,
		CountryCode: issuerCfg.CountryCode,
		MetaData:    metaData,
	})
	if err != nil {
		return repository.IssuerBin{}, repository.CardProduct{}, err
	}

	cardProduct, err := s.repository.EnsureCardProduct(ctx, repository.EnsureCardProductParams{
		IssuerBinID:         issuerBin.ID,
		Name:                issuerCfg.ProductName,
		PanLength:           int32(issuerCfg.PANLength),
		CvvLength:           int32(issuerCfg.CVVLength),
		ExpiryMonths:        int32(issuerCfg.ExpiryMonths),
		ServiceCode:         issuerCfg.ServiceCode,
		AllowMagneticStripe: issuerCfg.AllowMagneticStripe,
		MetaData:            metaData,
	})
	if err != nil {
		return repository.IssuerBin{}, repository.CardProduct{}, err
	}

	if !cardProduct.AllowMagneticStripe {
		return repository.IssuerBin{}, repository.CardProduct{}, errors.New("card product does not allow magnetic stripe issuance")
	}

	return issuerBin, cardProduct, nil
}

func (s *service) issueUniqueWalletCard(ctx context.Context, issuerCfg cardIssuerConfig, cardholderName string) (card_iso.IssuedCard, error) {
	for range maxCardIssuingAttempts {
		issued, err := card_iso.Issue(card_iso.IssueSpec{
			IssuerBIN:               issuerCfg.BIN,
			PANLength:               issuerCfg.PANLength,
			CVVLength:               issuerCfg.CVVLength,
			ExpiryMonths:            issuerCfg.ExpiryMonths,
			ServiceCode:             issuerCfg.ServiceCode,
			CardholderName:          cardholderName,
			HMACSecret:              issuerCfg.HMACSecret,
			DiscretionaryDataLength: issuerCfg.DiscretionaryDataLength,
		})
		if err != nil {
			return card_iso.IssuedCard{}, err
		}

		_, err = s.repository.GetWalletCardByPanFingerprint(ctx, issued.PANFingerprint)
		if errors.Is(err, pgx.ErrNoRows) {
			return issued, nil
		}
		if err != nil {
			return card_iso.IssuedCard{}, err
		}
	}

	return card_iso.IssuedCard{}, errors.New("failed to issue unique card after multiple attempts")
}

func loadCardIssuerConfig(cfg *config.Configuration, applicationName string) (cardIssuerConfig, error) {
	if cfg == nil {
		return cardIssuerConfig{}, errors.New("wallet configuration is required")
	}

	brand, err := parseCardBrand(cfg.CardBrand)
	if err != nil {
		return cardIssuerConfig{}, err
	}

	secret := strings.TrimSpace(cfg.CardHMACSecret)
	if secret == "" {
		secret = strings.TrimSpace(cfg.TokenSymmetricKey)
	}
	if secret == "" {
		return cardIssuerConfig{}, errors.New("card hmac secret is required")
	}

	issuerName := strings.TrimSpace(cfg.CardIssuerName)
	if issuerName == "" {
		issuerName = strings.ToUpper(strings.TrimSpace(applicationName))
		if issuerName == "" {
			issuerName = defaultCardIssuerName
		}
	}

	return cardIssuerConfig{
		BIN:                     defaultString(cfg.CardIssuerBin, defaultCardIssuerBIN),
		IssuerName:              issuerName,
		CountryCode:             strings.ToUpper(defaultString(cfg.CardIssuerCountryCode, defaultCardIssuerCountryCode)),
		Brand:                   brand,
		ProductName:             defaultString(cfg.CardProductName, defaultCardProductName),
		PANLength:               defaultInt(cfg.CardPanLength, defaultCardPANLength),
		CVVLength:               defaultInt(cfg.CardCvvLength, defaultCardCVVLength),
		ExpiryMonths:            defaultInt(cfg.CardExpiryMonths, defaultCardExpiryMonths),
		ServiceCode:             defaultString(cfg.CardServiceCode, defaultCardServiceCode),
		DiscretionaryDataLength: defaultInt(cfg.CardDiscretionaryDataLength, defaultCardDiscretionaryDataLength),
		HMACSecret:              secret,
		AllowMagneticStripe:     true,
	}, nil
}

func deriveCardholderName(user *models.SafeUserModel) (string, error) {
	if user == nil {
		return "", errors.New("user is required")
	}

	parts := []string{}
	if user.Profile != nil {
		if firstName := strings.TrimSpace(user.Profile.FirstName); firstName != "" {
			parts = append(parts, firstName)
		}
		if lastName := strings.TrimSpace(user.Profile.LastName); lastName != "" {
			parts = append(parts, lastName)
		}
	}

	rawName := strings.TrimSpace(strings.Join(parts, " "))
	if rawName == "" {
		rawName = user.Identifier
	}

	return card_iso.NormalizeCardholderName(rawName)
}

func parseCardBrand(value string) (repository.CardNetwork, error) {
	switch strings.ToUpper(strings.TrimSpace(defaultString(value, defaultCardBrand))) {
	case "LOCAL":
		return repository.CardNetworkLOCAL, nil
	case "VISA":
		return repository.CardNetworkVISA, nil
	case "MASTERCARD":
		return repository.CardNetworkMASTERCARD, nil
	case "AMEX":
		return repository.CardNetworkAMEX, nil
	default:
		return "", fmt.Errorf("unsupported card brand %q", value)
	}
}

func defaultString(value, fallback string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fallback
	}
	return trimmed
}

func defaultInt(value, fallback int) int {
	if value == 0 {
		return fallback
	}
	return value
}
