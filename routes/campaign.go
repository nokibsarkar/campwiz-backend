package routes

import (
	"errors"
	"nokib/campwiz/models"
	"nokib/campwiz/repository/cache"
	"nokib/campwiz/services"
	"time"

	"github.com/gin-gonic/gin"
)

func listAllCampaigns(c *gin.Context) ([]models.Campaign, error) {
	qry := &models.CampaignFilter{}
	err := c.ShouldBindQuery(qry)
	if err != nil {
		return nil, err
	}
	campaignService := services.NewCampaignService()
	var campaignList []models.Campaign
	if qry.IsHidden != nil && *qry.IsHidden {
		sess := GetSession(c)
		if sess == nil {
			return nil, errors.New("errors.unAuthenticated")
		}
		// if !sess.Permission.HasPermission(consts.PermissionOtherProjectAccess) {
		// 	// user is not an admin
		// 	if qry.ProjectID == "" {
		// 		c.JSON(400, models.ResponseError{Detail: "Invalid request : Project ID is required when isHidden is true and user is not an admin"})
		// 		return
		// 	}
		// }
		// if qry.ProjectID != "" {
		// 	// project id is provided
		// 	// check if the user is allowed to access this project
		// 	currentUser := GetCurrentUser(c)
		// 	if currentUser == nil {
		// 		c.JSON(400, models.ResponseError{Detail: "Invalid request : User not found"})
		// 		return
		// 	}
		// 	if currentUser.LeadingProjectID == nil {
		// 		c.JSON(400, models.ResponseError{Detail: "Invalid request : User is not leading any project"})
		// 		return
		// 	}
		// 	if *currentUser.LeadingProjectID != qry.ProjectID {
		// 		// the user is not an admin and the project ID does not match
		// 		// cross project access is allowed only for jury and coordinators
		// 		conn, close, err  := repository.GetDB()
		// 		defer close()
		// 		userRepo := repository.NewUserRepository()
		// 		roleFilter := &models.RoleFilter{ProjectID: qry.ProjectID, UserID: &currentUser.UserID}
		// 		roles, err := userRepo.FetchRoles(conn, roleFilter)
		// 		if err != nil {
		// 			c.JSON(400, models.ResponseError{Detail: "Invalid request : " + err.Error()})
		// 			return
		// 		}
		// 		hasRoles := false
		// 		for _, role := range roles {
		// 			if role.Type == models.RoleTypeCoordinator || role.Type == models.RoleTypeJury {
		// 				if role.DeletedAt == nil {
		// 					hasRoles = true
		// 					break
		// 				}
		// 			}
		// 		}
		// 		if !hasRoles {
		// 			c.JSON(400, models.ResponseError{Detail: "Invalid request : User does not have permission to access this project"})
		// 			return
		// 		}
		// 	}
		// }
		campaignList = campaignService.ListPrivateCampaigns(c, sess, qry)
	} else {
		campaignList = campaignService.GetAllCampaigns(c, qry)
	}
	return campaignList, nil
}

// ListAllCampaigns godoc
// @Summary List all campaigns
// @Description get all campaigns
// @Produce  json
// @Success 200 {object} models.ResponseList[models.Campaign]
// @Router /campaign/ [get]
// @param CampaignFilter query models.CampaignFilter false "Filter the campaigns"
// @Tags Campaign
// @Error 400 {object} models.ResponseError
func ListAllCampaigns(c *gin.Context) {
	campaignList, err := listAllCampaigns(c)
	if err != nil {
		c.JSON(400, models.ResponseError{Detail: "Invalid request : " + err.Error()})
		return
	}
	if len(campaignList) == 0 {
		c.JSON(200, models.ResponseList[models.Campaign]{Data: []models.Campaign{}})
		return
	}
	c.JSON(200, models.ResponseList[models.Campaign]{Data: campaignList})
}

/*
This function will return all the timelines of all the campaigns
*/
func GetAllCampaignTimeLine(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Hello, World!",
	})
}

