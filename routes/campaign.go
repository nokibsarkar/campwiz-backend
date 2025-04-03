package routes

import (
	"nokib/campwiz/consts"
	"nokib/campwiz/models"
	"nokib/campwiz/repository"
	"nokib/campwiz/repository/cache"
	"nokib/campwiz/services"
	"time"

	"github.com/gin-gonic/gin"
)

// ListAllCampaigns godoc
// @Summary List all campaigns
// @Description get all campaigns
// @Produce  json
// @Success 200 {object} ResponseList[models.Campaign]
// @Router /campaign/ [get]
// @param CampaignFilter query models.CampaignFilter false "Filter the campaigns"
// @Tags Campaign
// @Error 400 {object} ResponseError
func ListAllCampaigns(c *gin.Context) {
	qry := &models.CampaignFilter{}
	err := c.ShouldBindQuery(qry)
	if err != nil {
		c.JSON(400, ResponseError{Detail: "Invalid request : " + err.Error()})
		return
	}
	campaignService := services.NewCampaignService()
	if qry.IsHidden != nil && *qry.IsHidden {
		sess := GetSession(c)
		if sess == nil {
			c.JSON(400, ResponseError{Detail: "Invalid request : User Must be logged in to get hidden campaigns"})
			return
		}
		if !sess.Permission.HasPermission(consts.PermissionOtherProjectAccess) {
			// user is not an admin
			if qry.ProjectID == "" {
				c.JSON(400, ResponseError{Detail: "Invalid request : Project ID is required when isHidden is true and user is not an admin"})
				return
			}
		}
		if qry.ProjectID != "" {
			// project id is provided
			// check if the user is allowed to access this project
			currentUser := GetCurrentUser(c)
			if currentUser == nil {
				c.JSON(400, ResponseError{Detail: "Invalid request : User not found"})
				return
			}
			if currentUser.LeadingProjectID == nil {
				c.JSON(400, ResponseError{Detail: "Invalid request : User is not leading any project"})
				return
			}
			if *currentUser.LeadingProjectID != qry.ProjectID {
				// the user is not an admin and the project ID does not match
				// cross project access is allowed only for jury and coordinators
				conn, close, err := repository.GetDB()
				if err != nil {
					c.JSON(400, ResponseError{Detail: "Database Error: " + err.Error()})
					return
				}
				defer close()
				userRepo := repository.NewUserRepository()
				roleFilter := &models.RoleFilter{ProjectID: qry.ProjectID, UserID: &currentUser.UserID}
				roles, err := userRepo.FetchRoles(conn, roleFilter)
				if err != nil {
					c.JSON(400, ResponseError{Detail: "Invalid request : " + err.Error()})
					return
				}
				hasRoles := false
				for _, role := range roles {
					if role.Type == models.RoleTypeCoordinator || role.Type == models.RoleTypeJury {
						if role.DeletedAt == nil {
							hasRoles = true
							break
						}
					}
				}
				if !hasRoles {
					c.JSON(400, ResponseError{Detail: "Invalid request : User does not have permission to access this project"})
					return
				}
			}
		}
	}
	campaignList := campaignService.GetAllCampaigns(qry)
	c.JSON(200, ResponseList[models.Campaign]{Data: campaignList})
}

