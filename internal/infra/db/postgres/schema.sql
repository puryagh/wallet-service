-- SQL dump generated using DBML (dbml.dbdiagram.io)
-- Database: PostgreSQL
-- Generated at: 2026-03-08T21:48:11.934Z

CREATE TYPE "account_status" AS ENUM (
  'PENDING',
  'ISSUED',
  'ACTIVE'
);

CREATE TYPE "wallet_account_status" AS ENUM (
  'PENDING',
  'ACTIVE',
  'INACTIVE',
  'BANNED',
  'LOCKED',
  'LINKED_LOCKED'
);

CREATE TYPE "card_network" AS ENUM (
  'LOCAL',
  'VISA',
  'MASTERCARD',
  'AMEX'
);

CREATE TYPE "wallet_card_status" AS ENUM (
  'PENDING',
  'ACTIVE',
  'INACTIVE',
  'BLOCKED',
  'EXPIRED',
  'REISSUED'
);

CREATE TYPE "wallet_card_event_type" AS ENUM (
  'ISSUED',
  'AUTHENTICATED',
  'AUTHENTICATION_FAILED',
  'STATUS_CHANGED'
);

CREATE TYPE "wallet_asset_unit" AS ENUM (
  'NONE',
  'MILLIGRAM',
  'GRAM',
  'KILOGRAM',
  'TONNE',
  'PIECE'
);

CREATE TABLE "accounts" (
  "id" bigserial PRIMARY KEY,
  "identifier" ulid UNIQUE NOT NULL DEFAULT (gen_monotonic_ulid()),
  "title" varchar(128) NOT NULL,
  "description" varchar(256),
  "status" account_status NOT NULL DEFAULT 'ACTIVE',
  "banned" boolean NOT NULL DEFAULT false,
  "user_identifier" varchar(32) UNIQUE NOT NULL,
  "base_asset_identifier" varchar(32) NOT NULL,
  "meta_data" jsonb NOT NULL DEFAULT '{}',
  "expires_at" timestamptz,
  "created_at" timestamptz NOT NULL DEFAULT (CURRENT_TIMESTAMP),
  "updated_at" timestamptz,
  "deleted_at" timestamptz
);

CREATE TABLE "wallets" (
  "id" bigserial PRIMARY KEY,
  "identifier" ulid UNIQUE NOT NULL DEFAULT (gen_monotonic_ulid()),
  "user_identifier" varchar(32) NOT NULL,
  "account_identifier" varchar(32) NOT NULL,
  "account_id" bigserial NOT NULL,
  "asset_identifier" varchar(32) NOT NULL,
  "wallet_asset_id" bigserial NOT NULL,
  "meta_data" jsonb NOT NULL DEFAULT '{}',
  "ledger_account_id" bigint NOT NULL,
  "ledger_account_code" INTEGER NOT NULL,
  "primary_account_number" varchar(24) NOT NULL,
  "iban" varchar(34),
  "cvv" varchar(6),
  "cvv_two" varchar(6),
  "expire_date" varchar(6),
  "pin_code" varchar(256),
  "status" wallet_account_status NOT NULL DEFAULT 'ACTIVE',
  "transaction_totp_secret" varchar(64),
  "transaction_totp_expires_at" timestamptz,
  "created_at" timestamptz NOT NULL DEFAULT (CURRENT_TIMESTAMP),
  "updated_at" timestamptz,
  "deleted_at" timestamptz
);

CREATE TABLE "wallet_assets" (
  "id" bigserial PRIMARY KEY,
  "identifier" ulid UNIQUE NOT NULL DEFAULT (gen_monotonic_ulid()),
  "active" boolean NOT NULL DEFAULT false,
  "code" varchar(256) NOT NULL,
  "symbol" varchar(10) UNIQUE NOT NULL,
  "title" varchar(128) NOT NULL,
  "description" varchar(256),
  "predefined" boolean NOT NULL DEFAULT false,
  "unit" wallet_asset_unit NOT NULL DEFAULT 'NONE',
  "unit_title" varchar(128),
  "decimals" int NOT NULL DEFAULT 0,
  "network" varchar(256),
  "icon_url" varchar(256),
  "meta_data" jsonb NOT NULL DEFAULT '{}',
  "ledger_code" INTEGER NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (CURRENT_TIMESTAMP),
  "updated_at" timestamptz,
  "deleted_at" timestamptz
);

