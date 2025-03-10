package database

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
}
type RoleRepository struct{}

func NewRoleRepository() *RoleRepository {
	return &RoleRepository{}
}
func (r *RoleRepository) CreateRole(tx *gorm.DB, jury *Role) error {
	result := tx.Create(jury)
	return result.Error
}
func (r *RoleRepository) CreateRoles(tx *gorm.DB, juries []Role) error {
	result := tx.Create(juries)
	return result.Error
}
func (r *RoleRepository) FindRoleByID(tx *gorm.DB, juryID IDType) (*Role, error) {
	jury := &Role{}
	result := tx.First(jury, &Role{RoleID: juryID})
	return jury, result.Error
}
func (r *RoleRepository) ListAllRoles(tx *gorm.DB, filter *RoleFilter) ([]Role, error) {
	var juries []Role
	stmt := tx
	if filter != nil {
		if filter.ProjectID != "" {
			stmt = stmt.Where(&Role{ProjectID: filter.ProjectID})
		}
		if filter.TargetProjectID != nil {
			stmt = stmt.Where(&Role{TargetProjectID: filter.TargetProjectID})
		}
		if filter.CampaignID != nil {
			stmt = stmt.Where(&Role{CampaignID: filter.CampaignID})
		}
		if filter.RoundID != nil {
			stmt = stmt.Where(&Role{RoundID: filter.RoundID})
		}
		if filter.Type != nil {
			stmt = stmt.Where(&Role{Type: *filter.Type})
		}

	}
	result := stmt.Find(&juries)
	return juries, result.Error
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
