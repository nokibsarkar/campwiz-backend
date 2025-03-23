package importsources

import (
	"nokib/campwiz/models"
	"nokib/campwiz/repository"
)

type RoundCategoryListSource struct {
	Scores       []float64
	RoundId      string
	currentIndex string
	limit        int
	round_repo   *repository.RoundRepository
}

// ImportImageResults imports images from previous rounds
// For Each invocation it will import images from a single round
// If all rounds are imported it will return nil
func (c *RoundCategoryListSource) ImportImageResults(failedImageReason *map[string]string) ([]models.MediaResult, *map[string]string) {
	imageResults := []models.MediaResult{}
	q, close := repository.GetDBWithGen()
	defer close()
	Submission := q.Submission
	err := Submission.
		Select(Submission.Name.As("Name"),
			Submission.SubmissionID.As("SubmissionID"),
			Submission.URL.As("URL"),
			Submission.Author.As("UploaderUsername"),
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
		Where(Submission.RoundID.Eq(c.RoundId)).
		Where(Submission.Score.In(c.Scores...)).
		Limit(c.limit).
		Scan(&imageResults)
	if err != nil {
		(*failedImageReason)["*"] = err.Error()
		return nil, failedImageReason
	}
	if len(imageResults) == 0 {
		return nil, failedImageReason
	}
	c.currentIndex = imageResults[len(imageResults)-1].SubmissionID.String()
	return imageResults, failedImageReason
}
func NewRoundCategoryListSource(scores []models.ScoreType, roundId models.IDType) *RoundCategoryListSource {
	res := &RoundCategoryListSource{
		Scores:       []float64{},
		RoundId:      roundId.String(),
		currentIndex: "",
		round_repo:   repository.NewRoundRepository(),
		limit:        100,
	}
	for _, score := range scores {
		res.Scores = append(res.Scores, float64(score))
	}
	return res
}
