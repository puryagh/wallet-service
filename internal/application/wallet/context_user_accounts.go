package wallet

import (
	"context"
	"fmt"

	"github.com/liveutil/go-lib/contextutil"
	"github.com/liveutil/wallet-service/internal/abstract/pb"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ContextUserAccounts implements [pb.WalletServiceServer].
func (s *service) ContextUserAccounts(ctx context.Context, req *emptypb.Empty) (*pb.ContextUserWalletResponse, error) {
	contextUser := &contextutil.ContextUser{}
	if err := contextutil.CatchUser(ctx, contextUser); err != nil {
		return nil, err
	}

	user, err := s.userServiceMeshClient.GetUserByID(ctx, uint64(contextUser.ID))
	if err != nil {
		return nil, err
	}

	fmt.Println(user)

	return &pb.ContextUserWalletResponse{
		Error:   false,
		Message: "success",
	}, nil
}
