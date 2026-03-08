## Magnetic Card Issuing and Validation Proposal

### Purpose

Define a secure, standards-aware, extensible backend design for issuing one magnetic card per wallet while allowing future support for replacement cards, multiple card products, and stronger card controls.

### Scope

- PAN generation and validation
- Luhn check digit generation and validation
- expiry date generation and validation
- CVV1/CVV2 handling
- magnetic stripe track data generation policy
- secure storage, masking, access control, and auditability
- API, schema, domain, and migration recommendations

### Current Codebase Review Summary

Observed in the current backend:

- `internal/application/wallet/context_user_accounts.go` auto-creates wallets when missing.
- wallet creation currently sets `PrimaryAccountNumber` using `ulid.Make().String()[:24]`, which is not ISO card PAN generation and is not Luhn-based.
- `internal/infra/db/postgres/schema.sql` stores `primary_account_number`, `cvv`, `cvv_two`, `expire_date`, and `pin_code` directly on `wallets`.
- `proto/rpc_wallet_account.model.proto` exposes raw PAN/CVV/CVV2/PIN fields in API models.
- `internal/infra/db/postgres/queries/wallets.sql` has query/schema drift, including a duplicated `primary_account_number` column in `CreateWallet` and mismatched join columns in `GetUserAssetWallet`.
- there is no dedicated card lifecycle model for issuance, activation, block, reissue, renewal, replacement, reveal policy, or audit trail.
- there is no current domain/service boundary for magnetic stripe track generation or secure card secret management.

### Key Risks in the Current Design

- insecure PAN generation
- direct exposure of sensitive card data via protobuf responses
- overloading `wallets` with card-issuer concerns
- lack of separation between public card metadata and highly sensitive card secrets
- no explicit PCI-DSS handling boundaries
- no support for multiple cards over the lifetime of one wallet
- existing SQL drift raises implementation risk if card features are added without first stabilizing data access

### Requirements

#### Functional requirements

- issue a card for a wallet belonging to a user account
- support at least one active physical magnetic card per wallet by policy
- support future replacement/reissue without changing wallet identity
- generate PAN from configured issuer BIN/IIN ranges and product rules
- generate Luhn check digit and validate PAN by Luhn
- generate expiry month/year from product policy
- generate CVV1 for track usage and CVV2 for card-not-present usage
- generate Track 1 / Track 2 values for authorized personalization flows
- return masked card details for normal API reads
- allow privileged one-time reveal flows only when explicitly required and strongly authenticated
- support card status transitions such as `PENDING`, `ISSUED`, `PERSONALIZED`, `ACTIVE`, `BLOCKED`, `EXPIRED`, `REPLACED`, `CLOSED`
- support audit logging for issuance, reveal, validation failures, block/unblock, and lifecycle changes

#### Non-functional requirements

- extensible to multiple card products and BINs
- safe to evolve toward EMV/chip or virtual cards later
- idempotent issuance operations
- uniqueness guarantees for PAN and per-wallet active card policy
- operational observability without leaking secrets
- clear domain separation between wallet ledger concerns and card issuance concerns

#### Security and compliance requirements

- do not expose full PAN, CVV1, CVV2, PIN, or full track data in normal wallet APIs
- do not log full PAN, CVV1, CVV2, PIN, or track data
- encrypt PAN and any retained sensitive card data with KMS/HSM-backed keys or an isolated vault service
- do not persist full track data except in tightly controlled short-lived issuance/personalization flows
- do not persist CVV2 after issuance/reveal if avoidable; prefer one-time generation/reveal and immediate discard
- treat magnetic stripe Track 1 / Track 2 and CVV1 as highly sensitive authentication data
- require step-up authentication and authorization for any sensitive reveal or personalization operation
- maintain audit trails for every privileged access to card-sensitive data

### Recommended Terminology

Use standard card terminology in the new design:

- `PAN`: primary account number
- `CVV1`: value encoded for magnetic stripe verification
- `CVV2`: printed/manual verification value for card-not-present scenarios
- `PIN`: separate secret, never returned after setup
- `Track1` / `Track2`: derived personalization outputs, not general-purpose stored fields

Avoid keeping generic fields named only `cvv` and `cvv_two` as the long-term domain model.

### Proposed Domain Model

Do not extend `wallets` further for card issuance. Introduce dedicated card entities.

#### 1. `wallet_cards`

Primary card record linked to `wallets` and `accounts`.

Suggested fields:

- `id`
- `identifier` (ULID)
- `wallet_identifier`
- `account_identifier`
- `user_identifier`
- `card_product_identifier`
- `bin_identifier`
- `pan_last4`
- `pan_fingerprint` (deterministic lookup token, not raw PAN)
- `masked_pan`
- `expiry_month`
- `expiry_year`
- `status`
- `cardholder_name`
- `service_code`
- `sequence_number`
- `issued_at`
- `activated_at`
- `blocked_at`
- `replaced_by_card_identifier`
- `metadata`
- timestamps / soft delete

#### 2. `wallet_card_secrets`

Highly restricted secret material, ideally replaced by an external vault or HSM-backed card vault.

Suggested fields if temporary DB persistence is unavoidable:

- `card_identifier`
- `encrypted_pan`
- `encrypted_pin_block_or_reference`
- `encrypted_track1` only if temporarily required
- `encrypted_track2` only if temporarily required
- `cvv2_reveal_expires_at`
- `created_at`

