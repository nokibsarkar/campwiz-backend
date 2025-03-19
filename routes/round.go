package routes

import (
	"encoding/csv"
	"fmt"
	"nokib/campwiz/consts"
	"nokib/campwiz/models"
	"nokib/campwiz/repository/cache"
	"nokib/campwiz/services"

	"github.com/gin-gonic/gin"
)

// CreateRound godoc
// @Summary Create a new round
// @Description Create a new round for a campaign
// @Produce  json
// @Success 200 {object} ResponseSingle[models.Round]
// @Router /round/ [post]
// @Param roundRequest body services.RoundRequest true "The round request"
// @Tags Round
// @Error 400 {object} ResponseError
func CreateRound(c *gin.Context, sess *cache.Session) {
	defer HandleError("Create Round")
	requestedRounds := services.RoundRequest{
		CreatedByID: sess.UserID,
	}
	err := c.ShouldBindJSON(&requestedRounds)
	if err != nil {
		c.JSON(400, ResponseError{Detail: "Invalid request : " + err.Error()})
		return
	}
	if len(requestedRounds.Juries) == 0 {
		c.JSON(400, ResponseError{Detail: "Invalid request : At least one jury is required"})
		return
	}
	round_service := services.NewRoundService()
	round, err := round_service.CreateRound(&requestedRounds)
	if err != nil {
		c.JSON(400, ResponseError{Detail: "Error creating round : " + err.Error()})
		return
	}
	c.JSON(200, ResponseSingle[models.Round]{Data: *round})
}

// ListAllRounds godoc
// @Summary List all rounds
// @Description get all rounds
// @Produce  json
// @Success 200 {object} ResponseList[models.Round]
// @Router /round/ [get]
// @param RoundFilter query models.RoundFilter false "Filter the rounds"
// @Tags Round
// @Error 400 {object} ResponseError
func ListAllRounds(c *gin.Context, sess *cache.Session) {
	defer HandleError("ListAllRounds")
	filter := &models.RoundFilter{}
	err := c.ShouldBindQuery(filter)
	if err != nil {
		c.JSON(400, ResponseError{Detail: "Invalid request : " + err.Error()})
		return
	}
	round_service := services.NewRoundService()
	rounds, err := round_service.ListAllRounds(filter)
	if err != nil {
		c.JSON(400, ResponseError{Detail: "Error listing rounds : " + err.Error()})
		return
	}
	c.JSON(200, ResponseList[models.Round]{Data: rounds})
}

// ImportFromCommons godoc
// @Summary Import images from commons
// @Description The user would provide a round ID and a list of commons categories and the system would import images from those categories
// @Produce  json
// @Success 200 {object} ResponseSingle[models.Task]
// @Router /round/import/{roundId}/commons [post]
// @Param roundId path string true "The round ID"
// @Param ImportFromCommons body services.ImportFromCommonsPayload true "The import from commons request"
// @Tags Round
// @Error 400 {object} ResponseError
func ImportFromCommons(c *gin.Context, sess *cache.Session) {
	roundId := c.Param("roundId")
	if roundId == "" {
		c.JSON(400, ResponseError{Detail: "Invalid request : Round ID is required"})
	}
	req := &services.ImportFromCommonsPayload{}
	err := c.ShouldBindBodyWithJSON(req)
	if err != nil {
		c.JSON(400, ResponseError{Detail: "Error Decoding : " + err.Error()})
		return
	}
	round_service := services.NewRoundService()
	if len(req.Categories) == 0 {
		c.JSON(400, ResponseError{Detail: "Invalid request : No categories provided"})
		return
	}
	task, err := round_service.ImportFromCommons(models.IDType(roundId), req.Categories)
	if err != nil {
		c.JSON(400, ResponseError{Detail: "Failed to import images : " + err.Error()})
		return
	}
	c.JSON(200, ResponseSingle[*models.Task]{Data: task})
}

