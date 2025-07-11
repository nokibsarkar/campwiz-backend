package routes

import (
	"encoding/csv"
	"fmt"
	"log"
	"nokib/campwiz/consts"
	"nokib/campwiz/models"
	"nokib/campwiz/repository/cache"
	"nokib/campwiz/services"

	"github.com/gin-gonic/gin"
)

type RoundDeletedResponse struct {
	RoundID models.IDType `json:"roundId"`
}

// CreateRound godoc
// @Summary Create a new round
// @Description Create a new round for a campaign
// @Produce  json
// @Success 200 {object} models.ResponseSingle[models.Round]
// @Router /round/ [post]
// @Security ApiKeyAuth
// @Param roundRequest body services.RoundRequest true "The round request"
// @Tags Round
// @Error 400 {object} models.ResponseError
func CreateRound(c *gin.Context, sess *cache.Session) {
	// defer HandleError("Create Round")
	requestedRounds := services.RoundRequest{
		CreatedByID: sess.UserID,
	}
	err := c.ShouldBindJSON(&requestedRounds)
	if err != nil {
		c.JSON(400, models.ResponseError{Detail: "Invalid request : " + err.Error()})
		return
	}
	if !requestedRounds.IsPublicJury && len(requestedRounds.Juries) == 0 {
		c.JSON(400, models.ResponseError{Detail: "Invalid request : At least one jury is required for private jury"})
		return
	}
	round_service := services.NewRoundService()
	round, err := round_service.CreateRound(c, &requestedRounds)
	if err != nil {
		c.JSON(400, models.ResponseError{Detail: "Error creating round : " + err.Error()})
		return
	}
	c.JSON(200, models.ResponseSingle[models.Round]{Data: *round})
}

// ListAllRounds godoc
// @Summary List all rounds
// @Description get all rounds
// @Produce  json
// @Success 200 {object} models.ResponseList[models.Round]
// @Router /round/ [get]
// @param RoundFilter query models.RoundFilter false "Filter the rounds"
// @Tags Round
// @Security ApiKeyAuth
// @Error 400 {object} models.ResponseError
func ListAllRounds(c *gin.Context, sess *cache.Session) {
	defer HandleError("ListAllRounds")
	filter := &models.RoundFilter{}
	err := c.ShouldBindQuery(filter)
	if err != nil {
		c.JSON(400, models.ResponseError{Detail: "Invalid request : " + err.Error()})
		return
	}
	round_service := services.NewRoundService()
	rounds, err := round_service.ListAllRounds(c, filter)
	if err != nil {
		c.JSON(400, models.ResponseError{Detail: "Error listing rounds : " + err.Error()})
		return
	}
	c.JSON(200, models.ResponseList[models.Round]{Data: rounds})
}

// ImportFromCommons godoc
// @Summary Import images from commons
// @Description The user would provide a round ID and a list of commons categories and the system would import images from those categories
// @Produce  json
// @Success 200 {object} models.ResponseSingle[models.Task]
// @Router /round/import/{roundId}/commons [post]
// @Param roundId path string true "The round ID"
// @Param ImportFromCommons body services.ImportFromCommonsPayload true "The import from commons request"
// @Tags Round
// @Security ApiKeyAuth
// @Error 400 {object} models.ResponseError
func ImportFromCommons(c *gin.Context, sess *cache.Session) {
	roundId := c.Param("roundId")
	if roundId == "" {
		c.JSON(400, models.ResponseError{Detail: "Invalid request : Round ID is required"})
	}
	req := &services.ImportFromCommonsPayload{}
	err := c.ShouldBindBodyWithJSON(req)
	if err != nil {
		c.JSON(400, models.ResponseError{Detail: "Error Decoding : " + err.Error()})
		return
	}
	round_service := services.NewRoundService()
	if len(req.Categories) == 0 {
		c.JSON(400, models.ResponseError{Detail: "Invalid request : No categories provided"})
		return
	}
	task, err := round_service.ImportFromCommons(c, models.IDType(roundId), req.Categories)
	if err != nil {
		c.JSON(400, models.ResponseError{Detail: "Failed to import images : " + err.Error()})
		return
	}
	c.JSON(200, models.ResponseSingle[*models.Task]{Data: task})
}

