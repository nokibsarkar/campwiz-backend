package services

import (
	"errors"
	"log"
	"nokib/campwiz/models"
	"nokib/campwiz/models/types"
	"nokib/campwiz/query"
	"nokib/campwiz/repository"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CategoryService struct {
}

func NewCategoryService() *CategoryService {
	return &CategoryService{}
}
func (s *CategoryService) calculateCategoryDifference(conn *gorm.DB, ctx *gin.Context, submissionID types.SubmissionIDType, categories []string) (*models.CategoryResponse, *models.Submission, *models.Round, *models.Campaign, models.CategoryMap, error) {
	// First, it would validate all the provided data
	// first fetch the submission, round, campaign, and user

	if submissionID == "" {
		return nil, nil, nil, nil, nil, errors.New("submissionIDNotFound")
	}
	submission_repo := repository.NewSubmissionRepository()
	submission, err := submission_repo.FindSubmissionByID(conn.Preload("Round").Preload("Round.Campaign"), submissionID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil, nil, nil, nil, errors.New("submissionNotFound")
		}
		return nil, nil, nil, nil, nil, err
	}
	if submission == nil {
		return nil, nil, nil, nil, nil, errors.New("submissionNotFound")
	}
	round := submission.Round
	if round == nil {
		return nil, nil, nil, nil, nil, errors.New("roundNotFound")
	}
	campaign := round.Campaign
	if campaign == nil {
		return nil, nil, nil, nil, nil, errors.New("campaignNotFound")
	}
	pageID := submission.PageID
	commons_repo := repository.NewCommonsRepository()
	if pageID == 0 {
		return nil, nil, nil, nil, nil, errors.New("submissionPageIDNotFound")
	}
	latestRevision, err := commons_repo.GetLatestPageRevisionByPageID(ctx, pageID)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	if latestRevision == nil {
		return nil, nil, nil, nil, nil, errors.New("latestRevisionNotFound")
	}
	content := latestRevision.Slots.Main.Content
	tokens := content.SplitIntoTokens()
	categoryMap := content.GetCategoryMappingFromTokenList(tokens)
	log.Printf("Category map: %v", categoryMap)
	added := map[string]struct{}{}
	response := &models.CategoryResponse{
		PageTitle: latestRevision.Page.Title,
		Added:     []string{},
		Removed:   []string{},
	}
	for _, category := range categories {
		if category == "" {
			continue
		}
		if _, ok := (*categoryMap)[category]; ok {
			log.Printf("Category already exists: %s", category)
			continue
		}
		categoryMap, _ = categoryMap.Add(category)
		response.Added = append(response.Added, category)
		added[category] = struct{}{}
		log.Printf("Added category: %s", category)
	}
	for existingCategory := range *categoryMap {
		if _, ok := added[existingCategory]; !ok && strings.TrimSpace(existingCategory) != "" {
			response.Removed = append(response.Removed, existingCategory)
			log.Printf("Removed category: %s", existingCategory)
		}
	}
	return response, submission, round, campaign, *categoryMap, nil
}
func (s *CategoryService) SubmitCategoriesPreview(ctx *gin.Context, submissionID types.SubmissionIDType, categories []string, userID models.IDType) (*models.CategoryResponse, error) {
	// First, it would validate all the provided data
	// first fetch the submission, round, campaign, and user
	conn, close, err := repository.GetDB(ctx)
	if err != nil {
		return nil, err
	}
	defer close()

	response, _, _, _, _, err := s.calculateCategoryDifference(conn, ctx, submissionID, categories)
	if err != nil {
		return nil, err
	}
	response.Executed = false
	return response, nil
}
func (s *CategoryService) SubmitCategories(ctx *gin.Context, submissionID types.SubmissionIDType, categories []string, summary string, userID models.IDType) (*models.CategoryResponse, error) {
	// First, it would validate all the provided data
	// first fetch the submission, round, campaign, and user
	conn, close, err := repository.GetDB(ctx)
	if err != nil {
		return nil, err
	}
	defer close()

	response, submission, round, campaign, campaignMap, err := s.calculateCategoryDifference(conn, ctx, submissionID, categories)
	if err != nil {
		return nil, err
	}
	// fetch all the suggested categories from the campaign
	q := query.Use(conn)
	suggestedCategories, err := q.Category.Where(q.Category.SubmissionID.Eq(submission.SubmissionID.String())).Find()
	if err != nil {
		log.Println("Error fetching suggested categories:", err)
		return nil, err
	}
	log.Printf("Suggested categories: %v", suggestedCategories)
	// Now, we update which categories were approved or removed
	for _, addedCategory := range suggestedCategories {
		categoryName := addedCategory.CategoryName
		if _, ok := (campaignMap)[categoryName]; ok {
			// If the category is in the campaign map, it means it was approved
			log.Printf("Category %s was approved", categoryName)
			// response.Added = append(response.Added, categoryName)
			// We can also remove it from the campaign map to avoid duplicates
			delete(campaignMap, categoryName)
		} else {
			// If the category is not in the campaign map, it means it was removed
			log.Printf("Category %s was removed", categoryName)
			// response.Removed = append(response.Removed, categoryName)
		}
	}
	content := string(*campaignMap.GetContent())
	log.Printf("Final content after processing categories: %s", content)
	log.Printf("%+v %+v %+v %s", submission, round, campaign, campaignMap.GetContent())
	response.Executed = true
	return response, nil
}