// ImportFromPreviousRound godoc
// @Summary Import images from previous round
// @Description The user would provide a round ID and a list of scores and the system would import images from the previous round with those scores
// @Produce  json
// @Success 200 {object} ResponseSingle[models.Task]
// @Router /round/import/{targetRoundId}/previous [post]
// @Param targetRoundId path string true "The target round ID, where the images will be imported"
// @Param ImportFromPreviousRoundPayload body services.ImportFromPreviousRoundPayload true "The import from previous round request"
// @Tags Round
// @Error 400 {object} ResponseError
func ImportFromPreviousRound(c *gin.Context, sess *cache.Session) {
	roundId := c.Param("roundId")
	if roundId == "" {
		c.JSON(400, ResponseError{Detail: "Invalid request : Round ID is required"})
	}
	req := &services.ImportFromPreviousRoundPayload{}
	err := c.ShouldBindBodyWithJSON(req)
	if err != nil {
		c.JSON(400, ResponseError{Detail: "Error Decoding : " + err.Error()})
		return
	}
	if len(req.Scores) == 0 {
		c.JSON(400, ResponseError{Detail: "Invalid request : No scores provided"})
		return
	}
	if req.RoundID == "" {
		c.JSON(400, ResponseError{Detail: "Invalid request : No round ID provided"})
		return
	}
	round_service := services.NewRoundService()
	if len(req.Scores) == 0 {
		c.JSON(400, ResponseError{Detail: "Invalid request : No scores provided"})
		return
	}
	task, err := round_service.ImportFromPreviousRound(sess.UserID, models.IDType(roundId), req)
	if err != nil {
		c.JSON(400, ResponseError{Detail: "Failed to import images : " + err.Error()})
		return
	}
	c.JSON(200, ResponseSingle[*models.Task]{Data: task})
}

// UpdateRoundDetails godoc
// @Summary Update the details of a round
// @Description Update the details of a round
// @Produce  json
// @Success 200 {object} ResponseSingle[models.Round]
// @Router /round/{roundId} [post]
// @Param roundId path string true "The round ID"
// @Param roundRequest body services.RoundRequest true "The round request"
// @Tags Round
// @Error 400 {object} ResponseError
func UpdateRoundDetails(c *gin.Context, sess *cache.Session) {
	roundId := c.Param("roundId")
	if roundId == "" {
		c.JSON(400, ResponseError{Detail: "Invalid request : Round ID is required"})
	}
	q := &models.SingleCampaaignFilter{
		IncludeRoundRoles:      true,
		IncludeRoundRolesUsers: true,
	}
	// err := c.ShouldBindQuery(q)
	// if err != nil {
	// 	c.JSON(400, ResponseError{Detail: "Invalid request : " + err.Error()})
	// 	return
	// }
	req := &services.RoundRequest{}
	err := c.ShouldBindBodyWithJSON(req)
	if err != nil {
		c.JSON(400, ResponseError{Detail: "Error Decoding : " + err.Error()})
		return
	}
	round_service := services.NewRoundService()
	round, err := round_service.UpdateRoundDetails(models.IDType(roundId), req, q)
	if err != nil {
		c.JSON(400, ResponseError{Detail: "Failed to update round : " + err.Error()})
		return
	}
	c.JSON(200, ResponseSingle[*models.Round]{Data: round})
}

type UpdateStatusRequest struct {
	Status models.RoundStatus `json:"status"`
}

// UpdateStatus godoc
// @Summary Update the status of a round
// @Description Update the status of a round
// @Produce  json
// @Success 200 {object} ResponseSingle[models.Round]
// @Router /round/{roundId}/status [post]
// @Param roundId path string true "The round ID"
// @Param UpdateStatusRequest body UpdateStatusRequest true "The status request"
// @Tags Round
// @Error 400 {object} ResponseError
// @Security ApiKeyAuth
func UpdateStatus(c *gin.Context, sess *cache.Session) {
	roundId := c.Param("roundId")
	if roundId == "" {
		c.JSON(400, ResponseError{Detail: "Invalid request : Round ID is required"})
	}
	req := &UpdateStatusRequest{}
	err := c.ShouldBind(req)
	if err != nil {
		c.JSON(400, ResponseError{Detail: "Error Decoding : " + err.Error()})
		return
	}
	round_service := services.NewRoundService()
	round, err := round_service.UpdateStatus(sess.UserID, models.IDType(roundId), req.Status)
	if err != nil {
		c.JSON(400, ResponseError{Detail: "Failed to update round : " + err.Error()})
		return
	}
	c.JSON(200, ResponseSingle[*models.Round]{Data: round})

}