// ImportFromPreviousRound godoc
// @Summary Import images from previous round
// @Description The user would provide a round ID and a list of scores and the system would import images from the previous round with those scores
// @Produce  json
// @Success 200 {object} models.ResponseSingle[models.Task]
// @Router /round/import/{targetRoundId}/previous [post]
// @Param targetRoundId path string true "The target round ID, where the images will be imported"
// @Param ImportFromPreviousRoundPayload body services.ImportFromPreviousRoundPayload true "The import from previous round request"
// @Tags Round
// @Security ApiKeyAuth
// @Error 400 {object} models.ResponseError
func ImportFromPreviousRound(c *gin.Context, sess *cache.Session) {
	roundId := c.Param("roundId")
	if roundId == "" {
		c.JSON(400, models.ResponseError{Detail: "Invalid request : Round ID is required"})
	}
	req := &services.ImportFromPreviousRoundPayload{}
	err := c.ShouldBindBodyWithJSON(req)
	if err != nil {
		c.JSON(400, models.ResponseError{Detail: "Error Decoding : " + err.Error()})
		return
	}
	if len(req.Scores) == 0 {
		c.JSON(400, models.ResponseError{Detail: "Invalid request : No scores provided"})
		return
	}
	if req.RoundID == "" {
		c.JSON(400, models.ResponseError{Detail: "Invalid request : No round ID provided"})
		return
	}
	round_service := services.NewRoundService()
	if len(req.Scores) == 0 {
		c.JSON(400, models.ResponseError{Detail: "Invalid request : No scores provided"})
		return
	}
	task, err := round_service.ImportFromPreviousRound(c, sess.UserID, models.IDType(roundId), req)
	if err != nil {
		c.JSON(400, models.ResponseError{Detail: "Failed to import images : " + err.Error()})
		return
	}
	c.JSON(200, models.ResponseSingle[*models.Task]{Data: task})
}

// ImportFromCSV godoc
// @Summary Import images from CSV
// @Description The user would provide a round ID and a CSV file path and the system would import images from that CSV file
// @Consumes multipart/form-data
// @Produce  json
// @Success 200 {object} models.ResponseSingle[models.Task]
// @Router /round/import/{roundId}/csv [post]
// @Param roundId path string true "The round ID"
// @Param ImportFromCSVRequest formData services.ImportFromCSVRequest true "The import from CSV request"
// @Param file formData file true "The CSV file (upto 10MB CSV)"
// @Tags Round
// @Security ApiKeyAuth
// @Error 400 {object} models.ResponseError
func ImportFromCSV(c *gin.Context, sess *cache.Session) {
	roundId := c.Param("roundId")
	if roundId == "" {
		c.JSON(400, models.ResponseError{Detail: "Invalid request : Round ID is required"})
		return
	}
	req := &services.ImportFromCSVRequest{}
	err := c.ShouldBind(req)
	if err != nil {
		c.JSON(400, models.ResponseError{Detail: "Error Decoding : " + err.Error()})
		return
	}
	file := req.File
	if file == nil {
		c.JSON(400, models.ResponseError{Detail: "Invalid request : No file provided"})
		return
	}
	size := file.Size
	if size > consts.MAX_CSV_FILE_SIZE {
		c.JSON(400, models.ResponseError{Detail: "Invalid request : File size exceeds " + fmt.Sprintf("%dMB", consts.MAX_CSV_FILE_SIZE>>20)})
		return
	}

	round_service := services.NewRoundService()
	task, err := round_service.ImportFromCSV(c, models.IDType(roundId), req)
	if err != nil {
		c.JSON(400, models.ResponseError{Detail: "Failed to import images : " + err.Error()})
		return
	}
	c.JSON(200, models.ResponseSingle[*models.Task]{Data: task})
}

