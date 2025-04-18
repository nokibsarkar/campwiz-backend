package distributionstrategy

import (
	"context"
	"log"
	"nokib/campwiz/models"
	"nokib/campwiz/query"
	"nokib/campwiz/repository"
	"nokib/campwiz/repository/cache"
	idgenerator "nokib/campwiz/services/idGenerator"
	"sort"

	"gorm.io/gorm"
)

type RoundRobinDistributionStrategy struct {
	TaskId           models.IDType
	AssignableJuries []models.WikimediaUsernameType
	RoundId          models.IDType
}
type DistributionResultV3 struct {
	TotalWorkLoad             WorkLoadType                   `json:"totalWorkLoad"`
	AvergaeWorkLoad           WorkLoadType                   `json:"averageWorkLoad"`
	TotalWorkloadDistribution map[models.IDType]WorkLoadType `json:"totalworkloadDistribution"`
}
type WorkLoadType int

// Juror represents a juror with workload and ID
type JurorV3 struct {
	JudgeID models.IDType
	Count   WorkLoadType
}

type TaskDistributionResultV3 struct {
	TotalWorkLoad        WorkLoadType   `json:"totalWorkLoad"`
	AvergaeWorkLoad      WorkLoadType   `json:"averageWorkLoad"`
	WorkloadDistribution []WorkLoadType `json:"workloadDistribution"`
}

// MinimumWorkloadHeap is a priority queue of Jurors
type MinimumWorkloadHeap []JurorV3

func (h MinimumWorkloadHeap) Len() int           { return len(h) }
func (h MinimumWorkloadHeap) Less(i, j int) bool { return h[i].Count < h[j].Count }
func (h MinimumWorkloadHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *MinimumWorkloadHeap) Push(x any) {
	*h = append(*h, x.(JurorV3))
}
func (h *MinimumWorkloadHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}
func (d *DistributorServer) DistributeWithRoundRobin(ctx context.Context, req *models.DistributeWithRoundRobinRequest) (*models.DistributeWithRoundRobinResponse, error) {
	// Implement the round robin distribution logic here
	taskId := req.TaskId
	assignableJuries := []models.WikimediaUsernameType{}
	for _, jury := range req.JuryUsernames {
		assignableJuries = append(assignableJuries, models.WikimediaUsernameType(jury))
	}
	strategy := &RoundRobinDistributionStrategy{
		TaskId:           models.IDType(taskId),
		AssignableJuries: assignableJuries,
		RoundId:          models.IDType(req.RoundId),
	}
	go strategy.AssignJuries()
	return &models.DistributeWithRoundRobinResponse{
		TaskId: req.TaskId,
	}, nil
}

