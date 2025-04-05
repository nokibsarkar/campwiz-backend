package routes

import (
	"nokib/campwiz/models"
	"nokib/campwiz/models/types"
	"nokib/campwiz/services"

	"fmt"

	"github.com/gin-gonic/gin"
)

// ListAllSubmissions godoc
// @Summary List all submissions
// @Description get all submissions
// @Produce  json
// @Success 200 {object} models.ResponseList[models.Submission]
// @Router /submission/ [get]
// @param SubmissionListFilter query models.SubmissionListFilter false "Filter the submissions"
// @Tags Submission
// @Error 400 {object} models.ResponseError
func ListAllSubmissions(c *gin.Context) {
	filter := &models.SubmissionListFilter{}
	err := c.ShouldBindQuery(filter)
	if err != nil {
		c.JSON(400, models.ResponseError{
			Detail: "Invalid query",
		})
		return
	}
	if filter.Limit == 0 {
		filter.Limit = 10
	}
	submission_service := services.NewSubmissionService()
	submissions, err := submission_service.ListAllSubmissions(filter)

	if err != nil {
		c.JSON(400, models.ResponseError{
			Detail: "Error listing submissions",
		})
		return
	}
	continueToken := ""
	previousToken := ""
	if len(submissions) > 0 {
		continueToken = fmt.Sprint(submissions[len(submissions)-1].SubmissionID)
		previousToken = fmt.Sprint(submissions[0].SubmissionID)
	}
	c.JSON(200, models.ResponseList[models.Submission]{
		Data:          submissions,
		ContinueToken: continueToken,
		PreviousToken: previousToken,
	})

}
func CreateDraftSubmission(c *gin.Context) {
	// ...
}
func CreateLateDraftSubmission(c *gin.Context) {
	// ...
}
func GetDraftSubmission(c *gin.Context) {
	// ...
}

// GetSubmission godoc
// @Summary Get a submission
// @Description get a submission
// @Produce  json
// @Success 200 {object} models.Submission
// @Router /submission/{id} [get]
// @Param id path string true "Submission ID"
// @Tags Submission
// @Error 400 {object} models.ResponseError
// @Error 404 {object} models.ResponseError
func GetSubmission(c *gin.Context) {
	idString := c.Param("submissionId")
	submission_service := services.NewSubmissionService()
	submission, err := submission_service.GetSubmission(types.SubmissionIDType(idString))
	if err != nil {
		c.JSON(404, models.ResponseError{
			Detail: "Submission not found",
		})
		return
	}
	c.JSON(200, models.ResponseSingle[*models.Submission]{
		Data: submission,
	})
}
func DeleteSubmission(c *gin.Context) {
	// ...
}
func GetEvaluation(c *gin.Context) {
	// ...
}
func CreateSubmission(c *gin.Context) {
	// ...
}
func CreateLateSubmission(c *gin.Context) {
	// ...
}
func EvaluateSubmission(c *gin.Context) {
	// ...
}