type ImportFromFountainRequest struct {
	Code string `json:"code" binding:"required"`
}

// ImportFromFountain godoc
// @Summary Import images from Fountain
// @Description The user would provide a round ID and a code and the system would import images from that Fountain
// @Produce  json
// @Success 200 {object} models.ResponseSingle[models.Task]
// @Router /round/import/{roundId}/fountain [post]
// @Param roundId path string true "The round ID"
// @Param ImportFromFountainRequest body ImportFromFountainRequest true "The import from Fountain request"
// @Tags Round
// @Security ApiKeyAuth
// @Error 400 {object} models.ResponseError
func ImportFromFountain(c *gin.Context, sess *cache.Session) {
	roundId := c.Param("roundId")
	if roundId == "" {
		c.JSON(400, models.ResponseError{Detail: "Invalid request : Round ID is required"})
		return
	}
	req := &ImportFromFountainRequest{}
	err := c.ShouldBind(req)
	if err != nil {
		c.JSON(400, models.ResponseError{Detail: "Error Decoding : " + err.Error()})
		return
	}
	code := req.Code
	if code == "" {
		c.JSON(400, models.ResponseError{Detail: "Invalid request : No code provided"})
		return
	}

	round_service := services.NewRoundService()
	task, err := round_service.ImportFromFountain(c, models.IDType(roundId), code)
	if err != nil {
		c.JSON(400, models.ResponseError{Detail: "Failed to import images : " + err.Error()})
		return
	}
	c.JSON(200, models.ResponseSingle[*models.Task]{Data: task})
}

type ImportFromCampWizV1Request struct {
	// The path to the CampWiz V1 database file, must be accessible by the server
	FromFile string `json:"fromFile" binding:"required"`
	// From Which Campaign ID to import
	FromCampaignId int32 `json:"fromCampaignId" binding:"required"`
	// To which round in our system to import the images
	ToRoundId models.IDType `json:"toRoundId" binding:"required"`
}

// ImportFromCampWizV1 godoc
// @Summary Import images from CampWiz V1
// @Description The user would provide a round ID and a CampWiz V1 database file and the system would import images from that file
// @Produce  json
// @Success 200 {object} models.ResponseSingle[models.Task]
// @Router /round/import/{roundId}/campwizv1 [post]
// @Param roundId path string true "The round ID"
// @Param ImportFromCampWizV1Request body ImportFromCampWizV1Request true "The import from CampWiz V1 request"
// @Tags Round
// @Security ApiKeyAuth
// @Error 400 {object} models.ResponseError
func ImportFromCampWizV1(c *gin.Context, sess *cache.Session) {
	roundId := c.Param("roundId")
	if roundId == "" {
		c.JSON(400, models.ResponseError{Detail: "Invalid request : Round ID is required"})
		return
	}
	req := &ImportFromCampWizV1Request{}
	err := c.ShouldBind(req)
	if err != nil {
		c.JSON(400, models.ResponseError{Detail: "Error Decoding : " + err.Error()})
		return
	}
	if req.FromFile == "" {
		c.JSON(400, models.ResponseError{Detail: "Invalid request : No file provided"})
		return
	}
	if req.FromCampaignId <= 0 {
		c.JSON(400, models.ResponseError{Detail: "Invalid request : From Campaign ID must be positive"})
		return
	}
	if req.ToRoundId == "" {
		c.JSON(400, models.ResponseError{Detail: "Invalid request : No To Round ID provided"})
		return
	}
	round_service := services.NewRoundService()
	task, err := round_service.ImportFromCampWizV1(c, req.FromFile, req.FromCampaignId, req.ToRoundId)
	if err != nil {
		c.JSON(400, models.ResponseError{Detail: "Failed to import images : " + err.Error()})
		return
	}
	c.JSON(200, models.ResponseSingle[*models.Task]{Data: task})
}

