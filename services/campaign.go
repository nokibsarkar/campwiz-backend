package services

import (
	"fmt"
	"log"
	"nokib/campwiz/consts"
	"nokib/campwiz/models"
	"nokib/campwiz/repository"
	"nokib/campwiz/repository/cache"
	idgenerator "nokib/campwiz/services/idGenerator"
)

// WikimediaUsernameType is a type for jury user name

type CampaignService struct{}
type CampaignCreateRequest struct {
	models.CampaignWithWriteableFields
	CreatedByID  models.IDType                  `json:"-"`
	Coordinators []models.WikimediaUsernameType `json:"coordinators"`
	Organizers   []models.WikimediaUsernameType `json:"organizers"`
}
type CampaignUpdateRequest struct {
	CampaignCreateRequest
}

func NewCampaignService() *CampaignService {
	return &CampaignService{}
}

func (service *CampaignService) CreateCampaign(campaignRequest *CampaignCreateRequest) (*models.Campaign, error) {
	// if endDate.Before(time.Now()) {
	// 	return nil, fmt.Errorf("End date should be in the future")
	// }
	if campaignRequest.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if campaignRequest.StartDate.After(campaignRequest.EndDate) {
		return nil, fmt.Errorf("start date should be before end date")
	}
	user_repo := repository.NewUserRepository()
	campaign_repo := repository.NewCampaignRepository()
	// user_repo := repository.NewUserRepository()
	role_service := NewRoleService()
	conn, close, err := repository.GetDB()
	if err != nil {
		return nil, err
	}
	defer close()
	tx := conn.Begin()
	currentUser, err := user_repo.FindByID(tx, campaignRequest.CreatedByID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	// if !currentUser.Permission.HasPermission(consts.PermissionCreateCampaign) {
	// 	tx.Rollback()
	// 	return nil, fmt.Errorf("user does not have permission to create campaign")
	// }
	if campaignRequest.ProjectID == "" {
		log.Println("Project ID is not provided")
		if currentUser.LeadingProjectID == nil {
			tx.Rollback()
			return nil, fmt.Errorf("user is not leading any project. So project id is required")
		} else {
			campaignRequest.ProjectID = *currentUser.LeadingProjectID
		}
	} else {
		log.Println("Project ID is provided" + campaignRequest.ProjectID)
		// project id is provided
		if currentUser.LeadingProjectID == nil && !currentUser.Permission.HasPermission(consts.PermissionOtherProjectAccess) {
			tx.Rollback()
			return nil, fmt.Errorf("user is not leading any project and does not have permission to create campaign in other's project")
		} else if currentUser.LeadingProjectID != nil && *currentUser.LeadingProjectID != campaignRequest.ProjectID && !currentUser.Permission.HasPermission(consts.PermissionOtherProjectAccess) {
			tx.Rollback()
			return nil, fmt.Errorf("user does not have permission to create campaign in other's project")
		}
	}
	if *currentUser.LeadingProjectID != campaignRequest.ProjectID && !currentUser.Permission.HasPermission(consts.PermissionOtherProjectAccess) {
		tx.Rollback()
		return nil, fmt.Errorf("user does not have permission to create campaign in other's project")
	}
	campaign := &models.Campaign{
		CampaignID: idgenerator.GenerateID("c"),
		CampaignWithWriteableFields: models.CampaignWithWriteableFields{
			Name:        campaignRequest.Name,
			Description: campaignRequest.Description,
			StartDate:   campaignRequest.StartDate.UTC(),
			EndDate:     campaignRequest.EndDate.UTC(),
			Language:    campaignRequest.Language,
			Rules:       campaignRequest.Rules,
			Image:       campaignRequest.Image,
			ProjectID:   campaignRequest.ProjectID,
			IsPublic:    campaignRequest.IsPublic,
			Status:      models.RoundStatusActive,
		},
		CreatedByID: campaignRequest.CreatedByID,
	}

	err = campaign_repo.Create(tx.Preload("Roles"), campaign)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	_, _, err = role_service.FetchChangeRoles(tx, models.RoleTypeCoordinator, campaign.ProjectID, nil, &campaign.CampaignID, nil, campaignRequest.Coordinators)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	// _, err = role_service.FetchChangeRoles(tx, models.RoleTypeOrganizer, campaign.CampaignID, "", campaignRequest.Organizers)
	// if err != nil {
	// 	tx.Rollback()
	// 	return nil, err
	// }
	tx.Commit()
	return campaign, nil
}
func (service *CampaignService) GetAllCampaigns(query *models.CampaignFilter) []models.Campaign {
	conn, close, err := repository.GetDB()
	if err != nil {
		log.Println("Error: ", err)
		return []models.Campaign{}
	}
	defer close()
	campaign_repo := repository.NewCampaignRepository()

	campaigns, err := campaign_repo.ListAllCampaigns(conn, query)
	if err != nil {
		log.Println("Error: ", err)
		return []models.Campaign{}
	}
	return campaigns
}
func (service *CampaignService) ListPrivateCampaigns(sess *cache.Session, qry *models.CampaignFilter) []models.Campaign {
	q, close := repository.GetDBWithGen()
	defer close()
	campaigns := []models.Campaign{}
	stmt := q.Campaign.Where(q.Campaign.IsPublic.Not()).Join(q.Role, q.Role.CampaignID.EqCol(q.Campaign.CampaignID)).
		Where(q.Role.UserID.Eq(sess.UserID.String()))
	if qry.ProjectID != "" {
		stmt = stmt.Where(q.Campaign.ProjectID.Eq(qry.ProjectID.String()))
	}
	stmt = stmt.Group(q.Campaign.CampaignID).Limit(qry.Limit)
	stmt.Scan(&campaigns)
	return campaigns
}

type SingleCampaignQuery struct {
	IncludeRounds     bool `form:"includeRounds"`
	IncludeRoles      bool `form:"includeRoles"`
	IncludeProject    bool `form:"includeProject"`
	IncludeRoundRoles bool `form:"includeRoundRoles"`
}

func (service *CampaignService) GetCampaignByID(id models.IDType, query *SingleCampaignQuery) (*models.Campaign, error) {
	conn, close, err := repository.GetDB()
	if err != nil {
		log.Println("Error: ", err)
		return nil, err
	}
	defer close()
	if query != nil {
		if query.IncludeRounds {
			conn = conn.Preload("Rounds")
			if query.IncludeRoundRoles {
				conn = conn.Preload("Rounds.Roles").Preload("Rounds.Roles.User")
			}
		}
		if query.IncludeRoles {
			conn = conn.Preload("Roles").Preload("Roles.User")
		}
		if query.IncludeProject {
			conn = conn.Preload("Project")
		}
	}
	campaign_repo := repository.NewCampaignRepository()
	campaign, err := campaign_repo.FindByID(conn, id)
	if err != nil {
		log.Println("Error: ", err)
		return nil, err
	}
	return campaign, nil
}

func (service *CampaignService) UpdateCampaign(ID models.IDType, campaignRequest *CampaignUpdateRequest) (*models.Campaign, error) {
	conn, close, err := repository.GetDB()
	if err != nil {
		log.Println("Error: ", err)
		return nil, err
	}
	defer close()
	campaign_repo := repository.NewCampaignRepository()
	campaign, err := campaign_repo.FindByID(conn, ID)
	if err != nil {
		return nil, err
	}
	roleService := NewRoleService()
	campaign.Name = campaignRequest.Name
	campaign.Description = campaignRequest.Description
	campaign.StartDate = campaignRequest.StartDate
	campaign.EndDate = campaignRequest.EndDate
	// campaign.Language = campaignRequest.Language
	campaign.Rules = campaignRequest.Rules
	campaign.Image = campaignRequest.Image
	tx := conn.Begin()
	err = campaign_repo.Update(tx, campaign)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	// _, err = roleService.FetchChangeRoles(tx, models.RoleTypeOrganizer, campaign.CampaignID, "", campaignRequest.Organizers)
	// if err != nil {
	// 	tx.Rollback()
	// 	return nil, err
	// }
	_, _, err = roleService.FetchChangeRoles(tx, models.RoleTypeCoordinator, campaign.ProjectID, nil, &campaign.CampaignID, nil, campaignRequest.Coordinators)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()
	return campaign, nil
}
func (service *CampaignService) UpdateCampaignStatus(usrId models.IDType, ID models.IDType, IsArchived bool) (*models.Campaign, error) {
	conn, close, err := repository.GetDB()
	if err != nil {
		log.Println("Error: ", err)
		return nil, err
	}
	defer close()
	user_repo := repository.NewUserRepository()
	campaign_repo := repository.NewCampaignRepository()
	campaign, err := campaign_repo.FindByID(conn, ID)
	if err != nil {
		log.Println("Error: ", err)
		return nil, err
	}
	currentUser, err := user_repo.FindByID(conn, usrId)
	if err != nil {
		log.Println("Error: ", err)
		return nil, err
	}
	if !currentUser.Permission.HasPermission(consts.PermissionOtherProjectAccess) &&
		currentUser.LeadingProjectID == nil {
		log.Println("Error: ", err)
		return nil, fmt.Errorf("user does not have permission to archive cross-project campaigns")
	}
	if currentUser.LeadingProjectID != nil && *currentUser.LeadingProjectID != campaign.ProjectID {
		log.Println("Error: ", err)
		return nil, fmt.Errorf("user is not leading this project")
	}
	if !currentUser.Permission.HasPermission(consts.PermissionUpdateCampaignStatus) {
		log.Println("Error: ", err)
		return nil, fmt.Errorf("user does not have permission to update campaign status")
	}
	if IsArchived {
		err = campaign_repo.ArchiveCampaign(conn, ID)
	} else {
		err = campaign_repo.UnArchiveCampaign(conn, ID)
	}
	if err != nil {
		log.Println("Error: ", err)
		return nil, err
	}
	campaign, err = campaign_repo.FindByID(conn.Unscoped(), ID)
	if err != nil {
		log.Println("Error: ", err)
		return nil, err
	}
	return campaign, nil
}
