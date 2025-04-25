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
