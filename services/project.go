package services

import (
	"fmt"
	"log"
	"nokib/campwiz/models"
	"nokib/campwiz/repository"
	idgenerator "nokib/campwiz/services/idGenerator"

	"gorm.io/gorm"
)

type ProjectService struct{}

func NewProjectService() *ProjectService {
	return &ProjectService{}
}
func (p *ProjectService) GetProjectByID(id models.IDType, includeProjectLeads bool) (*models.ProjectExtended, error) {
	project_repo := repository.NewProjectRepository()
	conn, close := repository.GetDB()
	defer close()
	project, err := project_repo.FindProjectByID(conn, id)
	if err != nil {
		return nil, err
	}
	px := &models.ProjectExtended{Project: *project}
	if includeProjectLeads {
		user_repo := repository.NewUserRepository()
		leads, err := user_repo.FindProjectLeads(conn, &id)
		if err != nil {
			return nil, err
		}
		px.Leads = []models.WikimediaUsernameType{}
		for _, lead := range leads {
			px.Leads = append(px.Leads, lead.Username)
		}
	}
	return px, nil
}
func (p *ProjectService) CreateProject(projectReq *models.ProjectRequest, includeProjectLeads bool) (*models.ProjectExtended, error) {
	project_repo := repository.NewProjectRepository()
	conn, close := repository.GetDB()
	defer close()
	tx := conn.Begin()
	project := &models.Project{
		ProjectID:   projectReq.ProjectID,
		Name:        projectReq.Name,
		LogoURL:     projectReq.LogoURL,
		Link:        projectReq.Link,
		CreatedByID: projectReq.CreatedByID,
	}
	err := project_repo.CreateProject(tx, project)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	currentLeads, err := p.AssignProjectLead(tx, projectReq)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	res := tx.Commit()
	if res.Error != nil {
		return nil, res.Error
	}
	px := &models.ProjectExtended{Project: *project}
	px.Leads = []models.WikimediaUsernameType{}
	if includeProjectLeads {
		for _, lead := range currentLeads {
			px.Leads = append(px.Leads, lead.Username)
		}
	}
	return px, nil
}
func (p *ProjectService) AssignProjectLead(tx *gorm.DB, projectReq *models.ProjectRequest) (currentLeads []models.User, err error) {
	user_repo := repository.NewUserRepository()
	username2RandomId := map[models.WikimediaUsernameType]models.IDType{}
	for _, username := range projectReq.ProjectLeads {
		username2RandomId[username] = idgenerator.GenerateID("u")
	}
	username2RealIds, err := user_repo.EnsureExists(tx, username2RandomId)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	previousLeads, err := user_repo.FindProjectLeads(tx, &projectReq.ProjectID)
	if err != nil {
		return nil, err
	}
	previousLeadsSet := map[models.IDType]models.User{}
	for _, lead := range previousLeads {
		previousLeadsSet[lead.UserID] = lead
	}

	for _, realId := range username2RealIds {
		if _, ok := previousLeadsSet[realId]; ok {
			log.Printf("User with ID %v is already a lead of project %v\n", realId, projectReq.ProjectID)
			currentLeads = append(currentLeads, previousLeadsSet[realId])
			delete(previousLeadsSet, realId)
			continue
		}
		user, err := user_repo.FindByID(tx, realId)
		if err != nil {
			return nil, err
		}
		if user.LeadingProjectID != nil {
			return nil, fmt.Errorf(repository.ErrUserAlreadyLeadsProject, user.Username, *user.LeadingProjectID)
		}
		res := tx.Updates(&models.User{UserID: realId, LeadingProjectID: &projectReq.ProjectID})
		if res.Error != nil {
			return nil, res.Error
		}
		currentLeads = append(currentLeads, *user)
	}
	for _, lead := range previousLeadsSet {
		log.Printf("Removing user with ID %v from project %v\n", lead.UserID, projectReq.ProjectID)
		res := tx.Model(lead).Update("leading_project_id", nil)
		if res.Error != nil {
			return nil, res.Error
		}
	}
	if len(currentLeads) == 0 {
		return nil, fmt.Errorf(repository.ErrNoProjectLeads)
	}
	return currentLeads, nil
}
func (p *ProjectService) UpdateProject(projectReq *models.ProjectRequest) (*models.ProjectExtended, error) {
	project_repo := repository.NewProjectRepository()
	conn, close := repository.GetDB()
	defer close()
	tx := conn.Begin()
	project, err := project_repo.FindProjectByID(tx, projectReq.ProjectID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	project.Name = projectReq.Name
	project.LogoURL = projectReq.LogoURL
	project.Link = projectReq.Link
	err = project_repo.UpdateProject(tx.Preload("Roles"), project)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	currentLeads, err := p.AssignProjectLead(tx, projectReq)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	res := tx.Commit()
	if res.Error != nil {
		return nil, res.Error
	}
	px := &models.ProjectExtended{Project: *project}
	px.Leads = []models.WikimediaUsernameType{}
	for _, lead := range currentLeads {
		px.Leads = append(px.Leads, lead.Username)
	}
	return px, nil
}