// UpdateRoundDetails godoc
// @Summary Update the details of a round
// @Description Update the details of a round
// @Produce  json
// @Success 200 {object} models.ResponseSingle[models.Round]
// @Router /round/{roundId} [post]
// @Param roundId path string true "The round ID"
// @Param roundRequest body services.RoundRequest true "The round request"
// @Tags Round
// @Security ApiKeyAuth
// @Error 400 {object} models.ResponseError
func UpdateRoundDetails(c *gin.Context, sess *cache.Session) {
	roundId := c.Param("roundId")
	if roundId == "" {
		c.JSON(400, models.ResponseError{Detail: "Invalid request : Round ID is required"})
	}
	q := &models.SingleCampaignFilter{
		IncludeRoundRoles:      true,
		IncludeRoundRolesUsers: true,
	}
	// err := c.ShouldBindQuery(q)
	// if err != nil {
	// 	c.JSON(400, models.ResponseError{Detail: "Invalid request : " + err.Error()})
	// 	return
	// }
	req := &services.RoundRequest{}
	err := c.ShouldBindBodyWithJSON(req)
	if err != nil {
		c.JSON(400, models.ResponseError{Detail: "Error Decoding : " + err.Error()})
		return
	}
	log.Printf("Request : %+v", req)
	round_service := services.NewRoundService()
	round, err := round_service.UpdateRoundDetails(c, models.IDType(roundId), req, q)
	if err != nil {
		c.JSON(400, models.ResponseError{Detail: "Failed to update round : " + err.Error()})
		return
	}
	c.JSON(200, models.ResponseSingle[*models.Round]{Data: round})
}

type UpdateStatusRequest struct {
	Status models.RoundStatus `json:"status"`
}

// UpdateStatus godoc
// @Summary Update the status of a round
// @Description Update the status of a round
// @Produce  json
// @Success 200 {object} models.ResponseSingle[models.Round]
// @Router /round/{roundId}/status [post]
// @Param roundId path string true "The round ID"
// @Param UpdateStatusRequest body UpdateStatusRequest true "The status request"
// @Tags Round
// @Security ApiKeyAuth
// @Error 400 {object} models.ResponseError
// @Security ApiKeyAuth
func UpdateStatus(c *gin.Context, sess *cache.Session) {
	roundId := c.Param("roundId")
	if roundId == "" {
		c.JSON(400, models.ResponseError{Detail: "Invalid request : Round ID is required"})
	}
	req := &UpdateStatusRequest{}
	err := c.ShouldBind(req)
	if err != nil {
		c.JSON(400, models.ResponseError{Detail: "Error Decoding : " + err.Error()})
		return
	}
	round_service := services.NewRoundService()
	round, err := round_service.UpdateStatus(c, sess.UserID, models.IDType(roundId), req.Status)
	if err != nil {
		c.JSON(400, models.ResponseError{Detail: "Failed to update round : " + err.Error()})
		return
	}
	c.JSON(200, models.ResponseSingle[*models.Round]{Data: round})

}

// DistributeEvaluations godoc
// @Summary Distribute evaluations to juries
// @Description Distribute evaluations to juries
// @Produce  json
// @Success 200 {object} models.ResponseSingle[models.Task]
// @Router /round/distribute/{roundId} [post]
// @Param roundId path string true "The round ID"
// @Param DistributionRequest body services.DistributionRequest true "The distribution request"
// @Tags Round
// @Security ApiKeyAuth
// @Error 400 {object} models.ResponseError
func DistributeEvaluations(c *gin.Context, sess *cache.Session) {
	roundId := c.Param("roundId")
	if roundId == "" {
		c.JSON(400, models.ResponseError{Detail: "Invalid request : Round ID is required"})
	}
	distributionReq := &services.DistributionRequest{}
	err := c.ShouldBind(distributionReq)
	if err != nil {
		c.JSON(400, models.ResponseError{Detail: "Error Decoding : " + err.Error()})
		return
	}
	round_service := services.NewRoundService()
	task, err := round_service.DistributeEvaluations(c, sess.UserID, models.IDType(roundId), distributionReq)
	if err != nil {
		c.JSON(400, models.ResponseError{Detail: "Failed to distribute evaluations : " + err.Error()})
		return
	}
	c.JSON(200, models.ResponseSingle[*models.Task]{Data: task})
}

