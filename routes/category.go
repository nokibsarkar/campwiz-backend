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

// SubmitCategories godoc
// @Summary Submit categories for a submission
// @Description Submit categories for a submission
// @Produce json
// @Success 200 {object} models.ResponseList[models.CategoryResponse]
// @Router /category/{submissionId} [post]
// @Param submissionId path string true "The submission ID"
// @Param categories body []string true "The categories to submit"
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

	var categories []string
	if err := ctx.ShouldBindJSON(&categories); err != nil {
		ctx.JSON(400, models.ResponseError{Detail: "Invalid request: " + err.Error()})
		return
	}

	categoryService := services.NewCategoryService()
	resp, err := categoryService.SubmitCategories(ctx, types.SubmissionIDType(submissionID), categories, sess.UserID)
	if err != nil {
		ctx.JSON(400, models.ResponseError{Detail: err.Error()})
		return
	}

	ctx.JSON(200, models.ResponseList[*models.CategoryResponse]{Data: resp})
}
