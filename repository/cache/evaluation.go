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
	// UPDATE `evaluations` e
	//  JOIN (
	//     SELECT
	//         `evaluation_id`
	//     FROM
	//         `evaluations` e2
	//     WHERE e2.submission_id NOT IN (
	//     SELECT
	//         `submission_id`
	//     FROM
	//         `evaluations`
	//     WHERE
	//         `judge_id` = @my_judge_id
	//             AND
	//         `round_id` = @round_id
	//     UNION
	//         SELECT
	//             `submission_id`
	//         FROM
	//             `submissions` s
	//         WHERE
	//             `s`.`round_id` = @round_id
	//                 AND
	//                 `s`.`submitted_by_id` = @my_user_id
	//     ) AND
	//         e2.round_id = @round_id
	//      AND
	//         e2.score IS NULL
	//     AND
	//         e2.evaluated_at IS NULL
	//     AND
	//         (
	//             e2.distribution_task_id IS NULL
	//                 OR
	//             e2.distribution_task_id <> @task_id
	//     ) GROUP BY
	//         e2.submission_id
	//     LIMIT @N
	//  ) AS `e2`
	//     ON e.evaluation_id = e2.evaluation_id
	//     SET
	//     e.judge_id = @my_judge_id,
	//     e.distribution_task_id = @task_id
	// LIMIT @N;
	DistributeAssignmentsFromSelectedSource(my_judge_id models.IDType, my_user_id models.IDType, round_id string, reassignable_judges []string, task_id models.IDType, N int) (gen.RowsAffected, error)
	// 	UPDATE `evaluations` e
	//  JOIN (
	//     SELECT
	//         `evaluation_id`
	//     FROM
	//         `evaluations` e2
	//     WHERE e2.submission_id NOT IN (
	//     SELECT
	//         `submission_id`
	//     FROM
	//         `evaluations`
	//     WHERE
	//         `judge_id` = @my_judge_id
	//             AND
	//         `round_id` = @round_id
	//     UNION
	//         SELECT
	//             `submission_id`
	//         FROM
	//             `submissions` s
	//         WHERE
	//             `s`.`round_id` = @round_id
	//                 AND
	//                 `s`.`submitted_by_id` = @my_user_id
	//     ) AND
	//         e2.round_id = @round_id
	//     AND
	//         e2.score IS NULL
	//     AND
	//         e2.evaluated_at IS NULL
	//     AND
	//         (
	//             e2.distribution_task_id IS NULL
	//                 OR
	//             e2.distribution_task_id <> @task_id
	//     ) GROUP BY
	//         e2.submission_id
	//     LIMIT @N
	//  ) AS `e2`
	//     ON e.evaluation_id = e2.evaluation_id
	//     SET
	//     e.judge_id = @my_judge_id,
	//     e.distribution_task_id = @task_id
	// LIMIT @N;
	DistributeAssignmentsIncludingUnassigned(my_judge_id models.IDType, my_user_id models.IDType, round_id string, task_id models.IDType, N int) (gen.RowsAffected, error)
	// UPDATE `evaluations` e1
	// SET e1.judge_id = (
	// 		SELECT role_id FROM roles
	// 		WHERE role_id NOT IN (SELECT judge_id FROM evaluations WHERE submission_id = e1.submission_id AND round_id = @round_id)
	// 		AND round_id = @round_id
	// 		ORDER BY RAND()
	// 		LIMIT 1
	// ), e1.distribution_task_id = @task_id
	// WHERE e1.score IS NULL
	// AND e1.evaluated_at IS NULL
	// AND e1.judge_id IS NULL
	// AND e1.round_id = @round_id;
	DistributeTheLastRemainingEvaluations(task_id models.IDType, round_id string) (gen.RowsAffected, error)
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
