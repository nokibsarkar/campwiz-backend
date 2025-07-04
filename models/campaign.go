package models

import (
	"nokib/campwiz/consts"
	"time"

	"gorm.io/gorm"
)

type CampaignType string

const (
	// Commons Campaigns are campaigns that are used in Wikimedia Commons
	CampaignTypeCommons CampaignType = "commons"
	// Wikipedia Campaigns are campaigns that are used in Wikipedia
	CampaignTypeWikipedia CampaignType = "wikipedia"
	// Wikidata Campaigns are campaigns that are used in Wikidata
	CampaignTypeWikidata CampaignType = "wikidata"
	// Categorization is the type of campaigns whre categories are added or removed from submissions
	CampaignTypeCategorization CampaignType = "categorization"
	// Reference Campaigns are campaigns that are used to add references to articles
	CampaignTypeReference CampaignType = "reference"
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
	Tags      []string    `json:"tags,omitempty" gorm:"-"`
	// The type of the campaign, it should be one of the CampaignType constants
	CampaignType CampaignType `json:"campaignType" gorm:"type:ENUM('commons', 'wikipedia', 'wikidata', 'categorization', 'reference');default:'commons';not null;index" binding:"required"`
}
type Campaign struct {
	// A unique identifier for the campaign, it should be custom defined
	CampaignID IDType `gorm:"primaryKey;column:campaign_id;type:varchar(255)" json:"campaignId"`
	// The time the campaign was created, it would be set automatically
	CreatedAt   *time.Time `json:"createdAt" gorm:"-<-:create"`
	CreatedByID IDType     `json:"createdById" gorm:"index"`
	CampaignWithWriteableFields
	CampaignTags  []Tag           `json:"-"`
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
type SortOrder string

const (
	SortOrderAsc  SortOrder = "asc"
	SortOrderDesc SortOrder = "desc"
)

type CampaignFilter struct {
	IDs []IDType `form:"ids,omitEmpty"`
	// Whether the campaign is hidden from the public list
	// If isHidden is true, then projectID is required
	IsHidden *bool `form:"isHidden"`
	/*
		This projectID is the project that campaigns belong to.then ProjectID is required
		If the person is not an admin, then the project ID must match the project ID of the user
	*/
	ProjectID IDType    `form:"projectId"`
	SortOrder SortOrder `form:"sortOrder"`
	// Whether the campaign is closed (result have been finalized)
	IsClosed *bool `form:"isClosed"`
	// Tags are used to filter campaigns by tags
	// If tags are provided, then only campaigns with the given tags will be returned
	Tags []string `form:"tags,comma"` // Tags are used to filter campaigns by tags
	CommonFilter
}
type SingleCampaignFilter struct {
	IncludeRounds          bool `form:"includeRounds"`
	IncludeRoundRoles      bool `form:"includeRoundRoles"`
	IncludeRoundRolesUsers bool `form:"includeRoundRolesUsers"`
	IncludeProject         bool `form:"includeProject"`
}
type CampaignUpdateStatusRequest struct {
	// The status of the campaign
	IsArchived bool `json:"isArchived"`
}

func (t CampaignType) Value() (any, error) {
	return t.String(), nil
}
func (t CampaignType) String() string {
	if t == "" {
		return "commons"
	}
	return string(t)
}
