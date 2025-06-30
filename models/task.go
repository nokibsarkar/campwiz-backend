package models

import (
	"time"

	"gorm.io/datatypes"
)

type TaskStatus string
type TaskType string

const (
	TaskTypeImportFromCommons       TaskType = "submissions.import.commons"
	TaskTypeImportFromPreviousRound TaskType = "submissions.import.previous"
	TaskTypeImportFromCSV           TaskType = "submissions.import.csv"
	TaskTypeDistributeEvaluations   TaskType = "assignments.distribute"
	TaskTypeRandomizeAssignments    TaskType = "assignments.randomize"
)
const (
	TaskStatusPending TaskStatus = "pending"
	TaskStatusRunning TaskStatus = "running"
	TaskStatusSuccess TaskStatus = "success"
	TaskStatusFailed  TaskStatus = "failed"
)

type Task struct {
	TaskID               IDType                                 `json:"taskId" gorm:"primaryKey"`
	Type                 TaskType                               `json:"type"`
	Status               TaskStatus                             `json:"status"`
	AssociatedRoundID    *IDType                                `json:"roundId" gorm:"index;null;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	AssociatedCampaignID *IDType                                `json:"campaignId" gorm:"index;null;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	AssociatedUserID     *IDType                                `json:"userId" gorm:"index;null;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Data                 *datatypes.JSON                        `json:"data"`
	CreatedAt            time.Time                              `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt            time.Time                              `json:"updatedAt" gorm:"autoUpdateTime"`
	SuccessCount         int                                    `json:"successCount"`
	FailedCount          int                                    `json:"failedCount"`
	FailedIds            *datatypes.JSONType[map[string]string] `json:"failedIds"`
	RemainingCount       int                                    `json:"remainingCount"`
	CreatedByID          IDType                                 `json:"createdById" gorm:"index;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Submittor            User                                   `json:"-" gorm:"foreignKey:CreatedByID;references:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	TaskData             []TaskData                             `json:"taskData,omitempty"`
}
type TaskData struct {
	DataID IDType `json:"dataId" gorm:"primaryKey"`
	// The task ID that this data belongs to
	TaskID IDType  `json:"taskId"`
	Key    *string `json:"key,omitempty" gorm:"index;null"`
	Value  string  `json:"value"`
	// Whether this data is an input or output of the task
	// For example, if the task is to import data from commons, then the input would be the
	// commons data and the output would be the rejection reason
	IsOutput bool `json:"isOutput"`
	// The time the data was created, it would be set automatically
	Task *Task `json:"-" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

// `id`        INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
//
//	`pageid`    INTEGER NOT NULL, -- pageid of the page on the target wiki
//	`campaign_id`       INTEGER NOT NULL, -- campaign id
//	`title`     TEXT NOT NULL, -- title of the page on the target wiki
//	`oldid`     INTEGER NOT NULL, -- revision id of the page on the home wiki
//	`target_wiki`       TEXT NOT NULL, -- target wiki language code
//	`submitted_at`      TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- submission time
//	`submitted_by_id`   INTEGER NOT NULL, -- user id of the submitter
//	`submitted_by_username`     TEXT NOT NULL, -- username of the submitter
//	`created_at`        TEXT NOT NULL, -- creation time of the page on the target wiki
//	`created_by_id`     INTEGER NOT NULL, -- user id of the created_by
//	`created_by_username`       TEXT NOT NULL, -- username of the created_by
//	`total_bytes`       INTEGER NOT NULL, -- total bytes of the page on the target wiki
//	`total_words`       INTEGER  NULL DEFAULT NULL, -- total words of the page on the target wiki
//	`added_bytes`       INTEGER  NULL DEFAULT NULL, -- added bytes of the page on the target wiki
//	`added_words`       INTEGER  NULL DEFAULT NULL, -- added words of the page on the target wiki
//	`positive_votes`    INTEGER NOT NULL DEFAULT 0, -- positive votes of the submission
//	`negative_votes`    INTEGER NOT NULL DEFAULT 0, -- negative votes of the submission
//	`total_votes`       INTEGER NOT NULL DEFAULT 0, -- total votes of the submission
//	`points`    INTEGER NOT NULL DEFAULT 0, -- points of the submission
//	`judgable`  BOOLEAN NOT NULL DEFAULT TRUE, -- whether the submission is judgable
//	`newly_created`     BOOLEAN NOT NULL DEFAULT FALSE, -- whether the submission is newly created
type CampWizV1Submission struct {
	// ID is the primary key of the submission
	SubmissionID int `json:"id" gorm:"primaryKey"`
	// PageID is the page ID of the submission on the target wiki
	PageID int `json:"pageId" gorm:"index;not null;column:pageid"` // Use `column:pageid` to match the SQLite schema
	// CampaignID is the ID of the campaign this submission belongs to
	CampaignID          int       `json:"campaignId" gorm:"index;not null"`
	Title               string    `json:"title" gorm:"not null"`
	OldID               int       `json:"oldId" gorm:"not null"`
	TargetWiki          string    `json:"targetWiki" gorm:"not null"`
	SubmittedAt         time.Time `json:"submittedAt"`
	SubmittedByID       int       `json:"submittedById" gorm:"index;not null"`
	SubmittedByUsername string    `json:"submittedByUsername" gorm:"not null"`
	// CreatedAt           time.Time `json:"createdAt" gorm:"not null;column:created_at"` // Use `column:created_at` to match the SQLite schema
	CreatedByID       int    `json:"createdById" gorm:"index;not null"`
	CreatedByUsername string `json:"createdByUsername" gorm:"not null"`
	TotalBytes        int    `json:"totalBytes" gorm:"not null"`
	TotalWords        int    `json:"totalWords" gorm:"default:NULL"`
	AddedBytes        int    `json:"addedBytes" gorm:"default:NULL"`
	AddedWords        int    `json:"addedWords" gorm:"default:NULL"`
	PositiveVotes     int    `json:"positiveVotes" gorm:"default:0"`
	NegativeVotes     int    `json:"negativeVotes" gorm:"default:0"`
	TotalVotes        int    `json:"totalVotes" gorm:"default:0"`
	Points            int    `json:"points" gorm:"default:0"`
	Judgable          bool   `json:"judgable" gorm:"default:true"`
	NewlyCreated      bool   `json:"newlyCreated" gorm:"default:false"`
}

// `jury_id`   INTEGER NOT NULL, -- user id of the jury
//
//	`submission_id`     INTEGER NOT NULL, -- submission id
//	`campaign_id` INTEGER NOT NULL, -- Campaign id
//	`created_at`        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, -- creation time of the vote
//	`vote`      INTEGER NOT NULL, note text null default null, -- vote value
//	CONSTRAINT `PK_jury_vote` PRIMARY KEY (`jury_id`,`submission_id`),
//	CONSTRAINT `jury_vote_submission_id_fkey` FOREIGN KEY(`submission_id`) REFERENCES `submission`(`id`) ON DELETE CASCADE ON UPDATE CASCADE,
//	CONSTRAINT `jury_vote_campaign__user_id_fkey` FOREIGN KEY(`jury_id`,`campaign_id`) REFERENCES `jury`(`user_id`,`campaign_id`) ON DELETE CASCADE ON UPDATE CASCADE,
//	CONSTRAINT `jury_vote_campaign_id_fkey` FOREIGN KEY(`campaign_id`) REFERENCES `campaign`(`id`) ON DELETE CASCADE ON UPDATE CASCADE
type CampWizV1Evaluation struct {
	// ID is the primary key of the evaluation
	EvaluationId int `json:"id" gorm:"primaryKey"`
	// SubmissionID is the ID of the submission being evaluated
	SubmissionID int `json:"submissionId" gorm:"index;not null"`
	// CampaignID is the ID of the campaign this evaluation belongs to
	CampaignID int `json:"campaignId" gorm:"index;not null"`
	// JuryID is the ID of the jury member who created the evaluation
	JuryID int `json:"juryId" gorm:"index;not null"`
	// CreatedAt is the timestamp when the evaluation was created
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
	// Vote is the vote value given by the jury member
	Vote int `json:"vote" gorm:"not null"`
}
