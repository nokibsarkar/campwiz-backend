package models

import (
	"nokib/campwiz/consts"
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
	IsAllowed       bool                   `json:"isAllowed" gorm:"default:false"`
	TotalAssigned   int                    `json:"totalAssigned"`
	TotalEvaluated  int                    `json:"totalEvaluated"`
	TotalScore      int                    `json:"totalScore"`
	Permission      consts.PermissionGroup `json:"permission" gorm:"not null;default:0"`
	Campaign        *Campaign              `json:"-"  gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	User            *User                  `json:"-" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Round           *Round                 `json:"-"  gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Project         *Project               `json:"-"  gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
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
	// UPDATE `jury` SET
	UpdateJuryStatistics(roundID string) error
	// SELECT COUNT(*) AS TotalAssigned, SUM(IF(evaluated_at IS NOT NULL, 1, 0)) AS TotalEvaluated, judge_id FROM `evaluations` WHERE round_id = @roundID GROUP BY judge_id
	GetJuryStatistics(roundID string) ([]JuryStatistics, error)
}
