package repository

import (
	"nokib/campwiz/models"

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
func (r *ProjectRepository) ListProjects(tx *gorm.DB, filter *models.ProjectFilter) ([]models.Project, error) {
	projects := []models.Project{}
	result := tx.Find(&projects)
	return projects, result.Error
}