func ListAllCampaignsV2(c *gin.Context, sess *cache.Session) {
	qry := &models.CampaignFilter{}
	err := c.ShouldBindQuery(qry)
	if err != nil {
		c.JSON(400, ResponseError{Detail: "Invalid request : " + err.Error()})
		return
	}
	campaignService := services.NewCampaignService()
	var campaignList []models.Campaign
	if qry.IsHidden != nil && *qry.IsHidden {
		sess := GetSession(c)
		if sess == nil {
			c.JSON(400, ResponseError{Detail: "Invalid request : User Must be logged in to get hidden campaigns"})
			return
		}
		// if !sess.Permission.HasPermission(consts.PermissionOtherProjectAccess) {
		// 	// user is not an admin
		// 	if qry.ProjectID == "" {
		// 		c.JSON(400, ResponseError{Detail: "Invalid request : Project ID is required when isHidden is true and user is not an admin"})
		// 		return
		// 	}
		// }
		// if qry.ProjectID != "" {
		// 	// project id is provided
		// 	// check if the user is allowed to access this project
		// 	currentUser := GetCurrentUser(c)
		// 	if currentUser == nil {
		// 		c.JSON(400, ResponseError{Detail: "Invalid request : User not found"})
		// 		return
		// 	}
		// 	if currentUser.LeadingProjectID == nil {
		// 		c.JSON(400, ResponseError{Detail: "Invalid request : User is not leading any project"})
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
		// 			c.JSON(400, ResponseError{Detail: "Invalid request : " + err.Error()})
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
		// 			c.JSON(400, ResponseError{Detail: "Invalid request : User does not have permission to access this project"})
		// 			return
		// 		}
		// 	}
		// }
		campaignList = campaignService.ListPrivateCampaigns(sess, qry)
	} else {
		campaignList = campaignService.GetAllCampaigns(qry)
	}
	c.JSON(200, ResponseList[models.Campaign]{Data: campaignList})
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
// @Success 200 {object} ResponseSingle[models.CampaignExtended]
// @Router /campaign/{campaignId} [get]
// @Param campaignId path string true "The campaign ID"
// @Param campaignQuery query services.SingleCampaignQuery false "The query for the campaign"
// @Tags Campaign
func GetSingleCampaign(c *gin.Context, sess *cache.Session) {
	campaignId := c.Param("campaignId")
	if campaignId == "" {
		c.JSON(400, ResponseError{Detail: "Invalid request : Campaign ID is required"})
	}
	q := &services.SingleCampaignQuery{}
	err := c.ShouldBindQuery(q)
	if err != nil {
		c.JSON(400, ResponseError{Detail: "Invalid request : " + err.Error()})
		return
	}
	campaign_service := services.NewCampaignService()
	campaign, err := campaign_service.GetCampaignByID(models.IDType(campaignId), q)
	if err != nil {
		c.JSON(404, ResponseError{Detail: "Failed to get campaign : " + err.Error()})
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
	c.JSON(200, ResponseSingle[*models.CampaignExtended]{Data: ex})
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
// @success 200 {object} models.Campaign
// @router /campaign/ [post]
func CreateCampaign(c *gin.Context, sess *cache.Session) {
	createRequest := &services.CampaignCreateRequest{}
	err := c.ShouldBindBodyWithJSON(createRequest)
	if err != nil {
		c.JSON(400, ResponseError{Detail: "Invalid request : " + err.Error()})
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
	campaign, err := camapign_service.CreateCampaign(createRequest)
	if err != nil {
		c.JSON(400, ResponseError{Detail: "Failed to create campaign : " + err.Error()})
		return
	}
	c.JSON(200, ResponseSingle[*models.Campaign]{Data: campaign})
}

// UpdateCampaign godoc
// @Summary Update a campaign
// @Description Update a campaign
// @Produce  json
// @Success 200 {object} ResponseSingle[models.Campaign]
// @Router /campaign/{id} [post]
// @Tags Campaign
// @Param id path string true "The campaign ID"
// @Param campaignRequest body services.CampaignUpdateRequest true "The campaign request"
func UpdateCampaign(c *gin.Context, sess *cache.Session) {
	campaignId := c.Param("campaignId")
	if campaignId == "" {
		c.JSON(400, ResponseError{Detail: "Invalid request : Campaign ID is required"})
	}
	updateRequest := &services.CampaignUpdateRequest{}
	err := c.ShouldBindBodyWithJSON(updateRequest)
	if err != nil {
		c.JSON(400, ResponseError{Detail: "Invalid request : " + err.Error()})
		return
	}
	campaign_service := services.NewCampaignService()
	campaign, err := campaign_service.UpdateCampaign(models.IDType(campaignId), updateRequest)
	if err != nil {
		c.JSON(400, ResponseError{Detail: "Failed to update campaign : " + err.Error()})
		return
	}
	c.JSON(200, ResponseSingle[*models.Campaign]{Data: campaign})
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
