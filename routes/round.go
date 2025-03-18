package routes

import (
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

// GetResults godoc
// @Summary Get results of a round
// @Description Get results of a round
// @Produce  json
// @Success 200 {object} ResponseList[models.EvaluationResult]
// @Router /round/{roundId}/results [get]
// @Param roundId path string true "The round ID"
// @Tags Round
// @Error 400 {object} ResponseError
func GetResults(c *gin.Context, sess *cache.Session) {
	roundId := c.Param("roundId")
	if roundId == "" {
		c.JSON(400, ResponseError{Detail: "Invalid request : Round ID is required"})
	}
	round_service := services.NewRoundService()
	results, err := round_service.GetResults(models.IDType(roundId))
	if err != nil {
		c.JSON(400, ResponseError{Detail: "Failed to get round results : " + err.Error()})
		return
	}
	c.JSON(200, ResponseList[models.EvaluationResult]{Data: results})
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
	r.GET("/:roundId/results", WithSession(GetResults))
	r.POST("/:roundId/status", WithSession(UpdateStatus))
	r.POST("/", WithPermission(consts.PermissionCreateCampaign, CreateRound))
	r.POST("/:roundId", WithPermission(consts.PermissionCreateCampaign, UpdateRoundDetails))
	r.POST("/import/:roundId/commons", WithPermission(consts.PermissionLogin, ImportFromCommons))
	r.POST("/distribute/:roundId", WithPermission(consts.PermissionCreateCampaign, DistributeEvaluations))
}
