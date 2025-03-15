package repository

import (
	"nokib/campwiz/models"
	"nokib/campwiz/models/types"
	"nokib/campwiz/query"

	"gorm.io/gorm"
)

type SubmissionRepository struct{}

func NewSubmissionRepository() *SubmissionRepository {
	return &SubmissionRepository{}
}
func (r *SubmissionRepository) CreateSubmission(tx *gorm.DB, submission *models.Submission) error {
	result := tx.Create(submission)
	return result.Error
}
func (r *SubmissionRepository) FindSubmissionByID(tx *gorm.DB, submissionID types.SubmissionIDType) (*models.Submission, error) {
	submission := &models.Submission{}
	result := tx.First(submission, &models.Submission{SubmissionID: submissionID})
	return submission, result.Error
}
func (r *SubmissionRepository) ListAllSubmissions(tx *gorm.DB, filter *models.SubmissionListFilter) ([]models.Submission, error) {
	var submissions []models.Submission
	condition := &models.Submission{}
	if filter != nil {
		if filter.CampaignID != "" {
			condition.CampaignID = filter.CampaignID
		}
		if filter.RoundID != "" {
			condition.RoundID = filter.RoundID
		}
		if filter.ParticipantID != "" {
			condition.ParticipantID = filter.ParticipantID
		}
	}
	where := tx.Where(condition)
	if filter.ContinueToken != "" {
		where = where.Where("submission_id > ?", filter.ContinueToken)
	}

	stmt := where //.Order("submission_id")
	if filter.Limit > 0 {
		stmt = stmt.Limit(max(100, filter.Limit))
	}
	result := stmt.Find(&submissions)
	return submissions, result.Error
}
func (r *SubmissionRepository) GetSubmissionCount(tx *gorm.DB, filter *models.SubmissionListFilter) (int64, error) {
	condition := &models.Submission{}
	if filter != nil {
		if filter.CampaignID != "" {
			condition.CampaignID = filter.CampaignID
		}
		if filter.RoundID != "" {
			condition.RoundID = filter.RoundID
		}
		if filter.ParticipantID != "" {
			condition.ParticipantID = filter.ParticipantID
		}
	}
	var count int64
	result := tx.Model(&models.Submission{}).Where(condition).Count(&count)
	return count, result.Error
}
func (r *SubmissionRepository) ListAllSubmissionIDs(tx *gorm.DB, filter *models.SubmissionListFilter) ([]models.SubmissionSelectID, error) {
	var submissionIDs []models.SubmissionSelectID
	condition := &models.Submission{}
	if filter != nil {
		if filter.CampaignID != "" {
			condition.CampaignID = filter.CampaignID
		}
		if filter.RoundID != "" {
			condition.RoundID = filter.RoundID
		}
		if filter.ParticipantID != "" {
			condition.ParticipantID = filter.ParticipantID
		}
	}
	where := tx.Where(condition)
	if filter.ContinueToken != "" {
		where = where.Where("submission_id > ?", filter.ContinueToken)
	}

	stmt := where //.Order("submission_id")
	if filter.Limit > 0 {
		stmt = stmt.Limit(max(100, filter.Limit))
	}
	result := stmt.Model(&models.Submission{}).Find(&submissionIDs)
	return submissionIDs, result.Error
}
func (r *SubmissionRepository) FindNextUnevaluatedSubmissionForPublicJury(tx *gorm.DB, juryID *models.IDType, round *models.Round) (*models.Submission, error) {
	alreadyCoveredSubmissionIDs := []string{}
	if juryID != nil {
		type submissionID struct {
			SubmissionID types.SubmissionIDType
		}
		results := []submissionID{}
		stmt := tx.Model(&models.Evaluation{}).Find(&results, &models.Evaluation{JudgeID: juryID})
		if stmt.Error != nil {
			return nil, stmt.Error
		}
		for _, s := range results {
			alreadyCoveredSubmissionIDs = append(alreadyCoveredSubmissionIDs, s.SubmissionID.String())
		}
	}

	q := query.Use(tx)
	s := q.Submission
	submission, error := (s.
		Where(s.RoundID.Eq(round.RoundID.String())).
		Where(s.SubmissionID.NotIn(alreadyCoveredSubmissionIDs...)).
		Where(s.EvaluationCount.Lt(round.Quorum)).
		Order(s.EvaluationCount).
		First())
	return submission, error
}
