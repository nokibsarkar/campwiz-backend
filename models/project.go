package models

import "time"

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
type ProjectFilter struct {
	IncludeRoles         bool     `form:"includeRoles"`
	IncludeOtherProjects bool     `form:"includeOtherProjects"`
	IDs                  []IDType `form:"ids"`
	CommonFilter
}
