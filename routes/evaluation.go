package routes

import (
	"log"
	"nokib/campwiz/models"
	"nokib/campwiz/models/types"
	"nokib/campwiz/repository/cache"
	"nokib/campwiz/services"

	"github.com/gin-gonic/gin"
)

// List Evaluations godoc
// @Summary List all evaluations
// @Description get all evaluations
// @Produce  json
// @Success 200 {object} models.ResponseList[models.Evaluation]
// @Router /evaluation/ [get]
// @param EvaluationFilter query models.EvaluationFilter false "Filter the evaluations"
// @Tags Evaluation
// @Security ApiKeyAuth
// @Error 400 {object} models.ResponseError
func ListEvaluations(c *gin.Context, sess *cache.Session) {
	filter := &models.EvaluationFilter{}
	err := c.ShouldBindQuery(filter)
	if err != nil {
		c.JSON(400, models.ResponseError{Detail: "Invalid request : " + err.Error()})
		return
	}
	log.Printf("Requested evaluations : %+v", filter)
	evaluation_service := services.NewEvaluationService()
	evaluations, totalAssigned, totalEvaluated, err := evaluation_service.GetNextEvaluations(sess.UserID, filter)
	if err != nil {
		c.JSON(400, models.ResponseError{Detail: "Error listing evaluations : " + err.Error()})
		return
	}
	previousToken := ""
	nextToken := ""
	if len(evaluations) > 0 {
		previousToken = evaluations[0].EvaluationID.String()
		nextToken = evaluations[len(evaluations)-1].EvaluationID.String()
	}

	c.JSON(200, models.EvaluationListResponseWithCurrentStats{
		ResponseList:         models.ResponseList[*models.Evaluation]{Data: evaluations, ContinueToken: nextToken, PreviousToken: previousToken},
		TotalEvaluatedCount:  totalEvaluated,
		TotalAssignmentCount: totalAssigned,
	})
}

// Update Evaluation godoc
// @Summary Update an evaluation
// @Description Update an evaluation
// @Produce  json
// @Success 200 {object} models.ResponseSingle[models.Evaluation]
// @Router /evaluation/{evaluationId} [post]
// @Tags Evaluation
// @Param evaluationId path string true "The evaluation ID"
// @Param evaluationRequest body services.EvaluationRequest true "The evaluation request"
// @Security ApiKeyAuth
// @Error 400 {object} models.ResponseError
// @Error 403 {object} models.ResponseError
// @Error 404 {object} models.ResponseError
func UpdateEvaluation(c *gin.Context, sess *cache.Session) {
	evaluationId := c.Param("evaluationId")
	if evaluationId == "" {
		c.JSON(400, models.ResponseError{Detail: "Invalid request : Evaluation ID is required"})
		return
	}
	requestedEvaluation := services.EvaluationRequest{}
	err := c.ShouldBindJSON(&requestedEvaluation)
	if err != nil {
		c.JSON(400, models.ResponseError{Detail: "Invalid request : " + err.Error()})
		return
	}
	log.Println("Requested evaluation : ", requestedEvaluation)
	evaluation_service := services.NewEvaluationService()
	evaluation, err := evaluation_service.Evaluate(sess.UserID, models.IDType(evaluationId), &requestedEvaluation)
	if err != nil {
		c.JSON(400, models.ResponseError{Detail: "Error updating evaluation : " + err.Error()})
		return
	}
	c.JSON(200, models.ResponseSingle[models.Evaluation]{Data: *evaluation})
}

// Bulk Evaluate godoc
// @Summary Bulk evaluate
// @Description Bulk evaluate
// @Produce  json
// @Success 200 {object} models.ResponseList[models.Evaluation]
// @Router /evaluation/ [post]
// @Tags Evaluation
// @Security ApiKeyAuth
// @Param evaluationRequest body []services.EvaluationRequest true "The evaluation request"
// @Error 400 {object} models.ResponseError
// @Error 403 {object} models.ResponseError
// @Error 404 {object} models.ResponseError
func BulkEvaluate(c *gin.Context, sess *cache.Session) {
	requestedEvaluations := []services.EvaluationRequest{}
	err := c.ShouldBindJSON(&requestedEvaluations)
	if err != nil {
		c.JSON(400, models.ResponseError{Detail: "Invalid request : " + err.Error()})
		return
	}
	log.Println("Requested evaluations : ", requestedEvaluations)
	evaluation_service := services.NewEvaluationService()
	result, err := evaluation_service.BulkEvaluate(sess.UserID, requestedEvaluations)
	if err != nil {
		c.JSON(400, models.ResponseError{Detail: "Error updating evaluations : " + err.Error()})
		return
	}
	c.JSON(200, result)
}

