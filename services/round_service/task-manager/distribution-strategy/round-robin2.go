package distributionstrategy

import (
	"errors"
	"log"
	"nokib/campwiz/models"
	"nokib/campwiz/query"
	"nokib/campwiz/repository"
	"sort"

	"gorm.io/gorm"
)

// Prevent Self evaluation SQL: update evaluations u1 join (select judge_id, evaluation_id, name from evaluations join submissions join roles on evaluations.submission_id = submissions.submission_id and evaluations.judge_id = roles.role_id  where submitted_by_id=roles.user_id and evaluations.round_id='r2eczdvrjl2ps') u2 using(evaluation_id) set u1.judge_id = (select role_id from roles where role_id <> u2.judge_id and round_id='r2eczdvrjl2ps' and role_id not in (select judge_id from evaluations where submission_id = u1.submission_id) order by rand() limit 1) where round_id='r2eczdvrjl2ps' and score is null;
// This method would distribute all the evaluations to the juries in round robin fashion
func (strategy *RoundRobinDistributionStrategy) AssignJuries2() {
	taskRepo := repository.NewTaskRepository()
	submission_repo := repository.NewSubmissionRepository()
	conn, close, err := repository.GetDB()
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
	createdCount, err := strategy.createMissingEvaluations(conn, round.Type, round, submissions)
	if err != nil {
		task.Status = models.TaskStatusFailed
		log.Println("Error: ", err)
		return
	}
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
		log.Println("Error: ", err)
		return
	}
	targetRoleIds := make([]string, len(targetRoles))
	for i, targetRole := range targetRoles {
		targetRoleIds[i] = targetRole.RoleID.String()
	}
	// Calculate the workload
	newWorkload, err := strategy.calculateWorkloadV2(conn, strategy.RoundId, targetRoleIds)
	if err != nil {
		log.Println("Error: ", err)
		return
	}
	log.Printf("New workload: %+v", newWorkload)
	// Get the total number of evaluations
	tx := conn.Begin()
	q1 := query.Use(tx)
	for _, targetRole := range targetRoles {
		targetJudgeId := targetRole.RoleID
		workload, ok := newWorkload[targetJudgeId]
		if !ok || workload == 0 {
			log.Println("No workload found for target judge: ", targetJudgeId)
			continue
		}
		judgeUserId := targetRole.UserID
		log.Println("Judge ID: ", targetJudgeId)
		log.Println("User ID: ", judgeUserId)
		reassignment_count, err := q1.Evaluation.DistributeAssignmentsFromSelectedSource(targetJudgeId, judgeUserId, strategy.RoundId.String(), sourceRoleIds, strategy.TaskId, int(workload))
		if err != nil {
			log.Println("Error: ", err)
			tx.Rollback()
			return
		}
		log.Println("Reassigned: ", reassignment_count)
		task.SuccessCount += int(reassignment_count)
	}
	err = strategy.triggerStatisticsUpdateByRoundID(tx, round)
	if err != nil {
		task.Status = models.TaskStatusFailed
		log.Println("Error: ", err)
		tx.Rollback()
		return
	}
	if err := tx.Commit().Error; err != nil {
		task.Status = models.TaskStatusFailed
		log.Println("Error: ", err)
		return
	}
}
func (strategy *RoundRobinDistributionStrategy) calculateWorkloadV2(conn *gorm.DB, roundId models.IDType, targetJuryRoleIds []string) (fairNewWorkload map[models.IDType]WorkLoadType, err error) {
	// The Calculation is based on the following:
	// 1. Let the number of evaluated evaluation by Jury_i be E_i
	// 2. Total Number of evaluated Evaluations, E_total = E_1 + E_2 + ... + E_n
	// 3. Now, the number of reassignable evaluations (coming from th unevaluated assignments from source juries) is E_reassignable

	// This map would only have the juries that are in assignable jurors
	fairNewWorkload = map[models.IDType]WorkLoadType{}
	q := query.Use(conn)
	Assignment := q.Evaluation
	Role := q.Role

	assignmentCount := JurorV3{}
	// Get the total number of evaluations
	err = Assignment.Select(Assignment.EvaluationID.Count().As("Count")).
		Where(Assignment.RoundID.Eq(roundId.String())).Scan(&assignmentCount)
	if err != nil {
		return nil, err
	}
	grandtotalEvaluations := assignmentCount.Count
	log.Println("Total evaluations: ", grandtotalEvaluations)
	if grandtotalEvaluations == 0 {
		log.Println("No evaluations found")
		return nil, errors.New("noEvaluationFound")
	}
	// get total number of jury
	juryC := JurorV3{}
	err = Role.Select(Role.RoleID.Count().As("Count")).
		Where(Role.RoundID.Eq(roundId.String())).Scan(&juryC)
	if err != nil {
		return nil, err
	}
	totalJuryCount := juryC.Count
	log.Println("Total jury count: ", totalJuryCount)
	if totalJuryCount == 0 {
		log.Println("No jury found")
		return nil, errors.New("noJuryFound")
	}
	// Populate the existing workload
	existingWorkLoadMap := map[models.IDType]WorkLoadType{}
	// By default it would be zero, because
	// if
	for _, jurId := range targetJuryRoleIds {
		existingWorkLoadMap[models.IDType(jurId)] = 0
	}
	alreadyEvaluatedWorkloads := MinimumWorkloadHeap{}
	err = Assignment.Select(Assignment.JudgeID, Assignment.EvaluationID.Count().As("Count")).
		Where(Assignment.Score.IsNotNull()).Where(Assignment.EvaluatedAt.IsNotNull()).
		Where(Assignment.RoundID.Eq(roundId.String())).Where(Assignment.JudgeID.In(targetJuryRoleIds...)).
		Group(Assignment.JudgeID).Scan(&alreadyEvaluatedWorkloads)
	if err != nil {
		return nil, err
	}
	// totalJuryCount := len(targetJuryRoleIds)
	totalEvaluatedCount := WorkLoadType(0)
	for _, workload := range alreadyEvaluatedWorkloads {
		totalEvaluatedCount += workload.Count
		existingWorkLoadMap[workload.JudgeID] = WorkLoadType(workload.Count)
	}
	log.Printf("Existing workload: %+v", existingWorkLoadMap)

	averageWorkload := WorkLoadType(grandtotalEvaluations / totalJuryCount)
	extraWorkload := grandtotalEvaluations % totalJuryCount
	log.Println("Existing workload: ", alreadyEvaluatedWorkloads)
	log.Println("Total evaluations: ", grandtotalEvaluations)
	log.Println("Total jury count: ", totalJuryCount)
	log.Println("Average workload: ", averageWorkload)
	log.Println("Extra workload: ", extraWorkload)

	k := MinimumWorkloadHeap{}
	// Calculate the new workload for each juror
	for _, jurId := range targetJuryRoleIds {
		existingWorkload := existingWorkLoadMap[models.IDType(jurId)]
		newWorkload := averageWorkload - existingWorkload
		if newWorkload > 0 {
			fairNewWorkload[models.IDType(jurId)] = newWorkload
			k = append(k, JurorV3{
				JudgeID: models.IDType(jurId),
				Count:   newWorkload,
			})
		}
	}
	if extraWorkload > 0 {
		sort.SliceStable(k, func(i, j int) bool {
			return k[i].Count < k[j].Count
		})
		for _, workload := range k {
			if extraWorkload == 0 {
				break
			}
			fairNewWorkload[workload.JudgeID]++
			extraWorkload--
		}
	}
	return fairNewWorkload, nil
}
