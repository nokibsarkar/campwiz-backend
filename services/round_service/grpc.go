package round_service

import (
	"fmt"
	"nokib/campwiz/consts"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewGrpcClient() (*grpc.ClientConn, error) {
	serverAddr := fmt.Sprintf("%s:%s", consts.Config.TaskManager.Host, consts.Config.TaskManager.Port)
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	return grpc.NewClient(serverAddr, opts...)
}
