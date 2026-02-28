package wallet

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/liveutil/go-lib/contextutil"
	"github.com/liveutil/go-lib/pgutil"
	"github.com/liveutil/wallet-service/internal/abstract/pb"
	"github.com/liveutil/wallet-service/internal/infra/db/postgres/repository"
	"github.com/oklog/ulid/v2"
	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ContextUserAccounts implements [pb.WalletServiceServer].
func (s *service) ContextUserAccounts(ctx context.Context, req *emptypb.Empty) (*pb.ContextUserWalletResponse, error) {
	response := &pb.ContextUserWalletResponse{
		Wallets: []*pb.Wallet{},
	}

	contextUser := &contextutil.ContextUser{}
	if err := contextutil.CatchUser(ctx, contextUser); err != nil {
		return nil, err
	}

	user, err := s.userServiceMeshClient.GetUserByID(ctx, uint64(contextUser.ID))
	if err != nil {
		return nil, err
	}

	account, err := s.repository.GetUserAccount(ctx, user.Identifier)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			walletMetadata := map[string]string{
				"holder_name": fmt.Sprintf("%s %s", user.Profile.FirstName, user.Profile.LastName),
			}

			metadata, err := json.Marshal(walletMetadata)
			if err != nil {
				return nil, err
			}

			baseAsset, err := s.repository.GetWalletAssetBySymbol(ctx, s.config.AccountsBaseAssetSymbol)
			if err != nil {
				return nil, err
			}

			createAccountParams := repository.AccountsTxParams{
				CreateAccountParams: repository.CreateAccountParams{
					Title:               "Default Account",
					Description:         pgtype.Text{String: "Default Account", Valid: true},
					Status:              repository.AccountStatusISSUED,
					Banned:              false,
					UserIdentifier:      user.Identifier,
					BaseAssetIdentifier: baseAsset.Identifier.String(),
					MetaData:            metadata,
					ExpiresAt:           account.ExpiresAt,
				},
				AfterCreate: func(account repository.Account) error {
					return nil
				},
			}

			// create new account for user
			txTResult, err := s.repository.CreateAccountTx(ctx, createAccountParams)
			if err != nil {
				return nil, err
			}

			account = txTResult.Account

			for index, assetSymbol := range s.config.AccountsInitialAssets {
				asset, err := s.repository.GetWalletAssetBySymbol(ctx, assetSymbol)
				if err != nil {
					return nil, err
				}

				// generate ledger account id for wallet
				accountID := uint64(account.ID + int64(index))
				accountID += uint64(s.config.TigerbeetleReservedAccountNumber)

				ledgerAccountId := types.ToUint128(accountID)
				dbLedgerAccountId := ledgerAccountId.BigInt()

				createWalletParams := repository.WalletsTxParams{
					TigerBeetle: s.tigerbeetle,
					UserID:      user.ID,
					CreateWalletParams: repository.CreateWalletParams{
						UserIdentifier:       user.Identifier,
						AccountIdentifier:    account.Identifier.String(),
						AccountID:            account.ID,
						AssetIdentifier:      asset.Identifier.String(),
						WalletAssetID:        asset.ID,
						MetaData:             metadata,
						LedgerAccountCode:    asset.LedgerCode,
						Status:               repository.WalletAccountStatusACTIVE,
						PrimaryAccountNumber: ulid.Make().String()[:24],
						LedgerAccountID:      dbLedgerAccountId.Int64(),
					},
					AfterCreate: func(wallet repository.Wallet) error {
						return nil
					},
				}

				createWalletResult, err := s.repository.CreateWalletTx(ctx, createWalletParams)
				if err != nil {
					return nil, err
				}

				response.Wallets = append(response.Wallets, &pb.Wallet{
					CreatedAt:       timestamppb.New(createWalletResult.Wallet.CreatedAt),
					UpdatedAt:       timestamppb.New(createWalletResult.Wallet.UpdatedAt.Time),
					DeletedAt:       timestamppb.New(createWalletResult.Wallet.DeletedAt.Time),
					Banned:          createAccountParams.Banned,
					Identifier:      createWalletResult.Wallet.Identifier.String(),
					Title:           account.Title,
					Description:     account.Description.String,
					Status:          string(createWalletResult.Wallet.Status),
					UserIdentifier:  createWalletResult.Wallet.UserIdentifier,
					AssetIdentifier: createWalletResult.Wallet.AssetIdentifier,
					MetaData:        pgutil.JsonbToMap(createAccountParams.MetaData),
					Asset: &pb.Asset{
						Identifier:  asset.Identifier.String(),
						Code:        asset.Code,
						Symbol:      asset.Symbol,
						Title:       asset.Title,
						Description: asset.Description.String,
						Unit:        string(asset.Unit),
						UnitTitle:   asset.UnitTitle.String,
						Decimals:    asset.Decimals,
						Network:     asset.Network.String,
						IconUrl:     asset.IconUrl.String,
						MetaData:    pgutil.JsonbToMap(asset.MetaData),
						LedgerCode:  asset.LedgerCode,
					},
					Balance: &pb.WalletBalance{
						Credit:        "0",
						Debit:         "0",
						PendingCredit: "0",
						PendingDebit:  "0",
						Balance:       "0",
						Pending:       "0",
					},
				})

				response.Message = "success"
			}

			return response, nil
		} else {
			return nil, err
		}
	}

	if account.Banned {
		return nil, fmt.Errorf("account is banned")
	}

	if account.ExpiresAt.Valid {
		if account.ExpiresAt.Time.Before(time.Now()) {
			return nil, fmt.Errorf("account is expired")
		}
	}

	wallets, err := s.repository.GetUserWallets(ctx, repository.GetUserWalletsParams{
		UserIdentifier: user.Identifier,
		Limit:          100,
		Offset:         0,
	})
	if err != nil {
		return nil, err
	}

	for _, wallet := range wallets {
		asset, err := s.repository.GetWalletAssetByIdentifier(ctx, wallet.AssetIdentifier)
		if err != nil {
			return nil, err
		}

		ledgerAccountId := types.ToUint128(uint64(wallet.LedgerAccountID))

		ledgerAccounts, err := s.tigerbeetle.LookupAccounts([]types.Uint128{ledgerAccountId})
		if err != nil {
			return nil, err
		}

		if len(ledgerAccounts) == 1 {
			ledgerAccount := ledgerAccounts[0]
			credit := ledgerAccount.CreditsPosted.BigInt()
			debit := ledgerAccount.DebitsPosted.BigInt()
			balance := new(big.Int).Sub(&credit, &debit)

			pendingCredit := ledgerAccount.CreditsPending.BigInt()
			pendingDebit := ledgerAccount.DebitsPending.BigInt()
			pendingBalance := new(big.Int).Sub(&pendingCredit, &pendingDebit)

			response.Wallets = append(response.Wallets, &pb.Wallet{
				CreatedAt:       timestamppb.New(wallet.CreatedAt),
				UpdatedAt:       timestamppb.New(wallet.UpdatedAt.Time),
				DeletedAt:       timestamppb.New(wallet.DeletedAt.Time),
				Banned:          account.Banned,
				Identifier:      wallet.Identifier.String(),
				Title:           account.Title,
				Description:     account.Description.String,
				Status:          string(wallet.Status),
				UserIdentifier:  wallet.UserIdentifier,
				AssetIdentifier: wallet.AssetIdentifier,
				MetaData:        pgutil.JsonbToMap(wallet.MetaData),
				Asset: &pb.Asset{
					Identifier:  asset.Identifier.String(),
					Code:        asset.Code,
					Symbol:      asset.Symbol,
					Title:       asset.Title,
					Description: asset.Description.String,
					Unit:        string(asset.Unit),
					UnitTitle:   asset.UnitTitle.String,
					Decimals:    asset.Decimals,
					Network:     asset.Network.String,
					IconUrl:     asset.IconUrl.String,
					MetaData:    pgutil.JsonbToMap(asset.MetaData),
					LedgerCode:  asset.LedgerCode,
				},
				Balance: &pb.WalletBalance{
					Credit:        credit.String(),
					Debit:         debit.String(),
					PendingCredit: pendingCredit.String(),
					PendingDebit:  pendingDebit.String(),
					Balance:       balance.String(),
					Pending:       pendingBalance.String(),
				},
			})
		} else {
			// create ledger account if not exist
			userData := types.ToUint128(user.ID)
			ledgerAccountId := types.ToUint128(uint64(wallet.LedgerAccountID))
			// create tigerbeetle account
			res, err := s.tigerbeetle.CreateAccounts([]types.Account{
				{
					ID:          ledgerAccountId,
					UserData128: userData,
					UserData64:  user.ID,
					UserData32:  0,
					Ledger:      1,
					Code:        uint16(wallet.LedgerAccountCode),
					Flags:       types.AccountFlags{DebitsMustNotExceedCredits: true}.ToUint16(),
				},
			})
			if err != nil {
				return nil, err
			}

			for _, err := range res {
				return nil, errors.New(err.Result.String())
			}

			response.Wallets = append(response.Wallets, &pb.Wallet{
				CreatedAt:       timestamppb.New(wallet.CreatedAt),
				UpdatedAt:       timestamppb.New(wallet.UpdatedAt.Time),
				DeletedAt:       timestamppb.New(wallet.DeletedAt.Time),
				Banned:          account.Banned,
				Identifier:      wallet.Identifier.String(),
				Title:           account.Title,
				Description:     account.Description.String,
				Status:          string(wallet.Status),
				UserIdentifier:  wallet.UserIdentifier,
				AssetIdentifier: wallet.AssetIdentifier,
				MetaData:        pgutil.JsonbToMap(wallet.MetaData),
				Asset: &pb.Asset{
					Identifier:  asset.Identifier.String(),
					Code:        asset.Code,
					Symbol:      asset.Symbol,
					Title:       asset.Title,
					Description: asset.Description.String,
					Unit:        string(asset.Unit),
					UnitTitle:   asset.UnitTitle.String,
					Decimals:    asset.Decimals,
					Network:     asset.Network.String,
					IconUrl:     asset.IconUrl.String,
					MetaData:    pgutil.JsonbToMap(asset.MetaData),
					LedgerCode:  asset.LedgerCode,
				},
				Balance: &pb.WalletBalance{
					Credit:        "0",
					Debit:         "0",
					PendingCredit: "0",
					PendingDebit:  "0",
					Balance:       "0",
					Pending:       "0",
				},
			})
		}

		response.Message = "success"
	}

	return response, nil
}
