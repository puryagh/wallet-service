package wallet

import (
	kitlog "github.com/go-kit/log"
	"github.com/liveutil/go-lib/configuration"
	"github.com/liveutil/go-lib/framework/mesh"
	"github.com/liveutil/go-lib/paseto"
	"github.com/liveutil/wallet-service/internal/abstract/pb"
	"github.com/liveutil/wallet-service/internal/config"
	"github.com/liveutil/wallet-service/internal/infra/db/postgres/repository"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
)

// service is the implementation of the wallet service
type service struct {
	pb.UnimplementedWalletServiceServer

	repository      repository.Store
	baseConfig      *configuration.BaseConfiguration
	config          *config.Configuration
	redis           *redis.Client
	pasetoMaker     paseto.Maker
	nats            *nats.Conn
	schemaPath      string
	applicationName string
	logger          kitlog.Logger

	userServiceMeshClient *mesh.UsersServiceMeshClient
}

// NewService creates a new wallet service
func newService(opts *WalletServiceOptions) pb.WalletServiceServer {
	return &service{
		repository:            opts.Repository,
		baseConfig:            opts.BaseConfig,
		config:                opts.Config,
		redis:                 opts.Redis,
		pasetoMaker:           opts.PasetoMaker,
		nats:                  opts.NATS,
		schemaPath:            opts.SchemaPath,
		applicationName:       opts.ApplicationName,
		logger:                opts.Logger,
		userServiceMeshClient: opts.UserServiceMeshClient,
	}
}
