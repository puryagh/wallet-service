package models

import (
	"errors"

	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"
)

type LedgerAccount types.Account

// WalletAccount is the model for wallet account.
type WalletAccount struct {
	ID             int64  `json:"id" mapstructure:"id"`
	DebitsPending  uint64 `json:"debits_pending" mapstructure:"debits_pending"`
	DebitsPosted   uint64 `json:"debits_posted" mapstructure:"debits_posted"`
	CreditsPending uint64 `json:"credits_pending" mapstructure:"credits_pending"`
	CreditsPosted  uint64 `json:"credits_posted" mapstructure:"credits_posted"`
	UserData128    []byte `json:"user_data_128" mapstructure:"user_data_128"`
	UserData64     uint64 `json:"user_data_64" mapstructure:"user_data_64"`
	UserData32     uint32 `json:"user_data_32" mapstructure:"user_data_32"`
	Reserved       uint32 `json:"reserved" mapstructure:"reserved"`
	Ledger         uint32 `json:"ledger" mapstructure:"ledger"`
	Code           uint16 `json:"code" mapstructure:"code"`
	Flags          uint16 `json:"flags" mapstructure:"flags"`
	Timestamp      uint64 `json:"timestamp" mapstructure:"timestamp"`
}

var (
	ErrInvalidAccountID             = errors.New("invalid_account_id")
	ErrInvalidAccountDebitsPending  = errors.New("invalid_account_debits_pending")
	ErrInvalidAccountDebitsPosted   = errors.New("invalid_account_debits_posted")
	ErrInvalidAccountCreditsPending = errors.New("invalid_account_credits_pending")
	ErrInvalidAccountCreditsPosted  = errors.New("invalid_account_credits_posted")
)

// GetWalletAccount converts TigerBeetle account to WalletAccount.
func (a *LedgerAccount) GetWalletAccount() (*WalletAccount, error) {
	id := a.ID.BigInt()
	if !id.IsInt64() {
		return nil, ErrInvalidAccountID
	}

	debitsPending := a.DebitsPending.BigInt()
	if !debitsPending.IsUint64() {
		return nil, ErrInvalidAccountDebitsPending
	}

	debitsPosted := a.DebitsPosted.BigInt()
	if !debitsPosted.IsUint64() {
		return nil, ErrInvalidAccountDebitsPosted
	}

	creditsPending := a.CreditsPending.BigInt()
	if !creditsPending.IsUint64() {
		return nil, ErrInvalidAccountCreditsPending
	}

	creditsPosted := a.CreditsPosted.BigInt()
	if !creditsPosted.IsUint64() {
		return nil, ErrInvalidAccountCreditsPosted
	}

	walletAccount := &WalletAccount{
		ID:             id.Int64(),
		DebitsPending:  debitsPending.Uint64(),
		DebitsPosted:   debitsPosted.Uint64(),
		CreditsPending: creditsPending.Uint64(),
		CreditsPosted:  creditsPosted.Uint64(),
		UserData128:    a.UserData128[:],
		UserData64:     a.UserData64,
		UserData32:     a.UserData32,
		Reserved:       a.Reserved,
		Ledger:         a.Ledger,
		Code:           a.Code,
		Flags:          a.Flags,
		Timestamp:      a.Timestamp,
	}

	return walletAccount, nil
}
