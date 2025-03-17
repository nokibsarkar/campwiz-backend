package services

import (
	"errors"
	"fmt"
	"log"
	"math/rand/v2"
	"nokib/campwiz/models"
	"nokib/campwiz/repository"
	idgenerator "nokib/campwiz/services/idGenerator"
	importservice "nokib/campwiz/services/round/taskrunner"
	distributionstrategy "nokib/campwiz/services/round/taskrunner/distribution-strategy"
	importsources "nokib/campwiz/services/round/taskrunner/import-sources"
	"slices"

	"gorm.io/datatypes"
)

type RoundService struct {
}
type RoundRequest struct {
	CampaignID  models.IDType `json:"campaignId"`
	CreatedByID models.IDType `json:"-"`
	models.RoundWritable
	Juries []models.WikimediaUsernameType `json:"jury"`
}

type DistributionRequest struct {
	AmongJuries []models.IDType `json:"juries"`
}
type ImportFromCommonsPayload struct {
	// Categories from which images will be fetched
	Categories []string `json:"categories" binding:"required"`
}

type Jury struct {
	ID            uint64 `json:"id" gorm:"primaryKey"`
	totalAssigned int
}
type Evaluation struct {
	JuryID            uint64 `json:"juryId"`
	ImageID           uint64 `json:"imageId"`
	DistributionRound int    `json:"distributionRound"`
	Name              string `json:"name"`
}
type ByAssigned []*Jury

func (a ByAssigned) Len() int           { return len(a) }
func (a ByAssigned) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByAssigned) Less(i, j int) bool { return a[i].totalAssigned < a[j].totalAssigned }

func NewRoundService() *RoundService {
	return &RoundService{}
}
func (s *RoundService) CreateRound(request *RoundRequest) (*models.Round, error) {
	round_repo := repository.NewRoundRepository()
	campaign_repo := repository.NewCampaignRepository()
	role_service := NewRoleService()
	conn, close := repository.GetDB()
	defer close()
	tx := conn.Begin()
	campaign, err := campaign_repo.FindByID(tx, request.CampaignID)
	if err != nil {
		tx.Rollback()
		return nil, errors.New("campaign not found")
	}
	if campaign.Status != models.RoundStatusActive {
		tx.Rollback()
		return nil, errors.New("campaign is not active")
	}
	round := &models.Round{
		RoundID:       idgenerator.GenerateID("r"),
		CreatedByID:   request.CreatedByID,
		CampaignID:    campaign.CampaignID,
		RoundWritable: request.RoundWritable,
		Status:        models.RoundStatusPaused,
		ProjectID:     campaign.ProjectID,
	}
	round, err = round_repo.Create(tx, round)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	currentRoles, err := role_service.FetchChangeRoles(tx, models.RoleTypeJury, campaign.ProjectID, nil, &campaign.CampaignID, &round.RoundID, request.Juries)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	round.Roles = currentRoles
	tx.Commit()
	log.Println("Round created with ID: ", round.RoundID)
	log.Println("Roles: ", currentRoles)
	return round, nil
}
func (s *RoundService) ListAllRounds(filter *models.RoundFilter) ([]models.Round, error) {
	round_repo := repository.NewRoundRepository()
	conn, close := repository.GetDB()
	defer close()
	rounds, err := round_repo.FindAll(conn, filter)
	if err != nil {
		return nil, err
	}
	return rounds, nil
}

