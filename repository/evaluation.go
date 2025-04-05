package repository

import (
	"nokib/campwiz/models"
	"nokib/campwiz/models/types"

	"gorm.io/gorm"
)

type EvaluationRepository struct{}

func NewEvaluationRepository() *EvaluationRepository {
	return &EvaluationRepository{}
}
func (r *EvaluationRepository) CreateEvaluation(tx *gorm.DB, evaluation *models.Evaluation) error {
	result := tx.Create(evaluation)
	return result.Error
}
func (r *EvaluationRepository) FindEvaluationByID(tx *gorm.DB, evaluationID models.IDType) (*models.Evaluation, error) {
	evaluation := &models.Evaluation{}
	result := tx.First(evaluation, &models.Evaluation{EvaluationID: evaluationID})
	return evaluation, result.Error
}
func (r *EvaluationRepository) ListAllEvaluations(tx *gorm.DB, filter *models.EvaluationFilter) ([]*models.Evaluation, error) {
	var evaluations []*models.Evaluation
	condition := &models.Evaluation{}
	stmt := tx
	if filter != nil {
		if filter.IncludeSubmission {
			stmt = tx.Preload("Submission")
		}
		s := &models.Submission{}
		if filter.CampaignID != "" {
			if filter.CampaignID != "" {
				s.CampaignID = filter.CampaignID
			}
			stmt = stmt.Joins("Submission", tx.Where(s))
		}
		if filter.RoundID != "" {
			condition.RoundID = filter.RoundID
		}
		if filter.ParticipantID != "" {
			condition.ParticipantID = filter.ParticipantID
		}
		if filter.Type != "" {
			condition.Type = filter.Type
		}
		if filter.Evaluated != nil {
			if *filter.Evaluated {
				stmt = stmt.Where("evaluated_at IS NOT NULL")
			} else {
				stmt = stmt.Where("evaluated_at IS NULL")
			}
		}
		if filter.SubmissionID != "" {
			condition.SubmissionID = filter.SubmissionID
		}
		if filter.JuryRoleID != "" {
			condition.JudgeID = &filter.JuryRoleID
		}
		if filter.ContinueToken != "" {
			stmt = stmt.Where("evaluation_id > ?", filter.ContinueToken)
		}
		if filter.PreviousToken != "" {
			stmt = stmt.Where("evaluation_id < ?", filter.PreviousToken)
		}
		if filter.Limit > 0 {
			stmt = stmt.Limit(max(5, filter.Limit))
		}
	}
	result := stmt.Where(condition).Find(&evaluations)
	return evaluations, result.Error
}
func (r *EvaluationRepository) UpdateEvaluation(tx *gorm.DB, evaluation *models.Evaluation) error {
	result := tx.Updates(evaluation)
	return result.Error
}
func (r *EvaluationRepository) ListSubmissionIDWithEvaluationCount(tx *gorm.DB, filter *models.EvaluationFilter) ([]models.NewEvaluationRequest, error) {
	results := []models.NewEvaluationRequest{}
	stmt := tx.Model(&models.Evaluation{}).Select("submission_id, count(evaluation_id) as EvaluationCount").Group("submission_id")
	if filter != nil {
		stmt = stmt.Where(filter)
	}
	result := stmt.Find(&results)
	return results, result.Error
}

// This function would be used to trigger the evaluation score counting
func (e *EvaluationRepository) TriggerEvaluationScoreCount(tx *gorm.DB, submissionIds []types.SubmissionIDType) error {
	// This function would be used to trigger the evaluation score counting
	stringSubmissionIds := make([]string, len(submissionIds))
	for i, id := range submissionIds {
		stringSubmissionIds[i] = string(id)
	}
	submission_repo := NewSubmissionRepository()
	return submission_repo.TriggerSubmissionStatistics(tx, stringSubmissionIds)
}
