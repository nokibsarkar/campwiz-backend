package cache

import (
	"nokib/campwiz/models"
	"nokib/campwiz/models/types"
	idgenerator "nokib/campwiz/services/idGenerator"

	"gorm.io/gen"
	"gorm.io/gorm"
)

type Evaluation struct {
	EvaluationID       models.IDType          `gorm:"primaryKey"`
	SubmissionID       types.SubmissionIDType `gorm:"uniqueIndex:idx_submission_judge"`
	JudgeID            *models.IDType         `gorm:"index;null;uniqueIndex:idx_submission_judge"`
	DistributionTaskID models.IDType
}

type Dirtributor interface {
	// UPDATE `evaluations` SET `judge_id` = @judge_id WHERE `evaluations`.`judge_id` IS NULL AND `evaluations`.`evaluation_id` IN (SELECT MAX(`evaluation_id`) FROM `evaluations` WHERE `submission_id` NOT IN (SELECT DISTINCT submission_id FROM evaluations WHERE `judge_id` = @judge_id) AND `judge_id` IS NULL GROUP BY `submission_id` LIMIT @limit)
	DistributeAssigments(judge_id models.IDType, limit int) (gen.RowsAffected, error)
	// SELECT COUNT(`evaluation_id`) AS Count, `judge_id` FROM `evaluations` GROUP BY `judge_id`
	CountAssignedEvaluations() ([]gen.T, error)
	// SELECT * FROM `evaluations` WHERE `judge_id` NOT IN (SELECT judge_id FROM evaluations WHERE `submission_id` = @submission_id AND `judge_id` IS NOT NULL) LIMIT @limit
	SelectUnAssignedJudges(submission_id types.SubmissionIDType, limit int) ([]*Evaluation, error)
}

func ExportToCache(tx *gorm.DB, cacheTx *gorm.DB, filter *models.EvaluationFilter) ([]*Evaluation, error) {
	var evaluations []*Evaluation
	condition := &models.Evaluation{}
	stmt := tx.Model(&models.Evaluation{})
	if filter != nil {
		if filter.SubmissionID != "" {
			condition.SubmissionID = filter.SubmissionID
		}
		if filter.ParticipantID != "" {
			condition.ParticipantID = filter.ParticipantID
		}
		if filter.JuryRoleID != "" {
			condition.JudgeID = &filter.JuryRoleID
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
