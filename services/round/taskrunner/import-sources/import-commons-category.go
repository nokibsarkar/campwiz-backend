package importsources

import (
	"maps"
	"nokib/campwiz/models"
	"nokib/campwiz/repository"
	"strings"
)

type CommonsCategoryListSource struct {
	Categories           []string
	currentCategoryIndex int
	lastPageID           uint64
	round                *models.Round
	commons_repo         *repository.CommonsRepository
}

// ImportImageResults imports images from commons categories
// For Each invocation it will import images from a single category
// If all categories are imported it will return nil
// If there are no images in the category it will return nil
// If there are images in the category it will return the images
// If there are failed images in the category it will return the reason as value of the map
func (c *CommonsCategoryListSource) ImportImageResults(failedImageReason *map[string]string) ([]models.MediaResult, *map[string]string) {
	if c.currentCategoryIndex < len(c.Categories) {
		category := c.Categories[c.currentCategoryIndex]
		campaign := c.round.Campaign
		successMedia, currentfailedImages, lastPageID := c.commons_repo.GetImagesFromCommonsCategories2(category, c.lastPageID, c.round, campaign.StartDate, campaign.EndDate)
		if lastPageID == 0 {
			c.currentCategoryIndex++
		}
		c.lastPageID = lastPageID
		maps.Copy(*failedImageReason, currentfailedImages)
		return successMedia, failedImageReason
	}
	return nil, failedImageReason
}
func NewCommonsCategoryListSource(categories []string, round *models.Round) *CommonsCategoryListSource {
	ct := []string{}
	for _, category := range categories {
		kt := strings.Replace(category, " ", "_", -1)
		kt = strings.Replace(kt, "Category:", "", -1)
		ct = append(ct, kt)
	}
	return &CommonsCategoryListSource{
		Categories:           categories,
		currentCategoryIndex: 0,
		lastPageID:           0,
		commons_repo:         repository.NewCommonsRepository(),
		round:                round,
	}
}
