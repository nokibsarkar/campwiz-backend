package distributionstrategy

import (
	"nokib/campwiz/models"
)

type DistributorServer struct {
	// The server that will be used to distribute the tasks
	models.UnimplementedDistributorServer
}

func NewDistributorServer() models.DistributorServer {
	return &DistributorServer{}
}