// GetRound godoc
// @Summary Get a round
// @Description Get a round
// @Produce  json
// @Success 200 {object} models.ResponseSingle[models.Round]
// @Router /round/{roundId} [get]
// @Param roundId path string true "The round ID"
// @Tags Round
// @Security ApiKeyAuth
// @Error 400 {object} models.ResponseError
func GetRound(c *gin.Context) {
	roundId := c.Param("roundId")
	if roundId == "" {
		c.JSON(400, models.ResponseError{Detail: "Invalid request : Round ID is required"})
	}
	round_service := services.NewRoundService()
	round, err := round_service.GetById(c, models.IDType(roundId))
	if err != nil {
		c.JSON(400, models.ResponseError{Detail: "Failed to get round : " + err.Error()})
		return
	}
	c.JSON(200, models.ResponseSingle[*models.Round]{Data: round})
}

// GetResultSummary godoc
// @Summary Get results of a round
// @Description Get results of a round
// @Produce  json
// @Success 200 {object} models.ResponseList[models.EvaluationResult]
// @Router /round/{roundId}/results/summary [get]
// @Param roundId path string true "The round ID"
// @Tags Round
// @Security ApiKeyAuth
// @Error 400 {object} models.ResponseError
func GetResultSummary(c *gin.Context, sess *cache.Session) {
	roundId := c.Param("roundId")
	if roundId == "" {
		c.JSON(400, models.ResponseError{Detail: "Invalid request : Round ID is required"})
	}
	round_service := services.NewRoundService()
	results, err := round_service.GetResultSummary(c, models.IDType(roundId))
	if err != nil {
		c.JSON(400, models.ResponseError{Detail: "Failed to get round results : " + err.Error()})
		return
	}
	c.JSON(200, models.ResponseList[models.EvaluationResult]{Data: results})
}

// GetResults godoc
// @Summary Get results of a round
// @Description Get results of a round
// @Produce  json
// @Success 200 {object} models.ResponseList[models.SubmissionResult]
// @Produce  text/csv
// @Router /round/{roundId}/results/{format} [get]
// @Param roundId path string true "The round ID"
// @Param format path models.ResultExportFormat true "The format of the results"
// @Param SubmissionResultQuery query models.SubmissionResultQuery false "The query to filter the results"
// @Tags Round
// @Error 400 {object} models.ResponseError
// @Security ApiKeyAuth
func GetResults(c *gin.Context, sess *cache.Session) {
	roundId := c.Param("roundId")
	if roundId == "" {
		c.JSON(400, models.ResponseError{Detail: "Invalid request : Round ID is required"})
	}
	format := models.ResultExportFormatJSON
	formatString := c.Param("format")
	if formatString != "" {
		format = models.ResultExportFormat(formatString)
	}
	q := &models.SubmissionResultQuery{}
	err := c.ShouldBindQuery(&q)
	if err != nil {
		c.JSON(400, models.ResponseError{Detail: "Invalid request : " + err.Error()})
		return
	}

	round_service := services.NewRoundService()
	results, err := round_service.GetResults(c, sess.UserID, models.IDType(roundId), q)
	if err != nil {
		c.JSON(404, models.ResponseError{Detail: "Failed to get round results : " + err.Error()})
		return
	}
	switch format {
	case models.ResultExportFormatJSON:
		result := models.ResponseList[models.SubmissionResult]{Data: results}
		if len(results) > 0 {
			result.ContinueToken = results[len(results)-1].SubmissionID.String()
			result.PreviousToken = results[0].SubmissionID.String()
		}
		c.JSON(200, result)
	case models.ResultExportFormatCSV:
		c.Writer.Header().Set("Content-Type", "text/csv")
		c.Writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=round-%s-results.csv", roundId))
		csvWriter := csv.NewWriter(c.Writer)
		err := csvWriter.Write([]string{"Submission ID", "Name", "Score", "Author", "Evaluation Count", "Media Type"})
		if err != nil {
			c.JSON(400, models.ResponseError{Detail: "Failed to write CSV header : " + err.Error()})
			return
		}
		for _, result := range results {
			err = csvWriter.Write([]string{result.SubmissionID.String(),
				result.Name, fmt.Sprintf("%f", result.Score),
				result.Author,
				fmt.Sprintf("%d", result.EvaluationCount),
				string(result.MediaType)})
			if err != nil {
				c.JSON(400, models.ResponseError{Detail: "Failed to write CSV row : " + err.Error()})
				return
			}
		}
		csvWriter.Flush()
		if err := csvWriter.Error(); err != nil {
			c.JSON(400, models.ResponseError{Detail: "Failed to write CSV : " + err.Error()})
			return
		}
	default:
		c.JSON(400, models.ResponseError{Detail: "Invalid request : Invalid format"})
		return
	}
}

