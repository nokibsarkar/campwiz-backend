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
	JudgeID       *IDType                `json:"judgeId" gorm:"index;uniqueIndex:idx_submission_judge;index:idx_evaluation_judge_round_score"`
	ParticipantID IDType                 `json:"participantId"`
	RoundID       IDType                 `json:"roundId" gorm:"index;index:idx_evaluation_judge_round_score"`
	Type          EvaluationType         `json:"type"`
	Score         *ScoreType             `json:"score" gorm:"default:null;constraint:check:(score >= 0 AND score <= 100);index:idx_evaluation_judge_round_score"`
	Comment       string                 `json:"comment" gorm:"default:null"`
	Serial        uint                   `json:"serial"`
	Submission    *Submission            `json:"submission" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Participant   *User                  `json:"-" gorm:"foreignKey:ParticipantID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	// This would be null when any judge is deleted
	Judge       *Role      `json:"-" gorm:"foreignKey:JudgeID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	AssignedAt  *time.Time `json:"assignedAt" gorm:"autoCreateTime"`
	EvaluatedAt *time.Time `json:"evaluatedAt" gorm:"type:datetime"`
	// SkipExpirationAt is the time when the skip request will expire
	SkipExpirationAt *time.Time `json:"skipExpirationAt" gorm:"type:datetime"`
	// Round              *Round         `json:"-" gorm:"foreignKey:RoundID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	DistributionTaskID IDType `json:"distributionTaskId" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
type EvaluationFilter struct {
	Type          EvaluationType         `form:"type"`
	RoundID       IDType                 `form:"roundId"`
	CampaignID    IDType                 `form:"campaignId"`
	ParticipantID IDType                 `form:"userId"`
	SubmissionID  types.SubmissionIDType `form:"submissionId"`
	JuryRoleID    IDType                 `form:"juryId"`
	// whether to include the submissions that were evaluated
	IncludeEvaluated *bool `form:"includeEvaluated"`
	// whether to include the submissions that were skipped
	IncludeSkipped *bool `form:"includeSkipped"`
	// Whether to embed the submission object
	IncludeSubmission bool `form:"includeSubmission"`
	// Whether to include the non-evaluated submissions
	IncludeNonEvaluated *bool `form:"includeNonEvaluated,default=true"`
	Randomize           bool  `form:"randomize"`
	CommonFilter
}
type GetEvaluationQueryFilter struct {
	IncludeSkipped bool `form:"includeSkipped"`
	CommonFilter
}
type NewEvaluationRequest struct {
	SubmissionID IDType
	Times        int
}
type EvaluationListResponseWithCurrentStats struct {
	ResponseList[*Evaluation]
	TotalEvaluatedCount  int `json:"totalEvaluatedCount"`
	TotalAssignmentCount int `json:"totalAssignmentCount"`
}
type Evaluator interface {
	// DELETE FROM `evaluations` WHERE `evaluations`.`evaluated_at` IS NULL AND `round_id` = @roundID AND `submission_id` IN (SELECT `submission_id` FROM `submissions` WHERE `evaluation_count` <= @quorum` AND `round_id` = @roundID)
	RemoveRedundantEvaluation(roundID string, quorum int)
}
