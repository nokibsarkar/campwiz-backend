package distributionstrategy

import (
	"context"
	"log"
	"nokib/campwiz/models"
	"nokib/campwiz/query"
	"nokib/campwiz/repository"

	"gorm.io/gorm"
)

// Prevent Self evaluation SQL: update evaluations u1 join (select judge_id, evaluation_id, name from evaluations join submissions join roles on evaluations.submission_id = submissions.submission_id and evaluations.judge_id = roles.role_id  where submitted_by_id=roles.user_id and evaluations.round_id='r2eczdvrjl2ps') u2 using(evaluation_id) set u1.judge_id = (select role_id from roles where role_id <> u2.judge_id and round_id='r2eczdvrjl2ps' and role_id not in (select judge_id from evaluations where submission_id = u1.submission_id) order by rand() limit 1) where round_id='r2eczdvrjl2ps' and score is null;
// This method would distribute all the evaluations to the juries in round robin fashion
func (strategy *RoundRobinDistributionStrategy) AssignJuries2(ctx context.Context) {
	log.Println("Assigning juries in round robin fashion version 2")
	// parentSpan := sentry.StartSpan(ctx, "grpc.start", func(s *sentry.Span) {
	// 	s.SetTag("round_id", strategy.RoundId.String())
	// 	s.SetTag("task_id", strategy.TaskId.String())
	// 	s.SetData("source_juries", strategy.SourceJuries)
	// 	s.SetData("target_juries", strategy.TargetJuries)
	// 	s.Description = "Assigning juries in round robin fashion version 2"
	// 	s.SetData("strategy", "round-robin-v2")
	// })
	// defer parentSpan.Finish()

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
	// parentSpan.SetData("submission_count_missing_evaluations", len(submissions))
	createdCount, err := strategy.createMissingEvaluations(conn, round.Type, round, submissions)
	if err != nil {
		task.Status = models.TaskStatusFailed
		log.Println("Error: ", err)
		return
	}
	// parentSpan.SetData("submission_count_created_missing_evaluations", createdCount)
	log.Println("Created missing evaluations: ", createdCount)
	q := query.Use(conn)
	Role := q.Role
	User := q.User

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
	includeFromSourceOnly := len(sourceRoleIds) > 0
	// parentSpan.SetData("whether_include_from_source_only", includeFromSourceOnly)
	log.Println("Include from source only: ", includeFromSourceOnly)
	// // Calculate the workload
	// newWorkload, err := strategy.calculateWorkloadV2(conn, strategy.RoundId, sourceRoleIds, targetRoleIds)
	// if err != nil {
	// 	task.Status = models.TaskStatusFailed
	// 	log.Println("Error: ", err)
	// 	return
	// }
	// log.Printf("New workload: %+v", newWorkload)
	// Get the total number of evaluations
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
	// // first lock the negative workloads
	// for judgeId, workload := range newWorkload {
	// 	if workload < 0 {
	// 		log.Printf("Locking negative workload for judge %s: %d", judgeId, workload)
	// 		// Lock the workload in the database
	// 		Evaluation := q1.Evaluation
	// 		stmt := Evaluation.Where(Evaluation.JudgeID.Eq(judgeId.String())).
	// 			Where(Evaluation.RoundID.Eq(strategy.RoundId.String())).
	// 			Where(Evaluation.Score.IsNull()).
	// 			Where(Evaluation.EvaluatedAt.IsNull()).
	// 			Where(Evaluation.Where(Evaluation.DistributionTaskID.Neq(strategy.TaskId.String())).Or(Evaluation.DistributionTaskID.IsNull())).
	// 			Limit(-int(workload))
	// 		res, err := stmt.UpdateColumn(Evaluation.DistributionTaskID, strategy.TaskId)
	// 		if err != nil {
	// 			log.Println("Error locking workload: ", err)
	// 			task.Status = models.TaskStatusFailed
	// 			return
	// 		}
	// 		log.Printf("Rows affected by locking negative workload for judge %s: %d", judgeId, res.RowsAffected)
	// 		delete(newWorkload, judgeId) // Remove the negative workload from the map
	// 		log.Printf("Removed negative workload for judge %s: %d", judgeId, workload)
	// 	}
	// }

	// /////////////////////////////////////////////////////////
	Assignment := q1.Evaluation
	roundId := strategy.RoundId
	alreadyAssignedWorkloads := MinimumWorkloadHeap{}
	alreadyAssignedWorkflowMap := map[models.IDType]WorkLoadType{}
	err = Assignment.Select(Assignment.JudgeID, Assignment.EvaluationID.Count().As("Count")).
		// Where(Assignment.Score.IsNotNull()).Where(Assignment.EvaluatedAt.IsNotNull()).
		Where(Assignment.RoundID.Eq(roundId.String())).Where(Assignment.JudgeID.In(targetRoleIds...)).
		Group(Assignment.JudgeID).Scan(&alreadyAssignedWorkloads)
	if err != nil {
		return
	}
	totalAssignedCount := WorkLoadType(0)
	for _, workload := range alreadyAssignedWorkloads {
		totalAssignedCount += workload.Count
		alreadyAssignedWorkflowMap[workload.JudgeID] = WorkLoadType(workload.Count)
	}
	// parentSpan.SetData("total_assigned_count", totalAssignedCount)

	// get current number of evaluations to be distributed
	evaluatedAssignmentCount := []JurorV3{}
	stmt := Assignment.Select(Assignment.EvaluationID.Count().As("Count"), Assignment.JudgeID).
		Where(Assignment.RoundID.Eq(strategy.RoundId.String())).
		Where(Assignment.Score.IsNotNull()).
		Where(Assignment.EvaluatedAt.IsNotNull()).
		Group(Assignment.JudgeID).
		Where(Assignment.JudgeID.In(targetRoleIds...))

	evaluatedMap := map[models.IDType]WorkLoadType{}
	// Get the total number of unevaluated elligible assignments
	err = stmt.Scan(&evaluatedAssignmentCount)
	if err != gorm.ErrRecordNotFound && err != nil {
		log.Println("Error: ", err)
		return
	}
	// parentSpan.SetData("evaluated_assignment_count", len(evaluatedAssignmentCount))
	for _, juror := range evaluatedAssignmentCount {
		evaluatedMap[juror.JudgeID] = WorkLoadType(juror.Count)
	}
	transferableAssignmentCount := JurorV3{}
	stmt = Assignment.Select(Assignment.EvaluationID.Count().As("Count"), Assignment.JudgeID).
		Where(Assignment.RoundID.Eq(strategy.RoundId.String())).
		Where(Assignment.Score.IsNull()).
		Where(Assignment.EvaluatedAt.IsNull())
	if len(sourceRoleIds) > 0 {
		stmt = stmt.Where(Assignment.JudgeID.In(sourceRoleIds...))
	}
	// Get the total number of unevaluated elligible assignments
	err = stmt.Scan(&transferableAssignmentCount)
	if err != gorm.ErrRecordNotFound && err != nil {
		log.Println("Error: ", err)
		return
	}

	totalTransferableEvaluations := transferableAssignmentCount.Count
	// parentSpan.SetData("total_transferable_evaluations", totalTransferableEvaluations)
	log.Println("Total unevaluated evaluations: ", totalTransferableEvaluations)
	targetRoleCount := len(targetRoles)
	log.Println("Total target roles: ", targetRoleCount)
	log.Printf("Tranferable assignments: %+v", evaluatedMap)
	// parentSpan.SetData("target_role_count", targetRoleCount)
	// parentSpan.SetData("already_assigned_workflow_map", alreadyAssignedWorkflowMap)
	// for each of the target roles, we would be distributing the evaluations

	for i := range targetRoleCount {
		// childSpan := parentSpan.StartChild("assignment.distribution.round-robin-v2.target-role", func(s *sentry.Span) {
		// 	s.SetTag("target_role_index", fmt.Sprint(i))
		// 	s.SetTag("target_role_id", targetRoles[i].RoleID.String())
		// 	s.SetTag("target_role_user_id", targetRoles[i].UserID.String())
		// 	s.SetData("target_role_total_assigned", targetRoles[i].TotalAssigned)
		// })
		// defer childSpan.Finish()
		targetRole := targetRoles[i]
		targetJudgeId := targetRole.RoleID
		judgeUserId := targetRole.UserID
		log.Println("Target Judge ID: ", targetJudgeId)
		// childSpan.SetData("target_judge_id", targetJudgeId.String())
		// Get the average
		currentUnAssignedJuryCount := targetRoleCount - i
		// childSpan.SetData("current_unassigned_jury_count", currentUnAssignedJuryCount)
		workload := int(totalTransferableEvaluations) / currentUnAssignedJuryCount
		log.Printf("Average workload for target judge %s: %d", targetJudgeId, workload)
		// childSpan.SetData("average_workload", workload)

		//// DETERMINE THE WORKLOAD FOR THE JURY
		alreadyAssigned := alreadyAssignedWorkflowMap[targetJudgeId]
		nonTransferable := evaluatedMap[targetJudgeId]
		transferableCount := alreadyAssigned - nonTransferable
		// childSpan.SetData("already_assigned", alreadyAssigned)
		// childSpan.SetData("non_transferable", nonTransferable)
		// childSpan.SetData("transferable_count", transferableCount)
		log.Printf("Already assigned workload for target judge %s: %d", targetJudgeId, alreadyAssigned)
		log.Printf("Transferable workload for target judge %s: %d", targetJudgeId, transferableCount)
		log.Printf("Non-transferable for target judge %s: %d", targetJudgeId, nonTransferable)
		reassignment_count := int64(0)
		if workload > int(alreadyAssigned) {
			// If the workload is greater than the already assigned workload,
			// we need to make the difference to the workload
			totalTransferableEvaluations -= WorkLoadType(int(alreadyAssigned))
			workload = workload - int(alreadyAssigned)
			log.Printf("Workload for target judge %s is greater than already assigned workload, setting workload to %d", targetJudgeId, workload)
			if includeFromSourceOnly {
				reassignment_count, err = q1.Evaluation.WithContext(ctx).DistributeAssignmentsFromSelectedSource(targetJudgeId, judgeUserId, strategy.RoundId.String(), sourceRoleIds, strategy.TaskId, int(workload))
			} else {
				reassignment_count, err = q1.Evaluation.WithContext(ctx).DistributeAssignmentsIncludingUnassigned(targetJudgeId, judgeUserId, strategy.RoundId.String(), strategy.TaskId, int(workload))
			}
			if err != nil {
				log.Println("Error: ", err)
				return
			}
		} else if workload < int(alreadyAssigned) {
			// If the workload is less than the already assigned workload,
			// We need to lock the workload, no new assignments would be made
			log.Printf("Workload for target judge %s is less than already assigned workload :  %d", targetJudgeId, nonTransferable)
			if workload > int(nonTransferable) {
				// we should reduce the assignment count
				// upto the workload, lock upto the workload
				workload = int(nonTransferable) - workload
				log.Printf("Locking workload for target judge %s: %d", targetJudgeId, workload)
			} else {
				// If the workload is less than the non-transferable workload,
				// we do not need to lock aanything, we can just skip this role
				totalTransferableEvaluations -= WorkLoadType(workload)

				workload = 0
				log.Printf("Workload for target judge %s is less than non-transferable workload, skipping", targetJudgeId)

			}
		}

		///////////////////////////////////////
		// remove the current target role from the list
		// targetRoles = slices.Delete(targetRoles, i, i+1)
		if workload == 0 {
			// No workload for this judge, skip
			log.Printf("No workload for target judge %s, skipping", targetJudgeId)
			continue
		}

		if workload < 0 {
			// If the workload is negative, we need to lock the workload
			log.Printf("Locking negative workload for target judge %s: %d", targetJudgeId, workload)
			Evaluation := q1.Evaluation
			stmt := Evaluation.Where(Evaluation.JudgeID.Eq(targetJudgeId.String())).
				Where(Evaluation.RoundID.Eq(strategy.RoundId.String())).
				Where(Evaluation.Score.IsNull()).
				Where(Evaluation.EvaluatedAt.IsNull()).
				Where(Evaluation.Where(Evaluation.DistributionTaskID.Neq(strategy.TaskId.String())).Or(Evaluation.DistributionTaskID.IsNull())).
				Limit(-int(workload))
			res, err := stmt.UpdateColumn(Evaluation.DistributionTaskID, strategy.TaskId)
			if err != nil {
				log.Println("Error locking workload: ", err)
				task.Status = models.TaskStatusFailed
				return
			}
			reassignment_count = res.RowsAffected
			log.Printf("Rows affected by locking negative workload for target judge %s: %d", targetJudgeId, reassignment_count)
		}
		log.Printf("Reassigned %d evaluations to target judge %s", reassignment_count, targetJudgeId)
		totalTransferableEvaluations -= WorkLoadType(reassignment_count)
		task.SuccessCount += int(reassignment_count)
		log.Printf("Total unevaluated evaluations left: %d", totalTransferableEvaluations)
	}

	// for _, targetRole := range targetRoles {
	// 	targetJudgeId := targetRole.RoleID
	// 	workload, ok := newWorkload[targetJudgeId]
	// 	if !ok || workload == 0 {
	// 		log.Println("No workload found for target judge: ", targetJudgeId)
	// 		continue
	// 	}
	// 	judgeUserId := targetRole.UserID
	// 	log.Println("Judge ID: ", targetJudgeId)
	// 	log.Println("User ID: ", judgeUserId)
	// 	reassignment_count := int64(0)
	// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// 	defer cancel()
	// 	if includeFromSourceOnly {
	// 		reassignment_count, err = q1.Evaluation.WithContext(ctx).DistributeAssignmentsFromSelectedSource(targetJudgeId, judgeUserId, strategy.RoundId.String(), sourceRoleIds, strategy.TaskId, int(workload))
	// 	} else {
	// 		reassignment_count, err = q1.Evaluation.WithContext(ctx).DistributeAssignmentsIncludingUnassigned(targetJudgeId, judgeUserId, strategy.RoundId.String(), strategy.TaskId, int(workload))
	// 	}
	// 	if err != nil {
	// 		log.Println("Error: ", err)
	// 		return
	// 	}
	// 	log.Println("Reassigned: ", reassignment_count)
	// 	task.SuccessCount += int(reassignment_count)
	// }
	// ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	// defer cancel()
	// redistributeLastUnassigned, err := q.WithContext(ctx).Evaluation.DistributeTheLastRemainingEvaluations(strategy.TaskId, strategy.RoundId.String())
	// if err != nil {
	// 	log.Println("Error: ", err)
	// 	return
	// }
	// task.SuccessCount += int(redistributeLastUnassigned)
	// log.Println("Redistributed last unassigned evaluations: ", redistributeLastUnassigned)
	err = strategy.triggerStatisticsUpdateByRoundID(tx, round)
	if err != nil {
		task.Status = models.TaskStatusFailed
		log.Println("Error: ", err)
		tx.Rollback()
		return
	}
}

