package services

import (
	"errors"
	"fmt"
	"log"
	"math/rand/v2"
	"nokib/campwiz/consts"
	"nokib/campwiz/models"
	"nokib/campwiz/query"
	"nokib/campwiz/repository"
	"nokib/campwiz/repository/cache"
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
	AmongJuriesUsername []models.WikimediaUsernameType `json:"juries"`
}
type ImportFromCommonsPayload struct {
	// Categories from which images will be fetched
	Categories []string `json:"categories" binding:"required"`
}
type ImportFromPreviousRoundPayload struct {
	// RoundID from which images will be fetched
	RoundID models.IDType `json:"roundId" binding:"required"`
	// Scores of the images to be fetched
	Scores []models.ScoreType `json:"scores" binding:"required"`
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
	campaign, err := campaign_repo.FindByID(tx.Preload("LatestRound"), request.CampaignID)
	if err != nil {
		tx.Rollback()
		return nil, errors.New("campaign not found")
	}
	log.Println("Campaign found with ID: ", campaign.CampaignID)
	// if campaign.Status != models.RoundStatusActive {
	// 	tx.Rollback()
	// 	return nil, errors.New("campaign is not active")
	// }
	previousRound := campaign.LatestRound
	if previousRound != nil {
		log.Println("Previous round found with ID: ", previousRound.RoundID)
		// check previous rounds
		if previousRound.Status != models.RoundStatusCompleted {
			tx.Rollback()
			return nil, errors.New("previous round is not completed yet")
		}
		request.RoundWritable.Serial = previousRound.Serial + 1
		if request.IsPublicJury && !previousRound.IsPublicJury {
			tx.Rollback()
			return nil, errors.New("public jury cannot be created after private jury on the same campaign")
		}
		// current request is public if all previous rounds are public and the current request is for public
		request.RoundWritable.IsPublicJury = request.IsPublicJury && previousRound.IsPublicJury

		request.RoundWritable.DependsOnRoundID = &previousRound.RoundID
	} else {
		log.Println("No previous round found")
		request.Serial = 1
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
	q := query.Use(tx)
	stmt := q.Campaign.Where(q.Campaign.CampaignID.Eq(campaign.CampaignID.String()))
	if !request.IsPublicJury {
		_, err = stmt.Update(q.Campaign.IsPublic, campaign.IsPublic)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}
	campaign.LatestRoundID = &round.RoundID
	_, err = stmt.Update(q.Campaign.LatestRoundID, campaign.LatestRoundID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	// create roles for the juries

	currentRoles, _, err := role_service.FetchChangeRoles(tx, models.RoleTypeJury, campaign.ProjectID, nil, &campaign.CampaignID, &round.RoundID, request.Juries)
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
func (b *RoundService) ImportFromPreviousRound(currentUserId models.IDType, targetRoundId models.IDType, filter *ImportFromPreviousRoundPayload) (*models.Task, error) {
	round_repo := repository.NewRoundRepository()
	task_repo := repository.NewTaskRepository()
	conn, close := repository.GetDB()
	defer close()
	tx := conn.Begin()
	targetRound, err := round_repo.FindByID(tx.Preload("Campaign"), targetRoundId)
	if err != nil {
		tx.Rollback()
		return nil, err
	} else if targetRound == nil {
		tx.Rollback()
		return nil, fmt.Errorf("round not found")
	} else if targetRound.Campaign == nil {
		tx.Rollback()
		return nil, fmt.Errorf("campaign not found")
	}
	sourceRound, err := round_repo.FindByID(tx, filter.RoundID)
	if err != nil {
		tx.Rollback()
		return nil, err
	} else if sourceRound == nil {
		tx.Rollback()
		return nil, fmt.Errorf("source round not found")
	}
	if sourceRound.Status != models.RoundStatusCompleted {
		tx.Rollback()
		return nil, errors.New("source round is not completed")
	}
	if targetRound.Status != models.RoundStatusPaused {
		tx.Rollback()
		return nil, errors.New("target round is not paused")
	}
	if targetRound.CampaignID != sourceRound.CampaignID {
		tx.Rollback()
		return nil, errors.New("source and target rounds are not from the same campaign")
	}
	if targetRound.ProjectID != sourceRound.ProjectID {
		tx.Rollback()
		return nil, errors.New("source and target rounds are not from the same project")
	}
	if targetRound.Campaign.Status != models.RoundStatusActive {
		tx.Rollback()
		return nil, errors.New("campaign is not active")
	}

	taskReq := &models.Task{
		TaskID:               idgenerator.GenerateID("t"),
		Type:                 models.TaskTypeImportFromPreviousRound,
		Status:               models.TaskStatusPending,
		AssociatedRoundID:    &targetRoundId,
		AssociatedUserID:     &targetRound.CreatedByID,
		CreatedByID:          targetRound.CreatedByID,
		AssociatedCampaignID: &targetRound.CampaignID,
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
	previousRoundSource := importsources.NewRoundCategoryListSource(filter.Scores[0], sourceRound.RoundID)
	batch_processor := importservice.NewImportTaskRunner(task.TaskID, previousRoundSource)
	go batch_processor.Run()
	return task, nil
}
func (b *RoundService) GetById(roundId models.IDType) (*models.Round, error) {
	round_repo := repository.NewRoundRepository()
	conn, close := repository.GetDB()
	defer close()
	return round_repo.FindByID(conn, roundId)
}
func (b *RoundService) DistributeTaskAmongExistingJuries(images []models.MediaResult) {
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
				ImageID:           images[i].PageID,
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

func (r *RoundService) UpdateRoundDetails(roundID models.IDType, req *RoundRequest, qry *models.SingleCampaaignFilter) (*models.Round, error) {
	round_repo := repository.NewRoundRepository()
	role_service := NewRoleService()
	conn, close := repository.GetDB()
	defer close()
	tx := conn.Begin()
	q := query.Use(tx)
	round, err := round_repo.FindByID(tx.Preload("DependsOnRound"), roundID)
	if err != nil {
		tx.Rollback()
		return nil, err
	} else if round == nil {
		tx.Rollback()
		return nil, fmt.Errorf("round not found")
	}
	if round.Status != models.RoundStatusPaused {
		tx.Rollback()
		return nil, errors.New("round is not paused")
	}
	previousRound := round.DependsOnRound
	if previousRound != nil {
		log.Println("Previous round found with ID: ", previousRound.RoundID)
		if previousRound.Status != models.RoundStatusCompleted {
			tx.Rollback()
			return nil, errors.New("previous round is not completed yet")
		}
		if req.RoundWritable.Serial != previousRound.Serial+1 {
			tx.Rollback()
			return nil, errors.New("serial must be one more than the previous round")
		}
		if req.IsPublicJury && !previousRound.IsPublicJury {
			tx.Rollback()
			return nil, errors.New("public jury cannot be created after private jury on the same campaign")
		}
		log.Println("Previous round found with ID: ", previousRound.RoundID, req.IsPublicJury)

	}
	res, err := q.Campaign.Where(q.Campaign.CampaignID.Eq(round.CampaignID.String())).Update(q.Campaign.IsPublic, req.IsPublicJury)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	if res.Error != nil {
		tx.Rollback()
		return nil, res.Error
	}
	round.RoundWritable = req.RoundWritable
	round, err = round_repo.Update(tx, round)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	if !req.IsPublicJury {
		juryType := models.RoleTypeJury
		filter := &models.RoleFilter{
			RoundID:    &roundID,
			CampaignID: &round.CampaignID,
			Type:       &juryType,
			ProjectID:  round.ProjectID,
		}
		addedRoles, removedRoleIDs, err := role_service.CalculateRoleDifference(tx, models.RoleTypeJury, filter, req.Juries)
		if err != nil {
			log.Println(err)
			tx.Rollback()
			return nil, err
		}
		if len(addedRoles) > 0 {
			res := tx.Save(addedRoles)
			if res.Error != nil {
				tx.Rollback()
				return nil, res.Error
			}
		}
		if len(removedRoleIDs) > 0 {
			r := []string{}
			for _, roleID := range removedRoleIDs {
				res := tx.Delete(&models.Role{RoleID: roleID})
				if res.Error != nil {
					tx.Rollback()
					return nil, res.Error
				}
				r = append(r, roleID.String())
			}
			// make all the unevaluated evaluations available for re-assignment to other juries
			_, err = q.Evaluation.Where(q.Evaluation.RoundID.Eq(roundID.String())).Where(q.Evaluation.JudgeID.In(r...)).
				Where(q.Evaluation.EvaluatedAt.IsNull()).Update(q.Evaluation.JudgeID, nil)
			if err != nil {
				tx.Rollback()
				return nil, err
			}
		}
	}
	if qry != nil {
		stmt := tx
		if qry.IncludeRoundRoles {
			stmt = stmt.Preload("Roles")
			if qry.IncludeRoundRolesUsers {
				stmt = stmt.Preload("Roles.User")
			}
			round, err = round_repo.FindByID(stmt, roundID)
			if err != nil {
				tx.Rollback()
				return nil, err
			}
			if qry.IncludeRoundRolesUsers {
				round.Jury = map[models.IDType]models.WikimediaUsernameType{}
				for _, role := range round.Roles {
					round.Jury[role.UserID] = role.User.Username
				}
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
	strategy := distributionstrategy.NewRoundRobinDistributionStrategy(task.TaskID, distributionReq.AmongJuriesUsername)
	runner := importservice.NewDistributionTaskRunner(task.TaskID, strategy)
	go runner.Run()
	return task, nil
}
func (r *RoundService) GetResultSummary(roundID models.IDType) (results []models.EvaluationResult, err error) {
	round_repo := repository.NewRoundRepository()
	conn, close := repository.GetDB()
	defer close()
	results, err = round_repo.GetResultSummary(conn, roundID)
	if err != nil {
		return nil, err
	}
	return
}
func (e *RoundService) GetNextUnevaluatedSubmissionForPublicJury(userID models.IDType, filter *models.EvaluationFilter) ([]*models.Submission, error) {
	conn, close := repository.GetDB()
	defer close()
	submission_repo := repository.NewSubmissionRepository()
	round_repo := repository.NewRoundRepository()
	role_repo := repository.NewRoleRepository()
	round, err := round_repo.FindByID(conn, filter.RoundID)
	if err != nil {
		return nil, err
	}
	if round == nil {
		return nil, errors.New("round not found")
	}
	var roleID *models.IDType
	role, err := role_repo.FindRoleByUserIDAndRoundID(conn, userID, filter.RoundID, models.RoleTypeJury)
	if err != nil {
		if err.Error() != "record not found" {
			return nil, err
		}
	}
	if role.RoleID != "" {
		roleID = &role.RoleID
	}
	filter.JuryRoleID = *roleID
	submissions, err := submission_repo.FindNextUnevaluatedSubmissionForPublicJury(conn, filter, round)
	if err != nil {
		return nil, err
	}
	return submissions, nil
}
func (e *RoundService) UpdateStatus(currenUserID models.IDType, roundID models.IDType, status models.RoundStatus) (*models.Round, error) {
	round_repo := repository.NewRoundRepository()
	role_repo := repository.NewRoleRepository()
	conn, close := repository.GetDB()
	defer close()
	tx := conn.Begin()

	round, err := round_repo.FindByID(tx.Preload("Campaign"), roundID)
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
	campaign := round.Campaign
	if campaign == nil {
		tx.Rollback()
		return nil, errors.New("campaign not found")
	}
	// if campaign.Status == models.
	// 	tx.Rollback()
	// 	return nil, errors.New("campaign is not active")
	// }
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
	if coordinatorRole.DeletedAt != nil {
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
	if status == models.RoundStatusActive || status == models.RoundStatusPaused || status == models.RoundStatusCompleted {
		qm := query.Use(tx)
		campaignStatus := models.RoundStatusActive
		if status == models.RoundStatusPaused {
			campaignStatus = models.RoundStatusPaused
		}
		Campaign := qm.Campaign
		_, err = Campaign.Where(Campaign.CampaignID.Eq(round.CampaignID.String())).Update(Campaign.Status, campaignStatus)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}
	round, err = round_repo.Update(tx, &models.Round{
		RoundID: roundID,
		Status:  status,
		RoundWritable: models.RoundWritable{
			IsPublicJury: round.IsPublicJury,
			IsOpen:       round.IsOpen,
		},
	})
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()
	return round, nil
}

func (e *RoundService) GetResults(currentUserID models.IDType, roundID models.IDType, q *models.SubmissionResultQuery) ([]models.SubmissionResult, error) {
	round_repo := repository.NewRoundRepository()
	conn, close := repository.GetDB()
	defer close()
	round, err := round_repo.FindByID(conn, roundID)
	if err != nil {
		return nil, err
	}
	if round == nil {
		return nil, errors.New("round not found")
	}
	if round.Status != models.RoundStatusCompleted {
		return nil, errors.New("round is not completed")
	}
	role_repo := repository.NewRoleRepository()
	coordinatorType := models.RoleTypeCoordinator
	role, err := role_repo.ListAllRoles(conn, &models.RoleFilter{
		UserID:     &currentUserID,
		CampaignID: &round.CampaignID,
		Type:       &coordinatorType,
	})
	if err != nil {
		return nil, err
	}
	if len(role) == 0 {
		return nil, errors.New("user is not a coordinator")
	}
	return round_repo.GetResults(conn, roundID, q)
}
func (e *RoundService) DeleteRound(sess *cache.Session, roundID models.IDType) error {
	round_repo := repository.NewRoundRepository()
	role_repo := repository.NewRoleRepository()
	conn, close := repository.GetDB()
	defer close()
	tx := conn.Begin()
	round, err := round_repo.FindByID(tx.Preload("Campaign"), roundID)
	if err != nil {
		tx.Rollback()
		return err
	}
	if round == nil {
		tx.Rollback()
		return errors.New("round not found")
	}
	if round.Status == models.RoundStatusActive {
		tx.Rollback()
		return errors.New("active round cannot be deleted")
	}
	campaign := round.Campaign
	if campaign == nil {
		tx.Rollback()
		return errors.New("campaign not found")
	}
	if !sess.Permission.HasPermission(consts.PermissionDeleteRound) && !CheckAccess(consts.PermissionDeleteRound, campaign, &sess.UserID, tx) {
		tx.Rollback()
		return errors.New("user does not have permission to delete the round")
	}
	err = role_repo.DeleteRolesByRoundID(tx, roundID)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = round_repo.Delete(tx, roundID)
	if err != nil {
		tx.Rollback()
		return err
	}

	campaign_repo := repository.NewCampaignRepository()
	err = campaign_repo.UpdateLatestRound(tx, campaign.CampaignID)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}
func (e *RoundService) AddMyselfAsJury(currentUserID models.IDType, roundID models.IDType) (*models.Role, error) {
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
	if round.Status != models.RoundStatusActive {
		tx.Rollback()
		return nil, errors.New("only active rounds can have juries")
	}
	role, err := role_repo.FindRoleByUserIDAndRoundID(tx, currentUserID, roundID, models.RoleTypeJury)
	if err != nil && err.Error() != "record not found" {
		tx.Rollback()
		return nil, err
	}
	if role.RoleID != "" {
		tx.Rollback()
		return nil, errors.New("user is already a jury")
	}
	role = &models.Role{
		RoleID:    idgenerator.GenerateID("r"),
		UserID:    currentUserID,
		RoundID:   &roundID,
		ProjectID: round.ProjectID,
		Type:      models.RoleTypeJury,
	}
	err = role_repo.CreateRole(tx, role)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()
	return role, nil
}
