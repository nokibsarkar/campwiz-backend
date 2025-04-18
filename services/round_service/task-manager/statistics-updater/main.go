package statisticsupdater

import "nokib/campwiz/models"

type StatisticsUpdaterServer struct {
	models.UnimplementedStatisticsUpdaterServer
}

func NewStatisticsUpdaterServer() models.StatisticsUpdaterServer {
	return &StatisticsUpdaterServer{}
}
