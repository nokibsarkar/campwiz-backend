package routes

import (
	"nokib/campwiz/consts"
	"nokib/campwiz/models"
	"nokib/campwiz/repository/cache"
	"nokib/campwiz/services"

	"github.com/gin-gonic/gin"
)

// List Evaluations godoc
// @Summary List all evaluations
// @Description get all evaluations
// @Produce  json
// @Success 200 {object} ResponseList[models.Evaluation]
// @Router /evaluation/ [get]
// @param EvaluationFilter query models.EvaluationFilter false "Filter the evaluations"
// @Tags Evaluation
// @Security ApiKeyAuth
// @Error 400 {object} ResponseError
func ListEvaluations(c *gin.Context, sess *cache.Session) {
	filter := &models.EvaluationFilter{}
	err := c.ShouldBindQuery(filter)
	if err != nil {
		c.JSON(400, ResponseError{Detail: "Invalid request : " + err.Error()})
		return
	}
	evaluation_service := services.NewEvaluationService()
	evaluations, err := evaluation_service.ListEvaluations(filter)
	if err != nil {
		c.JSON(400, ResponseError{Detail: "Error listing evaluations : " + err.Error()})
		return
	}
	c.JSON(200, ResponseList[models.Evaluation]{Data: evaluations})
}

// Update Evaluation godoc
// @Summary Update an evaluation
// @Description Update an evaluation
// @Produce  json
// @Success 200 {object} ResponseSingle[models.Evaluation]
// @Router /evaluation/{evaluationId} [post]
// @Tags Evaluation
// @Param evaluationId path string true "The evaluation ID"
// @Param evaluationRequest body services.EvaluationRequest true "The evaluation request"
// @Security ApiKeyAuth
// @Error 400 {object} ResponseError
// @Error 403 {object} ResponseError
// @Error 404 {object} ResponseError
func UpdateEvaluation(c *gin.Context, sess *cache.Session) {
	evaluationId := c.Param("evaluationId")
	if evaluationId == "" {
		c.JSON(400, ResponseError{Detail: "Invalid request : Evaluation ID is required"})
		return
	}
	requestedEvaluation := services.EvaluationRequest{}
	err := c.ShouldBindJSON(&requestedEvaluation)
	if err != nil {
		c.JSON(400, ResponseError{Detail: "Invalid request : " + err.Error()})
		return
	}
	evaluation_service := services.NewEvaluationService()
	evaluation, err := evaluation_service.Evaluate(sess.UserID, models.IDType(evaluationId), &requestedEvaluation)
	if err != nil {
		c.JSON(400, ResponseError{Detail: "Error updating evaluation : " + err.Error()})
		return
	}
	c.JSON(200, ResponseSingle[models.Evaluation]{Data: *evaluation})
}

// Bulk Evaluate godoc
// @Summary Bulk evaluate
// @Description Bulk evaluate
// @Produce  json
// @Success 200 {object} ResponseList[models.Evaluation]
// @Router /evaluation/ [post]
// @Tags Evaluation
// @Security ApiKeyAuth
// @Param evaluationRequest body []services.EvaluationRequest true "The evaluation request"
// @Error 400 {object} ResponseError
// @Error 403 {object} ResponseError
// @Error 404 {object} ResponseError
func BulkEvaluate(c *gin.Context, sess *cache.Session) {
	requestedEvaluations := []services.EvaluationRequest{}
	err := c.ShouldBindJSON(&requestedEvaluations)
	if err != nil {
		c.JSON(400, ResponseError{Detail: "Invalid request : " + err.Error()})
		return
	}
	evaluation_service := services.NewEvaluationService()
	evaluations, err := evaluation_service.BulkEvaluate(sess.UserID, requestedEvaluations)
	if err != nil {
		c.JSON(400, ResponseError{Detail: "Error updating evaluations : " + err.Error()})
		return
	}
	c.JSON(200, ResponseList[*models.Evaluation]{Data: evaluations})
}

func NewEvaluationRoutes(r *gin.RouterGroup) {
	route := r.Group("/evaluation")
	route.GET("/", WithSession(ListEvaluations))
	route.POST("/", WithPermission(consts.PermissionCreateCampaign, BulkEvaluate))
	route.POST("/:evaluationId", WithPermission(consts.PermissionCreateCampaign, UpdateEvaluation))
}