// func (strategy *RoundRobinDistributionStrategy) calculateWorkloadV2(conn *gorm.DB, roundId models.IDType, sourceRoleIds, targetJuryRoleIds []string) (fairNewWorkload map[models.IDType]WorkLoadType, err error) {
// 	// The Calculation is based on the following:
// 	// 1. Let the number of evaluated evaluation by Jury_i be E_i
// 	// 2. Total Number of evaluated Evaluations, E_total = E_1 + E_2 + ... + E_n
// 	// 3. Now, the number of reassignable evaluations (coming from th unevaluated assignments from source juries) is E_reassignable

// 	// This map would only have the juries that are in assignable jurors
// 	fairNewWorkload = map[models.IDType]WorkLoadType{}
// 	q := query.Use(conn)
// 	Assignment := q.Evaluation

// 	elligibleAssignmentCount := []JurorV3{}
// 	// Populate the existing workload
// 	alreadyAssignedWorkflowMap := map[models.IDType]WorkLoadType{}
// 	// By default it would be zero, because
// 	// we would be assigning the evaluations to the juries
// 	for _, jurId := range targetJuryRoleIds {
// 		alreadyAssignedWorkflowMap[models.IDType(jurId)] = 0
// 	}
// 	stmt := Assignment.Select(Assignment.EvaluationID.Count().As("Count"), Assignment.JudgeID).
// 		Where(Assignment.RoundID.Eq(roundId.String())).
// 		Where(Assignment.Score.IsNull()).
// 		Where(Assignment.EvaluatedAt.IsNull()).
// 		Group(Assignment.JudgeID)
// 	if len(sourceRoleIds) > 0 {
// 		stmt = stmt.Where(Assignment.JudgeID.In(sourceRoleIds...))
// 	}
// 	// Get the total number of unevaluated elligible assignments
// 	err = stmt.Scan(&elligibleAssignmentCount)
// 	if err != gorm.ErrRecordNotFound && err != nil {
// 		log.Println("Error: ", err)
// 		return nil, err
// 	}
// 	log.Printf("Eligible assignments count: %+v", elligibleAssignmentCount)
// 	totalElligibleUnevaluatedEvaluations := WorkLoadType(0)
// 	for _, juror := range elligibleAssignmentCount {
// 		if juror.Count > 0 {
// 			if _, ok := alreadyAssignedWorkflowMap[juror.JudgeID]; ok {
// 				continue // Skip if the juror is in the target jury list
// 			}
// 			totalElligibleUnevaluatedEvaluations += WorkLoadType(juror.Count)
// 		}
// 	}
// 	log.Println("Total unevaluated evaluations: ", totalElligibleUnevaluatedEvaluations)

