package database

import (
	"time"

	"gorm.io/gorm"
)

type Project struct {
	ProjectID IDType  `json:"projectId" gorm:"primaryKey"`
	Name      string  `json:"name"`
	LogoURL   *string `json:"logoUrl" gorm:"null"`
	// The URL of the project's website
	Link        *string    `json:"url"`
	CreatedByID IDType     `json:"createdById" gorm:"index;not null"`
	CreatedAt   *time.Time `json:"createdAt" gorm:"-<-:create;autoCreateTime"`
}
type ProjectExtended struct {
	Project
	Leads []WikimediaUsernameType `json:"projectLeads"`
}
type ProjectRequest struct {
	ProjectID    IDType                  `json:"projectId"`
	Name         string                  `json:"name" binding:"required"`
	LogoURL      *string                 `json:"logoUrl"`
	Link         *string                 `json:"url"`
	CreatedByID  IDType                  `json:"-"`
	ProjectLeads []WikimediaUsernameType `json:"projectLeads"  gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Campaigns    []*Campaign             `json:"-"  gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
type ProjectRepository struct{}

func NewProjectRepository() *ProjectRepository {
	return &ProjectRepository{}
}
func (r *ProjectRepository) CreateProject(tx *gorm.DB, project *Project) error {
	result := tx.Create(project)
	return result.Error
}
func (r *ProjectRepository) FindProjectByID(tx *gorm.DB, projectID IDType) (*Project, error) {
	project := &Project{}
	result := tx.First(project, &Project{ProjectID: projectID})
	return project, result.Error
}
func (r *ProjectRepository) UpdateProject(tx *gorm.DB, project *Project) error {
	result := tx.Updates(project)
	return result.Error
}