CREATE TABLE "issuer_bins" (
  "id" bigserial PRIMARY KEY,
  "identifier" ulid UNIQUE NOT NULL DEFAULT (gen_monotonic_ulid()),
  "active" boolean NOT NULL DEFAULT true,
  "bin" varchar(8) UNIQUE NOT NULL,
  "brand" card_network NOT NULL DEFAULT 'LOCAL',
  "issuer_name" varchar(128) NOT NULL,
  "country_code" varchar(2) NOT NULL DEFAULT 'IR',
  "meta_data" jsonb NOT NULL DEFAULT '{}',
  "created_at" timestamptz NOT NULL DEFAULT (CURRENT_TIMESTAMP),
  "updated_at" timestamptz,
  "deleted_at" timestamptz
);

CREATE TABLE "card_products" (
  "id" bigserial PRIMARY KEY,
  "identifier" ulid UNIQUE NOT NULL DEFAULT (gen_monotonic_ulid()),
  "issuer_bin_id" bigserial NOT NULL,
  "name" varchar(64) UNIQUE NOT NULL,
  "pan_length" int NOT NULL DEFAULT 16,
  "cvv_length" int NOT NULL DEFAULT 3,
  "expiry_months" int NOT NULL DEFAULT 36,
  "service_code" varchar(3) NOT NULL DEFAULT '201',
  "allow_magnetic_stripe" boolean NOT NULL DEFAULT true,
  "meta_data" jsonb NOT NULL DEFAULT '{}',
  "created_at" timestamptz NOT NULL DEFAULT (CURRENT_TIMESTAMP),
  "updated_at" timestamptz,
  "deleted_at" timestamptz
);

CREATE TABLE "wallet_cards" (
  "id" bigserial PRIMARY KEY,
  "identifier" ulid UNIQUE NOT NULL DEFAULT (gen_monotonic_ulid()),
  "wallet_id" bigserial NOT NULL,
  "wallet_identifier" varchar(32) NOT NULL,
  "user_identifier" varchar(32) NOT NULL,
  "card_product_id" bigserial NOT NULL,
  "issuer_bin" varchar(8) NOT NULL,
  "brand" card_network NOT NULL DEFAULT 'LOCAL',
  "pan_fingerprint" varchar(128) UNIQUE NOT NULL,
  "masked_pan" varchar(32) NOT NULL,
  "pan_last4" varchar(4) NOT NULL,
  "expiry_month" int NOT NULL,
  "expiry_year" int NOT NULL,
  "cardholder_name" varchar(64) NOT NULL,
  "service_code" varchar(3) NOT NULL,
  "cvv_digest" varchar(128) NOT NULL,
  "cvv_two_digest" varchar(128) NOT NULL,
  "track1_digest" varchar(128),
  "track2_digest" varchar(128) NOT NULL,
  "pin_digest" varchar(256),
  "last_authenticated_at" timestamptz,
  "status" wallet_card_status NOT NULL DEFAULT 'ACTIVE',
  "meta_data" jsonb NOT NULL DEFAULT '{}',
  "created_at" timestamptz NOT NULL DEFAULT (CURRENT_TIMESTAMP),
  "updated_at" timestamptz,
  "deleted_at" timestamptz
);

