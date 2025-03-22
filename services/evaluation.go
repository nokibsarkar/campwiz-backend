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
	"gorm.io/gorm"
)

type EvaluationService struct{}

func NewEvaluationService() *EvaluationService {
	return &EvaluationService{}
}

type EvaluationRequest struct {
	Comment      string            `json:"comment"`
	Score        *models.ScoreType `json:"score"`
	EvaluationID models.IDType     `json:"evaluationId"`
}

func (e *EvaluationService) BulkEvaluate(currentUserID models.IDType, evaluationRequests []EvaluationRequest) ([]*models.Evaluation, error) {
	// ev_repo := repository.NewEvaluationRepository()
	user_repo := repository.NewUserRepository()
	round_repo := repository.NewRoundRepository()

	// jury_repo := repository.NewRoleRepository()
	conn, close := repository.GetDB()
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
	if err := e.triggerEvaluationScoreCount(tx, currentRound.RoundID, submissionIds); err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()
	return evaluations, nil
}

// This function would be used to trigger the evaluation score counting
func (e *EvaluationService) triggerEvaluationScoreCount(tx *gorm.DB, roundID models.IDType, submissionIds []types.SubmissionIDType) error {
	// This function would be used to trigger the evaluation score counting
	q := query.Use(tx)
	submission := q.Submission
	evaluation := q.Evaluation
	stringSubmissionIds := make([]string, len(submissionIds))
	for i, id := range submissionIds {
		stringSubmissionIds[i] = string(id)
	}
	averageScore := evaluation.Select(q.Evaluation.Score.Avg()).Where(evaluation.SubmissionID.EqCol(submission.SubmissionID))
	stmt, err := submission.WithContext(context.Background()).Where(submission.SubmissionID.In(stringSubmissionIds...)).Update(submission.Score, averageScore)
	if err != nil {
		return err
	}
	if stmt.Error != nil {
		return stmt.Error
	}
	evaluated_count := evaluation.Select(evaluation.EvaluationID.Count()).Where(evaluation.SubmissionID.EqCol(submission.SubmissionID)).Where(evaluation.Score.IsNotNull()).Where(evaluation.EvaluatedAt.IsNotNull())
	stmt, err = submission.WithContext(context.Background()).Where(submission.SubmissionID.In(stringSubmissionIds...)).Update(submission.EvaluationCount, evaluated_count)
	if err != nil {
		return err
	}
	if stmt.Error != nil {
		return stmt.Error
	}
	err = q.JuryStatistics.TriggerByRoundID(roundID.String())
	if err != nil {
		return err
	}
	log.Println("triggerEvaluationScoreCount", stmt.RowsAffected)
	return nil
}

func (e *EvaluationService) Evaluate(currentUserID models.IDType, evaluationID models.IDType, evaluationRequest *EvaluationRequest) (*models.Evaluation, error) {
	ev_repo := repository.NewEvaluationRepository()
	user_repo := repository.NewUserRepository()
	jury_repo := repository.NewRoleRepository()
	conn, close := repository.GetDB()
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
	if err := e.triggerEvaluationScoreCount(tx, round.RoundID, []types.SubmissionIDType{evaluation.SubmissionID}); err != nil {
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
	conn, close := repository.GetDB()
	defer close()
	tx := conn.Begin()
	submision, err := submission_repo.FindSubmissionByID(tx.Preload("Round"), submissionID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	if submision == nil {
		tx.Rollback()
		return nil, errors.New("submission not found")
	}
	// first check if the round exists
	round := submision.Round
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
		SubmissionID:  submision.SubmissionID,
		JudgeID:       &juryRole.RoleID,
		Score:         evaluationRequest.Score,
		Comment:       evaluationRequest.Comment,
		Type:          models.EvaluationTypeScore,
		EvaluatedAt:   &now,
		ParticipantID: submision.ParticipantID,
		RoundID:       submision.RoundID,
	}
	res := tx.Save(evaluation)
	if res.Error != nil {
		tx.Rollback()
		return nil, res.Error
	}
	// trigger submission score counting
	if err := e.triggerEvaluationScoreCount(tx, submision.RoundID, []types.SubmissionIDType{submision.SubmissionID}); err != nil {
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
	conn, close := repository.GetDB()
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
	conn, close := repository.GetDB()
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