func (b *RoundService) ImportFromCommons(roundId models.IDType, categories []string) (*models.Task, error) {
	round_repo := repository.NewRoundRepository()
	task_repo := repository.NewTaskRepository()
	conn, close := repository.GetDB()
	defer close()
	tx := conn.Begin()
	round, err := round_repo.FindByID(tx, roundId)
	if err != nil {
		tx.Rollback()
		return nil, err
	} else if round == nil {
		tx.Rollback()
		return nil, fmt.Errorf("round not found")
	}
	taskReq := &models.Task{
		TaskID:               idgenerator.GenerateID("t"),
		Type:                 models.TaskTypeImportFromCommons,
		Status:               models.TaskStatusPending,
		AssociatedRoundID:    &roundId,
		AssociatedUserID:     &round.CreatedByID,
		CreatedByID:          round.CreatedByID,
		AssociatedCampaignID: &round.CampaignID,
		SuccessCount:         0,
		FailedCount:          0,
		FailedIds:            &datatypes.JSONType[map[string]string]{},
		RemainingCount:       0,
	}
	task, err := task_repo.Create(tx, taskReq)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()
	log.Println("Task created with ID: ", task.TaskID)
	commonsCategorySource := importsources.NewCommonsCategoryListSource(categories)
	batch_processor := importservice.NewImportTaskRunner(task.TaskID, commonsCategorySource)
	go batch_processor.Run()
	return task, nil
}
func (b *RoundService) GetById(roundId models.IDType) (*models.Round, error) {
	round_repo := repository.NewRoundRepository()
	conn, close := repository.GetDB()
	defer close()
	return round_repo.FindByID(conn, roundId)
}
func (b *RoundService) DistributeTaskAmongExistingJuries(images []models.ImageResult) {
	juries := []*Jury{}
	for i := 1; i <= 100; i++ {
		juries = append(juries, &Jury{ID: uint64(i), totalAssigned: rand.IntN(100)})
	}
	evaluations := []Evaluation{}
	imageCount, juryCount, evaluationCountRequired := len(images), len(juries), 10
	// datasetIndex := 0
	toleranceCount := 100
	if toleranceCount == 0 {
		log.Println("Tolerance count cannot be zero. Setting it to 1")
		toleranceCount = 1
	}
	sortedJuryByAssigned := ByAssigned(juries)
	slices.SortStableFunc(sortedJuryByAssigned, func(a, b *Jury) int {
		if a.totalAssigned < b.totalAssigned {
			return -1
		}
		if a.totalAssigned > b.totalAssigned {
			return 1
		}
		return 0
	})
	for i := 0; i < imageCount; i++ {
		// check if the last considered jury has been assigned the maximum number of images
		if evaluationCountRequired < juryCount && i%toleranceCount == 0 {
			firstUnassignedJuryIndex := evaluationCountRequired
			swapped := false
			for pivot := firstUnassignedJuryIndex; pivot > 0; pivot-- {
				for k := pivot; k < juryCount; k++ {
					if sortedJuryByAssigned[k-1].totalAssigned < sortedJuryByAssigned[k].totalAssigned {
						break
					}
					// swap the juries
					sortedJuryByAssigned[k-1], sortedJuryByAssigned[k] = sortedJuryByAssigned[k], sortedJuryByAssigned[k-1]
					swapped = true
				}
				if !swapped {
					break
				}
			}
		}
		for j := 0; j < evaluationCountRequired; j++ {
			evaluations = append(evaluations, Evaluation{
				JuryID:            sortedJuryByAssigned[j].ID,
				ImageID:           images[i].ID,
				Name:              images[i].Name,
				DistributionRound: j + 1,
			})
			sortedJuryByAssigned[j].totalAssigned++
		}
	}
	groupByJuryID := make(map[uint64][]Evaluation)
	for _, evaluation := range evaluations {
		groupByJuryID[evaluation.JuryID] = append(groupByJuryID[evaluation.JuryID], evaluation)
	}
	for j := range juryCount {
		log.Printf("Jury %d has %d images\n", sortedJuryByAssigned[j].ID, len(groupByJuryID[sortedJuryByAssigned[j].ID]))
	}
}