CREATE TABLE "wallet_card_events" (
  "id" bigserial PRIMARY KEY,
  "identifier" ulid UNIQUE NOT NULL DEFAULT (gen_monotonic_ulid()),
  "wallet_card_id" bigserial NOT NULL,
  "wallet_card_identifier" varchar(32) NOT NULL,
  "user_identifier" varchar(32) NOT NULL,
  "event_type" wallet_card_event_type NOT NULL,
  "success" boolean NOT NULL DEFAULT true,
  "remote_ip" varchar(64),
  "meta_data" jsonb NOT NULL DEFAULT '{}',
  "created_at" timestamptz NOT NULL DEFAULT (CURRENT_TIMESTAMP)
);

CREATE INDEX ON "accounts" ("identifier");

CREATE INDEX ON "accounts" ("deleted_at");

CREATE INDEX ON "accounts" ("identifier", "deleted_at");

CREATE INDEX ON "accounts" ("identifier", "banned", "deleted_at");

CREATE INDEX ON "accounts" ("identifier", "status", "deleted_at");

CREATE INDEX ON "accounts" ("banned");

CREATE INDEX ON "accounts" ("status");

CREATE INDEX ON "wallets" ("identifier");

CREATE INDEX ON "wallets" ("deleted_at");

CREATE INDEX ON "wallets" ("identifier", "deleted_at");

CREATE INDEX ON "wallets" ("identifier", "asset_identifier");

CREATE INDEX ON "wallets" ("identifier", "account_identifier", "asset_identifier");

CREATE INDEX ON "wallet_assets" ("identifier");

CREATE INDEX ON "wallet_assets" ("deleted_at");

CREATE INDEX ON "wallet_assets" ("identifier", "deleted_at");

CREATE INDEX ON "wallet_assets" ("identifier", "active", "deleted_at");

CREATE INDEX ON "wallet_assets" ("active");

CREATE INDEX ON "issuer_bins" ("identifier");

CREATE INDEX ON "issuer_bins" ("deleted_at");

CREATE INDEX ON "issuer_bins" ("identifier", "deleted_at");

CREATE INDEX ON "issuer_bins" ("bin", "deleted_at");

CREATE INDEX ON "issuer_bins" ("active");

CREATE INDEX ON "card_products" ("identifier");

CREATE INDEX ON "card_products" ("deleted_at");

CREATE INDEX ON "card_products" ("identifier", "deleted_at");

CREATE INDEX ON "card_products" ("issuer_bin_id", "deleted_at");

CREATE INDEX ON "card_products" ("name", "deleted_at");

CREATE INDEX ON "wallet_cards" ("identifier");

CREATE INDEX ON "wallet_cards" ("deleted_at");

CREATE INDEX ON "wallet_cards" ("identifier", "deleted_at");

CREATE INDEX ON "wallet_cards" ("wallet_identifier", "deleted_at");

CREATE INDEX ON "wallet_cards" ("user_identifier", "deleted_at");

CREATE INDEX ON "wallet_cards" ("wallet_id", "deleted_at");

CREATE INDEX ON "wallet_cards" ("pan_fingerprint", "deleted_at");

CREATE INDEX ON "wallet_cards" ("status");

CREATE INDEX ON "wallet_card_events" ("identifier");

CREATE INDEX ON "wallet_card_events" ("wallet_card_id", "created_at");

CREATE INDEX ON "wallet_card_events" ("wallet_card_identifier", "created_at");

CREATE INDEX ON "wallet_card_events" ("user_identifier", "created_at");

CREATE INDEX ON "wallet_card_events" ("event_type", "created_at");

COMMENT ON COLUMN "accounts"."id" IS 'account unique id';

COMMENT ON COLUMN "accounts"."identifier" IS 'unique external identifier for inter system internal-external identifier separation';

COMMENT ON COLUMN "accounts"."title" IS 'account title';

COMMENT ON COLUMN "accounts"."description" IS 'account description';

COMMENT ON COLUMN "accounts"."status" IS 'account status';

COMMENT ON COLUMN "accounts"."banned" IS 'is account banned or no';

