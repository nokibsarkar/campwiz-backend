// File: routes/category.go
// Package routes provides the routing of a subtool where users can approve or reject categories of entries.
// Each Image would be considered a submission, and users can approve or reject categories for each submission.
package routes

import (
	"nokib/campwiz/models"
	"nokib/campwiz/models/types"
	"nokib/campwiz/repository/cache"
	"nokib/campwiz/services"

	"github.com/gin-gonic/gin"
)

type ConfirmSubmitCategory struct {
	// The Categories you want to set for the submission.
	Categories []string `json:"categories" binding:"required"`
	// The Summary to be added when your edit is submitted.
	// This is a required field.
	Summary string `json:"summary" binding:"required"`
}

// SubmitCategories godoc
// @Summary Submit categories for a submission
// @Description Submit categories for a submission. The Tool would edit on commons usng the token provided in the session. But the token would never be stored in the database.
// @Produce json
// @Success 200 {object} models.ResponseSingle[models.CategoryResponse]
// @Router /category/{submissionId} [post]
// @Param submissionId path string true "The submission ID"
// @Param Request body ConfirmSubmitCategory true "The categories to be set and the summary"
// @Tags Categories
// @Security ApiKeyAuth
// @Error 400 {object} models.ResponseError
// @Error 500 {object} models.ResponseError
// SubmitCategories handles the submission of categories for a given submission ID.
// It expects a JSON body containing an array of categories with writeable fields.
// It validates the submission ID and the categories, and then calls the CategoryService to process the submission.
// If successful, it returns a list of category responses.
// If there are any errors, it returns an appropriate error response.
func SubmitCategories(ctx *gin.Context, sess *cache.Session) {
	submissionID := ctx.Param("submissionId")
	if submissionID == "" {
		ctx.JSON(400, models.ResponseError{Detail: "Submission ID is required"})
		return
	}
	body := &ConfirmSubmitCategory{}
	if err := ctx.ShouldBindJSON(body); err != nil {
		ctx.JSON(400, models.ResponseError{Detail: "Invalid request: " + err.Error()})
		return
	}

	categoryService := services.NewCategoryService()
	resp, err := categoryService.SubmitCategories(ctx, types.SubmissionIDType(submissionID), body.Categories, body.Summary, sess.UserID)
	if err != nil {
		// Handle different types of errors with appropriate HTTP status codes
		switch err.Error() {
		case "authenticationRequired":
			ctx.JSON(401, models.ResponseError{
				Detail: "Authentication required. Please login with Wikimedia OAuth2 again.",
			})
		case "refreshTokenInvalid":
			ctx.JSON(401, models.ResponseError{
				Detail: "Authentication expired. Please login again.",
			})
		case "noCookieFound":
			ctx.JSON(401, models.ResponseError{
				Detail: "No authentication found. Please login with Wikimedia OAuth2.",
			})
		default:
			ctx.JSON(400, models.ResponseError{Detail: err.Error()})
		}
		return
	}

	ctx.JSON(200, models.ResponseSingle[*models.CategoryResponse]{Data: resp})
}

// SubmitCategoriesPreview godoc
// @Summary Submit categories preview for a submission
// @Description Submit categories preview for a submission
// @Produce json
// @Success 200 {object} models.ResponseSingle[models.CategoryResponse]
// @Router /category/{submissionId}/preview [post]
// @Param submissionId path string true "The submission ID"
// @Param categories body []string true "The categories to preview"
// @Tags Categories
// @Security ApiKeyAuth
// @Error 400 {object} models.ResponseError
// @Error 500 {object} models.ResponseError
// SubmitCategoriesPreview handles the preview submission of categories for a given submission ID.
// It expects a JSON body containing an array of categories with writeable fields.
// It validates the submission ID and the categories, and then calls the CategoryService to process the preview submission.
// If successful, it returns a single category response.
func SubmitCategoriesPreview(ctx *gin.Context, sess *cache.Session) {
	submissionID := ctx.Param("submissionId")
	if submissionID == "" {
		ctx.JSON(400, models.ResponseError{Detail: "Submission ID is required"})
		return
	}

	var categories []string
	if err := ctx.ShouldBindJSON(&categories); err != nil {
		ctx.JSON(400, models.ResponseError{Detail: "Invalid request: " + err.Error()})
		return
	}

	categoryService := services.NewCategoryService()
	resp, err := categoryService.SubmitCategoriesPreview(ctx, types.SubmissionIDType(submissionID), categories, sess.UserID)
	if err != nil {
		// Handle different types of errors with appropriate HTTP status codes
		switch err.Error() {
		case "authenticationRequired":
			ctx.JSON(401, models.ResponseError{
				Detail: "Authentication required. Please login with Wikimedia OAuth2 again.",
			})
		case "refreshTokenInvalid":
			ctx.JSON(401, models.ResponseError{
				Detail: "Authentication expired. Please login again.",
			})
		case "noCookieFound":
			ctx.JSON(401, models.ResponseError{
				Detail: "No authentication found. Please login with Wikimedia OAuth2.",
			})
		default:
			ctx.JSON(400, models.ResponseError{Detail: err.Error()})
		}
		return
	}

	ctx.JSON(200, models.ResponseSingle[*models.CategoryResponse]{Data: resp})
}