// 	totalJuryCount := WorkLoadType(len(targetJuryRoleIds))
// 	log.Println("Total target jury count: ", totalJuryCount)
// 	if totalJuryCount == 0 {
// 		log.Println("No jury found")
// 		return nil, errors.New("noJuryFound")
// 	}

// 	alreadyAssignedWorkloads := MinimumWorkloadHeap{}
// 	err = Assignment.Select(Assignment.JudgeID, Assignment.EvaluationID.Count().As("Count")).
// 		// Where(Assignment.Score.IsNotNull()).Where(Assignment.EvaluatedAt.IsNotNull()).
// 		Where(Assignment.RoundID.Eq(roundId.String())).Where(Assignment.JudgeID.In(targetJuryRoleIds...)).
// 		Group(Assignment.JudgeID).Scan(&alreadyAssignedWorkloads)
// 	if err != nil {
// 		return nil, err
// 	}
// 	totalAssignedCount := WorkLoadType(0)
// 	for _, workload := range alreadyAssignedWorkloads {
// 		totalAssignedCount += workload.Count
// 		alreadyAssignedWorkflowMap[workload.JudgeID] = WorkLoadType(workload.Count)
// 	}
// 	// Already evaluated workloads
// 	alreadyEvaluatedWorkloads := MinimumWorkloadHeap{}
// 	err = Assignment.Select(Assignment.JudgeID, Assignment.EvaluationID.Count().As("Count")).
// 		Where(Assignment.RoundID.Eq(roundId.String())).
// 		Where(Assignment.Score.IsNotNull()).Where(Assignment.EvaluatedAt.IsNotNull()).
// 		Where(Assignment.JudgeID.In(targetJuryRoleIds...)).
// 		Group(Assignment.JudgeID).Scan(&alreadyEvaluatedWorkloads)
// 	if err != nil {
// 		log.Println("Error: ", err)
// 		return nil, err
// 	}
// 	alreadyEvaluatedWorkloadMap := map[models.IDType]WorkLoadType{}
// 	log.Printf("Already evaluated workloads: %+v", alreadyEvaluatedWorkloads)
// 	// Update the existing workload map with the already evaluated workloads
// 	for _, workload := range alreadyEvaluatedWorkloads {
// 		alreadyEvaluatedWorkloadMap[workload.JudgeID] = workload.Count
// 	}
// 	log.Println("Existing evaluated workload map: ", alreadyEvaluatedWorkloadMap)
// 	log.Printf("Existing assigned workload: %+v", alreadyAssignedWorkflowMap)
// 	totalEvaluationsTobeConsideredForFairDistribution := totalElligibleUnevaluatedEvaluations + totalAssignedCount
// 	// if totalAssignedCount == 0 {
// 	// 	log.Println("No elligible evaluations found")
// 	// 	return nil, errors.New("noElligibleEvaluationFound")
// 	// }
// 	averageWorkload := totalEvaluationsTobeConsideredForFairDistribution / totalJuryCount
// 	extraWorkload := totalEvaluationsTobeConsideredForFairDistribution % totalJuryCount
// 	log.Println("Existing workload: ", alreadyAssignedWorkloads)
// 	log.Println("Total evaluations: ", totalEvaluationsTobeConsideredForFairDistribution)
// 	log.Println("Total jury count: ", totalJuryCount)
// 	log.Println("Average workload: ", averageWorkload)
// 	log.Println("Extra workload: ", extraWorkload)
// 	// 10 13 8

