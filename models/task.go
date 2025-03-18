package models

import (
	"time"

	"gorm.io/datatypes"
)

type TaskStatus string
type TaskType string

const (
	TaskTypeImportFromCommons       TaskType = "import.commons"
	TaskTypeImportFromPreviousRound TaskType = "import.previous.round"
	TaskTypeDistributeEvaluations   TaskType = "distribute.evaluations"
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
}
