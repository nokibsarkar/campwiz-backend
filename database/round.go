package database

import (
	"time"

	"gorm.io/gorm"
)

// These are the restrictions that are applied to the articles that are submitted to the campaign
type RoundCommonRestrictions struct {
	AllowJuryToParticipate bool   `json:"allowJuryToParticipate"`
	AllowMultipleJudgement bool   `json:"allowMultipleJudgement"`
	SecretBallot           bool   `json:"secretBallot"`
	Blacklist              string `json:"blacklist"`
}
type RoundStatus string

const (
	// RoundStatusPending is the status when the round is created but not yet approved by the admin
	RoundStatusPending RoundStatus = "PENDING"
	// RoundStatusImporting is the status when the round is approved and importing images from commons is in progress
	RoundStatusImporting RoundStatus = "IMPORTING"
	// RoundStatusDistributing is the status when the images are imported and being distributed among juries
	RoundStatusDistributing RoundStatus = "DISTRIBUTING"
	// RoundStatusEvaluating is the status when the images are distributed and juries are evaluating the images.
	RoundStatusEvaluating RoundStatus = "EVALUATING"
	// RoundStatusRejected is the status when the round is rejected by the admin
	RoundStatusRejected RoundStatus = "REJECTED"
	// RoundStatusCancelled is the status when the round is cancelled by the admin or the creator
	RoundStatusCancelled RoundStatus = "CANCELLED"
	// RoundStatusPaused is the status when the round is paused by the admin
	RoundStatusPaused RoundStatus = "PAUSED"
	// RoundStatusScheduled is the status when the round is scheduled to start at a later time
	RoundStatusScheduled RoundStatus = "SCHEDULED"
	// RoundStatusActive is the status when the round is active and juries are evaluating the images
	RoundStatusActive RoundStatus = "ACTIVE"
	// RoundStatusCompleted is the status when the round is completed and the results are ready
	RoundStatusCompleted RoundStatus = "COMPLETED"
)

// These are the restrictions that are applied to the audio and video that are submitted to the campaign
type RoundAudioRestrictions struct {
	AudioMinimumDurationMilliseconds int `json:"audioMinimumDurationMilliseconds" gorm:"default:0"`
	AudioMinimumSizeBytes            int `json:"audioMinimumSizeBytes" gorm:"default:0"`
}
type RoundVideoRestrictions struct {
	VideoMinimumDurationMilliseconds int `json:"videoMinimumDurationMilliseconds" gorm:"default:0"`
	VideoMinimumSizeBytes            int `json:"videoMinimumSizeBytes" gorm:"default:0"`
	VideoMinimumResolution           int `json:"videoMinimumResolution" gorm:"default:0"`
}

// These are the restrictions that are applied to the images that are submitted to the campaign
type RoundImageRestrictions struct {
	ImageMinimumResolution int `json:"imageMinimumResolution" gorm:"default:0"`
	ImageMinimumSizeBytes  int `json:"imageMinimumSizeBytes" gorm:"default:0"`
}
type RoundArticleRestrictions struct {
	MaximumSubmissionOfSameArticle int  `json:"articleMaximumSubmissionOfSameArticle"`
	ArticleAllowExpansions         bool `json:"articleAllowExpansions"`
	ArticleAllowCreations          bool `json:"articleAllowCreations"`
	ArticleMinimumTotalBytes       int  `json:"articleMinimumTotalBytes"`
	ArticleMinimumTotalWords       int  `json:"articleMinimumTotalWords"`
	ArticleMinimumAddedBytes       int  `json:"articleMinimumAddedBytes"`
	ArticleMinimumAddedWords       int  `json:"articleMinimumAddedWords"`
}
type RoundMediaRestrictions struct {
	RoundImageRestrictions
	RoundAudioRestrictions
	RoundVideoRestrictions
}

// these are the restrictions that are applied to
type RoundRestrictions struct {
	RoundCommonRestrictions
	RoundMediaRestrictions
	RoundArticleRestrictions
	AllowedMediaTypes MediaTypeSet `json:"allowedMediaTypes" gore:"type:text;not null;default:'ARTICLE'"`
}
type RoundWritable struct {
	Name             string         `json:"name"`
	Description      string         `json:"description" gorm:"type:text"`
	StartDate        time.Time      `json:"startDate" gorm:"type:datetime"`
	EndDate          time.Time      `json:"endDate" gorm:"type:datetime"`
	IsOpen           bool           `json:"isOpen" gorm:"default:true"`
	IsPublic         bool           `json:"isPublic" gorm:"default:false"`
	DependsOnRoundID *string        `json:"dependsOnRoundId" gorm:"default:null"`
	DependsOnRound   *Round         `json:"-" gorm:"foreignKey:DependsOnRoundID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Serial           int            `json:"serial" gorm:"default:0"`
	Type             EvaluationType `json:"type"`
	RoundRestrictions
}
type Round struct {
	RoundID                  IDType      `json:"roundId" gorm:"primaryKey"`
	CampaignID               IDType      `json:"campaignId" gorm:"index;cascade:OnUpdate,OnDelete"`
	ProjectID                IDType      `json:"projectId" gorm:"index;cascade:OnUpdate,OnDelete"`
	CreatedAt                *time.Time  `json:"createdAt" gorm:"-<-:create"`
	CreatedByID              IDType      `json:"createdById" gorm:"index"`
	TotalSubmissions         int         `json:"totalSubmissions" gorm:"default:0"`
	Status                   RoundStatus `json:"status"`
	Campaign                 *Campaign   `json:"-"  gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Creator                  *User       `json:"-" gorm:"foreignKey:CreatedByID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	LatestDistributionTaskID *IDType     `json:"latestTaskId" gorm:"default:null"`
	RoundWritable
	Roles []Role `json:"roles"`
	// Project Project `json:"-" gorm:"foreignKey:ProjectID"`
}
type RoundFilter struct {
	CampaignID IDType      `form:"campaignId"`
	Status     RoundStatus `form:"status"`
	CommonFilter
}
type RoundRepository struct{}

func NewRoundRepository() *RoundRepository {
	return &RoundRepository{}
}
func (r *RoundRepository) Create(conn *gorm.DB, round *Round) (*Round, error) {
	result := conn.Create(round)
	if result.Error != nil {
		return nil, result.Error
	}
	return round, nil
}
func (r *RoundRepository) Update(conn *gorm.DB, round *Round) (*Round, error) {
	result := conn.Save(round)
	if result.Error != nil {
		return nil, result.Error
	}
	return round, nil
}
func (r *RoundRepository) FindAll(conn *gorm.DB, filter *RoundFilter) ([]Round, error) {
	var rounds []Round
	where := &Round{}
	if filter != nil {
		if filter.CampaignID != "" {
			where.CampaignID = filter.CampaignID
		}
	}
	stmt := conn.Where(where)
	if filter.Limit > 0 {
		stmt = stmt.Limit(filter.Limit)
	}
	result := stmt.Find(&rounds)
	return rounds, result.Error
}
func (r *RoundRepository) FindByID(conn *gorm.DB, id IDType) (*Round, error) {
	round := &Round{}
	where := &Round{RoundID: IDType(id)}
	result := conn.First(round, where)
	return round, result.Error
}