// DistributeEvaluations godoc
// @Summary Distribute evaluations to juries
// @Description Distribute evaluations to juries
// @Produce  json
// @Success 200 {object} ResponseSingle[models.Task]
// @Router /round/distribute/{roundId} [post]
// @Param roundId path string true "The round ID"
// @Param DistributionRequest body services.DistributionRequest true "The distribution request"
// @Tags Round
// @Error 400 {object} ResponseError
func DistributeEvaluations(c *gin.Context, sess *cache.Session) {
	roundId := c.Param("roundId")
	if roundId == "" {
		c.JSON(400, ResponseError{Detail: "Invalid request : Round ID is required"})
	}
	distributionReq := &services.DistributionRequest{}
	err := c.ShouldBind(distributionReq)
	if err != nil {
		c.JSON(400, ResponseError{Detail: "Error Decoding : " + err.Error()})
		return
	}
	round_service := services.NewRoundService()
	task, err := round_service.DistributeEvaluations(sess.UserID, models.IDType(roundId), distributionReq)
	if err != nil {
		c.JSON(400, ResponseError{Detail: "Failed to distribute evaluations : " + err.Error()})
		return
	}
	c.JSON(200, ResponseSingle[*models.Task]{Data: task})
}

// GetRound godoc
// @Summary Get a round
// @Description Get a round
// @Produce  json
// @Success 200 {object} ResponseSingle[models.Round]
// @Router /round/{roundId} [get]
// @Param roundId path string true "The round ID"
// @Tags Round
// @Error 400 {object} ResponseError
func GetRound(c *gin.Context) {
	roundId := c.Param("roundId")
	if roundId == "" {
		c.JSON(400, ResponseError{Detail: "Invalid request : Round ID is required"})
	}
	round_service := services.NewRoundService()
	round, err := round_service.GetById(models.IDType(roundId))
	if err != nil {
		c.JSON(400, ResponseError{Detail: "Failed to get round : " + err.Error()})
		return
	}
	c.JSON(200, ResponseSingle[*models.Round]{Data: round})
}

// GetResultSummary godoc
// @Summary Get results of a round
// @Description Get results of a round
// @Produce  json
// @Success 200 {object} ResponseList[models.EvaluationResult]
// @Router /round/{roundId}/results/summary [get]
// @Param roundId path string true "The round ID"
// @Tags Round
// @Error 400 {object} ResponseError
func GetResultSummary(c *gin.Context, sess *cache.Session) {
	roundId := c.Param("roundId")
	if roundId == "" {
		c.JSON(400, ResponseError{Detail: "Invalid request : Round ID is required"})
	}
	round_service := services.NewRoundService()
	results, err := round_service.GetResultSummary(models.IDType(roundId))
	if err != nil {
		c.JSON(400, ResponseError{Detail: "Failed to get round results : " + err.Error()})
		return
	}
	c.JSON(200, ResponseList[models.EvaluationResult]{Data: results})
}

