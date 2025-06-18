package services

import (
	"errors"
	"log"
	"net/http"
	"nokib/campwiz/consts"
	"nokib/campwiz/models"
	"nokib/campwiz/models/types"
	"nokib/campwiz/query"
	"nokib/campwiz/repository"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
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
	commons_repo := repository.NewCommonsRepository(nil)
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
	log.Printf("Latest revision content:\n%s", content)
	tokens := content.SplitIntoTokens()
	categoryMap := content.GetCategoryMappingFromTokenList(tokens)
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
		canonicalName := models.ToCanonicalName(category)
		added[canonicalName] = struct{}{}
		if _, ok := (*categoryMap)[canonicalName]; ok {
			log.Printf("Category already exists: %s", category)
			continue
		}
		categoryMap, _ = categoryMap.Add(category)
		response.Added = append(response.Added, category)
		log.Printf("Added category: %s", category)
	}
	for canonicalName := range *categoryMap {
		if _, ok := added[canonicalName]; !ok && canonicalName != "" {
			response.Removed = append(response.Removed, models.FromCanonicalName(canonicalName))
			log.Printf("Removed category: %s", canonicalName)
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
func (s *CategoryService) handleAuthorization(ctx *gin.Context) (cl *http.Client, err error) {
	authConfig := consts.Config.Auth.GetOAuth2ReadWriteOauthConfig()
	oauth2Service := NewOAuth2Service(ctx, authConfig, consts.Config.Auth.Oauth2WriteAccess.RedirectPath)
	token := &oauth2.Token{
		TokenType: "Bearer",
	}
	// Check if the user is authorized to submit categories for this submission
	authorizationCookie, err := ctx.Request.Cookie(consts.ReadWriteAuthenticationCookieName)
	if err == nil && authorizationCookie != nil && authorizationCookie.Value != "" {
		token.AccessToken = authorizationCookie.Value
		// Set expiry from cookie if available, otherwise assume it might be expired
		if !authorizationCookie.Expires.IsZero() {
			token.Expiry = authorizationCookie.Expires
		} else {
			// If no expiry set, assume token might be expired to trigger refresh
			token.Expiry = time.Now().Add(-time.Hour)
		}
	}

	refreshCookie, err := ctx.Request.Cookie(consts.ReadWriteRefreshCookieName)
	if err == nil && refreshCookie != nil && refreshCookie.Value != "" {
		token.RefreshToken = refreshCookie.Value
	}

	if token.RefreshToken == "" && token.AccessToken == "" {
		return nil, errors.New("noCookieFound")
	}

	// Log token status for debugging
	log.Printf("Token status - AccessToken present: %v, RefreshToken present: %v, Expiry: %v",
		token.AccessToken != "", token.RefreshToken != "", token.Expiry)

	// Create HTTP client with token source that handles refresh automatically
	tokenSource := oauth2Service.Config.TokenSource(ctx, token)

	// Test if token is valid by getting a fresh token
	freshToken, err := tokenSource.Token()
	if err != nil {
		log.Printf("Token refresh failed: %v", err)

		// // Clear invalid cookies to force re-authentication
		// ctx.SetCookie(consts.ReadWriteAuthenticationCookieName, "", -1, "/", "", true, true)
		// ctx.SetCookie(consts.ReadWriteRefreshCookieName, "", -1, "/", "", true, true)

		// Return a more specific error for the frontend to handle
		if strings.Contains(err.Error(), "invalid_request") || strings.Contains(err.Error(), "invalid") {
			return nil, errors.New("authenticationRequired")
		}
		return nil, errors.New("refreshTokenInvalid")
	}

	// Update cookies with fresh token if they were refreshed
	if freshToken.AccessToken != token.AccessToken {
		log.Println("Token was refreshed, updating cookies")

		// Calculate expiry for cookie (OAuth2 tokens usually expire in 1 hour)
		expiresIn := int(time.Until(freshToken.Expiry).Seconds())
		// if expiresIn <= 0 {
		// 	expiresIn = 3600 // Default to 1 hour if expiry calculation fails
		// }

		// Update access token cookie
		ctx.SetSameSite(http.SameSiteNoneMode)
		ctx.SetCookie(consts.ReadWriteAuthenticationCookieName, freshToken.AccessToken, expiresIn, "/", "", true, true)

		// Update refresh token cookie if it changed
		if freshToken.RefreshToken != "" && freshToken.RefreshToken != token.RefreshToken {
			ctx.SetCookie(consts.ReadWriteRefreshCookieName, freshToken.RefreshToken, expiresIn+7*24*3600, "/", "", true, true)
		}
	}

	return oauth2Service.Config.Client(ctx, freshToken), nil

}
func (s *CategoryService) SubmitCategories(ctx *gin.Context, submissionID types.SubmissionIDType, categories []string, summary string, userID models.IDType) (*models.CategoryResponse, error) {
	httpClient, err := s.handleAuthorization(ctx)
	if err != nil {
		log.Println("Error handling authorization:", err)
		return nil, err
	}
	// First, it would validate all the provided data
	// first fetch the submission, round, campaign, and user
	conn, close, err := repository.GetDB(ctx)
	if err != nil {
		return nil, err
	}
	defer close()

	response, submission, _, _, campaignMap, err := s.calculateCategoryDifference(conn, ctx, submissionID, categories)
	if err != nil {
		return nil, err
	}
	commons_repo := repository.NewCommonsRepository(httpClient)
	// fetch all the suggested categories from the campaign
	q := query.Use(conn)
	suggestedCategories, err := q.Category.Where(q.Category.SubmissionID.Eq(submission.SubmissionID.String())).Find()
	if err != nil {
		log.Println("Error fetching suggested categories:", err)
		return nil, err
	}
	log.Printf("Suggested categories: %v", suggestedCategories)
	approvedList := []string{}
	rejectedList := []string{}
	// Now, we update which categories were approved or removed
	for _, addedCategory := range suggestedCategories {
		categoryName := addedCategory.CategoryName
		if _, ok := (campaignMap)[categoryName]; ok {
			// If the category is in the campaign map, it means it was approved
			approvedList = append(approvedList, categoryName)
			// We can also remove it from the campaign map to avoid duplicates
			delete(campaignMap, categoryName)
		} else {
			// If the category is not in the campaign map, it means it was removed
			rejectedList = append(rejectedList, categoryName)
		}
	}
	// log.Printf("CategoryMap after processing: %+v", campaignMap)
	content := string(*campaignMap.GetContent())
	if err := commons_repo.EditPageContent(ctx, submission.PageID, content, summary); err != nil {
		log.Println("Error editing page content:", err)
		return nil, err
	}
	if len(approvedList) > 0 {
		// Now Add the approved categories to the page
		res, err := q.Category.Where(q.Category.SubmissionID.Eq(submission.SubmissionID.String())).
			Where(q.Category.CategoryName.In(approvedList...)).
			Limit(len(approvedList)).
			Update(q.Category.AddedByID, userID.String())
		if err != nil {
			log.Println("Error updating categories:", err)
			return nil, err
		}
		if res.RowsAffected == 0 {
			log.Println("No categories were updated, possibly none matched the approved list")
		} else {
			log.Printf("Updated %d categories with user ID %s", res.RowsAffected, userID)
		}
	}
	if len(approvedList) > 0 {
		// Now we remove the rejected categories from the page
		res, err := q.Category.Where(q.Category.SubmissionID.Eq(submission.SubmissionID.String())).
			Where(q.Category.CategoryName.In(rejectedList...)).
			Limit(len(rejectedList)).
			Delete()
		if err != nil {
			log.Println("Error deleting categories:", err)
			return nil, err
		}
		if res.RowsAffected == 0 {
			log.Println("No categories were deleted, possibly none matched the rejected list")
		} else {
			log.Printf("Deleted %d categories from submission %s", res.RowsAffected, submission.SubmissionID)
		}
	}

	response.Executed = true
	return response, nil
}
func (s *CategoryService) GetCategoriesForSubmission(ctx *gin.Context, submissionID types.SubmissionIDType) (*models.SubmissionWithCategoryList, error) {
	_, err := s.handleAuthorization(ctx)
	if err != nil {
		log.Println("Error handling authorization:", err)
		return nil, err
	}
	// First, it would validate all the provided data
	// first fetch the submission, round, campaign, and user
	conn, close, err := repository.GetDB(ctx)
	if err != nil {
		return nil, err
	}
	defer close()
	// here all the removed categories would be calculated, these are actually categories that are present in the page
	response, submission, _, _, catMap, err := s.calculateCategoryDifference(conn, ctx, submissionID, []string{})
	if err != nil {
		return nil, err
	}
	// fetch all the suggested categories from the campaign
	q := query.Use(conn)
	suggestedCategories, err := q.Category.Where(q.Category.SubmissionID.Eq(submissionID.String())).Find()
	if err != nil {
		log.Println("Error fetching suggested categories:", err)
		return nil, err
	}
	result := &models.SubmissionWithCategoryList{
		Submission: *submission,
		Categories: []models.PageCategory{},
	}
	for _, category := range response.Removed {
		result.Categories = append(result.Categories, models.PageCategory{
			Name:  category,
			Fixed: true, // Already present in the page, so it is fixed
		})
	}
	for _, category := range suggestedCategories {
		if _, ok := catMap[models.ToCanonicalName(category.CategoryName)]; ok {
			// If the category is already present in the page, we don't add it again
			continue
		}
		result.Categories = append(result.Categories, models.PageCategory{
			Name:  category.CategoryName,
			Fixed: false, // Not present in the page, so it is not fixed
		})
	}
	return result, nil
}
