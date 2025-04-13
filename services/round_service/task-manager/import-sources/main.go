package importsources

import (
	"nokib/campwiz/models"
	"nokib/campwiz/query"

	"gorm.io/gorm"
)

type ImporterServer struct {
	models.UnimplementedImporterServer
}

func (imp *ImporterServer) updateStatistics(tx *gorm.DB, round *models.Round, successCount, failedCount int) error {
	type Result struct {
		TotalSubmissions          int
		TotalEvaluatedSubmissions int
	}
	var result Result
	q := query.Use(tx)
	Submission := q.Submission
	err := Submission.Select(Submission.SubmissionID.Count().As("TotalSubmissions"), Submission.AssignmentCount.Sum().
		As("TotalEvaluatedSubmissions")).Where(Submission.RoundID.Eq(round.RoundID.String())).Scan(&result)
	if err != nil {
		return err
	}
	res := tx.Updates(&models.Round{
		RoundID:                   round.RoundID,
		TotalSubmissions:          result.TotalSubmissions,
		TotalEvaluatedSubmissions: result.TotalEvaluatedSubmissions,
	})
	return res.Error
}
func NewImporterServer() *ImporterServer {
	return &ImporterServer{}
}