// GetResults godoc
// @Summary Get results of a round
// @Description Get results of a round
// @Produce  json
// @Success 200 {object} ResponseList[models.SubmissionResult]
// @Produce  text/csv
// @Router /round/{roundId}/results/{format} [get]
// @Param roundId path string true "The round ID"
// @Param format path models.ResultExportFormat true "The format of the results"
// @Param SubmissionResultQuery query models.SubmissionResultQuery false "The query to filter the results"
// @Tags Round
// @Error 400 {object} ResponseError
// @Security ApiKeyAuth
func GetResults(c *gin.Context, sess *cache.Session) {
	roundId := c.Param("roundId")
	if roundId == "" {
		c.JSON(400, ResponseError{Detail: "Invalid request : Round ID is required"})
	}
	format := models.ResultExportFormatJSON
	formatString := c.Param("format")
	if formatString != "" {
		format = models.ResultExportFormat(formatString)
	}
	q := &models.SubmissionResultQuery{}
	err := c.ShouldBindQuery(&q)
	if err != nil {
		c.JSON(400, ResponseError{Detail: "Invalid request : " + err.Error()})
		return
	}

	round_service := services.NewRoundService()
	results, err := round_service.GetResults(sess.UserID, models.IDType(roundId), q)
	if err != nil {
		c.JSON(404, ResponseError{Detail: "Failed to get round results : " + err.Error()})
		return
	}
	if format == models.ResultExportFormatJSON {
		result := ResponseList[models.SubmissionResult]{Data: results}
		if len(results) > 0 {
			result.ContinueToken = results[len(results)-1].SubmissionID.String()
			result.PreviousToken = results[0].SubmissionID.String()
		}
		c.JSON(200, result)
	} else if format == models.ResultExportFormatCSV {
		c.Writer.Header().Set("Content-Type", "text/csv")
		c.Writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=round-%s-results.csv", roundId))
		csvWriter := csv.NewWriter(c.Writer)
		csvWriter.Write([]string{"Submission ID", "Name", "Score", "Author", "Evaluation Count", "Media Type"})
		for _, result := range results {
			csvWriter.Write([]string{result.SubmissionID.String(),
				result.Name, fmt.Sprintf("%f", result.Score),
				result.Author,
				fmt.Sprintf("%d", result.EvaluationCount),
				string(result.MediaType)})

		}
		csvWriter.Flush()
		if err := csvWriter.Error(); err != nil {
			c.JSON(400, ResponseError{Detail: "Failed to write CSV : " + err.Error()})
			return
		}
	} else {
		c.JSON(400, ResponseError{Detail: "Invalid request : Invalid format"})
		return
	}
}

// NextPublicSubmission godoc
// @Summary Get the next public submission
// @Description Get the next public submission for a jury
// @Produce  json
// @Success 200 {object} ResponseSingle[models.Submission]
// @Router /round/{roundId}/next/public [get]
// @Param roundId path string true "The round ID"
// @Tags Round
// @Error 400 {object} ResponseError
func NextPublicSubmission(c *gin.Context, sess *cache.Session) {
	roundID := c.Param("roundId")
	if roundID == "" {
		c.JSON(400, ResponseError{Detail: "Invalid request : Round ID is required"})
	}
	u := GetCurrentUser(c)
	round_service := services.NewRoundService()
	submission, err := round_service.GetNextUnevaluatedSubmissionForPublicJury(u.UserID, models.IDType(roundID))
	if err != nil {
		c.JSON(400, ResponseError{Detail: "Failed to get next submission : " + err.Error()})
		return
	}
	c.JSON(200, ResponseSingle[*models.Submission]{Data: submission})
}
func NextSubmissionEvaluation(c *gin.Context, sess *cache.Session) {
	roundID := c.Param("roundId")
	if roundID == "" {
		c.JSON(400, ResponseError{Detail: "Invalid request : Round ID is required"})
	}
	q := models.GetEvaluationQueryFilter{}
	err := c.ShouldBindQuery(&q)
	if err != nil {
		c.JSON(400, ResponseError{Detail: "Invalid request : " + err.Error()})
		return
	}
	ListEvaluations(c, sess)
}
func NewRoundRoutes(parent *gin.RouterGroup) {
	r := parent.Group("/round")
	r.GET("/", WithSession(ListAllRounds))
	r.GET("/:roundId", GetRound)
	r.GET("/:roundId/next/public", WithSession(NextPublicSubmission))
	r.GET("/:roundId/results/summary", WithSession(GetResultSummary))
	r.GET("/:roundId/results/:format", WithSession(GetResults))
	r.POST("/:roundId/status", WithSession(UpdateStatus))
	r.POST("/", WithPermission(consts.PermissionCreateCampaign, CreateRound))
	r.POST("/:roundId", WithSession(UpdateRoundDetails))
	r.POST("/import/:roundId/commons", WithPermission(consts.PermissionLogin, ImportFromCommons))
	r.POST("/import/:roundId/previous", WithPermission(consts.PermissionLogin, ImportFromPreviousRound))
	r.POST("/distribute/:roundId", WithPermission(consts.PermissionCreateCampaign, DistributeEvaluations))
}
