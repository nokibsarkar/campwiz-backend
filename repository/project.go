package repository

import (
	"nokib/campwiz/models"
	"nokib/campwiz/query"

	"gorm.io/gorm"
)

type ProjectRepository struct{}

func NewProjectRepository() *ProjectRepository {
	return &ProjectRepository{}
}
func (r *ProjectRepository) CreateProject(tx *gorm.DB, project *models.Project) error {
	result := tx.Create(project)
	return result.Error
}
func (r *ProjectRepository) FindProjectByID(tx *gorm.DB, projectID models.IDType) (*models.Project, error) {
	project := &models.Project{}
	result := tx.First(project, &models.Project{ProjectID: projectID})
	return project, result.Error
}
func (r *ProjectRepository) UpdateProject(tx *gorm.DB, project *models.Project) error {
	result := tx.Updates(project)
	return result.Error
}
func (r *ProjectRepository) ListProjects(tx *gorm.DB, filter *models.ProjectFilter) ([]models.ProjectExtended, error) {
	projects := []models.Project{}
	q := query.Use(tx)
	stmt := q.Project.Select(q.Project.ALL)
	if filter != nil {

		if filter.IDs != nil {
			strs := []string{}
			for _, id := range filter.IDs {
				strs = append(strs, string(id))
			}
			stmt = stmt.Where(q.Project.ProjectID.In(strs...))
		}
		if filter.Limit > 0 {
			stmt = stmt.Limit(filter.Limit)
		}
		if filter.ContinueToken != "" {
			stmt = stmt.Where(q.Project.ProjectID.Gt(filter.ContinueToken))
		}
		if filter.PreviousToken != "" {
			stmt = stmt.Where(q.Project.ProjectID.Lt(filter.PreviousToken))
		}
	}
	err := stmt.Scan(&projects)
	pxs := []models.ProjectExtended{}
	if filter != nil && filter.IncludeRoles {
		for _, p := range projects {
			// add all the leades
			users, err := q.User.Select(q.User.Username).Where(q.User.LeadingProjectID.Eq(p.ProjectID.String())).Find()
			if err != nil {
				return nil, err
			}
			px := models.ProjectExtended{Project: p}
			px.Leads = []models.WikimediaUsernameType{}
			for _, u := range users {
				px.Leads = append(px.Leads, u.Username)
			}
			pxs = append(pxs, px)
		}
	}
	return pxs, err
}
