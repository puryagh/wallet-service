-- SQL dump generated using DBML (dbml.dbdiagram.io)
-- Database: PostgreSQL
-- Generated at: 2026-01-23T07:42:40.094Z

CREATE TYPE "contact_type" AS ENUM (
  'EMAIL',
  'MOBILE'
);

CREATE TYPE "profile_status" AS ENUM (
  'PENDING',
  'FILLED',
  'REJECTED',
  'APPROVED',
  'LOCKED'
);

CREATE TYPE "profile_type" AS ENUM (
  'NATURAL',
  'LEGAL'
);

CREATE TABLE "users" (
  "id" bigserial PRIMARY KEY,
  "identifier" ulid UNIQUE NOT NULL DEFAULT (gen_monotonic_ulid()),
  "approved" boolean NOT NULL DEFAULT false,
  "banned" boolean NOT NULL DEFAULT false,
  "meta_data" jsonb NOT NULL DEFAULT '{}',
  "roles" text[] NOT NULL DEFAULT '{USER}',
  "expires_at" timestamptz,
  "created_at" timestamptz NOT NULL DEFAULT (CURRENT_TIMESTAMP),
  "updated_at" timestamptz,
  "deleted_at" timestamptz
);

CREATE TABLE "contacts" (
  "id" bigserial PRIMARY KEY,
  "identifier" ulid UNIQUE NOT NULL DEFAULT (gen_monotonic_ulid()),
  "contact_type" contact_type NOT NULL DEFAULT 'MOBILE',
  "user_id" bigserial NOT NULL,
  "mobile" varchar(16) UNIQUE NOT NULL,
  "mobile_totp" varchar(256),
  "is_mobile_verified" boolean NOT NULL DEFAULT false,
  "mobile_totp_expires_at" timestamptz,
  "email" varchar(256) UNIQUE NOT NULL,
  "email_totp" varchar(256),
  "is_email_verified" boolean NOT NULL DEFAULT false,
  "email_totp_expires_at" timestamptz,
  "meta_data" jsonb NOT NULL DEFAULT '{}',
  "created_at" timestamptz NOT NULL DEFAULT (CURRENT_TIMESTAMP),
  "updated_at" timestamptz,
  "deleted_at" timestamptz
);

CREATE TABLE "profiles" (
  "id" bigserial PRIMARY KEY,
  "identifier" ulid UNIQUE NOT NULL DEFAULT (gen_monotonic_ulid()),
  "user_id" bigserial NOT NULL,
  "profile_type" profile_type NOT NULL,
  "first_name" varchar(128) NOT NULL,
  "last_name" varchar(128) NOT NULL,
  "national_id" varchar(32) UNIQUE NOT NULL,
  "status" profile_status NOT NULL DEFAULT 'PENDING',
  "meta_data" jsonb NOT NULL DEFAULT '{}',
  "birth_date" timestamptz,
  "created_at" timestamptz NOT NULL DEFAULT (CURRENT_TIMESTAMP),
  "updated_at" timestamptz,
  "deleted_at" timestamptz
);

CREATE TABLE "sessions" (
  "id" bigserial PRIMARY KEY NOT NULL,
  "host" varchar(256) NOT NULL,
  "device_id" varchar(128) NOT NULL,
  "device_name" varchar(256) NOT NULL,
  "device_user_agent" jsonb NOT NULL,
  "device_notification_token" varchar(256),
  "identifier" varchar(64) UNIQUE NOT NULL DEFAULT (gen_monotonic_ulid()),
  "user_id" bigserial NOT NULL,
  "expires_in" timestamptz NOT NULL,
  "notification" varchar(256) DEFAULT '',
  "meta_data" jsonb DEFAULT '{}',
  "created_at" timestamptz NOT NULL DEFAULT (CURRENT_TIMESTAMP),
  "deleted_at" timestamptz,
  "updated_at" timestamptz
);

CREATE INDEX ON "users" ("id");

CREATE INDEX ON "users" ("identifier");

CREATE INDEX ON "users" ("deleted_at");

CREATE INDEX ON "users" ("id", "identifier", "deleted_at");

CREATE INDEX ON "users" ("id", "identifier", "banned", "approved", "deleted_at");

CREATE INDEX ON "users" ("banned", "approved");

CREATE INDEX ON "contacts" ("id");

CREATE INDEX ON "contacts" ("identifier");

CREATE INDEX ON "contacts" ("deleted_at");

CREATE INDEX ON "contacts" ("id", "identifier", "deleted_at");

CREATE INDEX ON "contacts" ("id", "identifier", "contact_type", "mobile", "email");

CREATE INDEX ON "profiles" ("id");

CREATE INDEX ON "profiles" ("identifier");

CREATE INDEX ON "profiles" ("deleted_at");

CREATE INDEX ON "profiles" ("id", "identifier", "deleted_at");

CREATE INDEX ON "profiles" ("id", "identifier", "profile_type", "status", "national_id");

CREATE INDEX ON "sessions" ("deleted_at");

CREATE INDEX ON "sessions" ("host");

CREATE INDEX ON "sessions" ("id", "deleted_at");

CREATE INDEX ON "sessions" ("user_id", "deleted_at");