COMMENT ON COLUMN "accounts"."user_identifier" IS 'related user identifier to determining account owner account';

COMMENT ON COLUMN "accounts"."base_asset_identifier" IS 'related base asset identifier to determining account base asset';

COMMENT ON COLUMN "accounts"."meta_data" IS 'account meta data';

COMMENT ON COLUMN "accounts"."expires_at" IS 'expire time of account, if not sets then user valid for unlimited time';

COMMENT ON COLUMN "accounts"."created_at" IS 'when account was created';

COMMENT ON COLUMN "accounts"."updated_at" IS 'when account was updated';

COMMENT ON COLUMN "accounts"."deleted_at" IS 'when account was deleted';

COMMENT ON COLUMN "wallets"."id" IS 'wallet unique id';

COMMENT ON COLUMN "wallets"."identifier" IS 'unique external identifier for inter system internal-external identifier separation';

COMMENT ON COLUMN "wallets"."user_identifier" IS 'related user identifier to determining account owner';

COMMENT ON COLUMN "wallets"."account_identifier" IS 'related account identifier to determining account of wallet';

COMMENT ON COLUMN "wallets"."account_id" IS 'related account id to determining account of wallet';

COMMENT ON COLUMN "wallets"."asset_identifier" IS 'related asset identifier to determining asset of wallet';

COMMENT ON COLUMN "wallets"."wallet_asset_id" IS 'related wallet asset id to determining asset of wallet';

COMMENT ON COLUMN "wallets"."meta_data" IS 'wallet account meta data';

COMMENT ON COLUMN "wallets"."ledger_account_id" IS 'related ledger account id to determining "ledger account" of wallet in ledger core';

COMMENT ON COLUMN "wallets"."ledger_account_code" IS 'related ledger account code to determining service code for wallet in ledger core';

COMMENT ON COLUMN "wallets"."primary_account_number" IS 'primary account number of wallet account';

COMMENT ON COLUMN "wallets"."iban" IS 'IBAN of wallet account';

COMMENT ON COLUMN "wallets"."cvv" IS 'CVV of wallet account';

COMMENT ON COLUMN "wallets"."cvv_two" IS 'CVV2 of wallet account';

COMMENT ON COLUMN "wallets"."expire_date" IS 'expire date of wallet account';

COMMENT ON COLUMN "wallets"."pin_code" IS 'PIN code of wallet account (bcrypted)';

COMMENT ON COLUMN "wallets"."status" IS 'wallet account status';

COMMENT ON COLUMN "wallets"."transaction_totp_secret" IS 'transaction TOTP secret for wallet account (bcrypted)';

COMMENT ON COLUMN "wallets"."transaction_totp_expires_at" IS 'transaction TOTP secret expire time for wallet account';

COMMENT ON COLUMN "wallets"."created_at" IS 'when wallet account was created';

COMMENT ON COLUMN "wallets"."updated_at" IS 'when wallet account was updated';

COMMENT ON COLUMN "wallets"."deleted_at" IS 'when wallet account was deleted';

COMMENT ON COLUMN "wallet_assets"."id" IS 'wallet_asset unique id';

COMMENT ON COLUMN "wallet_assets"."identifier" IS 'unique external identifier for inter system internal-external identifier separation';

COMMENT ON COLUMN "wallet_assets"."active" IS 'is wallet asset active or no';

COMMENT ON COLUMN "wallet_assets"."code" IS 'asset code of wallet asset';

COMMENT ON COLUMN "wallet_assets"."symbol" IS 'asset symbol of wallet asset';

COMMENT ON COLUMN "wallet_assets"."title" IS 'wallet asset title';

COMMENT ON COLUMN "wallet_assets"."description" IS 'wallet asset description';

COMMENT ON COLUMN "wallet_assets"."predefined" IS 'is wallet asset predefined for account initialization or no';

COMMENT ON COLUMN "wallet_assets"."unit" IS 'wallet asset unit';

