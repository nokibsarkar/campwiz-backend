package models

import (
	"nokib/campwiz/consts"
	"time"

	"gorm.io/gorm"
)

type CampaignWithWriteableFields struct {
	Name        string          `json:"name" binding:"required"`
	Description string          `json:"description"`
	StartDate   time.Time       `json:"startDate" binding:"required"`
	EndDate     time.Time       `json:"endDate" binding:"required"`
	Language    consts.Language `json:"language" binding:"required"`
	Rules       string          `json:"rules"`
	Image       string          `json:"image"`
	// Whether the campaign is shown in the public list
	IsPublic  bool        `json:"isPublic"`
	ProjectID IDType      `json:"projectId"`
	Status    RoundStatus `json:"status"`
}
type Campaign struct {
	// A unique identifier for the campaign, it should be custom defined
	CampaignID IDType `gorm:"primaryKey" json:"campaignId"`
	// The time the campaign was created, it would be set automatically
	CreatedAt   *time.Time `json:"createdAt" gorm:"-<-:create"`
	CreatedByID IDType     `json:"createdById" gorm:"index"`
	CampaignWithWriteableFields
	CreatedBy     *User           `json:"-" gorm:"foreignKey:CreatedByID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	Roles         []Role          `json:"-" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Rounds        []Round         `json:"rounds" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Project       *Project        `json:"project" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	LatestRoundID *IDType         `json:"latestRoundId" gorm:"default:null"`
	LatestRound   *Round          `json:"-" gorm:"foreignKey:LatestRoundID;constraint:OnUpdate:SET NULL,OnDelete:SET NULL"`
	ArchivedAt    *gorm.DeletedAt `json:"archivedAt" gorm:"index;default:null"`
}
type CampaignExtended struct {
	Campaign
	Coordinators []WikimediaUsernameType `json:"coordinators"`
}

type CampaignFilter struct {
	IDs []IDType `form:"ids,omitEmpty"`
	// Whether the campaign is hidden from the public list
	// If isHidden is true, then projectID is required
	IsHidden *bool `form:"isHidden"`
	/*
		This projectID is the project that campaigns belong to.then ProjectID is required
		If the person is not an admin, then the project ID must match the project ID of the user
	*/
	ProjectID IDType `form:"projectId"`
	// Whether the campaign is closed (result have been finalized)
	IsClosed *bool `form:"isClosed"`
	CommonFilter
}
type SingleCampaaignFilter struct {
	IncludeRounds          bool `form:"includeRounds"`
	IncludeRoundRoles      bool `form:"includeRoundRoles"`
	IncludeRoundRolesUsers bool `form:"includeRoundRolesUsers"`
	IncludeProject         bool `form:"includeProject"`
}