// GetSingleCampaign godoc
// @Summary Get a single campaign
// @Description Get a single campaign
// @Produce  json
// @Success 200 {object} models.ResponseSingle[models.CampaignExtended]
// @Router /campaign/{campaignId} [get]
// @Security ApiKeyAuth
// @Param campaignId path string true "The campaign ID"
// @Param campaignQuery query services.SingleCampaignQuery false "The query for the campaign"
// @Tags Campaign
func GetSingleCampaign(c *gin.Context, sess *cache.Session) {
	campaignId := c.Param("campaignId")
	if campaignId == "" {
		c.JSON(400, models.ResponseError{Detail: "Invalid request : Campaign ID is required"})
	}
	q := &services.SingleCampaignQuery{}
	err := c.ShouldBindQuery(q)
	if err != nil {
		c.JSON(400, models.ResponseError{Detail: "Invalid request : " + err.Error()})
		return
	}
	campaign_service := services.NewCampaignService()
	campaign, err := campaign_service.GetCampaignByID(c, models.IDType(campaignId), q)
	if err != nil {
		c.JSON(404, models.ResponseError{Detail: "Failed to get campaign : " + err.Error()})
		return
	}
	ex := &models.CampaignExtended{Campaign: *campaign, Coordinators: []models.WikimediaUsernameType{}}
	if q.IncludeRoles {
		ex.Coordinators = []models.WikimediaUsernameType{}
		for _, role := range campaign.Roles {
			if role.Type == models.RoleTypeCoordinator {
				ex.Coordinators = append(ex.Coordinators, role.User.Username)
			}
		}
	}
	if q.IncludeRounds && q.IncludeRoundRoles {
		// if campaign.IsPublic {
		// 	for i, round := range campaign.Rounds {
		// 		// No need to show the roles for public campaigns
		// 		round.Roles = []models.Role{}
		// 		campaign.Rounds[i] = round
		// 	}
		// }
		for i, round := range campaign.Rounds {
			round.Jury = map[models.IDType]models.WikimediaUsernameType{}
			for _, role := range round.Roles {
				if role.Type == models.RoleTypeJury {
					round.Jury[role.UserID] = role.User.Username
				}
			}
			campaign.Rounds[i] = round
		}
	}
	c.JSON(200, models.ResponseSingle[*models.CampaignExtended]{Data: ex})
}

func ListAllJury(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Hello, World!",
	})
}

// CreateCampaign creates a new campaign
// @summary Create a new campaign
// @description Create a new campaign
// @tags Campaign
// @param campaignRequest body services.CampaignCreateRequest true "The campaign request"
// @produce json
// @Security ApiKeyAuth
// @success 200 {object} models.Campaign
// @router /campaign/ [post]
func CreateCampaign(c *gin.Context, sess *cache.Session) {
	createRequest := &services.CampaignCreateRequest{}
	err := c.ShouldBindBodyWithJSON(createRequest)
	if err != nil {
		c.JSON(400, models.ResponseError{Detail: "Invalid request : " + err.Error()})
		return
	}
	createRequest.CreatedByID = sess.UserID
	// Create a new campaign
	endDate := createRequest.EndDate
	startDate := createRequest.StartDate
	//make the start date 00:00:00 of the day
	createRequest.StartDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.UTC)
	// make the end date 23:59:59 of the day
	createRequest.EndDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 0, time.UTC)

	camapign_service := services.NewCampaignService()
	createRequest.Status = models.RoundStatusActive
	campaign, err := camapign_service.CreateCampaign(c, createRequest)
	if err != nil {
		c.JSON(400, models.ResponseError{Detail: "Failed to create campaign : " + err.Error()})
		return
	}
	c.JSON(200, models.ResponseSingle[*models.Campaign]{Data: campaign})
}