func (strategy *RoundRobinDistributionStrategy) AssignJuries() {
	taskRepo := repository.NewTaskRepository()
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

	tx := conn.Begin()
	previousRoundStatus := round.Status
	round.Status = models.RoundStatusDistributing
	if err := tx.Save(round).Error; err != nil {
		log.Println("Error: ", err)
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			task.Status = models.TaskStatusFailed
			log.Println("Error: ", err)
		} else {
			if _, err := round_repo.Update(tx, &models.Round{
				RoundID: round.RoundID,
				Status:  previousRoundStatus,
			}); err != nil {
				log.Println("Error: ", err)
				tx.Rollback()
			} else {
				tx.Commit()
			}
		}

		if _, err := taskRepo.Update(conn, &models.Task{
			TaskID:       task.TaskID,
			Status:       task.Status,
			SuccessCount: task.SuccessCount,
			FailedCount:  task.FailedCount,
		}); err != nil {
			log.Println("Error: ", err)
			return
		}
	}()
	submission_repo := repository.NewSubmissionRepository()
	jury_repo := repository.NewRoleRepository()
	j := models.RoleTypeJury
	fetchedJuries, err := jury_repo.FindRolesByUsername(conn, strategy.AssignableJuries, &models.RoleFilter{
		Type:       &j,
		RoundID:    &round.RoundID,
		CampaignID: &round.CampaignID,
	})
	if err != nil {
		task.Status = models.TaskStatusFailed
		log.Println("Error: ", err)
		return
	}
	if len(fetchedJuries) == 0 {
		log.Println("No juries found")
		task.Status = models.TaskStatusFailed
		return
	}
	assignableJuryRoleIDs := []models.IDType{}
	for _, jury := range fetchedJuries {
		assignableJuryRoleIDs = append(assignableJuryRoleIDs, jury.RoleID)
	}
	submissions, err := submission_repo.ListAllSubmissions(conn.Where("assignment_count < ?", round.Quorum), &models.SubmissionListFilter{
		RoundID:    round.RoundID,
		CampaignID: round.CampaignID,
	})
	if err != nil {
		task.Status = models.TaskStatusFailed
		log.Println("Error: ", err)
		return
	}
	created, err := strategy.createMissingEvaluations(tx, round.Type, round, submissions)
	if err != nil {
		task.Status = models.TaskStatusFailed
		log.Println("Error: ", err)
		return
	}
	log.Printf("Created %d missing evaluations", created)
	taskDB, closeTaskDB := cache.GetTaskCacheDB(strategy.TaskId)
	defer closeTaskDB()
	totalEvaluations, err := strategy.importToCache(conn, taskDB, round)
	if err != nil {
		task.Status = models.TaskStatusFailed
		log.Println("Error: ", err)
		return
	}

	workload, err := strategy.calculateWorkloadQuota(taskDB, totalEvaluations, assignableJuryRoleIDs, fetchedJuries)
	if err != nil {
		task.Status = models.TaskStatusFailed
		log.Println("Error: ", err)
		return
	}
	log.Println("Workload: ", workload)
	_, err = strategy.distribute(taskDB, workload)
	if err != nil {
		task.Status = models.TaskStatusFailed
		log.Println("Error: ", err)
		return
	}
	task.SuccessCount, task.FailedCount, err = strategy.exportFromCache2MainDB(taskDB, tx)
	if err != nil {
		task.Status = models.TaskStatusFailed
		log.Println("Error: ", err)
		return
	}
	err = strategy.triggerStatisticsUpdateByRoundID(tx, round)
	if err != nil {
		task.Status = models.TaskStatusFailed
		log.Println("Error: ", err)
		return
	}
}
func (strategy *RoundRobinDistributionStrategy) createMissingEvaluations(tx *gorm.DB, evtype models.EvaluationType, round *models.Round, req []models.Submission) (int, error) {
	evaluations := []models.Evaluation{}
	log.Println("Creating missing evaluations", len(req))
	updatedPrepared := tx.Session(&gorm.Session{PrepareStmt: false})
	cnt := 0
	for _, submission := range req {
		rest := round.Quorum - submission.AssignmentCount
		for _ = range rest {
			evID := idgenerator.GenerateID("e")
			evaluation := models.Evaluation{
				SubmissionID:       submission.SubmissionID,
				EvaluationID:       evID,
				Type:               evtype,
				DistributionTaskID: strategy.TaskId,
				ParticipantID:      submission.ParticipantID,
				RoundID:            round.RoundID,
			}
			evaluations = append(evaluations, evaluation)
		}
		if res := updatedPrepared.Limit(1).Updates(&models.Submission{
			SubmissionID:    submission.SubmissionID,
			AssignmentCount: submission.AssignmentCount + rest,
		}); res.Error != nil {
			return 0, res.Error
		}
		cnt++
		// if cnt%163800 == 0 {
		// 	log.Println("Prepared statement count: ", cnt)
		// 	if res := updatedPrepared.SavePoint(fmt.Sprintf("sv1%d", cnt)); res.Error != nil {
		// 		// updatedPrepared.Rollback()
		// 		return 0, res.Error
		// 	}
		// 	cnt = 0
		// 	updatedPrepared = tx.Session(&gorm.Session{PrepareStmt: false})
		// }
	}
	// if res := updatedPrepared.Commit(); res.Error != nil {
	// 	// updatedPrepared.Rollback()
	// 	return 0, res.Error
	// }
	if len(evaluations) == 0 {
		return 0, nil
	}
	batchSze := 5000
	batch := len(evaluations)/batchSze + 1
	for i := range batch {
		start := i * batchSze
		end := min(start+batchSze, len(evaluations))
		ev := evaluations[start:end]
		res := tx.Create(&ev)
		if res.Error != nil {
			return 0, res.Error
		}
	}

	return len(evaluations), nil
}

