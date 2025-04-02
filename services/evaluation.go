package services

import (
	"errors"
	"fmt"
	"log"
	"nokib/campwiz/models"
	"nokib/campwiz/models/types"
	"nokib/campwiz/query"
	"nokib/campwiz/repository"
	idgenerator "nokib/campwiz/services/idGenerator"
	"time"

	"golang.org/x/net/context"
)

type EvaluationService struct{}

func NewEvaluationService() *EvaluationService {
	return &EvaluationService{}
}

type EvaluationRequest struct {
	Comment      string            `json:"comment"`
	Score        *models.ScoreType `json:"score"`
	EvaluationID models.IDType     `json:"evaluationId,omitempty"`
	SubmissionID models.IDType     `json:"submissionId,omitempty"`
}

func (e *EvaluationService) BulkEvaluate(currentUserID models.IDType, evaluationRequests []EvaluationRequest) ([]*models.Evaluation, error) {
	// ev_repo := repository.NewEvaluationRepository()
	user_repo := repository.NewUserRepository()
	round_repo := repository.NewRoundRepository()
	evaluation_repo := repository.NewEvaluationRepository()

	// jury_repo := repository.NewRoleRepository()
	conn, close, err := repository.GetDB()
	if err != nil {
		return nil, err
	}
	defer close()
	tx := conn.Begin()
	currentUser, err := user_repo.FindByID(tx, currentUserID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	if currentUser == nil {
		tx.Rollback()
		return nil, errors.New("user not found")
	}
	evaluations := []*models.Evaluation{}
	evaluationIDs := []models.IDType{}
	submissionIds := []types.SubmissionIDType{}
	evaluationRequestMap := map[models.IDType]EvaluationRequest{}
	var currentRound *models.Round
	var campaign *models.Campaign
	var juryRole *models.Role
	for _, evaluationRequest := range evaluationRequests {
		if evaluationRequest.Score == nil {
			tx.Rollback()
			return nil, fmt.Errorf("no score is given for evaluation %s", evaluationRequest.EvaluationID)
		}
		if *evaluationRequest.Score > models.MAXIMUM_EVALUATION_SCORE {
			tx.Rollback()
			return nil, fmt.Errorf("score is greater than %v", models.MAXIMUM_EVALUATION_SCORE)
		}

		evaluationRequestMap[evaluationRequest.EvaluationID] = evaluationRequest
		evaluationIDs = append(evaluationIDs, evaluationRequest.EvaluationID)
	}
	res := tx.Preload("Submission").Where("evaluation_id IN ?", evaluationIDs).Find(&evaluations)
	if res.Error != nil {
		tx.Rollback()
		return nil, res.Error
	}
	for _, evaluation := range evaluations {
		evaluationRequest, ok := evaluationRequestMap[evaluation.EvaluationID]
		if !ok {
			tx.Rollback()
			return nil, errors.New("evaluation not found in request")
		}
		submission := evaluation.Submission
		if submission.SubmittedByID == currentUser.UserID {
			tx.Rollback()
			return nil, errors.New("user can't evaluate his/her own submission")
		}
		if currentRound == nil {
			currentRound, err = round_repo.FindByID(tx.Preload("Campaign").Preload(("Roles")), submission.RoundID)
			if err != nil {
				tx.Rollback()
				return nil, err
			}
			campaign = currentRound.Campaign
			if campaign == nil {
				tx.Rollback()
				return nil, errors.New("campaign not found")
			}
			if campaign.Status != models.RoundStatusActive {
				tx.Rollback()
				return nil, errors.New("campaign is not active")
			}
			roles := currentRound.Roles
			for _, role := range roles {
				if role.UserID == currentUser.UserID {
					juryRole = &role
					break
				}
			}
			if juryRole == nil {
				tx.Rollback()
				return nil, errors.New("user is not a jury")
			}
			if juryRole.DeletedAt != nil {
				tx.Rollback()
				return nil, errors.New("user is not allowed to evaluate")
			}
		}
		if submission.RoundID != currentRound.RoundID {
			tx.Rollback()
			return nil, errors.New("all submissions must be from the same round")
		}
		if evaluationRequest.Score == nil {
			tx.Rollback()
			return nil, errors.New("no score is given")
		}
		if evaluation.Type == models.EvaluationTypeBinary {
			log.Println("Binary evaluation")
		} else if evaluation.Type == models.EvaluationTypeRanking {
			log.Println("Ranking evaluation")
		} else if evaluation.Type == models.EvaluationTypeScore {
			log.Println("Score evaluation")
		}
		now := time.Now().UTC()
		res = tx.Updates(&models.Evaluation{
			EvaluationID: evaluationRequest.EvaluationID,
			Score:        evaluationRequest.Score,
			Comment:      evaluationRequest.Comment,
			EvaluatedAt:  &now,
		})

		if res.Error != nil {
			tx.Rollback()
			return nil, res.Error
		}
		submissionIds = append(submissionIds, submission.SubmissionID)

	}
	if currentRound == nil {
		tx.Rollback()
		return nil, errors.New("no evaluations found")
	}
	// trigger submission score counting
	if err := evaluation_repo.TriggerEvaluationScoreCount(tx, submissionIds); err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()
	return evaluations, nil
}
func (e *EvaluationService) PublicBulkEvaluate(currentUserID models.IDType, evaluationRequests []EvaluationRequest) ([]*models.Evaluation, error) {
	// ev_repo := repository.NewEvaluationRepository()
	user_repo := repository.NewUserRepository()
	round_repo := repository.NewRoundRepository()
	evaluations := []*models.Evaluation{}

	// jury_repo := repository.NewRoleRepository()
	conn, close, err := repository.GetDB()
	if err != nil {
		return nil, err
	}
	defer close()
	tx := conn.Begin()
	currentUser, err := user_repo.FindByID(tx, currentUserID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	if currentUser == nil {
		tx.Rollback()
		return nil, errors.New("user not found")
	}
	existingEvaluationIDs := []string{}
	var currentRound *models.Round
	var campaign *models.Campaign
	var juryRole *models.Role
	newEvaluationRequests := []EvaluationRequest{}
	existingEvaluationRequestMap := map[models.IDType]EvaluationRequest{}
	newEvaluationSubmissionMap := map[types.SubmissionIDType]EvaluationRequest{}
	newEvalutionSubmissionIds := []string{}
	combinedSubmissionIds := []types.SubmissionIDType{}
	q := query.Use(tx)
	for _, evaluationRequest := range evaluationRequests {
		if evaluationRequest.Score == nil {
			tx.Rollback()
			return nil, fmt.Errorf("no score is given for evaluation %s", evaluationRequest.EvaluationID)
		}
		if *evaluationRequest.Score > models.MAXIMUM_EVALUATION_SCORE {
			tx.Rollback()
			return nil, fmt.Errorf("score is greater than %v", models.MAXIMUM_EVALUATION_SCORE)
		}
		if evaluationRequest.EvaluationID == "" {
			// No evaluation ID is given
			if evaluationRequest.SubmissionID == "" {
				// No submission ID is given
				tx.Rollback()
				return nil, errors.New("either evaluationId or submission Id must be provided")
			} else {
				// Submission ID is given
				evaluationRequest.EvaluationID = idgenerator.GenerateID("e")
				newEvaluationRequests = append(newEvaluationRequests, evaluationRequest)
				newEvaluationSubmissionMap[types.SubmissionIDType(evaluationRequest.SubmissionID)] = evaluationRequest
				newEvalutionSubmissionIds = append(newEvalutionSubmissionIds, evaluationRequest.SubmissionID.String())
			}
		} else {
			// evaluationId is given
			existingEvaluationIDs = append(existingEvaluationIDs, evaluationRequest.EvaluationID.String())

			existingEvaluationRequestMap[evaluationRequest.EvaluationID] = evaluationRequest
		}
	}
	Evaluation := q.Evaluation
	Submission := q.Submission

	now := time.Now().UTC()
	if len(newEvaluationRequests) > 0 {
		submissions, err := Submission.Where(Submission.SubmissionID.In(newEvalutionSubmissionIds...)).Find()
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		newEvaluations := []*models.Evaluation{}
		for _, submission := range submissions {
			if currentRound == nil {
				currentRound, err = round_repo.FindByID(tx.Preload("Roles").Preload("Campaign"), submission.RoundID)
				if err != nil {
					tx.Rollback()
					return nil, err
				}
				campaign = currentRound.Campaign
				if campaign == nil {
					tx.Rollback()
					return nil, errors.New("campaign not found")
				}

				roles := currentRound.Roles
				for _, role := range roles {
					if role.UserID == currentUser.UserID {
						juryRole = &role
						break
					}
				}
				if juryRole == nil {
					tx.Rollback()
					return nil, errors.New("user is not a jury")
				}
				if juryRole.DeletedAt != nil {
					tx.Rollback()
					return nil, errors.New("user is not allowed to evaluate")
				}
			}
			log.Println(newEvaluationSubmissionMap, submission.SubmissionID)
			eReq, ok := newEvaluationSubmissionMap[submission.SubmissionID]
			if !ok {
				tx.Rollback()
				return nil, errors.New("eReq not found")
			}
			ev := &models.Evaluation{
				EvaluationID:  eReq.EvaluationID,
				SubmissionID:  types.SubmissionIDType(eReq.SubmissionID),
				JudgeID:       &juryRole.RoleID,
				RoundID:       currentRound.RoundID,
				Type:          currentRound.Type,
				Score:         eReq.Score,
				Judge:         juryRole,
				Comment:       eReq.Comment,
				ParticipantID: submission.ParticipantID,
				EvaluatedAt:   &now,
			}
			newEvaluations = append(newEvaluations, ev)
			combinedSubmissionIds = append(combinedSubmissionIds, submission.SubmissionID)
		}
		ctx, closeCtx := context.WithTimeout(context.Background(), 5*time.Second)
		defer closeCtx()
		res := tx.WithContext(ctx).Create(newEvaluations)
		if res.Error != nil {
			return nil, res.Error
		}
		evaluations = append(evaluations, newEvaluations...)
	}
	if len(existingEvaluationIDs) > 0 {
		existingEvaluations := []*models.Evaluation{}
		err := Evaluation.Preload(Evaluation.Submission).Where(Evaluation.EvaluationID.In(existingEvaluationIDs...)).Scan(&existingEvaluations)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		for _, evaluation := range existingEvaluations {
			evaluationRequest, ok := existingEvaluationRequestMap[evaluation.EvaluationID]
			if !ok {
				tx.Rollback()
				return nil, errors.New("evaluation not found in request")
			}
			submission := evaluation.Submission
			if submission.SubmittedByID == currentUser.UserID {
				tx.Rollback()
				return nil, errors.New("user can't evaluate his/her own submission")
			}
			if currentRound == nil {
				currentRound, err = round_repo.FindByID(tx.Preload("Campaign").Preload(("Roles")), submission.RoundID)
				if err != nil {
					tx.Rollback()
					return nil, err
				}
				campaign = currentRound.Campaign
				if campaign == nil {
					tx.Rollback()
					return nil, errors.New("campaign not found")
				}
				if campaign.Status != models.RoundStatusActive {
					tx.Rollback()
					return nil, errors.New("campaign is not active")
				}
				roles := currentRound.Roles
				for _, role := range roles {
					if role.UserID == currentUser.UserID {
						juryRole = &role
						break
					}
				}
				if juryRole == nil {
					tx.Rollback()
					return nil, errors.New("user is not a jury")
				}
				if juryRole.DeletedAt != nil {
					tx.Rollback()
					return nil, errors.New("user is not allowed to evaluate")
				}
			}
			if submission.RoundID != currentRound.RoundID {
				tx.Rollback()
				return nil, errors.New("all submissions must be from the same round")
			}
			if evaluationRequest.Score == nil {
				tx.Rollback()
				return nil, errors.New("no score is given")
			}
			if evaluation.Type == models.EvaluationTypeBinary {
				log.Println("Binary evaluation")
			} else if evaluation.Type == models.EvaluationTypeRanking {
				log.Println("Ranking evaluation")
			} else if evaluation.Type == models.EvaluationTypeScore {
				log.Println("Score evaluation")
			}
			res := tx.Updates(&models.Evaluation{
				EvaluationID: evaluationRequest.EvaluationID,
				Score:        evaluationRequest.Score,
				Comment:      evaluationRequest.Comment,
				EvaluatedAt:  &now,
			})

			if res.Error != nil {
				tx.Rollback()
				return nil, res.Error
			}
			combinedSubmissionIds = append(combinedSubmissionIds, submission.SubmissionID)
			evaluations = append(evaluations, evaluation)
		}
	}
	if currentRound == nil {
		tx.Rollback()
		return nil, errors.New("no evaluations found")
	}
	evaluation_repo := repository.NewEvaluationRepository()
	// trigger submission score counting
	if err := evaluation_repo.TriggerEvaluationScoreCount(tx, combinedSubmissionIds); err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()
	return evaluations, nil
}

func (e *EvaluationService) Evaluate(currentUserID models.IDType, evaluationID models.IDType, evaluationRequest *EvaluationRequest) (*models.Evaluation, error) {
	ev_repo := repository.NewEvaluationRepository()
	user_repo := repository.NewUserRepository()
	jury_repo := repository.NewRoleRepository()
	conn, close, err := repository.GetDB()
	if err != nil {
		return nil, err
	}
	defer close()
	tx := conn.Begin()
	// first check if user
	evaluation, err := ev_repo.FindEvaluationByID(tx.Preload("Submission").Preload("Submission.Round").Preload("Submission.Round.Campaign"), evaluationID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	if evaluation == nil {
		tx.Rollback()
		return nil, errors.New("evaluation not found")
	}
	if evaluation.Type == models.EvaluationTypeBinary && evaluationRequest.Score == nil {
		tx.Rollback()
		return nil, errors.New("votePassed is required for binary evaluation")
	} else if evaluation.Type == models.EvaluationTypeRanking && evaluationRequest.Score == nil {
		tx.Rollback()
		return nil, errors.New("votePosition is required for positional evaluation")
	} else if evaluation.Type == models.EvaluationTypeScore && evaluationRequest.Score == nil {
		tx.Rollback()
		return nil, errors.New("voteScore is required for score evaluation")
	}

	submission := evaluation.Submission
	currentUser, err := user_repo.FindByID(tx, currentUserID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	if currentUser == nil {
		tx.Rollback()
		return nil, errors.New("user not found")
	}
	if submission.SubmittedByID == currentUser.UserID {
		tx.Rollback()
		return nil, errors.New("user can't evaluate his/her own submission")
	}
	round := submission.Round
	campaign := round.Campaign
	juries, err := jury_repo.ListAllRoles(tx, &models.RoleFilter{RoundID: &round.RoundID, CampaignID: &campaign.CampaignID})
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	juryMap := map[models.IDType]*models.Role{}
	for _, jury := range juries {
		juryMap[jury.RoleID] = &jury
	}
	// if _, ok := juryMap[currentUser.UserID]; !ok {
	// 	tx.Rollback()
	// 	return nil, errors.New("user is not a jury")
	// }
	if evaluation.Type == models.EvaluationTypeBinary {
		log.Println("Binary evaluation")
	} else if evaluation.Type == models.EvaluationTypeRanking {
		log.Println("Ranking evaluation")
	} else if evaluation.Type == models.EvaluationTypeScore {
		log.Println("Score evaluation")
	}
	if evaluationRequest.Comment != "" {
		evaluation.Comment = evaluationRequest.Comment
	}
	res := tx.Updates(&evaluation)
	if res.Error != nil {
		tx.Rollback()
		return nil, res.Error
	}
	// trigger submission score counting
	if err := ev_repo.TriggerEvaluationScoreCount(tx, []types.SubmissionIDType{evaluation.SubmissionID}); err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()
	return evaluation, nil
}
func (e *EvaluationService) PublicEvaluate(currentUserID models.IDType, submissionID types.SubmissionIDType, evaluationRequest *EvaluationRequest) (*models.Evaluation, error) {
	if evaluationRequest == nil {
		return nil, errors.New("evaluation request is required")
	}
	if evaluationRequest.Score == nil {
		return nil, errors.New("score is required")
	}
	if evaluationRequest.EvaluationID != "" {
		return e.Evaluate(currentUserID, evaluationRequest.EvaluationID, evaluationRequest)
	}
	submission_repo := repository.NewSubmissionRepository()
	jury_repo := repository.NewRoleRepository()
	conn, close, err := repository.GetDB()
	if err != nil {
		return nil, err
	}
	defer close()
	tx := conn.Begin()
	submission, err := submission_repo.FindSubmissionByID(tx.Preload("Round"), submissionID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	if submission == nil {
		tx.Rollback()
		return nil, errors.New("submission not found")
	}
	// first check if the round exists
	round := submission.Round
	// now round exists, check if jury exists
	juryRole, err := jury_repo.FindRoleByUserIDAndRoundID(tx, currentUserID, round.RoundID, models.RoleTypeJury)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	if juryRole == nil {
		// if not, create a jury
		juryRole = &models.Role{
			UserID:         currentUserID,
			RoundID:        &round.RoundID,
			Type:           models.RoleTypeJury,
			RoleID:         idgenerator.GenerateID("j"),
			ProjectID:      round.ProjectID,
			CampaignID:     &round.CampaignID,
			TotalAssigned:  1,
			TotalEvaluated: 1,
			TotalScore:     0,
		}
		res := tx.Create(juryRole)
		if res.Error != nil {
			tx.Rollback()
			return nil, res.Error
		}
	}
	now := time.Now().UTC()
	// then create evaluation with the user
	evaluation := &models.Evaluation{
		EvaluationID:  idgenerator.GenerateID("e"),
		SubmissionID:  submission.SubmissionID,
		JudgeID:       &juryRole.RoleID,
		Score:         evaluationRequest.Score,
		Comment:       evaluationRequest.Comment,
		Type:          models.EvaluationTypeScore,
		EvaluatedAt:   &now,
		ParticipantID: submission.ParticipantID,
		RoundID:       submission.RoundID,
	}
	res := tx.Save(evaluation)
	if res.Error != nil {
		tx.Rollback()
		return nil, res.Error
	}
	ev_repo := repository.NewEvaluationRepository()
	// trigger submission score counting
	if err := ev_repo.TriggerEvaluationScoreCount(tx, []types.SubmissionIDType{submission.SubmissionID}); err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()
	return evaluation, nil
}
func (e *EvaluationService) GetEvaluationById() {
}

func (e *EvaluationService) ListEvaluations(filter *models.EvaluationFilter) ([]models.Evaluation, error) {
	ev_repo := repository.NewEvaluationRepository()
	conn, close, err := repository.GetDB()
	if err != nil {
		return nil, err
	}
	defer close()
	return ev_repo.ListAllEvaluations(conn, filter)
}

// First get the roleID and round ID of the current user
// if not found, return error
// check if the role has judge permission
// if not, return error
// now list all the submissions where the roundID is the same as the current round and
// - if includeEvaluated is true,  evaluated_at is not null
// - if includeEvaluated is false, evaluated_at is null
// - if includeEvaluated is nil, no condition
// - if includeSkipped is true, include skipped submissions
// - if includeSkipped is false, exclude skipped submissions
func (e *EvaluationService) GetNextEvaluations(currenUserID models.IDType, filter *models.EvaluationFilter) ([]models.Evaluation, error) {
	ev_repo := repository.NewEvaluationRepository()
	roleRepo := repository.NewRoleRepository()
	conn, close, err := repository.GetDB()
	if err != nil {
		return nil, err
	}
	defer close()
	juryType := models.RoleTypeJury
	roles, err := roleRepo.ListAllRoles(conn, &models.RoleFilter{UserID: &currenUserID, RoundID: &filter.RoundID, Type: &juryType})
	if err != nil {
		return nil, err
	}
	if len(roles) == 0 {
		return nil, errors.New("user is not a jury")
	}
	juryRole := roles[0]
	if juryRole.DeletedAt != nil {
		return nil, errors.New("user is not allowed to evaluate")
	}
	juryRoleID := juryRole.RoleID
	filter.JuryRoleID = juryRoleID
	falsey := false
	filter.Evaluated = &falsey
	return ev_repo.ListAllEvaluations(conn, filter)
}
