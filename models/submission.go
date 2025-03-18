package models

import (
	"nokib/campwiz/models/types"
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
	Width      uint64 `json:"width"`
	Height     uint64 `json:"height"`
	Resolution uint64 `json:"resolution"`
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
	SubmissionID types.SubmissionIDType `json:"submissionId" gorm:"primaryKey"`
	Name         string                 `json:"title"`
	CampaignID   IDType                 `json:"campaignId" gorm:"null;index;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	URL          string                 `json:"url"`
	// The Average Score of the submission
	Score ScoreType `json:"score" gorm:"default:0"`
	// The Actual Author in the Wikimedia
	Author WikimediaUsernameType `json:"author"`
	// The User who submitted the article on behalf of the participant
	SubmittedByID      IDType     `json:"submittedById" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	ParticipantID      IDType     `json:"participantId" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	RoundID            IDType     `json:"currentRoundId" gorm:"index"`
	SubmittedAt        time.Time  `json:"submittedAt" gorm:"type:datetime"`
	Participant        User       `json:"-" gorm:"foreignKey:ParticipantID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Submitter          User       `json:"-" gorm:"foreignKey:SubmittedByID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Campaign           *Campaign  `json:"-"  gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	CreatedAtExternal  *time.Time `json:"createdAtServer"`
	Round              *Round     `json:"-" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	DistributionTaskID *IDType    `json:"distributionTaskId" gorm:"null"`
	ImportTaskID       IDType     `json:"importTaskId" gorm:"null"`
	// The number of times the submission has been assigned to the juries
	AssignmentCount uint `json:"assignmentCount" gorm:"default:0"`
	// The number of times the submission has been evaluated by the juries
	EvaluationCount uint `json:"evaluationCount" gorm:"default:0"`
	// The task that was used to distribute the submission to the juries
	DistributionTask *Task `json:"-" gorm:"foreignKey:DistributionTaskID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	// The task that was used to import the submission from the external source
	ImportTask *Task `json:"-" gorm:"foreignKey:ImportTaskID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	MediaSubmission
}
type SubmissionSelectID struct {
	SubmissionID types.SubmissionIDType
}
type SubmissionResult struct {
	SubmissionID    IDType    `json:"submissionId"`
	Name            string    `json:"name"`
	Author          string    `json:"author"`
	Score           ScoreType `json:"score"`
	EvaluationCount int       `json:"juryCount"`
	MediaType       MediaType `json:"type"`
}
type SubmissionResultQuery struct {
	CommonFilter
	Type []MediaType `form:"type" collectionFormat:"multi"`
}
type SubmissionStatistics struct {
	SubmissionID    types.SubmissionIDType
	AssignmentCount int
	EvaluationCount int
}
type SubmissionStatisticsFetcher interface {
	// SELECT COUNT(*) AS `AssignmentCount`, SUM(`score` IS NOT NULL) AS EvaluationCount, `submission_id`  FROM `evaluations`  WHERE `round_id` = @round_id GROUP BY `submission_id`
	FetchByRoundID(round_id string) ([]SubmissionStatistics, error)
}