func (strategy *RoundRobinDistributionStrategy) importToCache(tx *gorm.DB, taskCacheDB *gorm.DB, round *models.Round) (totalEvaluations int, err error) {
	var evaluations []*cache.Evaluation
	q := query.Use(tx)
	log.Println("Importing to cache")
	evs, err := (q.Evaluation.Where(q.Evaluation.RoundID.Eq(round.RoundID.String())).
		// Where(q.Evaluation.EvaluatedAt.IsNull()).
		// Where(q.Evaluation.Score.IsNull()).
		// Order(q.Evaluation.EvaluationID).
		Find())
	if err != nil {
		return 0, err
	}
	if len(evs) == 0 {
		return 0, nil
	}

	for _, evaluation := range evs {
		totalEvaluations++
		j := evaluation.JudgeID
		if evaluation.Score == nil {
			// Non evaluated means no judge
			j = nil
		}
		evaluations = append(evaluations, &cache.Evaluation{
			EvaluationID: evaluation.EvaluationID,
			SubmissionID: evaluation.SubmissionID,
			JudgeID:      j,
			Score:        evaluation.Score,
		})
	}
	batchSize := 1000
	batch := len(evaluations)/batchSize + 1
	for i := range batch {
		start := i * batchSize
		end := min(start+batchSize, len(evaluations))
		ev := evaluations[start:end]
		res := taskCacheDB.Create(&ev)
		if res.Error != nil {
			return 0, res.Error
		}
	}
	log.Println("Total evaluations: ", totalEvaluations)
	return totalEvaluations, nil
}

