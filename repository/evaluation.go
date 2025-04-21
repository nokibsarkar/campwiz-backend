package repository

import (
	"math/rand"
	"nokib/campwiz/models"
	"nokib/campwiz/models/types"
	"nokib/campwiz/query"

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
	q := query.Use(tx)
	Evaluation := q.Evaluation
	stmt1 := Evaluation.Select((Evaluation.ALL))
	if filter != nil {

		if filter.IncludeSubmission {
			stmt1 = stmt1.Preload(Evaluation.Submission)
		}
		if filter.IncludeEvaluated != nil {
			if *filter.IncludeEvaluated {
				stmt1 = stmt1.Where(Evaluation.Score.IsNotNull())
			} else {
				stmt1 = stmt1.Where(Evaluation.Score.IsNull())
			}
		}
		s := &models.Submission{}
		if filter.CampaignID != "" {

			if filter.CampaignID != "" {
				s.CampaignID = filter.CampaignID
				stmt1 = stmt1.Join(q.Submission, q.Submission.CampaignID.Eq(filter.CampaignID.String()))
			}
		}
		if filter.RoundID != "" {
			stmt1 = stmt1.Where(Evaluation.RoundID.Eq(filter.RoundID.String()))
		}
		if filter.ParticipantID != "" {
			stmt1 = stmt1.Where(Evaluation.ParticipantID.Eq(filter.ParticipantID.String()))
		}
		if filter.Type != "" {
			stmt1 = stmt1.Where(Evaluation.Type.Eq(string(filter.Type)))
		}

		if filter.IncludeSkipped != nil {
			if *filter.IncludeSkipped {
				stmt1 = stmt1.Where(Evaluation.SkipExpirationAt.IsNotNull())
			} else {
				stmt1 = stmt1.Where(Evaluation.SkipExpirationAt.IsNull())
			}
		}
		if filter.SubmissionID != "" {
			stmt1 = stmt1.Where(Evaluation.SubmissionID.Eq(filter.SubmissionID.String()))
		}
		if filter.JuryRoleID != "" {
			stmt1 = stmt1.Where(Evaluation.JudgeID.Eq(filter.JuryRoleID.String()))
		}
		if filter.ContinueToken != "" {
			stmt1 = stmt1.Where(Evaluation.EvaluationID.Gt(filter.ContinueToken))
		}
		if filter.PreviousToken != "" {
			stmt1 = stmt1.Where(Evaluation.EvaluationID.Lt(filter.PreviousToken))
		}
		if filter.Limit > 0 {
			stmt1 = stmt1.Limit(max(5, filter.Limit))
		}
	}
	evaluations, err := stmt1.Find()
	if err != nil {
		return nil, err
	}
	if filter.Randomize {
		rand.Shuffle(len(evaluations), func(i, j int) {
			evaluations[i], evaluations[j] = evaluations[j], evaluations[i]
		})
	}
	return evaluations, err
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
