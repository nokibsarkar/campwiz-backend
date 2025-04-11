package models

import (
	"crypto/md5"
	"fmt"
	"nokib/campwiz/models/types"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gen"
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
	PageID       uint64                 `json:"pageId"`
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
type CommonsSubmissionEntry struct {
	PageID      uint64 `json:"pageId"`
	PageTitle   string `json:"pageTitle"`
	UserName    string `json:"userName"`
	FrTimestamp uint64 `json:"frTimestamp"`
	FrHeight    uint64 `json:"frHeight"`
	FrWidth     uint64 `json:"frWidth"`
	FrSize      uint64 `json:"frSize"`
	FtMediaType string `json:"ftMediaType"`
}
type SubmissionStatisticsFetcher interface {
	// SELECT COUNT(*) AS `AssignmentCount`, SUM(`score` IS NOT NULL) AS EvaluationCount, `submission_id`  FROM `evaluations`  WHERE `round_id` = @round_id GROUP BY `submission_id`
	FetchByRoundID(round_id string) ([]SubmissionStatistics, error)
	// UPDATE `submissions` JOIN (SELECT AVG(`evaluations`.`score`) As `Score`, COUNT(`evaluations`.`evaluation_id`) AS `AssignmentCount`, SUM(`evaluations`.`score` IS NOT NULL) AS `EvaluationCount`,`evaluations`.`submission_id` FROM `evaluations` WHERE  `evaluations`.submission_id IN (@submissionIds) GROUP BY `evaluations`.`submission_id`) AS `e` ON `submissions`.`submission_id` = `e`.`submission_id` SET `submissions`.`assignment_count` = `e`.`AssignmentCount`, `submissions`.`evaluation_count` = `e`.`EvaluationCount`, `submissions`.`score` = `e`.`Score` WHERE `submissions`.`submission_id` = `e`.`submission_id`
	TriggerBySubmissionIds(submissionIds []string) (gen.RowsAffected, error)
	// UPDATE `submissions` JOIN (SELECT AVG(`evaluations`.`score`) As `Score`, COUNT(`evaluations`.`evaluation_id`) AS `AssignmentCount`, SUM(`evaluations`.`score` IS NOT NULL) AS `EvaluationCount`,`evaluations`.`submission_id` FROM `evaluations` WHERE `evaluations`.`round_id` = @round_id GROUP BY `evaluations`.`submission_id`) AS `e` ON `submissions`.`submission_id` = `e`.`submission_id` SET `submissions`.`assignment_count` = `e`.`AssignmentCount`, `submissions`.`evaluation_count` = `e`.`EvaluationCount`, `submissions`.`score` = `e`.`Score` WHERE `submissions`.`round_id` = @round_id
	TriggerByRoundId(round_id string) error
}

type SubmissionFetcher interface {
	/*
		SLOW OK
	*/
	//SELECT page_id, page_title, user_name, img_timestamp as fr_timestamp, img_height as fr_height, img_width as fr_width, img_size as fr_size, img_media_type as ft_media_type FROM categorylinks JOIN page JOIN image JOIN actor JOIN `user` ON user_id=actor_user and cl_from=page_id and img_name=page_title and actor_id=img_actor where cl_to=@categoryName and cl_from > @startPageID and @minimumTimestamp <= img_timestamp and img_timestamp < @maximumTimestamp ORDER BY `page_id` ASC LIMIT @limit;
	FetchSubmissionsFromCommonsDBByCategoryOld(categoryName string, startPageID uint64, minimumTimestamp uint64, maximumTimestamp uint64, limit int) ([]CommonsSubmissionEntry, error)
	// SELECT  page_id, page_title, user_name, fr_timestamp, fr_height, fr_width, fr_size, ft_media_type FROM categorylinks JOIN page JOIN file JOIN filerevision JOIN actor JOIN `user` JOIN filetypes ON ft_id = file_type AND fr_id=file_latest AND user_id=actor_user and cl_from=page_id and file_name=page_title and actor_id=fr_actor where cl_from > @startPageID and @minimumTimestamp <= fr_timestamp and fr_timestamp < @maximumTimestamp and cl_to=@categoryName and fr_deleted = false and file_deleted=false ORDER BY `page_id` ASC LIMIT @limit;
	FetchSubmissionsFromCommonsDBByCategory(categoryName string, startPageID uint64, minimumTimestamp uint64, maximumTimestamp uint64, limit int) ([]CommonsSubmissionEntry, error)
}

func (c *CommonsSubmissionEntry) GetURL() string {
	if c.PageTitle == "" {
		return ""
	}
	md5Hash := md5.Sum([]byte(c.PageTitle))
	md5HashHex := fmt.Sprintf("%x", md5Hash)
	folder1, folder2 := md5HashHex[:1], md5HashHex[:2]
	URL := fmt.Sprintf("https://upload.wikimedia.org/wikipedia/commons/%s/%s/%s", folder1, folder2, c.PageTitle)
	return URL
}
func (c *CommonsSubmissionEntry) GetThumbURL() string {
	if c.PageTitle == "" {
		return ""
	}
	md5Hash := md5.Sum([]byte(c.PageTitle))
	md5HashHex := fmt.Sprintf("%x", md5Hash)
	folder1, folder2 := md5HashHex[:1], md5HashHex[:2]
	URL := fmt.Sprintf("https://upload.wikimedia.org/wikipedia/commons/thumb/%s/%s/%s", folder1, folder2, c.PageTitle)
	return URL
}
func (c *CommonsSubmissionEntry) GetSubmittedAt() time.Time {
	const YEAR_MULTIPLIER = 10000000000
	const MONTH_MULTIPLIER = YEAR_MULTIPLIER / 100
	const DAY_MULTIPLIER = MONTH_MULTIPLIER / 100
	const HOUR_MULTIPLIER = DAY_MULTIPLIER / 100
	const MINUTE_MULTIPLIER = HOUR_MULTIPLIER / 100
	remaining := c.FrTimestamp
	year, remaining := remaining/YEAR_MULTIPLIER, remaining%YEAR_MULTIPLIER
	month, remaining := remaining/MONTH_MULTIPLIER, remaining%MONTH_MULTIPLIER
	day, remaining := remaining/DAY_MULTIPLIER, remaining%DAY_MULTIPLIER
	hour, remaining := remaining/HOUR_MULTIPLIER, remaining%HOUR_MULTIPLIER
	minute, second := remaining/MINUTE_MULTIPLIER, remaining%MINUTE_MULTIPLIER
	submittedAt := time.Date(int(year), time.Month(month), int(day), int(hour), int(minute), int(second), 0, time.UTC)
	return submittedAt
}
