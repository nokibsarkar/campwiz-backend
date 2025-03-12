package models

import "time"

type EvaluationType string

const (
	EvaluationTypeRanking EvaluationType = "ranking"
	EvaluationTypeScore   EvaluationType = "score"
	EvaluationTypeBinary  EvaluationType = "binary"
)

type Evaluation struct {
	EvaluationID  IDType         `json:"evaluationId" gorm:"primaryKey"`
	SubmissionID  IDType         `json:"submissionId" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	JudgeID       IDType         `json:"judgeId" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;uniqueIndex:idx_unique_vote_position"`
	ParticipantID IDType         `json:"participantId" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Type          EvaluationType `json:"type"`
	// Applicable if the evaluation type is score, it would be between 0-100
	VoteScore *int `json:"score" gorm:"null;:null"`
	// Applicable if the evaluation type is binary, it would be 0 to Number of submissions in this round
	// The pair (JudgeID, VotePosition) should be unique (i.e. a judge can't vote for the same position twice)
	VotePosition *int `json:"votePosition" gorm:"null;default:null;uniqueIndex:idx_unique_vote_position"`
	// Applicable if the evaluation type is binary, it would be true or false
	VotePassed         *bool       `json:"votePassed" gorm:"null;default:null;"`
	Comment            string      `json:"comment" gorm:"default:null"`
	Serial             uint        `json:"serial"`
	Submission         *Submission `json:"-" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Participant        *User       `json:"-" gorm:"foreignKey:ParticipantID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Judge              *Role       `json:"-" gorm:"foreignKey:JudgeID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	CreatedAt          *time.Time  `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt          *time.Time  `json:"updatedAt" gorm:"autoUpdateTime"`
	EvaluatedAt        *time.Time  `json:"evaluatedAt" gorm:"type:datetime"`
	DistributionTaskID IDType      `json:"distributionTaskId" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
type EvaluationFilter struct {
	Type          EvaluationType `form:"type"`
	RoundID       IDType         `form:"roundId"`
	CampaignID    IDType         `form:"campaignId"`
	ParticipantID IDType         `form:"userId"`
	Evaluated     *bool          `form:"status"`
	SubmissionID  IDType         `form:"submissionId"`
	JuryRoleID    IDType         `form:"juryId"`
	CommonFilter
}
