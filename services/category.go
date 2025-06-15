package services

import (
	"errors"
	"nokib/campwiz/models"
	"nokib/campwiz/models/types"
	"nokib/campwiz/repository"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CategoryService struct {
}

func NewCategoryService() *CategoryService {
	return &CategoryService{}
}

func (s *CategoryService) SubmitCategories(ctx *gin.Context, submissionID types.SubmissionIDType, categories []string, userID models.IDType) ([]*models.CategoryResponse, error) {
	// First, it would validate all the provided data
	// first fetch the submission, round, campaign, and user
	conn, close, err := repository.GetDB(ctx)
	if err != nil {
		return nil, err
	}
	defer close()
	submission_repo := repository.NewSubmissionRepository()
	submission, err := submission_repo.FindSubmissionByID(conn.Preload("Round").Preload("Round.Campaign"), submissionID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("submissionNotFound")
		}
		return nil, err
	}
	if submission == nil {
		return nil, errors.New("submissionNotFound")
	}
	round := submission.Round
	if round == nil {
		return nil, errors.New("roundNotFound")
	}
	campaign := round.Campaign
	if campaign == nil {
		return nil, errors.New("campaignNotFound")
	}
	var response []*models.CategoryResponse
	for _, category := range categories {
		response = append(response, &models.CategoryResponse{
			Added:   []string{category},
			Removed: []string{}, // Assuming no removals for simplicity
		})
	}
	return response, nil
}