func (r *RoundService) UpdateRoundDetails(roundID models.IDType, req *RoundRequest) (*models.Round, error) {
	round_repo := repository.NewRoundRepository()
	role_service := NewRoleService()
	conn, close := repository.GetDB()
	defer close()
	tx := conn.Begin()
	round, err := round_repo.FindByID(tx, roundID)
	if err != nil {
		tx.Rollback()
		return nil, err
	} else if round == nil {
		tx.Rollback()
		return nil, fmt.Errorf("round not found")
	}
	// round.CampaignID = req.CampaignID // CampaignID is not updatable
	round.RoundWritable = req.RoundWritable
	round, err = round_repo.Update(tx, round)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	juryType := models.RoleTypeJury
	filter := &models.RoleFilter{
		RoundID:    &roundID,
		CampaignID: &round.CampaignID,
		Type:       &juryType,
		ProjectID:  round.ProjectID,
	}
	addedRoles, removedRoles, err := role_service.CalculateRoleDifference(tx, models.RoleTypeJury, filter, req.Juries)
	if err != nil {
		log.Println(err)
		tx.Rollback()
		return nil, err
	}
	if len(addedRoles) > 0 {
		res := tx.Create(addedRoles)
		if res.Error != nil {
			tx.Rollback()
			return nil, res.Error
		}
	}
	if len(removedRoles) > 0 {
		for _, roleID := range removedRoles {
			log.Println("Banning role: ", roleID)
			res := tx.Updates(&models.Role{RoleID: roleID, IsAllowed: false})
			if res.Error != nil {
				tx.Rollback()
				return nil, res.Error
			}
		}
	}
	tx.Commit()
	return round, nil
}
func (r *RoundService) DistributeEvaluations(currentUserID models.IDType, roundId models.IDType, distributionReq *DistributionRequest) (*models.Task, error) {
	round_repo := repository.NewRoundRepository()
	task_repo := repository.NewTaskRepository()
	conn, close := repository.GetDB()
	defer close()
	tx := conn.Begin()
	round, err := round_repo.FindByID(tx, roundId)
	if err != nil {
		tx.Rollback()
		return nil, err
	} else if round == nil {
		tx.Rollback()
		return nil, fmt.Errorf("round not found")
	}
	if round.Status == models.RoundStatusActive {
		tx.Rollback()
		return nil, errors.New("please pause the round before distributing evaluations")
	}
	taskReq := &models.Task{
		TaskID:               idgenerator.GenerateID("t"),
		Type:                 models.TaskTypeDistributeEvaluations,
		Status:               models.TaskStatusPending,
		AssociatedRoundID:    &roundId,
		AssociatedUserID:     &currentUserID,
		CreatedByID:          currentUserID,
		AssociatedCampaignID: &round.CampaignID,
		SuccessCount:         0,
		FailedCount:          0,
		FailedIds:            &datatypes.JSONType[map[string]string]{},
		RemainingCount:       0,
	}
	task, err := task_repo.Create(tx, taskReq)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()
	log.Println("Task created with ID: ", task.TaskID)
	strategy := distributionstrategy.NewRoundRobinDistributionStrategyV3(task.TaskID)
	runner := importservice.NewDistributionTaskRunner(task.TaskID, strategy)
	go runner.Run()
	return task, nil
}
func (r *RoundService) GetResults(roundID models.IDType) (results []models.EvaluationResult, err error) {
	round_repo := repository.NewRoundRepository()
	conn, close := repository.GetDB()
	defer close()
	results, err = round_repo.GetResultsV2(conn, roundID)
	if err != nil {
		return nil, err
	}
	return
}
func (e *RoundService) GetNextUnevaluatedSubmissionForPublicJury(userID models.IDType, roundID models.IDType) (*models.Submission, error) {
	conn, close := repository.GetDB()
	defer close()
	submission_repo := repository.NewSubmissionRepository()
	round_repo := repository.NewRoundRepository()
	role_repo := repository.NewRoleRepository()
	round, err := round_repo.FindByID(conn, roundID)
	if err != nil {
		return nil, err
	}
	if round == nil {
		return nil, errors.New("round not found")
	}
	var roleID *models.IDType
	role, err := role_repo.FindRoleByUserIDAndRoundID(conn, userID, roundID, models.RoleTypeJury)
	if err != nil {
		if err.Error() != "record not found" {
			return nil, err
		}
	}
	if role.RoleID != "" {
		roleID = &role.RoleID
	}
	submission, err := submission_repo.FindNextUnevaluatedSubmissionForPublicJury(conn, roleID, round)
	if err != nil {
		return nil, err
	}
	return submission, nil
}
func (e *RoundService) UpdateStatus(currenUserID models.IDType, roundID models.IDType, status models.RoundStatus) (*models.Round, error) {
	round_repo := repository.NewRoundRepository()
	role_repo := repository.NewRoleRepository()
	conn, close := repository.GetDB()
	defer close()
	tx := conn.Begin()

	round, err := round_repo.FindByID(tx, roundID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	if round == nil {
		tx.Rollback()
		return nil, errors.New("round not found")
	}
	if round.Status == status {
		tx.Rollback()
		return round, nil
	}
	coordinatorType := models.RoleTypeCoordinator
	coordinatorRoles, err := role_repo.ListAllRoles(tx, &models.RoleFilter{
		UserID:     &currenUserID,
		Type:       &coordinatorType,
		ProjectID:  round.ProjectID,
		CampaignID: &round.CampaignID,
	})
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	if len(coordinatorRoles) == 0 {
		tx.Rollback()
		return nil, errors.New("user is not a coordinator")
	}
	coordinatorRole := coordinatorRoles[0]
	if !coordinatorRole.IsAllowed {
		tx.Rollback()
		return nil, errors.New("user is not allowed to update the round")
	}
	// if a round is completed, it cannot be set to any other status
	switch round.Status {
	case models.RoundStatusCompleted:
		tx.Rollback()
		return nil, errors.New("round is completed")
	case models.RoundStatusActive:
		if status != models.RoundStatusCompleted && status != models.RoundStatusPaused {
			tx.Rollback()
			return nil, errors.New("round can only be set to completed from the backend")
		}
	case models.RoundStatusPaused:
		if status != models.RoundStatusActive && status != models.RoundStatusCompleted {
			tx.Rollback()
			return nil, errors.New("round can only be set to active or completed from paused")
		}
	case models.RoundStatusImporting:
		tx.Rollback()
		return nil, errors.New("nothing from frontend can change the status of a round from importing")
	case models.RoundStatusDistributing:
		return nil, errors.New("nothing from frontend can change the status of a round from distributing")
	case models.RoundStatusEvaluating:
		if status != models.RoundStatusPaused && status != models.RoundStatusCompleted {
			tx.Rollback()
			return nil, errors.New("round can only be set to paused or completed from evaluating")
		}
	case models.RoundStatusPending:
		if status != models.RoundStatusActive && status != models.RoundStatusRejected {
			tx.Rollback()
			return nil, errors.New("round can only be set to active or rejected from pending")
		}
	case models.RoundStatusScheduled:
		if status != models.RoundStatusActive && status != models.RoundStatusCancelled {
			tx.Rollback()
			return nil, errors.New("round can only be set to active or cancelled from scheduled")
		}
	case models.RoundStatusRejected:
		tx.Rollback()
		return nil, errors.New("round cannot be set to any other status from rejected")
	case models.RoundStatusCancelled:
		tx.Rollback()
		return nil, errors.New("round cannot be set to any other status from cancelled")
	}
	round, err = round_repo.Update(tx, &models.Round{
		RoundID: roundID,
		Status:  status,
	})
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()
	return round, nil
}
