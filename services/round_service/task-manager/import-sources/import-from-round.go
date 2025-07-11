package importsources

import (
	"context"
	"fmt"
	"log"
	"nokib/campwiz/models"
	"nokib/campwiz/repository"
)

type RoundPreviousRound struct {
	Scores        []float64
	SourceRoundId string
	currentIndex  string
	limit         int
	round_repo    *repository.RoundRepository
}

func (t *ImporterServer) ImportFromPreviousRound(ctx context.Context, req *models.ImportFromPreviousRoundRequest) (*models.ImportResponse, error) {
	log.Println("Importing from previous round")

	currentRoundId := req.GetRoundId()
	sourceRoundId := req.GetSourceRoundId()
	if sourceRoundId == "" {
		return nil, fmt.Errorf("sourceRoundId is required")
	}
	if currentRoundId == "" {
		return nil, fmt.Errorf("roundId is required")
	}
	taskId := req.GetTaskId()
	if taskId == "" {
		return nil, fmt.Errorf("taskId is required")
	}

	scores := []float64{} // Ensure scores are in float64 format
	for _, score := range req.GetScores() {
		scores = append(scores, float64(score))
	}

	source := RoundPreviousRound{
		Scores:        scores,
		SourceRoundId: sourceRoundId,
		currentIndex:  "",
		round_repo:    repository.NewRoundRepository(),
		limit:         100,
	}
	go t.importFrom(context.Background(), &source, taskId, currentRoundId)

	return &models.ImportResponse{
		TaskId:  taskId,
		RoundId: currentRoundId,
	}, nil

}

// ImportImageResults imports images from previous rounds
// For Each invocation it will import images from a single round
// If all rounds are imported it will return nil
func (c *RoundPreviousRound) ImportImageResults(ctx context.Context, currentRound *models.Round, failedImageReason *map[string]string) ([]models.MediaResult, *map[string]string) {
	imageResults := []models.MediaResult{}
	q, close := repository.GetDBWithGen(ctx)
	defer close()
	log.Println("Importing images from previous round:", c.SourceRoundId, "with score:", c.Scores)
	Submission := q.Submission
	err := Submission.
		Select(Submission.Name.As("Name"),
			Submission.SubmissionID.As("SubmissionID"),
			Submission.URL.As("URL"),
			Submission.Author.As("CreatedByUsername"),
			Submission.Author.As("SubmittedByUsername"),
			Submission.Height.As("Height"),
			Submission.Width.As("Width"),
			Submission.Size.As("Size"),
			Submission.MediaType.As("MediaType"),
			Submission.Duration.As("Duration"),
			Submission.License.As("License"),
			Submission.Description.As("Description"),
			Submission.CreditHTML.As("CreditHTML"),
			Submission.ThumbHeight.As("ThumbHeight"),
			Submission.ThumbWidth.As("ThumbWidth"),
			Submission.ThumbURL.As("ThumbURL"),
			Submission.Resolution.As("Resolution"),
			Submission.SubmittedAt.As("SubmittedAt"),
			Submission.PageID.As("PageID"),
		).Where(Submission.SubmissionID.Gt(c.currentIndex)).
		Where(Submission.RoundID.Eq(c.SourceRoundId)).
		Where(Submission.Score.In(c.Scores...)).
		Limit(c.limit).
		Scan(&imageResults)
	if err != nil {
		log.Println("Error importing images from previous round:", err)
		(*failedImageReason)["*"] = err.Error()
		return nil, failedImageReason
	}
	if len(imageResults) == 0 {
		return nil, failedImageReason
	}
	c.currentIndex = imageResults[len(imageResults)-1].SubmissionID.String()
	return imageResults, failedImageReason
}
func NewRoundPreviousRound(scores []models.ScoreType, roundId models.IDType) *RoundPreviousRound {
	res := &RoundPreviousRound{
		Scores:        make([]float64, len(scores)),
		SourceRoundId: roundId.String(),
		currentIndex:  "",
		round_repo:    repository.NewRoundRepository(),
		limit:         100,
	}
	for i, score := range scores {
		res.Scores[i] = float64(score)
	}
	return res
}