COMMENT ON COLUMN "wallet_assets"."unit_title" IS 'wallet asset unit title';

COMMENT ON COLUMN "wallet_assets"."decimals" IS 'wallet asset decimals';

COMMENT ON COLUMN "wallet_assets"."network" IS 'wallet asset network';

COMMENT ON COLUMN "wallet_assets"."icon_url" IS 'wallet asset icon url';

COMMENT ON COLUMN "wallet_assets"."meta_data" IS 'wallet asset meta data';

COMMENT ON COLUMN "wallet_assets"."ledger_code" IS 'related ledger code to determining wallet in ledger';

COMMENT ON COLUMN "wallet_assets"."created_at" IS 'when wallet asset was created';

COMMENT ON COLUMN "wallet_assets"."updated_at" IS 'when wallet asset was updated';

COMMENT ON COLUMN "wallet_assets"."deleted_at" IS 'when wallet asset was deleted';

COMMENT ON COLUMN "issuer_bins"."id" IS 'issuer bin unique id';

COMMENT ON COLUMN "issuer_bins"."identifier" IS 'unique external identifier for issuer bin';

COMMENT ON COLUMN "issuer_bins"."active" IS 'is issuer bin active or no';

COMMENT ON COLUMN "issuer_bins"."bin" IS 'issuer identification number / bank identification number';

COMMENT ON COLUMN "issuer_bins"."brand" IS 'card network / brand';

COMMENT ON COLUMN "issuer_bins"."issuer_name" IS 'issuer display name';

COMMENT ON COLUMN "issuer_bins"."country_code" IS 'issuer ISO country code';

COMMENT ON COLUMN "issuer_bins"."meta_data" IS 'issuer bin meta data';

COMMENT ON COLUMN "issuer_bins"."created_at" IS 'when issuer bin was created';

COMMENT ON COLUMN "issuer_bins"."updated_at" IS 'when issuer bin was updated';

COMMENT ON COLUMN "issuer_bins"."deleted_at" IS 'when issuer bin was deleted';

COMMENT ON COLUMN "card_products"."id" IS 'card product unique id';

COMMENT ON COLUMN "card_products"."identifier" IS 'unique external identifier for card product';

COMMENT ON COLUMN "card_products"."issuer_bin_id" IS 'related issuer bin id for this card product';

COMMENT ON COLUMN "card_products"."name" IS 'card product name';

COMMENT ON COLUMN "card_products"."pan_length" IS 'configured PAN length for this product';

COMMENT ON COLUMN "card_products"."cvv_length" IS 'configured CVV/CVV2 length for this product';

COMMENT ON COLUMN "card_products"."expiry_months" IS 'default expiry policy in months';

COMMENT ON COLUMN "card_products"."service_code" IS 'magnetic stripe service code';

COMMENT ON COLUMN "card_products"."allow_magnetic_stripe" IS 'is magnetic stripe issuance allowed';

COMMENT ON COLUMN "card_products"."meta_data" IS 'card product meta data';

COMMENT ON COLUMN "card_products"."created_at" IS 'when card product was created';

COMMENT ON COLUMN "card_products"."updated_at" IS 'when card product was updated';

COMMENT ON COLUMN "card_products"."deleted_at" IS 'when card product was deleted';

COMMENT ON COLUMN "wallet_cards"."id" IS 'wallet card unique id';

COMMENT ON COLUMN "wallet_cards"."identifier" IS 'unique external identifier for wallet card';

COMMENT ON COLUMN "wallet_cards"."wallet_id" IS 'related wallet id';

COMMENT ON COLUMN "wallet_cards"."wallet_identifier" IS 'related wallet identifier';

COMMENT ON COLUMN "wallet_cards"."user_identifier" IS 'related user identifier';

COMMENT ON COLUMN "wallet_cards"."card_product_id" IS 'related card product id';

COMMENT ON COLUMN "wallet_cards"."issuer_bin" IS 'issuer bin used for card generation';

