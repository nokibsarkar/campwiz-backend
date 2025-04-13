package distributionstrategy

import (
	"context"
	"nokib/campwiz/models"
)

type DistributorServer struct {
	// The server that will be used to distribute the tasks
	models.UnimplementedDistributorServer
}

func NewDistributorServer() models.DistributorServer {
	return &DistributorServer{}
}
func (d *DistributorServer) DistributeWithRoundRobin(context.Context, *models.DistributeWithRoundRobinRequest) (*models.DistributeWithRoundRobinResponse, error) {
	// Implement the round robin distribution logic here
	return &models.DistributeWithRoundRobinResponse{}, nil
}
