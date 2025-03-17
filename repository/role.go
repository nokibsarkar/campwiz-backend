package repository

import (
	"log"
	"nokib/campwiz/models"

	"gorm.io/gorm"
)

type RoleRepository struct{}

func NewRoleRepository() *RoleRepository {
	return &RoleRepository{}
}
func (r *RoleRepository) CreateRole(tx *gorm.DB, jury *models.Role) error {
	result := tx.Create(jury)
	return result.Error
}
func (r *RoleRepository) CreateRoles(tx *gorm.DB, juries []models.Role) error {
	result := tx.Create(juries)
	return result.Error
}
func (r *RoleRepository) FindRoleByID(tx *gorm.DB, juryID models.IDType) (*models.Role, error) {
	jury := &models.Role{}
	result := tx.First(jury, &models.Role{RoleID: juryID})
	return jury, result.Error
}
func (r *RoleRepository) ListAllRoles(tx *gorm.DB, filter *models.RoleFilter) ([]models.Role, error) {
	var juries []models.Role
	stmt := tx
	if filter != nil {
		// if filter.IncludeDeleted != nil {
		// 	if *filter.IncludeDeleted {
		// 		stmt = stmt.Unscoped()
		// 	}
		// }
		if filter.ProjectID != "" {
			stmt = stmt.Where(&models.Role{ProjectID: filter.ProjectID})
		}
		if filter.TargetProjectID != nil {
			stmt = stmt.Where(&models.Role{TargetProjectID: filter.TargetProjectID})
		}
		if filter.CampaignID != nil {
			stmt = stmt.Where(&models.Role{CampaignID: filter.CampaignID})
		}
		if filter.RoundID != nil {
			stmt = stmt.Where(&models.Role{RoundID: filter.RoundID})
		}
		if filter.Type != nil {
			stmt = stmt.Where(&models.Role{Type: *filter.Type})
		}
		if filter.UserID != nil {
			stmt = stmt.Where(&models.Role{UserID: *filter.UserID})
		}
		if filter.Limit > 0 {
			stmt = stmt.Limit(filter.Limit)
		}

	}
	result := stmt.Find(&juries)
	log.Println("ListAllRoles: ", juries)
	return juries, result.Error
}
func (r *RoleRepository) FindRoleByUserIDAndRoundID(tx *gorm.DB, userID models.IDType, roundID models.IDType, roleType models.RoleType) (*models.Role, error) {
	jury := &models.Role{}
	result := tx.First(jury, &models.Role{UserID: userID, RoundID: &roundID, Type: roleType})
	return jury, result.Error
}