COMMENT ON COLUMN "wallet_cards"."brand" IS 'card network / brand';

COMMENT ON COLUMN "wallet_cards"."pan_fingerprint" IS 'stable HMAC-based PAN fingerprint used for lookup';

COMMENT ON COLUMN "wallet_cards"."masked_pan" IS 'masked PAN for display only';

COMMENT ON COLUMN "wallet_cards"."pan_last4" IS 'last 4 PAN digits';

COMMENT ON COLUMN "wallet_cards"."expiry_month" IS 'card expiry month in MM form';

COMMENT ON COLUMN "wallet_cards"."expiry_year" IS 'card expiry year in YYYY form';

COMMENT ON COLUMN "wallet_cards"."cardholder_name" IS 'normalized cardholder name for embossing/track generation';

COMMENT ON COLUMN "wallet_cards"."service_code" IS 'magnetic stripe service code';

COMMENT ON COLUMN "wallet_cards"."cvv_digest" IS 'HMAC digest of CVV verification material';

COMMENT ON COLUMN "wallet_cards"."cvv_two_digest" IS 'HMAC digest of CVV2 verification material';

COMMENT ON COLUMN "wallet_cards"."track1_digest" IS 'HMAC digest of generated track1 data';

COMMENT ON COLUMN "wallet_cards"."track2_digest" IS 'HMAC digest of generated track2 data';

COMMENT ON COLUMN "wallet_cards"."pin_digest" IS 'digest or delegated reference for PIN material';

COMMENT ON COLUMN "wallet_cards"."last_authenticated_at" IS 'when card was last authenticated successfully';

COMMENT ON COLUMN "wallet_cards"."status" IS 'wallet card lifecycle status';

COMMENT ON COLUMN "wallet_cards"."meta_data" IS 'wallet card meta data';

COMMENT ON COLUMN "wallet_cards"."created_at" IS 'when wallet card was created';

COMMENT ON COLUMN "wallet_cards"."updated_at" IS 'when wallet card was updated';

COMMENT ON COLUMN "wallet_cards"."deleted_at" IS 'when wallet card was deleted';

COMMENT ON COLUMN "wallet_card_events"."id" IS 'wallet card event unique id';

COMMENT ON COLUMN "wallet_card_events"."identifier" IS 'unique external identifier for wallet card event';

COMMENT ON COLUMN "wallet_card_events"."wallet_card_id" IS 'related wallet card id';

COMMENT ON COLUMN "wallet_card_events"."wallet_card_identifier" IS 'related wallet card identifier';

COMMENT ON COLUMN "wallet_card_events"."user_identifier" IS 'related user identifier';

COMMENT ON COLUMN "wallet_card_events"."event_type" IS 'wallet card event type';

COMMENT ON COLUMN "wallet_card_events"."success" IS 'event success flag';

COMMENT ON COLUMN "wallet_card_events"."remote_ip" IS 'remote ip address of requester if available';

COMMENT ON COLUMN "wallet_card_events"."meta_data" IS 'wallet card event meta data';

COMMENT ON COLUMN "wallet_card_events"."created_at" IS 'when wallet card event was created';

ALTER TABLE "wallets" ADD FOREIGN KEY ("account_id") REFERENCES "accounts" ("id");

ALTER TABLE "wallets" ADD FOREIGN KEY ("wallet_asset_id") REFERENCES "wallet_assets" ("id");

ALTER TABLE "card_products" ADD FOREIGN KEY ("issuer_bin_id") REFERENCES "issuer_bins" ("id");

ALTER TABLE "wallet_cards" ADD FOREIGN KEY ("wallet_id") REFERENCES "wallets" ("id");

ALTER TABLE "wallet_cards" ADD FOREIGN KEY ("card_product_id") REFERENCES "card_products" ("id");

ALTER TABLE "wallet_card_events" ADD FOREIGN KEY ("wallet_card_id") REFERENCES "wallet_cards" ("id");