// SubmitNewPublicEvaluation godoc
// @Summary Submit a new public evaluation
// @Description Submit a new public evaluation
// @Produce  json
// @Success 200 {object} models.ResponseSingle[models.Evaluation]
// @Router /evaluation/public/{roundId}/{submissionId} [post]
// @Tags Evaluation
// @Param roundId path string true "The round ID"
// @Param submissionId path string true "The submission ID"
// @Param evaluationRequest body services.EvaluationRequest true "The evaluation request"
// @Security ApiKeyAuth
// @Error 400 {object} models.ResponseError
// @Error 403 {object} models.ResponseError
// @Error 404 {object} models.ResponseError
func SubmitNewPublicEvaluation(c *gin.Context, sess *cache.Session) {
	roundId := c.Param("roundId")
	submissionId := c.Param("submissionId")
	if roundId == "" || submissionId == "" {
		c.JSON(400, models.ResponseError{Detail: "Invalid request : Round ID and Submission ID are required"})
		return
	}
	requestedEvaluation := services.EvaluationRequest{}
	err := c.ShouldBindJSON(&requestedEvaluation)
	if err != nil {
		c.JSON(400, models.ResponseError{Detail: "Invalid request : " + err.Error()})
		return
	}
	evaluation_service := services.NewEvaluationService()
	evaluation, err := evaluation_service.PublicEvaluate(sess.UserID, types.SubmissionIDType(submissionId), &requestedEvaluation)
	if err != nil {
		c.JSON(400, models.ResponseError{Detail: "Error updating evaluation : " + err.Error()})
		return
	}
	c.JSON(200, models.ResponseSingle[models.Evaluation]{Data: *evaluation})
}

// SubmitNewBulkPublicEvaluation godoc
// @Summary Submit a new bulk public evaluation
// @Description Submit a new bulk public evaluation
// @Produce  json
// @Success 200 {object} models.ResponseList[models.Evaluation]
// @Router /evaluation/public/{roundId} [post]
// @Tags Evaluation
// @Param roundId path string true "The round ID"
// @Param evaluationRequest body []services.EvaluationRequest true "The evaluation request"
// @Security ApiKeyAuth
// @Error 400 {object} models.ResponseError
// @Error 403 {object} models.ResponseError
// @Error 404 {object} models.ResponseError
func SubmitNewBulkPublicEvaluation(c *gin.Context, sess *cache.Session) {
	roundId := c.Param("roundId")
	if roundId == "" {
		c.JSON(400, models.ResponseError{Detail: "Invalid request : Round ID and Submission ID are required"})
		return
	}
	requestedEvaluation := []services.EvaluationRequest{}
	err := c.ShouldBindJSON(&requestedEvaluation)
	if err != nil {
		c.JSON(400, models.ResponseError{Detail: "Invalid request : " + err.Error()})
		return
	}
	evaluation_service := services.NewEvaluationService()
	evaluations, assignmentCount, evaluationCount, err := evaluation_service.PublicBulkEvaluate(sess.UserID, requestedEvaluation)
	if err != nil {
		c.JSON(400, models.ResponseError{Detail: "Error updating evaluation : " + err.Error()})
		return
	}
	result := models.EvaluationListResponseWithCurrentStats{
		ResponseList:         models.ResponseList[*models.Evaluation]{Data: evaluations},
		TotalEvaluatedCount:  evaluationCount,
		TotalAssignmentCount: assignmentCount,
	}
	if len(evaluations) > 0 {
		result.PreviousToken = evaluations[0].EvaluationID.String()
		result.ContinueToken = evaluations[len(evaluations)-1].EvaluationID.String()
	}
	c.JSON(200, result)
}
