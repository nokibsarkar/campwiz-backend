package importsources

import (
	"context"
	"log"
	"maps"
	"nokib/campwiz/models"
	"nokib/campwiz/repository"
	"strings"
)

type CommonsCategoryListSource struct {
	Categories           []string
	currentCategoryIndex int
	lastPageID           uint64
	commons_repo         *repository.CommonsRepository
}

func (t *ImporterServer) ImportFromCommonsCategory(ctx context.Context, req *models.ImportFromCommonsCategoryRequest) (*models.ImportResponse, error) {
	log.Printf("ImportFromCommonsCategory %v", req)

	commonsCategoryLister := NewCommonsCategoryListSource(req.CommonsCategory)
	go t.importFrom(commonsCategoryLister, req.TaskId, req.RoundId)
	return &models.ImportResponse{
		TaskId:  req.TaskId,
		RoundId: req.RoundId,
	}, nil
}

// ImportImageResults imports images from commons categories
// For Each invocation it will import images from a single category
// If all categories are imported it will return nil
// If there are no images in the category it will return nil
// If there are images in the category it will return the images
// If there are failed images in the category it will return the reason as value of the map
func (c *CommonsCategoryListSource) ImportImageResults(currentRound *models.Round, failedImageReason *map[string]string) ([]models.MediaResult, *map[string]string) {
	if c.currentCategoryIndex < len(c.Categories) {
		category := c.Categories[c.currentCategoryIndex]
		campaign := currentRound.Campaign
		successMedia, currentfailedImages, lastPageID := c.commons_repo.GetImagesFromCommonsCategories2(category, c.lastPageID, currentRound, campaign.StartDate, campaign.EndDate)
		if lastPageID == 0 {
			c.currentCategoryIndex++
		}
		c.lastPageID = lastPageID
		maps.Copy(*failedImageReason, currentfailedImages)
		return successMedia, failedImageReason
	}

	return nil, failedImageReason
}

func NewCommonsCategoryListSource(categories []string) *CommonsCategoryListSource {
	normalizedCategories := []string{}
	for _, category := range categories {
		categoryWithUnderscores := strings.ReplaceAll(category, " ", "_")
		categoryWithUnderscores = strings.ReplaceAll(categoryWithUnderscores, "Category:", "")
		normalizedCategories = append(normalizedCategories, categoryWithUnderscores)
	}
	return &CommonsCategoryListSource{
		Categories:           normalizedCategories,
		currentCategoryIndex: 0,
		lastPageID:           0,
		commons_repo:         repository.NewCommonsRepository(),
	}
}
