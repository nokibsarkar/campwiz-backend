package repository

import (
	"nokib/campwiz/models"
	"nokib/campwiz/query"

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
	var roles []models.Role
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
		stmt = stmt.Where(&models.Role{DeletedAt: nil})

	}
	result := stmt.Find(&roles)
	return roles, result.Error
}
func (r *RoleRepository) FindRoleByUserIDAndRoundID(tx *gorm.DB, userID models.IDType, roundID models.IDType, roleType models.RoleType) (*models.Role, error) {
	jury := &models.Role{}
	result := tx.First(jury, &models.Role{UserID: userID, RoundID: &roundID, Type: roleType})
	return jury, result.Error
}
func (r *RoleRepository) DeleteRolesByRoundID(tx *gorm.DB, roundID models.IDType) error {
	result := tx.Unscoped().Delete(&models.Role{}, &models.Role{
		RoundID: &roundID,
	})
	return result.Error
}
func (r *RoleRepository) FindRolesByUsername(tx *gorm.DB, usernames []models.WikimediaUsernameType, filter *models.RoleFilter) ([]models.Role, error) {
	roles := []models.Role{}
	Usernames := []string{}
	for _, username := range usernames {
		Usernames = append(Usernames, string(username))
	}
	q := query.Use(tx)
	Role := q.Role
	stmt := Role.Join(q.User, q.User.UserID.EqCol(Role.UserID)).
		Where(q.User.Username.In(Usernames...))
	if filter != nil {
		// if filter.IncludeDeleted != nil {
		// 	if *filter.IncludeDeleted {
		// 		stmt = stmt.Unscoped()
		// 	}
		// }
		if filter.ProjectID != "" {
			stmt = stmt.Where(Role.ProjectID.Eq(filter.ProjectID.String()))
		}
		if filter.TargetProjectID != nil {
			stmt = stmt.Where(Role.TargetProjectID.Eq(filter.TargetProjectID.String()))
		}
		if filter.CampaignID != nil {
			stmt = stmt.Where(Role.CampaignID.Eq(filter.CampaignID.String()))
		}
		if filter.RoundID != nil {
			stmt = stmt.Where(Role.RoundID.Eq(filter.RoundID.String()))
		}
		if filter.Type != nil {
			stmt = stmt.Where(Role.Type.Eq(string(*filter.Type)))
		}
		if filter.UserID != nil {
			stmt = stmt.Where(Role.UserID.Eq(filter.UserID.String()))
		}
		if filter.Limit > 0 {
			stmt = stmt.Limit(filter.Limit)
		}

	}
	err := stmt.Scan(&roles)
	return roles, err
}