// NextPublicSubmission godoc
// @Summary Get the next public submission
// @Description Get the next public submission for a jury
// @Produce  json
// @Success 200 {object} models.ResponseSingle[models.Submission]
// @Router /round/{roundId}/next/public [get]
// @Param roundId path string true "The round ID"
// @Tags Round
// @Security ApiKeyAuth
// @Error 400 {object} models.ResponseError
func NextPublicSubmission(c *gin.Context, sess *cache.Session) {
	roundID := c.Param("roundId")
	if roundID == "" {
		c.JSON(400, models.ResponseError{Detail: "Invalid request : Round ID is required"})
	}
	qry := &models.EvaluationFilter{}
	err := c.ShouldBindQuery(&qry)
	if err != nil {
		c.JSON(400, models.ResponseError{Detail: "Invalid request : " + err.Error()})
		return
	}
	u := GetCurrentUser(c)
	qry.RoundID = models.IDType(roundID)
	round_service := services.NewRoundService()
	round, err := round_service.GetById(c, models.IDType(roundID))
	if err != nil {
		c.JSON(400, models.ResponseError{Detail: "Failed to get round : " + err.Error()})
		return
	}
	if !round.IsPublicJury {
		c.JSON(400, models.ResponseError{Detail: "Invalid request : Round is not public jury"})
		return
	}
	submissions, role, err := round_service.GetNextUnevaluatedSubmissionForPublicJury(c, u.UserID, qry)
	if err != nil {
		c.JSON(400, models.ResponseError{Detail: "Failed to get next submission : " + err.Error()})
		return
	}
	ev := []*models.Evaluation{}
	for _, submission := range submissions {
		evaluation := &models.Evaluation{
			SubmissionID: submission.SubmissionID,
			JudgeID:      &role.RoleID,
			RoundID:      round.RoundID,
			// EvaluationID: idgenerator.GenerateID("e"),
			Submission: submission,
			Type:       round.Type,
		}
		ev = append(ev, evaluation)
	}
	result := models.EvaluationListResponseWithCurrentStats{
		ResponseList:         models.ResponseList[*models.Evaluation]{Data: ev},
		TotalEvaluatedCount:  role.TotalEvaluated,
		TotalAssignmentCount: role.TotalAssigned,
	}
	if len(ev) > 0 {
		result.ContinueToken = string(ev[len(ev)-1].SubmissionID)
		result.PreviousToken = string(ev[0].SubmissionID)
	}
	c.JSON(200, result)
}

