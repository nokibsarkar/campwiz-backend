package round_service

import (
	"nokib/campwiz/models"
	"time"
)

type TechnicalJudgeService struct {
	AllowedTypes      models.MediaTypeSet
	MinimumUploadDate time.Time
	MinimumResolution uint64
	MinimumSize       uint64
	MaximumUploadDate time.Time
	// This would be a list of persons who are not allowed to submit images
	// Thes include the banned users, judges, coordinators, moderators etc
	Blacklist []string
}

func NewTechnicalJudgeService(round *models.Round, campaign *models.Campaign) *TechnicalJudgeService {
	return &TechnicalJudgeService{
		AllowedTypes:      round.AllowedMediaTypes,
		MinimumUploadDate: campaign.StartDate,
		MaximumUploadDate: campaign.EndDate,
		MinimumResolution: uint64(round.ImageMinimumResolution),
		MinimumSize:       uint64(round.ArticleMinimumTotalBytes),
		Blacklist:         []string{},
	}
}

// This method would perform some basic checks to see if the image is prevented from submission
// It would consider
// - For images
//   - Minimum Upload Date
//   - Minimum Resolution
//   - Minimum Size
//   - Whether Image allowed or not
func (j *TechnicalJudgeService) RejectReason(submission models.MediaResult) string {
	if j.AllowedTypes != nil && !j.AllowedTypes.Contains(models.MediaType(submission.MediaType)) {
		// log.Printf("Image %s is not allowed because it is of type %s", img.Name, img.MediaType)
		return "not-allowed-type"
	}
	if !j.MinimumUploadDate.IsZero() && submission.SubmittedAt.Before(j.MinimumUploadDate) {
		// log.Printf("Image %s is not allowed because it was uploaded before %s", img.Name, j.MinimumUploadDate)
		return "before-minimum-upload-date"
	}
	if !j.MaximumUploadDate.IsZero() && submission.SubmittedAt.After(j.MaximumUploadDate) {
		// log.Printf("Image %s is not allowed because it was uploaded after %s", img.Name, j.MaximumUploadDate)
		return "after-maximum-upload-date"
	}
	mt := models.MediaType(submission.MediaType)
	if (mt == models.MediaTypeImage || mt == models.MediaTypeVideo) && submission.Resolution < j.MinimumResolution {
		// log.Printf("Image %s is not allowed because it has a resolution of %d which is less than %d", img.Name, img.Resolution, j.MinimumResolution)
		return "below-minimum-resolution"
	}
	if submission.Size < j.MinimumSize {
		// log.Printf("Image %s is not allowed because it has a size of %d which is less than %d", img.Name, img.Size, j.MinimumSize)
		return "below-minimum-size"
	}

	return ""
}