func (strategy *RoundRobinDistributionStrategy) distribute(cacheDB *gorm.DB, fairNewWorkload map[models.IDType]WorkLoadType) (int, error) {
	q := query.Use(cacheDB)
	Assignment := q.Evaluation
	for judgeID, workload := range fairNewWorkload {
		added, err := Assignment.DistributeAssigments(judgeID, int(workload))
		if err != nil {
			return 0, err
		}
		fairNewWorkload[judgeID] -= WorkLoadType(added)
		// if fairNewWorkload[judgeID] == 0 {
		// 	delete(fairNewWorkload, judgeID)
		// }
	}
	unassignedAssignments, err := Assignment.Where(Assignment.JudgeID.IsNull()).Find()
	if err != nil {
		return 0, err
	}
	log.Println("Unassigned evaluations: ", len(unassignedAssignments))
	for _, unassigned := range unassignedAssignments {
		ass, err := Assignment.SelectUnAssignedJudges(unassigned.SubmissionID, 1)
		if err != nil {
			return 0, nil
		}
		if len(ass) == 0 {
			log.Printf("Any replacement for %s cannot be found. %s has been assigned to all jury at least once", unassigned.EvaluationID, unassigned.SubmissionID)
			continue
		}
		selectedJury := ass[0]
		log.Println("Selected a replacement ", selectedJury)
		unassigned.JudgeID = selectedJury.JudgeID
		res, err := Assignment.Where(Assignment.EvaluationID.Eq(unassigned.EvaluationID.String())).Update(Assignment.JudgeID, selectedJury.JudgeID)
		if err != nil {
			log.Println(err)
			continue
		}
		if res.Error != nil {
			log.Println(res.Error)
		}
	}
	return 0, nil
}
func (strategy *RoundRobinDistributionStrategy) exportFromCache2MainDB(cache *gorm.DB, tx *gorm.DB) (successCount int, failedCount int, err error) {
	successCount = 0
	failedCount = 0
	lastId := ""
	limit := 1000
	q := query.Use(cache)
	Assignment := q.Evaluation
	for {
		assignments, err := Assignment.Where(Assignment.EvaluationID.Gt(lastId)).Limit(limit).Find()
		if err != nil {
			return 0, 0, err
		}
		if len(assignments) == 0 {
			return successCount, failedCount, err
		}
		log.Printf("Assignments %d\n", len(assignments))
		for _, assignment := range assignments {
			res := tx.Updates(&models.Evaluation{
				EvaluationID:       assignment.EvaluationID,
				JudgeID:            assignment.JudgeID,
				DistributionTaskID: strategy.TaskId,
			})
			if res.Error != nil {
				return 0, 0, res.Error
			}
			lastId = assignment.EvaluationID.String()
		}
	}
}
func (strategy *RoundRobinDistributionStrategy) calculateWorkloadQuota(cache *gorm.DB, totalEvaluations int, assignableJuries []models.IDType, juries []models.Role) (fairNewWorkload map[models.IDType]WorkLoadType, err error) {
	juryIDs := []string{}
	for _, jury := range juries {
		juryIDs = append(juryIDs, jury.RoleID.String())
	}
	fairNewWorkload = map[models.IDType]WorkLoadType{}
	existingWorkLoadMap := map[models.IDType]WorkLoadType{}
	for _, jurId := range juryIDs {
		existingWorkLoadMap[models.IDType(jurId)] = 0
	}
	q := query.Use(cache)
	Assignment := q.Evaluation

	workload := MinimumWorkloadHeap{}
	err = Assignment.Select(Assignment.JudgeID, Assignment.EvaluationID.Count().As("Count")).Where(Assignment.Score.IsNotNull()).Where(Assignment.JudgeID.In(juryIDs...)).Group(Assignment.JudgeID).Scan(&workload)
	if err != nil {
		return nil, err
	}
	totalJuryCount := len(juryIDs)
	totalAlreadyEvaluatedCount := WorkLoadType(0)
	for _, workload := range workload {
		totalAlreadyEvaluatedCount += workload.Count
		existingWorkLoadMap[workload.JudgeID] = WorkLoadType(workload.Count)
	}

	averageWorkload := WorkLoadType(totalEvaluations / totalJuryCount)
	extraWorkload := totalEvaluations % totalJuryCount
	log.Println("Exisiting workload: ", workload)
	log.Println("Total evaluations: ", totalEvaluations)
	log.Println("Total jury count: ", totalJuryCount)
	log.Println("Average workload: ", averageWorkload)
	log.Println("Extra workload: ", extraWorkload)

	k := MinimumWorkloadHeap{}
	// Calculate the new workload for each juror
	for _, jurId := range assignableJuries {
		existingWorkload := existingWorkLoadMap[jurId]
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
	return
}
func (strategy *RoundRobinDistributionStrategy) triggerStatisticsUpdateByRoundID(tx *gorm.DB, round *models.Round) error {
	q := query.Use(tx)
	submissionStatistics, err := q.SubmissionStatistics.FetchByRoundID(round.RoundID.String())
	if err != nil {
		return err
	}
	totalAssignments := 0
	totalEvaluations := 0
	totalEvaluatedSubmissions := 0
	for _, stat := range submissionStatistics {
		totalAssignments += stat.AssignmentCount
		totalEvaluations += stat.EvaluationCount
		if stat.EvaluationCount >= int(round.Quorum) {
			// The submission has been evaluated by at least quorum number of juries
			totalEvaluatedSubmissions++
		}
		// res := tx.Updates(&models.Submission{
		// 	SubmissionID:    stat.SubmissionID,
		// 	AssignmentCount: uint(stat.AssignmentCount),
		// 	EvaluationCount: uint(stat.EvaluationCount),
		// })
		// if res.Error != nil {
		// 	return res.Error
		// }
	}
	res := tx.Updates(&models.Round{
		RoundID:                   round.RoundID,
		TotalAssignments:          totalAssignments,
		TotalSubmissions:          len(submissionStatistics),
		TotalEvaluatedAssignments: totalEvaluations,
		TotalEvaluatedSubmissions: totalEvaluatedSubmissions,
		LatestDistributionTaskID:  &strategy.TaskId,
	})
	if res.Error != nil {
		return res.Error
	}
	jMap, err := q.JuryStatistics.GetJuryStatistics(round.RoundID.String())
	if err != nil {
		return err
	}
	log.Println(jMap)
	for _, stat := range jMap {
		log.Println("Updating jury statistics: ", stat)
		res := tx.Unscoped().Where(&models.Role{
			RoleID: stat.JudgeID,
		}).Updates(&models.Role{
			TotalAssigned:  stat.TotalAssigned,
			TotalEvaluated: stat.TotalEvaluated,
		})
		if res.Error != nil {
			return res.Error
		}
	}
	return nil

}
