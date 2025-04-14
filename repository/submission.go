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
func (r *SubmissionRepository) FindNextUnevaluatedSubmissionForPublicJury(tx *gorm.DB, filter *models.EvaluationFilter, round *models.Round) ([]*models.Submission, error) {
	q := query.Use(tx)
	Submission := q.Submission
	stmt := Submission.Where(Submission.RoundID.Eq(round.RoundID.String()))
	alreadyCoveredSubmissionIDs := []string{}
	includeEvaluated := false
	includeNonEvaluated := true
	if filter.JuryRoleID != "" {
		type submissionID struct {
			SubmissionID types.SubmissionIDType
		}
		results := []submissionID{}
		stmt := tx.Model(&models.Evaluation{}).Find(&results, &models.Evaluation{JudgeID: &filter.JuryRoleID})
		if stmt.Error != nil {
			return nil, stmt.Error
		}
		for _, s := range results {
			alreadyCoveredSubmissionIDs = append(alreadyCoveredSubmissionIDs, s.SubmissionID.String())
		}

		if filter.IncludeEvaluated != nil {
			includeEvaluated = *filter.IncludeEvaluated
		}
		if filter.IncludeNonEvaluated != nil {
			includeNonEvaluated = *filter.IncludeNonEvaluated
		}
	}
	if includeEvaluated != includeNonEvaluated {
		if includeEvaluated {
			stmt = stmt.Where(Submission.SubmissionID.In(alreadyCoveredSubmissionIDs...))
		} else if includeNonEvaluated {
			stmt = stmt.Where(Submission.SubmissionID.NotIn(alreadyCoveredSubmissionIDs...)).
				Where(Submission.EvaluationCount.Lt(round.Quorum))
		}
	}
	if filter.ContinueToken != "" {
		stmt = stmt.Where(Submission.SubmissionID.Gt(filter.ContinueToken))
	} else if filter.PreviousToken != "" {
		stmt = stmt.Where(Submission.SubmissionID.Lt(filter.PreviousToken))
	}
	if filter.Limit > 0 {
		stmt = stmt.Limit(filter.Limit)
	}
	submissions, error := (stmt.
		Order(Submission.EvaluationCount.Asc()).
		Find())
	return submissions, error
}
func (r *SubmissionRepository) GetPageIDsForWithout(tx *gorm.DB, roundID models.IDType) ([]uint64, error) {
	pageIds := []uint64{}
	q := query.Use(tx)
	s := q.Submission
	err := s.Select(s.PageID).Where(s.RoundID.Eq(roundID.String())).Scan(&pageIds)
	return pageIds, err
}
func (r *SubmissionRepository) TriggerSubmissionStatistics(tx *gorm.DB, submissionIds []string) error {
	if len(submissionIds) == 0 {
		return nil
	}

	q := query.Use(tx)
	rowsAffected, err := q.SubmissionStatistics.TriggerBySubmissionIds(submissionIds)
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		// Nothing changed, no need to trigger further statistics
		return nil
	}

	submissionQuery := q.Submission
	// Prepare to trigger upper lvl statistics
	// Get the first submission ID
	firstSubmissionID := submissionIds[0]
	if firstSubmissionID == "" {
		return nil
	}
	// Fetch First Submission
	firstSubmission, err := submissionQuery.Where(submissionQuery.SubmissionID.Eq(firstSubmissionID)).First()
	if err != nil {
		// Handle error
		return err
	}
	if firstSubmission == nil {
		// Handle case where submission is not found
		return nil
	}
	roundId := firstSubmission.RoundID
	// now trigger the round statistics
	round_repo := &RoundRepository{}
	err = round_repo.UpdateStatisticsByRoundID(tx, roundId)
	return err

}
func (r *SubmissionRepository) GetPageIDWithoutDescriptionByRoundID(tx *gorm.DB, roundID models.IDType, lastPageID uint64, limit int) ([]uint64, error) {
	pageIds := []uint64{}
	q := query.Use(tx)
	Submission := q.Submission
	err := Submission.Select(Submission.PageID).Where(Submission.RoundID.Eq(roundID.String())).Where(Submission.Description.Eq("")).
		Where(Submission.PageID.Gt(lastPageID)).Order(Submission.PageID.Asc()).Limit(limit).Scan(&pageIds)
	if err != nil {
		return nil, err
	}
	return pageIds, err
}