COMMENT ON COLUMN "users"."id" IS 'user unique id';

COMMENT ON COLUMN "users"."identifier" IS 'unique external identifier for inter system internal-external identifier separation';

COMMENT ON COLUMN "users"."approved" IS 'is user approved or no';

COMMENT ON COLUMN "users"."banned" IS 'is user banned or no';

COMMENT ON COLUMN "users"."meta_data" IS 'user meta data';

COMMENT ON COLUMN "users"."roles" IS 'user assigned roles for permission controls';

COMMENT ON COLUMN "users"."expires_at" IS 'expire time of user, if not sets then user valid for unlimited time';

COMMENT ON COLUMN "users"."created_at" IS 'when user was created';

COMMENT ON COLUMN "users"."updated_at" IS 'when user was updated';

COMMENT ON COLUMN "users"."deleted_at" IS 'when user was deleted';

COMMENT ON COLUMN "contacts"."id" IS 'contact unique id';

COMMENT ON COLUMN "contacts"."identifier" IS 'unique external identifier for inter system internal-external identifier separation';

COMMENT ON COLUMN "contacts"."contact_type" IS 'contact primary type';

COMMENT ON COLUMN "contacts"."user_id" IS 'related user id to determining session owner account';

COMMENT ON COLUMN "contacts"."mobile" IS 'contact primary mobile phone number for authorization use';

COMMENT ON COLUMN "contacts"."mobile_totp" IS 'holds TOTP bcrypted pass code';

COMMENT ON COLUMN "contacts"."is_mobile_verified" IS 'sets to true if user verified his mobile by first time otp verification';

COMMENT ON COLUMN "contacts"."mobile_totp_expires_at" IS 'holds by mobile OTP verification code expire time';

COMMENT ON COLUMN "contacts"."email" IS 'contact primary e-mail address';

COMMENT ON COLUMN "contacts"."email_totp" IS 'holds TOTP bcrypted pass code';

COMMENT ON COLUMN "contacts"."is_email_verified" IS 'sets to true if user verified his email by first time otp verification';

COMMENT ON COLUMN "contacts"."email_totp_expires_at" IS 'holds by e-mail OTP verification code expire time';

COMMENT ON COLUMN "contacts"."meta_data" IS 'contact meta data';

COMMENT ON COLUMN "contacts"."created_at" IS 'when contact was created';

COMMENT ON COLUMN "contacts"."updated_at" IS 'when contact was updated';

COMMENT ON COLUMN "contacts"."deleted_at" IS 'when contact was deleted';

COMMENT ON COLUMN "profiles"."id" IS 'profile unique id';

COMMENT ON COLUMN "profiles"."identifier" IS 'unique external identifier for inter system internal-external identifier separation';

COMMENT ON COLUMN "profiles"."user_id" IS 'related user id to determining session owner account';

COMMENT ON COLUMN "profiles"."profile_type" IS 'legal or natural person type definition';

COMMENT ON COLUMN "profiles"."first_name" IS 'user first name';

COMMENT ON COLUMN "profiles"."last_name" IS 'user last name';

COMMENT ON COLUMN "profiles"."national_id" IS 'user unique personal national id-code';

COMMENT ON COLUMN "profiles"."status" IS 'profile control status';

COMMENT ON COLUMN "profiles"."meta_data" IS 'profile meta data';

COMMENT ON COLUMN "profiles"."birth_date" IS 'user birth date information';

COMMENT ON COLUMN "profiles"."created_at" IS 'when profile was created';

COMMENT ON COLUMN "profiles"."updated_at" IS 'when profile was updated';

COMMENT ON COLUMN "profiles"."deleted_at" IS 'when profile was deleted';

COMMENT ON COLUMN "sessions"."id" IS 'session unique id';

COMMENT ON COLUMN "sessions"."host" IS 'session creation request host name (for SSO use)';

COMMENT ON COLUMN "sessions"."device_id" IS 'device unique identifier for device recognition';

COMMENT ON COLUMN "sessions"."device_name" IS 'device human friendly name';

COMMENT ON COLUMN "sessions"."device_user_agent" IS 'device user agent info in json format';

COMMENT ON COLUMN "sessions"."device_notification_token" IS 'device notification token for push notification sending';

COMMENT ON COLUMN "sessions"."identifier" IS 'unique external identifier for inter system internal-external identifier separation';

COMMENT ON COLUMN "sessions"."user_id" IS 'related user id to determining session owner account';

COMMENT ON COLUMN "sessions"."expires_in" IS 'when session expires (only refresh token would updates this field)';

COMMENT ON COLUMN "sessions"."notification" IS 'notification provider token for sending notifications by device';

COMMENT ON COLUMN "sessions"."meta_data" IS 'meta data of session like IP, UserAgent etc...';

COMMENT ON COLUMN "sessions"."created_at" IS 'when session created';

COMMENT ON COLUMN "sessions"."deleted_at" IS 'when session deleted';

COMMENT ON COLUMN "sessions"."updated_at" IS 'when session updated';

ALTER TABLE "contacts" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");

ALTER TABLE "profiles" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");

ALTER TABLE "sessions" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");
