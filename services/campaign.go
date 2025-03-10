package services

import (
	"fmt"
	"nokib/campwiz/consts"
	"nokib/campwiz/database"
	idgenerator "nokib/campwiz/services/idGenerator"
)

// WikimediaUsernameType is a type for jury user name

type CampaignService struct{}
type CampaignCreateRequest struct {
	database.CampaignWithWriteableFields
	CreatedByID  database.IDType                  `json:"-"`
	Coordinators []database.WikimediaUsernameType `json:"coordinators"`
	Organizers   []database.WikimediaUsernameType `json:"organizers"`
}
type CampaignUpdateRequest struct {
	CampaignCreateRequest
}

func NewCampaignService() *CampaignService {
	return &CampaignService{}
}

func (service *CampaignService) CreateCampaign(campaignRequest *CampaignCreateRequest) (*database.Campaign, error) {
	// if endDate.Before(time.Now()) {
	// 	return nil, fmt.Errorf("End date should be in the future")
	// }
	if campaignRequest.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if campaignRequest.StartDate.After(campaignRequest.EndDate) {
		return nil, fmt.Errorf("start date should be before end date")
	}
	user_repo := database.NewUserRepository()
	campaign_repo := database.NewCampaignRepository()
	// user_repo := database.NewUserRepository()
	role_service := NewRoleService()
	conn, close := database.GetDB()
	defer close()
	tx := conn.Begin()
	currentUser, err := user_repo.FindByID(tx, campaignRequest.CreatedByID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	if currentUser.Permission.HasPermission(consts.PermissionCreateCampaign) == false {
		tx.Rollback()
		return nil, fmt.Errorf("user does not have permission to create campaign")
	}
	if currentUser.LeadingProjectID == nil {
		tx.Rollback()
		return nil, fmt.Errorf("user is not leading any project")
	}
	campaign := &database.Campaign{
		CampaignID: idgenerator.GenerateID("c"),
		CampaignWithWriteableFields: database.CampaignWithWriteableFields{
			Name:        campaignRequest.Name,
			Description: campaignRequest.Description,
			StartDate:   campaignRequest.StartDate.UTC(),
			EndDate:     campaignRequest.EndDate.UTC(),
			Language:    campaignRequest.Language,
			Rules:       campaignRequest.Rules,
			Image:       campaignRequest.Image,
			ProjectID:   *currentUser.LeadingProjectID,
			IsPublic:    campaignRequest.IsPublic,
		},
		CreatedByID: campaignRequest.CreatedByID,
	}

	err = campaign_repo.Create(tx.Preload("Roles"), campaign)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	_, err = role_service.FetchChangeRoles(tx, database.RoleTypeCoordinator, campaign.ProjectID, nil, &campaign.CampaignID, nil, campaignRequest.Coordinators)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	// _, err = role_service.FetchChangeRoles(tx, database.RoleTypeOrganizer, campaign.CampaignID, "", campaignRequest.Organizers)
	// if err != nil {
	// 	tx.Rollback()
	// 	return nil, err
	// }
	tx.Commit()
	return campaign, nil
}
func (service *CampaignService) GetAllCampaigns(query *database.CampaignFilter) []database.Campaign {
	conn, close := database.GetDB()
	defer close()
	campaign_repo := database.NewCampaignRepository()

	campaigns, err := campaign_repo.ListAllCampaigns(conn, query)
	if err != nil {
		fmt.Println("Error: ", err)
		return []database.Campaign{}
	}
	return campaigns
}
func (service *CampaignService) GetCampaignByID(id database.IDType) (*database.Campaign, error) {
	conn, close := database.GetDB()
	defer close()
	campaign_repo := database.NewCampaignRepository()
	campaign, err := campaign_repo.FindByID(conn, id)
	if err != nil {
		fmt.Println("Error: ", err)
		return nil, err
	}
	return campaign, nil
}

func (service *CampaignService) UpdateCampaign(ID database.IDType, campaignRequest *CampaignUpdateRequest) (*database.Campaign, error) {
	conn, close := database.GetDB()
	defer close()
	campaign_repo := database.NewCampaignRepository()
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
	// _, err = roleService.FetchChangeRoles(tx, database.RoleTypeOrganizer, campaign.CampaignID, "", campaignRequest.Organizers)
	// if err != nil {
	// 	tx.Rollback()
	// 	return nil, err
	// }
	_, err = roleService.FetchChangeRoles(tx, database.RoleTypeCoordinator, campaign.ProjectID, nil, &campaign.CampaignID, nil, campaignRequest.Coordinators)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()
	return campaign, nil
}
