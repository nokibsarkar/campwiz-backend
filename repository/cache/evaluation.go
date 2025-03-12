package cache

import (
	"nokib/campwiz/models"
	idgenerator "nokib/campwiz/services/idGenerator"

	"gorm.io/gorm"
)

type Assignments struct {
	EvaluationID       models.IDType ` gorm:"primaryKey"`
	SubmissionID       models.IDType `gorm:"uniqueIndex:idx_submission_judge"`
	JudgeID            models.IDType `gorm:"uniqueIndex:idx_submission_judge"`
	DistributionTaskID models.IDType
}

func ExportToCache(tx *gorm.DB, cacheTx *gorm.DB, filter *models.EvaluationFilter) ([]*Assignments, error) {
	var evaluations []*Assignments
	condition := &models.Evaluation{}
	stmt := tx.Model(&models.Evaluation{})
	if filter != nil {
		if filter.RoundID != "" {
			condition.SubmissionID = filter.RoundID
		}
		if filter.ParticipantID != "" {
			condition.ParticipantID = filter.ParticipantID
		}
		if filter.JuryRoleID != "" {
			condition.JudgeID = filter.JuryRoleID
		}
		if filter.Limit > 0 {
			stmt = stmt.Limit(max(5, filter.Limit))
		}
	}
	result := stmt.Where(condition).Find(&evaluations)
	if result.Error != nil {
		return nil, result.Error
	}
	if len(evaluations) == 0 {
		return evaluations, nil
	}
	for _, evaluation := range evaluations {
		evaluation.DistributionTaskID = idgenerator.GenerateID("t")
	}
	cacheTx.Create(evaluations)
	return evaluations, result.Error
}
