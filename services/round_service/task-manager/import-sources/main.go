package importsources

import (
	"nokib/campwiz/models"
	"nokib/campwiz/query"

	"gorm.io/gorm"
)

type ImporterServer struct {
	models.UnimplementedImporterServer
}

// ImportService is an interface for importing data from different sources
// All the importer services should implement this interface
type IImportSource interface {
	// This method would be called in a loop to fetch each batch of images
	// It should return the images that were successfully imported and the images that failed to import
	// If there are no images to import it should return nil
	// If there are failed images it should return the reason as a map
	ImportImageResults(round *models.Round, failedImageReason *map[string]string) ([]models.MediaResult, *map[string]string)
}
type IDistributionStrategy interface {
	AssignJuries(tx *gorm.DB, round *models.Round, juries []models.Role) (success int, fail int, err error)
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
