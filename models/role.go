package models

import (
	"nokib/campwiz/consts"

	"gorm.io/gorm"
)

type RoleType string

const (
	RoleTypeAdmin       RoleType = "admin"
	RoleTypeProjectLead RoleType = "projectLead"
	RoleTypeCoordinator RoleType = "coordinator"
	RoleTypeJury        RoleType = "jury"
	RoleTypeParticipant RoleType = "participant"
)

type Role struct {
	RoleID          IDType                 `json:"roleId" gorm:"primaryKey"`
	Type            RoleType               `json:"type" gorm:"not null"`
	UserID          IDType                 `json:"userId" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	ProjectID       IDType                 `json:"projectId" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	TargetProjectID *IDType                `json:"targetProjectId" gorm:"null;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	CampaignID      *IDType                `json:"campaignId" gorm:"null;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	RoundID         *IDType                `json:"roundId" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	TotalAssigned   int                    `json:"totalAssigned"`
	TotalEvaluated  int                    `json:"totalEvaluated"`
	TotalScore      int                    `json:"totalScore"`
	Permission      consts.PermissionGroup `json:"permission" gorm:"not null;default:0"`
	Round           *Round                 `json:"-" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Campaign        *Campaign              `json:"-"  gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	User            *User                  `json:"-" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Project         *Project               `json:"-"  gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	DeletedAt       *gorm.DeletedAt        `json:"deletedAt"`
}
type RoleFilter struct {
	CommonFilter
	ProjectID       IDType                  `form:"projectId"`
	TargetProjectID *IDType                 `form:"targetProjectId"`
	CampaignID      *IDType                 `form:"campaignId"`
	RoundID         *IDType                 `form:"roundId"`
	Type            *RoleType               `form:"type"`
	Permission      *consts.PermissionGroup `form:"permission"`
	UserID          *IDType                 `form:"userId"`
}

func (r *RoleType) GetPermission() consts.PermissionGroup {
	switch *r {
	case RoleTypeAdmin:
		return consts.PermissionGroupADMIN
	case RoleTypeProjectLead:
		return consts.PermissionGroupLead
	case RoleTypeCoordinator:
		return consts.PermissionGroupCoordinator
	case RoleTypeJury:
		return consts.PermissionGroupJury
	}
	return consts.PermissionGroupUSER
}

type JuryStatistics struct {
	JudgeID        IDType `json:"judgeId"`
	TotalAssigned  int    `json:"totalAssigned"`
	TotalEvaluated int    `json:"totalEvaluated"`
}
type JuryStatisticsUpdater interface {
	// SELECT COUNT(*) AS TotalAssigned, SUM(IF(evaluated_at IS NOT NULL, 1, 0)) AS TotalEvaluated, judge_id FROM `evaluations` WHERE round_id = @roundID AND `judge_id` IS NOT NULL GROUP  BY judge_id
	GetJuryStatistics(roundID string) ([]JuryStatistics, error)
	// UPDATE roles AS jury JOIN ( SELECT judge_id, COUNT(*) AS c,    SUM(evaluated_at IS NOT NULL) AS ev  FROM evaluations  WHERE  round_id = @roundID  GROUP BY judge_id ) AS d ON jury.role_id = d.judge_id SET jury.total_evaluated = d.ev, jury.total_assigned = d.c WHERE jury.round_id = @roundID
	TriggerByRoundID(roundID string) error
}
