package statisticsupdater

import (
	"context"
	"log"
	"nokib/campwiz/models"
	"nokib/campwiz/repository"
)

var updatingSubmissionStatistics = false

func (s *StatisticsUpdaterServer) TriggerEvaluationScoreCount(ctx context.Context, req *models.UpdateStatisticsRequest) (*models.UpdateStatisticsResponse, error) {
	if !updatingSubmissionStatistics {
		go func() {
			updatingSubmissionStatistics = true
			defer func() {
				updatingSubmissionStatistics = false
			}()
			conn, close, err := repository.GetDB()
			if err != nil {
				return
			}
			defer close()
			submission_repo := repository.NewSubmissionRepository()
			err = submission_repo.TriggerSubmissionStatistics(conn, req.SubmissionIds)
			if err != nil {
				log.Println("Error while Updating Submission Statistics")
			}

		}()
	}
	return &models.UpdateStatisticsResponse{}, nil
}
