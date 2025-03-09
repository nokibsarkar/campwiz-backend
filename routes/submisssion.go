package routes

import (
	"nokib/campwiz/database"
	"nokib/campwiz/services"

	"github.com/gin-gonic/gin"
)

// ListAllSubmissions godoc
// @Summary List all submissions
// @Description get all submissions
// @Produce  json
// @Success 200 {object} ResponseList[database.Submission]
// @Router /submission/ [get]
// @param SubmissionListFilter query database.SubmissionListFilter false "Filter the submissions"
// @Tags Submission
// @Error 400 {object} ResponseError
func ListAllSubmissions(c *gin.Context) {
	filter := &database.SubmissionListFilter{}
	err := c.ShouldBindQuery(filter)
	if err != nil {
		c.JSON(400, ResponseError{
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
		c.JSON(400, ResponseError{
			Detail: "Error listing submissions",
		})
		return
	}
	continueToken := ""
	previousToken := ""
	if len(submissions) > 0 {
		continueToken = string(submissions[len(submissions)-1].SubmissionID)
		previousToken = string(submissions[0].SubmissionID)
	}
	c.JSON(200, ResponseList[database.Submission]{
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
// @Success 200 {object} database.Submission
// @Router /submission/{id} [get]
// @Param id path string true "Submission ID"
// @Tags Submission
// @Error 400 {object} ResponseError
// @Error 404 {object} ResponseError
func GetSubmission(c *gin.Context) {
	id := c.Param("submissionId")
	submission_service := services.NewSubmissionService()
	submission, err := submission_service.GetSubmission(database.IDType(id))
	if err != nil {
		c.JSON(404, ResponseError{
			Detail: "Submission not found",
		})
		return
	}
	c.JSON(200, ResponseSingle[*database.Submission]{
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

func NewSubmissionRoutes(parent *gin.RouterGroup) {
	r := parent.Group("/submission")
	r.GET("/", ListAllSubmissions)
	r.POST("/draft", CreateDraftSubmission)
	r.POST("/draft/late", CreateLateDraftSubmission)
	r.GET("/draft/:id", GetDraftSubmission)

	r.DELETE("/:id", DeleteSubmission)

	r.GET("/:submissionId", GetSubmission)
	// r.GET("/:submissionId/judge", GetEvaluation)
	r.POST("/", CreateSubmission)
	r.POST("/late", CreateLateSubmission)
	r.POST("/:submissionId/judge", EvaluateSubmission)
}
