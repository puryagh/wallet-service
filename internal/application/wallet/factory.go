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
	tb "github.com/tigerbeetle/tigerbeetle-go"
)

type (
	WalletServiceOptions struct {
		Repository            repository.Store
		BaseConfig            *configuration.BaseConfiguration
		Config                *config.Configuration
		Redis                 *redis.Client
		PasetoMaker           paseto.Maker
		NATS                  *nats.Conn
		SchemaPath            string
		ApplicationName       string
		Logger                kitlog.Logger
		UserServiceMeshClient mesh.UsersServiceMeshClient
		Tigerbeetle           tb.Client
	}
)

// NewWalletService creates a new wallet service with all middleware layers
func NewWalletService(opts *WalletServiceOptions) (pb.WalletServiceServer, error) {
	// Create base service
	svc := newService(opts)

	// Add middleware layers
	svc = NewAuthorizationMiddleware(svc)

	return svc, nil
}
