-- SQL dump generated using DBML (dbml.dbdiagram.io)
-- Database: PostgreSQL
-- Generated at: 2026-02-11T09:19:42.757Z

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

CREATE TABLE "accounts" (
  "identifier" ulid UNIQUE NOT NULL DEFAULT (gen_monotonic_ulid()),
  "title" varchar(128) NOT NULL,
  "description" varchar(256),
  "status" account_status NOT NULL DEFAULT 'ACTIVE',
  "banned" boolean NOT NULL DEFAULT false,
  "user_identifier" ulid NOT NULL,
  "base_asset_identifier" ulid NOT NULL,
  "meta_data" jsonb NOT NULL DEFAULT '{}',
  "expires_at" timestamptz,
  "created_at" timestamptz NOT NULL DEFAULT (CURRENT_TIMESTAMP),
  "updated_at" timestamptz,
  "deleted_at" timestamptz
);

CREATE TABLE "wallets" (
  "identifier" ulid UNIQUE NOT NULL DEFAULT (gen_monotonic_ulid()),
  "user_identifier" ulid NOT NULL,
  "account_identifier" ulid NOT NULL,
  "asset_identifier" bigserial NOT NULL,
  "meta_data" jsonb NOT NULL DEFAULT '{}',
  "ledger_account_id" "BYTEA(16)" NOT NULL,
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
  "identifier" ulid UNIQUE NOT NULL DEFAULT (gen_monotonic_ulid()),
  "active" boolean NOT NULL DEFAULT false,
  "code" varchar(256) NOT NULL,
  "symbol" varchar(10) NOT NULL,
  "title" varchar(128) NOT NULL,
  "description" varchar(256),
  "meta_data" jsonb NOT NULL DEFAULT '{}',
  "ledger_code" INTEGER NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (CURRENT_TIMESTAMP),
  "updated_at" timestamptz,
  "deleted_at" timestamptz
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

COMMENT ON COLUMN "wallets"."identifier" IS 'unique external identifier for inter system internal-external identifier separation';

COMMENT ON COLUMN "wallets"."user_identifier" IS 'related user identifier to determining account owner';

COMMENT ON COLUMN "wallets"."account_identifier" IS 'related account identifier to determining account of wallet';

COMMENT ON COLUMN "wallets"."asset_identifier" IS 'related asset identifier to determining asset of wallet';

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

COMMENT ON COLUMN "wallet_assets"."identifier" IS 'unique external identifier for inter system internal-external identifier separation';

COMMENT ON COLUMN "wallet_assets"."active" IS 'is wallet asset active or no';

COMMENT ON COLUMN "wallet_assets"."code" IS 'asset code of wallet asset';

COMMENT ON COLUMN "wallet_assets"."symbol" IS 'asset symbol of wallet asset';

COMMENT ON COLUMN "wallet_assets"."title" IS 'wallet asset title';

COMMENT ON COLUMN "wallet_assets"."description" IS 'wallet asset description';

COMMENT ON COLUMN "wallet_assets"."meta_data" IS 'wallet asset meta data';

COMMENT ON COLUMN "wallet_assets"."ledger_code" IS 'related ledger code to determining wallet in ledger';

COMMENT ON COLUMN "wallet_assets"."created_at" IS 'when wallet asset was created';

COMMENT ON COLUMN "wallet_assets"."updated_at" IS 'when wallet asset was updated';

COMMENT ON COLUMN "wallet_assets"."deleted_at" IS 'when wallet asset was deleted';

ALTER TABLE "wallets" ADD FOREIGN KEY ("account_identifier") REFERENCES "accounts" ("identifier");

ALTER TABLE "wallets" ADD FOREIGN KEY ("asset_identifier") REFERENCES "wallet_assets" ("identifier");