// 	k := MinimumWorkloadHeap{}
// 	// Calculate the new workload for each juror
// 	for _, jurId := range targetJuryRoleIds {
// 		existingWorkload := alreadyAssignedWorkflowMap[models.IDType(jurId)]
// 		newWorkload := averageWorkload - existingWorkload

// 		fairNewWorkload[models.IDType(jurId)] = newWorkload
// 		k = append(k, JurorV3{
// 			JudgeID: models.IDType(jurId),
// 			Count:   newWorkload,
// 		})
// 	}
// 	if extraWorkload > 0 {
// 		sort.SliceStable(k, func(i, j int) bool {
// 			return k[i].Count < k[j].Count
// 		})
// 		for _, workload := range k {
// 			if extraWorkload == 0 {
// 				break
// 			}
// 			fairNewWorkload[workload.JudgeID]++
// 			extraWorkload--
// 		}
// 	}
// 	// ঋণাত্মক সংখ্যাগুলোকে একটু বদলে দিতে হচ্ছে
// 	// এর মানে কতগুলো বিয়োগ করতে হবে, এর বদলে কতগুলো রাখতে হবে, সেটা
// 	// তবে চিহ্ন একই থাকবে
// 	// এতে চিহ্ন বিয়োগ থাকায় বোঝা সুবিধা হবে, অন্যদিকে কতগুলো রাখতে হবে,
// 	// সেটার ফলে বণ্টন আইডি শুধুমাত্র সেগুলোতেই দেয়া হবে। ফলে বাকিগুলো অন্য কেউ সহজেই দখল করতে পারবে।
// 	for jurId, workload := range fairNewWorkload {
// 		evaluated := alreadyEvaluatedWorkloadMap[jurId]
// 		assigned := alreadyAssignedWorkflowMap[jurId]
// 		toBeKept := assigned - 2*evaluated + workload

// 		if workload < 0 {
// 			log.Printf("Juror %s: Workload: %d, Evaluated: %d, Assigned: %d, To be kept: %d", jurId, workload, evaluated, assigned, toBeKept)
// 			if toBeKept >= 0 {
// 				// If the workload is negative, it means we need to keep that many assignments
// 				fairNewWorkload[jurId] = -toBeKept
// 			} else {
// 				// If the workload is negative and toBeKept is negative, we can set it to zero
// 				fairNewWorkload[jurId] = 0
// 				log.Printf("Juror %s: Workload is negative and toBeKept is negative, setting to zero", jurId)
// 			}

// 		}
// 		if fairNewWorkload[jurId] == 0 {
// 			// If the workload is zero, we can remove it from the map
// 			fairNewWorkload[jurId] = -assigned
// 		}
// 	}

// 	return fairNewWorkload, nil
// }