// UpdateCampaign godoc
// @Summary Update a campaign
// @Description Update a campaign
// @Produce  json
// @Success 200 {object} models.ResponseSingle[models.Campaign]
// @Router /campaign/{id} [post]
// @Security ApiKeyAuth
// @Tags Campaign
// @Param id path string true "The campaign ID"
// @Param campaignRequest body services.CampaignUpdateRequest true "The campaign request"
func UpdateCampaign(c *gin.Context, sess *cache.Session) {
	campaignId := c.Param("campaignId")
	if campaignId == "" {
		c.JSON(400, models.ResponseError{Detail: "Invalid request : Campaign ID is required"})
	}
	updateRequest := &services.CampaignUpdateRequest{}
	err := c.ShouldBindBodyWithJSON(updateRequest)
	if err != nil {
		c.JSON(400, models.ResponseError{Detail: "Invalid request : " + err.Error()})
		return
	}
	campaign_service := services.NewCampaignService()
	campaign, err := campaign_service.UpdateCampaign(c, models.IDType(campaignId), updateRequest)
	if err != nil {
		c.JSON(400, models.ResponseError{Detail: "Failed to update campaign : " + err.Error()})
		return
	}
	c.JSON(200, models.ResponseSingle[*models.Campaign]{Data: campaign})
}
func GetCampaignResultSummary(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Hello, World!",
	})
}
func GetCampaignResults(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Hello, World!",
	})
}
func GetCampaignSubmissions(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Hello, World!",
	})
}
func GetNextSubmission(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Hello, World!",
	})
}

func ApproveCampaign(c *gin.Context) {

	c.JSON(200, gin.H{
		"message": "Hello, World!",
	})
}
func ImportEvaluationFromFountain(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Hello, World!",
	})
}

// UpdateCampaignStatus godoc
// @Summary Update a campaign status
// @Description Update a campaign status
// @Produce  json
// @Success 200 {object} models.ResponseSingle[models.Campaign]
// @Router /campaign/{campaignId}/status [post]
// @Tags Campaign
// @Param campaignId path string true "The campaign ID"
// @Param campaignUpdateStatusRequest body models.CampaignUpdateStatusRequest true "The campaign update status request"
// @Security ApiKeyAuth
// @Error 400 {object} models.ResponseError
// @Error 403 {object} models.ResponseError
// @Error 404 {object} models.ResponseError
func UpdateCampaignStatus(c *gin.Context, sess *cache.Session) {
	campaignId := c.Param("campaignId")
	if campaignId == "" {
		c.JSON(400, models.ResponseError{Detail: "Invalid request : Campaign ID is required"})
		return
	}
	updateRequest := &models.CampaignUpdateStatusRequest{}
	err := c.ShouldBindBodyWithJSON(updateRequest)
	if err != nil {
		c.JSON(400, models.ResponseError{Detail: "Invalid request : " + err.Error()})
		return
	}
	campaign_service := services.NewCampaignService()
	campaign, err := campaign_service.UpdateCampaignStatus(c, sess.UserID, models.IDType(campaignId), updateRequest.IsArchived)
	if err != nil {
		c.JSON(400, models.ResponseError{Detail: "Failed to update campaign : " + err.Error()})
		return
	}
	c.JSON(200, models.ResponseSingle[*models.Campaign]{Data: campaign})
}

// FetchCampaignStatistics godoc
// @Summary Fetch campaign statistics
// @Description Fetch campaign statistics
// @Produce  json
// @Success 200 {object} models.ResponseList[models.RoundStatisticsView]
// @Router /campaign/statistics [get]
// @Tags Campaign
// @Param CampaignFilter query models.CampaignFilter false "Filter the campaigns"
// @Security ApiKeyAuth
// @Error 400 {object} models.ResponseError
// @Error 403 {object} models.ResponseError
// @Error 404 {object} models.ResponseError
func FetchCampaignStatistics(c *gin.Context) {
	campaignList, err := listAllCampaigns(c)
	if err != nil {
		c.JSON(400, models.ResponseError{Detail: "Invalid request : " + err.Error()})
		return
	}
	if len(campaignList) == 0 {
		c.JSON(200, models.ResponseList[models.RoundStatisticsView]{Data: []models.RoundStatisticsView{}})
		return
	}
	roundIds := make([]string, 0, len(campaignList))
	for _, campaign := range campaignList {
		roundIds = append(roundIds, campaign.LatestRoundID.String())
	}
	campaignService := services.NewCampaignService()
	statistics, err := campaignService.FetchCampaignStatistics(c, roundIds)
	if err != nil {
		c.JSON(400, models.ResponseError{Detail: "Failed to fetch campaign statistics : " + err.Error()})
		return
	}
	c.JSON(200, models.ResponseList[models.RoundStatisticsView]{Data: statistics})
}
