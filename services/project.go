package services

import (
	"nokib/campwiz/database"
	idgenerator "nokib/campwiz/services/idGenerator"
)

type ProjectService struct{}

func NewProjectService() *ProjectService {
	return &ProjectService{}
}
func (p *ProjectService) GetProjectByID(id database.IDType) (*database.Project, error) {
	project_repo := database.NewProjectRepository()
	conn, close := database.GetDB()
	defer close()
	return project_repo.FindProjectByID(conn, id)
}
func (p *ProjectService) CreateProject(projectReq *database.ProjectRequest) (*database.Project, error) {
	project_repo := database.NewProjectRepository()
	conn, close := database.GetDB()
	defer close()
	tx := conn.Begin()
	project := &database.Project{
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
	user_repo := database.NewUserRepository()
	username2RandomId := map[database.WikimediaUsernameType]database.IDType{}
	for _, username := range projectReq.ProjectLeads {
		username2RandomId[username] = idgenerator.GenerateID("u")
	}
	_, err = user_repo.EnsureExists(tx, username2RandomId)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	role_service := NewRoleService()
	_, err = role_service.FetchChangeRoles(tx, database.RoleTypeProjectLead, project.ProjectID, &project.ProjectID, nil, nil, projectReq.ProjectLeads)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	res := tx.Commit()
	if res.Error != nil {
		return nil, res.Error
	}
	return project, nil
}
func (p *ProjectService) UpdateProject(projectReq *database.ProjectRequest) (*database.Project, error) {
	project_repo := database.NewProjectRepository()
	conn, close := database.GetDB()
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
	user_repo := database.NewUserRepository()
	username2RandomId := map[database.WikimediaUsernameType]database.IDType{}
	for _, username := range projectReq.ProjectLeads {
		username2RandomId[username] = idgenerator.GenerateID("u")
	}
	_, err = user_repo.EnsureExists(tx, username2RandomId)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	role_service := NewRoleService()
	_, err = role_service.FetchChangeRoles(tx, database.RoleTypeProjectLead, project.ProjectID, &project.ProjectID, nil, nil, projectReq.ProjectLeads)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	res := tx.Commit()
	if res.Error != nil {
		return nil, res.Error
	}
	return project, nil
}