// NextSubmissionEvaluation godoc
// @Summary Get the next submission evaluation
// @Description Get the next submission evaluation for a jury
// @Produce  json
// @Success 200 {object} models.ResponseList[models.Evaluation]
// @Router /round/{roundId}/next/evaluation [get]
// @Param roundId path string true "The round ID"
// @Tags Round
// @Security ApiKeyAuth
// @Error 400 {object} models.ResponseError
func NextSubmissionEvaluation(c *gin.Context, sess *cache.Session) {
	roundID := c.Param("roundId")
	if roundID == "" {
		c.JSON(400, models.ResponseError{Detail: "Invalid request : Round ID is required"})
	}
	q := models.GetEvaluationQueryFilter{}
	err := c.ShouldBindQuery(&q)
	if err != nil {
		c.JSON(400, models.ResponseError{Detail: "Invalid request : " + err.Error()})
		return
	}
	ListEvaluations(c, sess)
}

// DeleteRound godoc
// @Summary Delete a round
// @Description Delete a round
// @Produce  json
// @Success 200 {object} models.ResponseSingle[RoundDeletedResponse]
// @Router /round/{roundId} [delete]
// @Param roundId path string true "The round ID"
// @Tags Round
// @Security ApiKeyAuth
// @Error 400 {object} models.ResponseError
func DeleteRound(c *gin.Context, sess *cache.Session) {
	roundID := c.Param("roundId")
	if roundID == "" {
		c.JSON(400, models.ResponseError{Detail: "Invalid request : Round ID is required"})
	}
	round_service := services.NewRoundService()
	err := round_service.DeleteRound(c, sess, models.IDType(roundID))
	if err != nil {
		c.JSON(400, models.ResponseError{Detail: "Failed to delete round : " + err.Error()})
		return
	}
	c.JSON(200, models.ResponseSingle[RoundDeletedResponse]{Data: RoundDeletedResponse{RoundID: models.IDType(roundID)}})
}

// AddMySelfAsJury godoc
// @Summary Add myself as a jury
// @Description Add the current user as a jury for the round
// @Router /round/{roundId}/jury [post]
// @Param roundId path string true "The round ID"
// @Tags Round
// @Security ApiKeyAuth
// @Success 200 {object} models.ResponseSingle[models.Role]
// @Param roundId path string true "The round ID"
// @Error 400 {object} models.ResponseError
// @Produce  json
func addMySelfAsJury(c *gin.Context, sess *cache.Session) {
	roundID := c.Param("roundId")
	if roundID == "" {
		c.JSON(400, models.ResponseError{Detail: "Invalid request : Round ID is required"})
	}
	round_service := services.NewRoundService()
	role, err := round_service.AddMyselfAsJury(c, sess.UserID, models.IDType(roundID))
	if err != nil {
		c.JSON(400, models.ResponseError{Detail: "Failed to add myself as jury : " + err.Error()})
		return
	}
	c.JSON(200, models.ResponseSingle[*models.Role]{Data: role})
}

// Randomize godoc
// @Summary Randomize the evaluation distribution
// @Description Randomize the evaluation distribution
// @Produce  json
// @Success 200 {object} models.ResponseSingle[models.Task]
// @Router /round/{roundId}/randomize [post]
// @Param roundId path string true "The round ID"
// @Tags Round
// @Error 400 {object} models.ResponseError
func Randomize(c *gin.Context, sess *cache.Session) {
	roundId := c.Param("roundId")
	if roundId == "" {
		c.JSON(400, models.ResponseError{Detail: "Invalid request : Round ID is required"})
		return
	}
	round_service := services.NewRoundService()
	task, err := round_service.Randomize(c, sess.UserID, models.IDType(roundId))
	if err != nil {
		c.JSON(400, models.ResponseError{Detail: "Failed to randomize evaluations : " + err.Error()})
		return
	}
	c.JSON(200, models.ResponseSingle[*models.Task]{Data: task})
}
