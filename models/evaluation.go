package models

import (
	"nokib/campwiz/models/types"
	"time"
)

type EvaluationType string

const MAXIMUM_EVALUATION_SCORE = ScoreType(100)
const (
	EvaluationTypeRanking EvaluationType = "ranking"
	EvaluationTypeScore   EvaluationType = "score"
	EvaluationTypeBinary  EvaluationType = "binary"
)

type Evaluation struct {
	EvaluationID  IDType                 `json:"evaluationId" gorm:"primaryKey"`
	SubmissionID  types.SubmissionIDType `json:"submissionId" gorm:"index;uniqueIndex:idx_submission_judge"`
	JudgeID       *IDType                `json:"judgeId" gorm:"index;uniqueIndex:idx_submission_judge"`
	ParticipantID IDType                 `json:"participantId"`
	RoundID       IDType                 `json:"roundId" gorm:"index"`
	Type          EvaluationType         `json:"type"`
	Score         *ScoreType             `json:"score" gorm:"default:null;constraint:check:(score >= 0 AND score <= 100)"`
	Comment       string                 `json:"comment" gorm:"default:null"`
	Serial        uint                   `json:"serial"`
	Submission    *Submission            `json:"-" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Participant   *User                  `json:"-" gorm:"foreignKey:ParticipantID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Judge         *Role                  `json:"-" gorm:"foreignKey:JudgeID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	CreatedAt     *time.Time             `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt     *time.Time             `json:"updatedAt" gorm:"autoUpdateTime"`
	EvaluatedAt   *time.Time             `json:"evaluatedAt" gorm:"type:datetime"`
	// Round              *Round         `json:"-" gorm:"foreignKey:RoundID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	DistributionTaskID IDType `json:"distributionTaskId" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
type EvaluationFilter struct {
	Type          EvaluationType         `form:"type"`
	RoundID       IDType                 `form:"roundId"`
	CampaignID    IDType                 `form:"campaignId"`
	ParticipantID IDType                 `form:"userId"`
	Evaluated     *bool                  `form:"status"`
	SubmissionID  types.SubmissionIDType `form:"submissionId"`
	JuryRoleID    IDType                 `form:"juryId"`
	CommonFilter
}
type NewEvaluationRequest struct {
	SubmissionID IDType
	Times        int
}
