package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/google/uuid"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/liveutil/go-lib/configuration"
	"github.com/liveutil/go-lib/env"
	"github.com/liveutil/go-lib/framework/healthcheck"
	"github.com/liveutil/go-lib/framework/healthcheck/healthpb"
	"github.com/liveutil/go-lib/framework/mesh"
	"github.com/liveutil/go-lib/fsutil"
	"github.com/liveutil/go-lib/grpcutil"
	"github.com/liveutil/go-lib/jsonschema"
	fl "github.com/liveutil/go-lib/logger"
	"github.com/liveutil/go-lib/paseto"
	"github.com/liveutil/go-lib/pgutil"
	"github.com/liveutil/go-lib/redisutils"
	"github.com/liveutil/go-lib/tracing"
	"github.com/liveutil/wallet-service/internal/abstract/pb"
	"github.com/liveutil/wallet-service/internal/application/wallet"
	"github.com/liveutil/wallet-service/internal/config"
	"github.com/liveutil/wallet-service/internal/infra/db/postgres/repository"
	"github.com/nats-io/nats.go"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
)

const VERSION string = "0.0.1"
const APPLICATION string = "wallet_service"

var fieldKeys = []string{"method"}

func main() {
	appContext, cancel := context.WithCancel(context.Background())
	defer cancel()

	environment := "dev"

	if env.IsProduction() {
		environment = "prod"
	}

	parentDir, err := fsutil.GetParentDirectory()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// determine which configuration file to use
	baseConfigEnvFile := env.GetStringDefault("LU_CFG_BASE_CONFIG_ENV_FILE", fmt.Sprintf("%s/liveutil-stack/env/stack.%s.env", parentDir, environment))

	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatal().Err(err).Msgf("cannot load config: %v", err)
	}
	appConfigEnvFile := env.GetStringDefault("LU_CFG_APP_CONFIG_ENV_FILE", fmt.Sprintf("%s/%s.env", currentDir, environment))

	// load configurations from environment variables file
	// baseConf, err := configuration.LoadEnvConfig[config.Configuration](".", appConfigFile, "env")
	baseConf, appConfig, err := configuration.LoadBaseConfig[config.Configuration](baseConfigEnvFile, appConfigEnvFile)
	if err != nil {
		log.Fatal().Err(err).Msgf("cannot load config: %v", err)
	}

	// initialize Jaeger tracer
	tp, err := tracing.InitJaegerTracing(&tracing.JaegerConfig{
		Endpoint:          baseConf.JaegerHost,
		ServiceName:       APPLICATION,
		ServiceVersion:    VERSION,
		ServiceInstanceID: uuid.NewString(),
	})
	if err != nil {
		log.Fatal().Err(err).Msgf("failed to initialize Jaeger tracer: %v", err)
	}

	defer func() {
		if err := tp.Shutdown(appContext); err != nil {
			log.Fatal().Err(err).Msgf("failed to shutdown Jaeger tracer: %v", err)
		}
	}()

	jaegerTracer := tp.Tracer(APPLICATION)

	// initialize nats connection to write logs to NATS in development mode and send events to NATS in production mode
	natsConn, err := nats.Connect(baseConf.NatsConnection)
	if err != nil {
		log.Fatal().Err(err).Msgf("failed to connect to nats: %v", err)
	}
	defer natsConn.Close()

	// initialize mongo client to write logs to MongoDB in development mode
	mongoClient, err := mongo.Connect(appContext, options.Client().ApplyURI(appConfig.LogsMongoUri))
	if err != nil {
		log.Fatal().Err(err).Msgf("failed to connect to mongo: %v", err)
	}
	defer mongoClient.Disconnect(appContext)

	// initialize logger to write logs to console in development mode
	if baseConf.Environment == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	// initialize logger
	logger := fl.InitKitLogger(&fl.LoggerOptions{
		LogDestination:       baseConf.LogDestination,
		MongoUri:             appConfig.LogsMongoUri,
		NatsConnection:       baseConf.NatsConnection,
		LogsDestinationLabel: baseConf.LogsDestinationLabel,
	})

	// initialize 'PASETO' token maker
	maker, err := paseto.NewPasetoMaker(appConfig.TokenSymmetricKey, appConfig.Issuer, appConfig.Audience)
	if err != nil {
		log.Fatal().Err(err).Msgf("failed to create 'PASETO' token maker: %v", err)
	}

	// initialize postgres pool connection.
	dbTimeout := time.Duration(baseConf.DBConnectionTimeout) * time.Second
	pg := pgutil.NewOrGetSingleton(baseConf.DBMaxPoolSize, baseConf.DBConnectionAttempt, dbTimeout, baseConf.DBSource, logger)
	defer pg.Close()

	// Register custom PostgreSQL types (ULID) with pgx
	// This fixes the "cannot scan unknown type (OID 16702)" error
	pg.Pool.Config().AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		// Register ULID type - OID 16702 is the custom ULID type from the PostgreSQL extension
		conn.TypeMap().RegisterType(&pgtype.Type{
			Name:  "ulid",
			OID:   16702,
			Codec: &pgtype.TextCodec{},
		})
		return nil
	}

	go func(pool *pgxpool.Pool) {
		for {
			if e := pool.Ping(appContext); e != nil {
				_ = logger.Log("component", "postgres_connection", "error", e)
				log.Err(e).Msgf("postgres_connection: %v", e)
			}
			time.Sleep(time.Second * 5)
		}
	}(pg.Pool)

	// initialize store using postgres pool
	repo := repository.NewStore(pg.Pool)

	// initialize redis client
	redisClient, err := redisutils.InitRedisClient(
		baseConf.RedisHost,
		baseConf.RedisPassword,
		VERSION,
		baseConf.RedisDB,
	)
	if err != nil {
		log.Fatal().Err(err).Msgf("failed to initialize Redis client: %v", err)
	}
	defer redisClient.Close()

	counter := kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Namespace: "grpc",
		Subsystem: APPLICATION,
		Name:      "request_count",
		Help:      "num of requests received.",
	}, fieldKeys)

	latency := kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
		Namespace: "grpc",
		Subsystem: APPLICATION,
		Name:      "request_latency_microseconds",
		Help:      "total duration of requests (ms).",
	}, fieldKeys)

	userServiceMeshClient := mesh.NewUsersServiceMeshClient(natsConn, dbTimeout)

	walletServiceOpts := &wallet.WalletServiceOptions{
		Repository:            repo,
		BaseConfig:            &baseConf,
		Config:                &appConfig,
		Redis:                 redisClient,
		PasetoMaker:           maker,
		NATS:                  natsConn,
		SchemaPath:            jsonschema.GetSchemaPath(APPLICATION),
		ApplicationName:       APPLICATION,
		Logger:                logger,
		UserServiceMeshClient: &userServiceMeshClient,
	}

	service, err := wallet.NewWalletService(walletServiceOpts)
	if err != nil {
		log.Fatal().Err(err).Msgf("failed to create user service: %v", err)
	}

	// initialize admin service
	// adminServiceOpts := &admin.UserAdminServiceOpts{
	// 	Repository: repo,
	// 	BaseConfig: &baseConf,
	// 	Config:     &appConfig,
	// 	Redis:      redisClient,
	// 	PASETO:     maker,
	// 	NATS:       natsConn,
	// 	Logger:     logger,
	// }
	// adminService := admin.NewService(adminServiceOpts)

	// logging interceptor
	loggingInterceptor := grpcutil.NewLoggingInterceptor(logger, jaegerTracer)

	// instrumented interceptor
	instrumentedInterceptor := grpcutil.NewInstrumentingInterceptor(counter, latency, jaegerTracer)

	// validation interceptor
	validationInterceptor, err := grpcutil.NewValidationInterceptor(jsonschema.GetSchemaPath(APPLICATION), jaegerTracer)
	if err != nil {
		log.Fatal().Err(err).Msgf("failed to create validation interceptor: %v", err)
	}

	// paseto auth interceptor
	authInterceptor := grpcutil.NewPasetoAuthInterceptor(redisClient, grpcutil.AuthenticationOptions{
		HealthCheckKey: appConfig.HealthCheckKey,
		Secret:         appConfig.TokenSymmetricKey,
		GetUserInfo: func(ctx context.Context, id int64) (any, error) {
			return userServiceMeshClient.GetUserByID(ctx, uint64(id))
		},
		Paseto: maker,
		AccessRoles: map[string][]string{
			"DepositAsset":                  {"USER", "ADMIN", "SUPER_ADMIN", "MERCHANT"},
			"WithdrawAsset":                 {"USER", "ADMIN", "SUPER_ADMIN", "MERCHANT"},
			"TransferAsset":                 {"USER", "ADMIN", "SUPER_ADMIN", "MERCHANT"},
			"ExchangeAsset":                 {"USER", "ADMIN", "SUPER_ADMIN", "MERCHANT"},
			"ContextUserAccounts":           {"USER", "ADMIN", "SUPER_ADMIN", "MERCHANT"},
			"ContextUserWalletTransactions": {"USER", "ADMIN", "SUPER_ADMIN", "MERCHANT"},
			"GetAccountTransfers":           {"USER", "ADMIN", "SUPER_ADMIN", "MERCHANT"},
			"QueryTransfers":                {"USER", "ADMIN", "SUPER_ADMIN", "MERCHANT"},
			"LookupTransfers":               {"USER", "ADMIN", "SUPER_ADMIN", "MERCHANT"},
			"hCheck":                        {},
			"Watch":                         {},
		},
		Tracer: jaegerTracer,
	})

	// initialize gRPC server
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			loggingInterceptor.LoggingInterceptor(),
			instrumentedInterceptor.InstrumentingInterceptor(),
			validationInterceptor.ValidationInterceptor(),
			authInterceptor.PasetoAuthInterceptor(),
		),
	)

	// register user service gRPC service implementation
	pb.RegisterWalletServiceServer(server, service)

	// register users administration gRPC service implementation
	// pb.RegisterUsersAdminServiceServer(server, adminService)

	// use framework healthcheck service for readiness and liveness probes
	healthCheckServiceOpts := &healthcheck.HealthServiceOpts{
		Database:   pg.Pool,
		Redis:      redisClient,
		Mongo:      nil, //mongoClient,
		NATS:       natsConn,
		DaprClient: nil, //daprClient
		Tracer:     jaegerTracer,
	}

	healthCheckService := healthcheck.NewHealthService(healthCheckServiceOpts)
	healthpb.RegisterFrameworkHealthServiceServer(server, healthCheckService)

	// enable gRPC server reflection in development mode
	if !env.IsProduction() {
		reflection.Register(server)
	}

	// initialize gRPC listener
	listener, err := net.Listen("tcp", appConfig.GrpcListenerHost)
	if err != nil {
		log.Fatal().Err(err).Msgf("failed to listen: %v", err)
	}

	// start gRPC server
	go func() {
		if err := server.Serve(listener); err != nil {
			log.Fatal().Err(err).Msgf("failed to serve: %v", err)
		}
	}()

	log.Info().Msgf("%s v(%s) gRPC server started on %s", APPLICATION, VERSION, appConfig.GrpcListenerHost)

	// Create HTTP gateway server
	if appConfig.HttpListenerHost != "" {
		// Create a new gRPC-Gateway mux with snake_case field names
		gatewayMux := runtime.NewServeMux(
			runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
				MarshalOptions: protojson.MarshalOptions{
					UseProtoNames:   true,
					EmitUnpopulated: true,
				},
			}),
		)

		// Register the gateway endpoints
		opts := []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		}
		err = pb.RegisterWalletServiceHandlerFromEndpoint(appContext, gatewayMux, appConfig.GrpcListenerHost, opts)
		if err != nil {
			log.Fatal().Err(err).Msgf("failed to register gateway: %v", err)
		}

		// CORS middleware function
		corsHandler := func(h http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Set CORS headers
				w.Header().Set("Access-Control-Allow-Origin", "*")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
				w.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Type")
				w.Header().Set("Access-Control-Max-Age", "86400")

				// Handle preflight OPTIONS request
				if r.Method == "OPTIONS" {
					w.WriteHeader(http.StatusOK)
					return
				}

				// Call the next handler
				h.ServeHTTP(w, r)
			})
		}

		// Create HTTP server with CORS enabled
		httpServer := &http.Server{
			Addr:    appConfig.HttpListenerHost,
			Handler: corsHandler(gatewayMux),
		}

		// Start HTTP gateway server
		go func() {
			log.Info().Msgf("%s v(%s) HTTP gateway server started on %s", APPLICATION, VERSION, appConfig.HttpListenerHost)
			if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatal().Err(err).Msgf("failed to serve HTTP gateway: %v", err)
			}
		}()
	}

	// Set up a signal channel to capture shutdown signals
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for a shutdown signal
	<-shutdownChan

	log.Info().Msg("shutdown signal received, stopping server...")
	// Stop the server gracefully
	server.GracefulStop()

	log.Info().Msgf("%s v(%s) gracefully stopped.", APPLICATION, VERSION)
}