Preferred approach: store only references/tokens to a dedicated secure secret store, not the raw encrypted values in the main service database.

#### 3. `card_products`

Defines rules per card type.

- product name
- PAN length
- allowed BIN/IIN range
- expiry policy in months
- CVV length
- track format policy
- service code
- allow_magnetic_stripe
- allow_virtual / allow_physical

#### 4. `issuer_bins`

Configuration of BIN/IIN inventory and sequencing.

- BIN/IIN value
- brand/network
- PAN length
- sequence counter / allocator state
- active flag

#### 5. `wallet_card_events`

Immutable audit trail.

- `card_identifier`
- `event_type`
- actor / system source
- reason
- metadata
- created_at

### Issuance Flow Proposal

1. validate wallet and account ownership/status
2. enforce issuance policy: one active physical magnetic card per wallet unless reissue flow
3. select card product and BIN/IIN
4. allocate PAN body sequence
5. compute Luhn check digit and final PAN
6. generate expiry month/year from product policy
7. generate CVV1 and CVV2 in secure boundary
8. derive Track 1 / Track 2 only inside secure boundary
9. persist public metadata in `wallet_cards`
10. persist secrets only in secure vault/reference store
11. return masked summary only
12. record audit event

### Validation Rules Proposal

#### PAN validation

- numeric only
- allowed length from product/BIN policy, typically 16 to 19 digits
- known configured BIN/IIN prefix
- Luhn valid
- not expired or blocked if used as a card credential

#### Expiry validation

- month must be `01` to `12`
- year stored as 4-digit internally; expose formatted month/year as needed
- consider card valid through the last day of the expiry month

#### CVV validation

- do not validate CVV2 in general wallet read flows
- perform CVV/CVV2 comparison only inside privileged authorization flows
- store comparison-safe value or use secure vault verification endpoint instead of retrieving the original secret

#### Track data validation

- validate format against product rules
- never expose full track data through ordinary API models
- only generate/reveal for authorized issuance/personalization channels

### API and Proto Proposal

#### Remove from normal wallet response models

- full PAN
- CVV1/CVV2
- PIN
- full track data

#### Add dedicated card response models

- `WalletCardSummary`
- `WalletCardStatus`
- `IssueWalletCardRequest/Response`
- `GetWalletCardRequest/Response`
- `ListWalletCardsRequest/Response`
- `BlockWalletCardRequest/Response`
- `ReplaceWalletCardRequest/Response`
- `RevealWalletCardSensitiveDataRequest/Response` with strict authorization and one-time semantics

Normal responses should contain only:

- card identifier
- wallet identifier
- masked PAN
- last4
- expiry month/year
- brand/product
- status
- created/issued timestamps

### Service and Package Structure Proposal

- `internal/domain/card`
- `internal/application/card`
- `internal/infra/db/postgres/queries/cards.sql`
- `internal/infra/db/postgres/repository/cards.sql.go`
- `internal/infra/cardvault` for secret storage abstraction
- `internal/infra/cardcrypto` for PAN/Luhn/track/CVV generation interfaces

Keep card issuance logic separate from `ContextUserAccounts`; wallet creation should not silently manufacture real production card credentials.

### Migration Strategy

#### Phase 0: prerequisite stabilization

- fix wallet SQL/schema drift before adding new card tables
- stop treating current `wallets.primary_account_number` as a standards-compliant PAN

#### Phase 1: introduce new schema

- add `wallet_cards`, `wallet_card_events`, `card_products`, `issuer_bins`
- add vault integration or `wallet_card_secrets` placeholder abstraction

#### Phase 2: introduce card APIs

- add masked card summary reads
- add privileged issuance and reveal flows
- stop exposing raw card fields in wallet/account protobuf models

#### Phase 3: transition existing data

- treat existing wallet-level PAN/CVV/CVV2 fields as legacy/untrusted
- migrate only what is valid and policy-approved
- backfill masked display values where possible

#### Phase 4: deprecate legacy wallet columns

- remove card-sensitive fields from `wallets` after migration and client rollout

### Testing Strategy

- unit tests for Luhn generation and validation
- unit tests for PAN format and issuer BIN selection
- unit tests for expiry generation/validation edge cases
- unit tests for masking and redaction
- integration tests for issuance idempotency and uniqueness
- authorization tests for sensitive reveal endpoints
- audit log tests
- negative tests ensuring logs and responses never contain full secrets

### Recommended Immediate Decisions

1. move card data out of `wallets`
2. stop returning raw PAN/CVV/CVV2/PIN from protobuf models
3. model `CVV1` and `CVV2` explicitly instead of `cvv` / `cvv_two`
4. use BIN/IIN + sequence + Luhn, not ULID substrings, for PAN generation
5. generate track data on demand in a secure boundary and avoid persistent storage
6. make wallet-to-card relationship extensible to multiple cards over time even if current business policy is one active card per wallet

### Conclusion

The current backend contains wallet-linked placeholders for card-like data, but it is not yet structured for secure magnetic card issuance. The safest and most extensible path is to create a dedicated card domain, isolate sensitive secrets from normal wallet records, expose only masked card data in normal APIs, and implement card generation/validation inside a controlled security boundary.