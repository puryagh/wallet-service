package config

import "time"

type Configuration struct {
	RefreshTokenDuration             int           `json:"refresh_token_duration" mapstructure:"LU_CFG_REFRESH_TOKEN_DURATION"`
	AccessTokenDuration              int           `json:"access_token_duration" mapstructure:"LU_CFG_ACCESS_TOKEN_DURATION"`
	VerificationDuration             int           `json:"verification_duration" mapstructure:"LU_CFG_VERIFICATION_DURATION"`
	TokenSymmetricKey                string        `json:"token_symmetric_key" mapstructure:"LU_CFG_TOKEN_SYMMETRIC_KEY"`
	Issuer                           string        `json:"issuer" mapstructure:"LU_CFG_ISSUER"`
	Audience                         string        `json:"audience" mapstructure:"LU_CFG_AUDIENCE"`
	NotificationTopic                string        `json:"notification_topic" mapstructure:"LU_CFG_NOTIFICATION_TOPIC"`
	LogsMongoUri                     string        `json:"logs_mongodb_uri" mapstructure:"LU_CFG_LOGS_MONGODB_URI"`
	HealthCheckKey                   string        `json:"health_check_key" mapstructure:"LU_CFG_HEALTH_CHECK_KEY"`
	Roles                            []string      `json:"roles" mapstructure:"LU_CFG_ROLES"`
	SchemaPath                       string        `json:"schema_path" mapstructure:"LU_CFG_SCHEMA_PATH"`
	GrpcListenerHost                 string        `json:"grpc_listener_host" mapstructure:"LU_CFG_GRPC_LISTENER_HOST"`
	HttpListenerHost                 string        `json:"http_listener_host" mapstructure:"LU_CFG_HTTP_LISTENER_HOST"`
	MeliPayamakURL                   string        `json:"meli_payamak_url" mapstructure:"LU_CFG_MELI_PAYAMAK_URL"`
	MeliPayamakApiKey                string        `json:"meli_payamak_api_key" mapstructure:"LU_CFG_MELI_PAYAMAK_API_KEY"`
	MeliPayamakTimeout               time.Duration `json:"meli_payamak_timeout" mapstructure:"LU_FCG_MELI_PAYAMAK_TIMEOUT"`
	MeliPayamakMaxWorkers            int           `json:"meli_payamak_max_workers" mapstructure:"LU_FCG_MELI_PAYAMAK_MAX_WORKERS"`
	MeliPayamakVerificationPatternID int           `json:"meli_payamak_verification_pattern_id" mapstructure:"LU_FCG_MELI_PAYAMAK_VERIFICATION_PATTERN_ID"`
	TigerbeetleClusterID             string        `json:"tigerbeetle_cluster_id" mapstructure:"LU_CFG_TIGERBEETLE_CLUSTER_ID"`
	TigerbeetleClusterAddresses      []string      `json:"tigerbeetle_cluster_addresses" mapstructure:"LU_CFG_TIGERBEETLE_CLUSTER_ADDRESSES"`
	TigerbeetleReservedAccountNumber int           `json:"tigerbeetle_reserved_account_number" mapstructure:"LU_CFG_TIGERBEETLE_RESERVED_ACCOUNTS_NUMBER"`
	AccountsInitialAssets            []string      `json:"accounts_initial_assets" mapstructure:"LU_CFG_ACCOUNTS_INITIAL_ASSETS"`
	AccountsBaseAssetSymbol          string        `json:"accounts_base_asset_symbol" mapstructure:"LU_CFG_ACCOUNTS_BASE_ASSET_SYMBOL"`
}
