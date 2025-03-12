package services

import (
	"errors"
	"log"
	"nokib/campwiz/models"
	"nokib/campwiz/repository"
)

type EvaluationService struct{}

func NewEvaluationService() *EvaluationService {
	return &EvaluationService{}
}

type EvaluationRequest struct {
	VoteScore    *int          `json:"voteScore"`
	Comment      string        `json:"comment"`
	VotePassed   *bool         `json:"votePassed"`
	VotePosition *int          `json:"votePosition"`
	EvaluationID models.IDType `json:"evaluationId"`
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
	evaluationRequestMap := map[models.IDType]EvaluationRequest{}
	var currentRound *models.Round
	var campaign *models.Campaign
	for _, evaluationRequest := range evaluationRequests {
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
			currentRound, err = round_repo.FindByID(tx.Preload("Campaign"), submission.CurrentRoundID)
			if err != nil {
				tx.Rollback()
				return nil, err
			}
			campaign = currentRound.Campaign
			if campaign == nil {
				tx.Rollback()
				return nil, errors.New("campaign not found")
			}
			if false {
				tx.Rollback()
				return nil, errors.New("campaign is not active")
			}
		}
		if submission.CurrentRoundID != currentRound.RoundID {
			tx.Rollback()
			return nil, errors.New("all submissions must be from the same round")
		}
		log.Println("Evaluating evaluation", evaluation.Type)
		if evaluation.Type == models.EvaluationTypeBinary {
			log.Println("Binary evaluation")
		} else if evaluation.Type == models.EvaluationTypeRanking {
			log.Println("Ranking evaluation")
		} else if evaluation.Type == models.EvaluationTypeScore {
			log.Println("Score evaluation")
		}
		log.Println(evaluationRequest)
	}

	return evaluations, nil
}

func (e *EvaluationService) Evaluate(currentUserID models.IDType, evaluationID models.IDType, evaluationRequest *EvaluationRequest) (*models.Evaluation, error) {
	ev_repo := repository.NewEvaluationRepository()
	user_repo := repository.NewUserRepository()
	jury_repo := repository.NewRoleRepository()
	conn, close := repository.GetDB()
	defer close()
	tx := conn.Begin()
	evaluation, err := ev_repo.FindEvaluationByID(tx.Preload("Submission").Preload("Submission.CurrentRound").Preload("Submission.CurrentRound.Campaign"), evaluationID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	if evaluation == nil {
		tx.Rollback()
		return nil, errors.New("evaluation not found")
	}
	if evaluation.Type == models.EvaluationTypeBinary && evaluationRequest.VotePassed == nil {
		tx.Rollback()
		return nil, errors.New("votePassed is required for binary evaluation")
	} else if evaluation.Type == models.EvaluationTypeRanking && evaluationRequest.VotePosition == nil {
		tx.Rollback()
		return nil, errors.New("votePosition is required for positional evaluation")
	} else if evaluation.Type == models.EvaluationTypeScore && evaluationRequest.VoteScore == nil {
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
	round := submission.CurrentRound
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
