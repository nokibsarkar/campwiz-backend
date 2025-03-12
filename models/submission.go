package models

import (
	"time"

	"gorm.io/datatypes"
)

type SubmissionListFilter struct {
	CampaignID    IDType `form:"campaignId"`
	RoundID       IDType `form:"roundId"`
	ParticipantID IDType `form:"participantId"`
	CommonFilter
}
type ArticleSubmission struct {
	Language   string `json:"language"`
	TotalBytes uint64 `json:"totalbytes" gorm:"default:0"`
	TotalWords uint64 `json:"totalwords" gorm:"default:0"`
	AddedBytes uint64 `json:"addedbytes" gorm:"default:0"`
	AddedWords uint64 `json:"addedwords" gorm:"default:0"`
}
type ImageSubmission struct {
	Width  uint64 `json:"width"`
	Height uint64 `json:"height"`
}
type AudioVideoSubmission struct {
	Duration uint64 `json:"duration"` // in milliseconds
	Bitrate  uint64 `json:"bitrate"`  // in kbps
	Size     uint64 `json:"size"`     // in bytes
}
type MediaSubmission struct {
	MediaType   MediaType      `json:"mediatype" gorm:"not null;default:'BITMAP'"`
	ThumbURL    string         `json:"thumburl"`
	ThumbWidth  uint64         `json:"thumbwidth"`
	ThumbHeight uint64         `json:"thumbheight"`
	License     string         `json:"license"`
	Description string         `json:"description"`
	CreditHTML  string         `json:"creditHTML"`
	Metadata    datatypes.JSON `json:"metadata" gorm:"type:json"`
	ImageSubmission
	AudioVideoSubmission
}
type Submission struct {
	SubmissionID IDType `json:"submissionId" gorm:"primaryKey"`
	Name         string `json:"title"`
	CampaignID   IDType `json:"campaignId" gorm:"null;index;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	URL          string `json:"url"`
	// The Average Score of the submission
	Score ScoreType `json:"score" gorm:"default:0"`
	// The Actual Author in the Wikimedia
	Author WikimediaUsernameType `json:"author"`
	// The User who submitted the article on behalf of the participant
	SubmittedByID      IDType     `json:"submittedById" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	ParticipantID      IDType     `json:"participantId" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	CurrentRoundID     IDType     `json:"currentRoundId" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	SubmittedAt        time.Time  `json:"submittedAt" gorm:"type:datetime"`
	Participant        User       `json:"-" gorm:"foreignKey:ParticipantID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Submitter          User       `json:"-" gorm:"foreignKey:SubmittedByID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Campaign           *Campaign  `json:"-"  gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	CreatedAtExternal  *time.Time `json:"createdAtServer"`
	CurrentRound       *Round     `json:"-" gorm:"foreignKey:CurrentRoundID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	DistributionTaskID *IDType    `json:"distributionTaskId" gorm:"null"`
	ImportTaskID       IDType     `json:"importTaskId" gorm:"null"`
	// The task that was used to distribute the submission to the juries
	DistributionTask *Task `json:"-" gorm:"foreignKey:DistributionTaskID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	// The task that was used to import the submission from the external source
	ImportTask *Task `json:"-" gorm:"foreignKey:ImportTaskID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	MediaSubmission
}
type SubmissionSelectID struct {
	SubmissionID IDType
}
