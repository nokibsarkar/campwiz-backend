package distributionstrategy

import (
	"bufio"
	"context"
	"log"
	"nokib/campwiz/models"
	"nokib/campwiz/query"
	"nokib/campwiz/repository"
	"os"

	"fmt"

	"slices"

	"github.com/getsentry/sentry-go"
	"gorm.io/gorm"
)

// Prevent Self evaluation SQL: update evaluations u1 join (select judge_id, evaluation_id, name from evaluations join submissions join roles on evaluations.submission_id = submissions.submission_id and evaluations.judge_id = roles.role_id  where submitted_by_id=roles.user_id and evaluations.round_id='r2eczdvrjl2ps') u2 using(evaluation_id) set u1.judge_id = (select role_id from roles where role_id <> u2.judge_id and round_id='r2eczdvrjl2ps' and role_id not in (select judge_id from evaluations where submission_id = u1.submission_id) order by rand() limit 1) where round_id='r2eczdvrjl2ps' and score is null;
// This method would distribute all the evaluations to the juries in round robin fashion
func (strategy *RoundRobinDistributionStrategy) AssignJuries2(ctx context.Context) {
	log.Println("Assigning juries in round robin fashion version 2")
	parentSpan := sentry.StartSpan(ctx, "grpc.start", func(s *sentry.Span) {
		s.SetTag("round_id", strategy.RoundId.String())
		s.SetTag("task_id", strategy.TaskId.String())
		s.SetData("source_juries", strategy.SourceJuries)
		s.SetData("target_juries", strategy.TargetJuries)
		s.Description = "Assigning juries in round robin fashion version 2"
		s.SetData("strategy", "round-robin-v2")
	})
	defer parentSpan.Finish()

	taskRepo := repository.NewTaskRepository()
	submission_repo := repository.NewSubmissionRepository()
	conn, close, err := repository.GetDB(ctx)
	if err != nil {
		log.Println(err)
		return
	}
	defer close()
	task, err := taskRepo.FindByID(conn, strategy.TaskId)
	if err != nil {
		return
	}
	if task == nil {
		log.Println("Task not found")
		return
	}
	round_repo := repository.NewRoundRepository()
	round, err := round_repo.FindByID(conn, strategy.RoundId)
	if err != nil {
		log.Println("Round not found")
		return
	}
	if round == nil {
		log.Println("Round not found")
		return
	}
	if round.RoundID != *task.AssociatedRoundID {
		log.Println("Round ID mismatch")
		return
	}

	previousRoundStatus := round.Status
	round.Status = models.RoundStatusDistributing
	if err := conn.Save(round).Error; err != nil {
		log.Println("Error: ", err)
		return
	}
	defer func() {
		if _, updateErr := round_repo.Update(conn, &models.Round{
			RoundID: round.RoundID,
			Status:  previousRoundStatus,
		}); updateErr != nil {
			log.Println("Error: ", updateErr)
			return
		}

		if _, updateErr := taskRepo.Update(conn, &models.Task{
			TaskID:       task.TaskID,
			Status:       task.Status,
			SuccessCount: task.SuccessCount,
			FailedCount:  task.FailedCount,
		}); updateErr != nil {
			log.Println("Error: ", updateErr)
			return
		}
	}()
	submissions, err := submission_repo.ListAllSubmissions(conn.Where("assignment_count < ?", round.Quorum), &models.SubmissionListFilter{
		RoundID:    round.RoundID,
		CampaignID: round.CampaignID,
	})
	if err != nil {
		task.Status = models.TaskStatusFailed
		log.Println("Error: ", err)
		return
	}
	parentSpan.SetData("submission_count_missing_evaluations", len(submissions))
	createdCount, err := strategy.createMissingEvaluations(conn, round.Type, round, submissions)
	if err != nil {
		task.Status = models.TaskStatusFailed
		log.Println("Error: ", err)
		return
	}
	parentSpan.SetData("submission_count_created_missing_evaluations", createdCount)
	log.Println("Created missing evaluations: ", createdCount)
	q := query.Use(conn)
	Role := q.Role
	User := q.User
	//Target

	targetUsernames := strategy.TargetJuries
	targetRoles, err := Role.Select(Role.RoleID, Role.UserID, Role.TotalAssigned).Join(User, Role.UserID.EqCol(User.UserID)).Where(Role.RoundID.Eq(strategy.RoundId.String()), User.Username.In(targetUsernames...)).Limit(len(targetUsernames)).Find()
	if err != nil {
		task.Status = models.TaskStatusFailed
		log.Println("Error: ", err)
		return
	}
	targetRoleIds := make([]string, len(targetRoles))
	for i, targetRole := range targetRoles {
		targetRoleIds[i] = targetRole.RoleID.String()
	}
	// Source
	// From these jury, we would be collecting the evaluations
	sourceUsernames := []string{}
	for _, sourceUserId := range strategy.SourceJuries {
		sourceUsernames = append(sourceUsernames, sourceUserId.String())
	}
	sourceRoles, err := Role.Select(Role.RoleID).Join(User, Role.UserID.EqCol(User.UserID)).Where(Role.RoundID.Eq(strategy.RoundId.String()), User.Username.In(sourceUsernames...)).Limit(len(sourceUsernames)).Find()
	if err != nil {
		log.Println("Error: ", err)
		task.Status = models.TaskStatusFailed
		return
	}
	sourceRoleIds := make([]string, len(sourceRoles))
	for i, sourceRole := range sourceRoles {
		sourceRoleIds[i] = sourceRole.RoleID.String()
	}

	includeFromSourceOnly := len(sourceRoleIds) > 0
	parentSpan.SetData("whether_include_from_source_only", includeFromSourceOnly)
	log.Println("Include from source only: ", includeFromSourceOnly)
	tx := conn.Begin()
	defer func() {
		log.Println("Distributing finished")
		if err != nil {
			tx.Rollback()
			task.Status = models.TaskStatusFailed
			log.Println("Error: ", err)
		} else {
			log.Println("Committing transaction")
			tx.Commit()
			task.Status = models.TaskStatusSuccess
		}
	}()
	q1 := query.Use(tx)

	// /////////////////////////////////////////////////////////
	Assignment := q1.Evaluation
	roundId := strategy.RoundId
	for range 1 {
		alreadyAssignedWorkloads := MinimumWorkloadHeap{}
		alreadyAssignedToTargetJury := map[models.IDType]WorkLoadType{}
		err = Assignment.Select(Assignment.JudgeID, Assignment.EvaluationID.Count().As("Count")).
			Where(Assignment.RoundID.Eq(roundId.String())).Where(Assignment.JudgeID.In(targetRoleIds...)).
			Group(Assignment.JudgeID).Scan(&alreadyAssignedWorkloads)
		if err != nil {
			return
		}
		totalAlreadyAssignedtoTargetJuryCount := WorkLoadType(0)
		for _, workload := range alreadyAssignedWorkloads {
			totalAlreadyAssignedtoTargetJuryCount += workload.Count
			alreadyAssignedToTargetJury[workload.JudgeID] = WorkLoadType(workload.Count)
		}
		parentSpan.SetData("total_assigned_count", totalAlreadyAssignedtoTargetJuryCount)

		// get current number of evaluations to be distributed
		evaluatedByTarget := []JurorV3{}
		stmt := Assignment.Select(Assignment.EvaluationID.Count().As("Count"), Assignment.JudgeID).
			Where(Assignment.RoundID.Eq(strategy.RoundId.String())).
			Where(Assignment.Score.IsNotNull()).
			Where(Assignment.EvaluatedAt.IsNotNull()).
			Group(Assignment.JudgeID).
			Where(Assignment.JudgeID.In(targetRoleIds...))

		alreadyEvaluatedByTargetMap := map[models.IDType]WorkLoadType{}
		// Get the total number of unevaluated elligible assignments
		err = stmt.Scan(&evaluatedByTarget)
		if err != gorm.ErrRecordNotFound && err != nil {
			log.Println("Error: ", err)
			return
		}
		parentSpan.SetData("evaluated_assignment_count", len(evaluatedByTarget))
		for _, juror := range evaluatedByTarget {
			alreadyEvaluatedByTargetMap[juror.JudgeID] = WorkLoadType(juror.Count)
		}
		transferableAssignmentCount := []JurorV3{}
		stmt = Assignment.Select(Assignment.EvaluationID.Count().As("Count"), Assignment.JudgeID).
			Where(Assignment.RoundID.Eq(strategy.RoundId.String())).
			Where(Assignment.Score.IsNull()).
			Where(Assignment.EvaluatedAt.IsNull()).
			Group(Assignment.JudgeID)
		if len(sourceRoleIds) > 0 {
			stmt = stmt.Where(Assignment.JudgeID.In(sourceRoleIds...))
		}
		// Get the total number of unevaluated elligible assignments
		err = stmt.Scan(&transferableAssignmentCount)
		if err != gorm.ErrRecordNotFound && err != nil {
			log.Println("Error: ", err)
			return
		}

		totalTransferableEvaluations := 0
		for _, juror := range transferableAssignmentCount {
			if _, ok := alreadyAssignedToTargetJury[juror.JudgeID]; !ok {
				totalTransferableEvaluations += int(juror.Count)
			}
		}
		totalReassignableEvaluations := totalAlreadyAssignedtoTargetJuryCount + WorkLoadType(totalTransferableEvaluations)

		parentSpan.SetData("total_transferable_evaluations", totalTransferableEvaluations)
		log.Println("Total unevaluated evaluations: ", totalTransferableEvaluations)
		targetRoleCount := len(targetRoles)
		log.Println("Total target roles: ", targetRoleCount)
		log.Printf("Tranferable assignments: %+v", alreadyEvaluatedByTargetMap)
		log.Printf("Already assigned to target jury: %+v", alreadyAssignedToTargetJury)
		parentSpan.SetData("target_role_count", targetRoleCount)
		parentSpan.SetData("already_assigned_workflow_map", alreadyAssignedToTargetJury)
		// for each of the target roles, we would be distributing the evaluations
		errorMargin := 0
		scanner := bufio.NewScanner(os.Stdin)
		for i := range targetRoleCount {
			// select locked
			b := JurorV3{}
			err = q1.Evaluation.Select(q1.Evaluation.EvaluationID.Count().As("Count")).
				Where(q1.Evaluation.RoundID.Eq(strategy.RoundId.String())).
				// Where(q1.Evaluation.JudgeID.Eq(targetJudgeId.String())).
				Where(q1.Evaluation.DistributionTaskID.Eq(strategy.TaskId.String())).Scan(&b)
			if err != nil {
				log.Println("Error: ", err)
				task.Status = models.TaskStatusFailed
				return
			}
			log.Printf("\n\n\tCurrently locked evaluations for target role %d: %d", i, b.Count)

			childSpan := parentSpan.StartChild("assignment.distribution.round-robin-v2.target-role", func(s *sentry.Span) {
				s.SetTag("target_role_index", fmt.Sprint(i))
				s.SetTag("target_role_id", targetRoles[i].RoleID.String())
				s.SetTag("target_role_user_id", targetRoles[i].UserID.String())
				s.SetData("target_role_total_assigned", targetRoles[i].TotalAssigned)
			})
			defer childSpan.Finish()

			targetRole := targetRoles[i]
			targetJudgeId := targetRole.RoleID
			judgeUserId := targetRole.UserID
			log.Println("Target Judge ID: ", targetJudgeId)
			childSpan.SetData("target_judge_id", targetJudgeId.String())
			// Get the average
			log.Printf("Total reassignable evaluations: %d", totalReassignableEvaluations)

			currentUnAssignedJuryCount := targetRoleCount - i
			childSpan.SetData("current_unassigned_jury_count", currentUnAssignedJuryCount)
			avgWorkload := int(totalReassignableEvaluations)/currentUnAssignedJuryCount + errorMargin
			log.Printf("Average workload for target judge %s: %d", targetJudgeId, avgWorkload)
			childSpan.SetData("average_workload", avgWorkload)

			//// DETERMINE THE WORKLOAD FOR THE JURY
			alreadyAssigned := alreadyAssignedToTargetJury[targetJudgeId]
			alreadyEvaluated := alreadyEvaluatedByTargetMap[targetJudgeId]
			Evaluation := q1.Evaluation

			keepUpto := min(WorkLoadType(avgWorkload), alreadyAssigned)
			lockableEvaluations := max(alreadyEvaluated, keepUpto) - alreadyEvaluated
			locked := int64(0)
			if lockableEvaluations > 0 {
				log.Printf("Locking assigned workload for target judge %s: %d", targetJudgeId, lockableEvaluations)
				stmt := Evaluation.Where(Evaluation.JudgeID.Eq(targetJudgeId.String())).
					Where(Evaluation.RoundID.Eq(strategy.RoundId.String())).
					Where(Evaluation.Score.IsNull()).
					Where(Evaluation.EvaluatedAt.IsNull()).
					Where(Evaluation.Where(Evaluation.DistributionTaskID.Neq(strategy.TaskId.String())).Or(Evaluation.DistributionTaskID.IsNull())).
					Limit(int(lockableEvaluations))
				res, err := stmt.UpdateColumn(Evaluation.DistributionTaskID, strategy.TaskId)
				if err != nil {
					log.Println("Error locking workload: ", err)
					task.Status = models.TaskStatusFailed
					return
				}
				locked = res.RowsAffected
			}
			if avgWorkload > int(alreadyAssigned) {
				// add the difference to the workload
				newlyAssignable := avgWorkload - int(alreadyAssigned)

				var newlyAssigned int64
				if includeFromSourceOnly {
					combinedSourceRoleIds := slices.Concat(sourceRoleIds, targetRoleIds[:i-1])
					newlyAssigned, err = q1.Evaluation.WithContext(ctx).DistributeAssignmentsFromSelectedSource(targetJudgeId, judgeUserId, strategy.RoundId.String(), combinedSourceRoleIds, strategy.TaskId, int(newlyAssignable))
				} else {
					newlyAssigned, err = q1.Evaluation.WithContext(ctx).DistributeAssignmentsIncludingUnassigned(targetJudgeId, judgeUserId, strategy.RoundId.String(), strategy.TaskId, int(newlyAssignable))
				}
				if err != nil {
					log.Println("Error: ", err)
					task.Status = models.TaskStatusFailed
					return
				}
				totalReassignableEvaluations -= alreadyEvaluated + WorkLoadType(locked) + WorkLoadType(newlyAssigned)
			} else if avgWorkload <= int(alreadyAssigned) {
				// If the workload is less than the already assigned workload,
				// the lock already made
				totalReassignableEvaluations -= alreadyEvaluated + WorkLoadType(locked)
			}
			// wait for confirmation
			childSpan.SetData("target_judge_locked_evaluations", locked)
			log.Printf("Press Enter to continue for target judge %s", targetJudgeId)
			scanner.Scan()
		}
	}
	err = strategy.triggerStatisticsUpdateByRoundID(tx, round)
	if err != nil {
		task.Status = models.TaskStatusFailed
		log.Println("Error: ", err)
		return
	}
	affected, err := q1.Evaluation.WithContext(ctx).DistributeTheLastRemainingEvaluations(strategy.TaskId, strategy.RoundId.String())
	if err != nil {
		log.Println("Error: ", err)
		task.Status = models.TaskStatusFailed
		return
	}
	log.Printf("Total affected evaluations: %d", affected)
	task.SuccessCount += int(affected)
}