// GetSubmissionWithCategoryList godoc
// @Summary Get categories for a submission
// @Description Get categories for a submission
// @Produce json
// @Success 200 {object} models.ResponseSingle[models.SubmissionWithCategoryList]
// @Router /category/{submissionId} [get]
// @Param submissionId path string true "The submission ID"
// @Tags Categories
// @Security ApiKeyAuth
// @Error 400 {object} models.ResponseError
// @Error 500 {object} models.ResponseError
func GetSubmissionWithCategoryList(ctx *gin.Context, sess *cache.Session) {
	submissionID := ctx.Param("submissionId")
	if submissionID == "" {
		ctx.JSON(400, models.ResponseError{Detail: "Submission ID is required"})
		return
	}
	categoryService := services.NewCategoryService()
	resp, err := categoryService.GetCategoriesForSubmission(ctx, types.SubmissionIDType(submissionID))
	if err != nil {
		// Handle different types of errors with appropriate HTTP status codes
		switch err.Error() {
		case "authenticationRequired":
			ctx.JSON(401, models.ResponseError{
				Detail: "Authentication required. Please login with Wikimedia OAuth2 again.",
			})
		case "refreshTokenInvalid":
			ctx.JSON(401, models.ResponseError{
				Detail: "Authentication expired. Please login again.",
			})
		case "noCookieFound":
			ctx.JSON(401, models.ResponseError{
				Detail: "No authentication found. Please login with Wikimedia OAuth2.",
			})
		default:
			ctx.JSON(400, models.ResponseError{Detail: err.Error()})
		}
		return
	}

	ctx.JSON(200, models.ResponseSingle[*models.SubmissionWithCategoryList]{Data: resp})
}

// GetUncategorizedSubmissions godoc
// @Summary Get uncategorized submissions
// @Description Get a list of uncategorized submissions
// @Produce json
// @Success 200 {object} models.ResponseList[models.Submission]
// @Router /category/uncategorized/{campaignId} [get]
// @Param campaignId path string true "The campaign ID"
// @Tags Categories
// @Security ApiKeyAuth
// @Error 400 {object} models.ResponseError
// @Error 500 {object} models.ResponseError
func GetUncategorizedSubmissions(ctx *gin.Context, sess *cache.Session) {
	campaignID := ctx.Param("campaignId")
	if campaignID == "" {
		ctx.JSON(400, models.ResponseError{Detail: "Campaign ID is required"})
		return
	}
	categoryService := services.NewCategoryService()
	resp, err := categoryService.GetUncategorizedSubmissions(ctx, campaignID, 100, "")
	if err != nil {
		// Handle different types of errors with appropriate HTTP status codes
		switch err.Error() {
		case "authenticationRequired":
			ctx.JSON(401, models.ResponseError{
				Detail: "noAuth-Authentication required. Please login with Wikimedia OAuth2 again.",
			})
		case "refreshTokenInvalid":
			ctx.JSON(401, models.ResponseError{
				Detail: "noAuth-Authentication expired. Please login again.",
			})
		case "noCookieFound":
			ctx.JSON(401, models.ResponseError{
				Detail: "noAuth-No authentication found. Please login with Wikimedia OAuth2.",
			})
		default:
			ctx.JSON(400, models.ResponseError{Detail: err.Error()})
		}
		return
	}

	ctx.JSON(200, models.ResponseList[*models.Submission]{Data: resp})
}
