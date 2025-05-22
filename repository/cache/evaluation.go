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
	Score              *models.ScoreType `gorm:"type:FLOAT"`
}

type Dirtributor interface {
	// UPDATE `evaluations` SET `judge_id` = @judge_id WHERE `evaluations`.`judge_id` IS NULL AND `evaluations`.`evaluation_id` IN (SELECT `evaluation_id` FROM `evaluations` WHERE `submission_id` NOT IN (SELECT DISTINCT submission_id FROM evaluations WHERE `judge_id` = @judge_id) AND `judge_id` IS NULL GROUP BY `submission_id` LIMIT @limit)
	DistributeAssigments(judge_id models.IDType, limit int) (gen.RowsAffected, error)
	// SELECT COUNT(`evaluation_id`) AS Count, `judge_id` FROM `evaluations` GROUP BY `judge_id`
	CountAssignedEvaluations() ([]gen.T, error)
	// SELECT * FROM `evaluations` WHERE `judge_id` NOT IN (SELECT judge_id FROM evaluations WHERE `submission_id` = @submission_id AND `judge_id` IS NOT NULL) LIMIT @limit
	SelectUnAssignedJudges(submission_id types.SubmissionIDType, limit int) ([]*Evaluation, error)
	//update evaluations join (select max(evaluation_id) as evaluation_id from evaluations JOIN submissions using(submission_id) where evaluations.score is null and evaluations.evaluated_at is null and (evaluations.judge_id in (@reassignable_judges) OR evaluations.judge_id IS NULL) and evaluations.round_id = @round_id and submissions.round_id = @round_id and submissions.submitted_by_id <> @my_user_id and evaluations.submission_id not in (SELECT submission_id FROM evaluations where round_id = @round_id and judge_id = @judge_id) GROUP BY `evaluations`.`submission_id` LIMIT @N) as target_evals using (evaluation_id) set evaluations.judge_id = @judge_id, evaluations.distribution_task_id = @task_id where evaluations.round_id = @round_id and evaluations.distribution_task_id <> @task_id LIMIT @N
	DistributeAssignmentsFromSelectedSource(judge_id models.IDType, my_user_id models.IDType, round_id string, reassignable_judges []string, task_id models.IDType, N int) (gen.RowsAffected, error)
	//update evaluations join (select max(evaluation_id) as evaluation_id from evaluations JOIN submissions using(submission_id) where evaluations.score is null and evaluations.evaluated_at is null and evaluations.round_id = @round_id and submissions.round_id = @round_id and submissions.submitted_by_id <> @my_user_id and evaluations.submission_id not in (SELECT submission_id FROM evaluations where round_id = @round_id and judge_id = @judge_id) LIMIT @N) as target_evals using (evaluation_id) set evaluations.judge_id = @judge_id, evaluations.distribution_task_id = @task_id where evaluations.round_id = @round_id and evaluations.distribution_task_id <> @task_id LIMIT @N
	DistributeAssignmentsIncludingUnassigned(judge_id models.IDType, my_user_id models.IDType, round_id string, task_id models.IDType, N int) (gen.RowsAffected, error)
}

// update evaluations join (select min(evaluation_id) from evaluations JOIN submissions using(submission_id) where evaluations.score is null and evaluations.evaluated_at is null and evaluations.judge_id in ('r2eayfc854mio','r2au5d42wsb9c','r2audoxx0tatc') and evaluations.round_id = 'r2an2bnorpxc0' and submissions.round_id = 'r2an2bnorpxc0' and submissions.submitted_by_id <> 'u2aspyfke718g' and evaluations.submission_id not in (SELECT submission_id FROM evaluations where round_id = 'r2an2bnorpxc0' and judge_id = 'r2b6za303pcsg') LIMIT 3
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
