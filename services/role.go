package services

import (
	"log"
	"nokib/campwiz/consts"
	"nokib/campwiz/models"
	"nokib/campwiz/repository"
	idgenerator "nokib/campwiz/services/idGenerator"

	"gorm.io/gorm"
)

type RoleService struct{}

func NewRoleService() *RoleService {
	return &RoleService{}
}
func (r *RoleService) CalculateRoleDifferenceWithRole(tx *gorm.DB, Type models.RoleType, filter *models.RoleFilter, updatedRoleUsernames []models.WikimediaUsernameType) (addedRoles []models.Role, removedRoles []*models.Role, err error) {
	role_repo := repository.NewRoleRepository()
	user_repo := repository.NewUserRepository()
	existingRoles, err := role_repo.ListAllRoles(tx.Unscoped(), filter)
	if err != nil {
		return nil, nil, err
	}
	username2IDMap := map[models.WikimediaUsernameType]models.IDType{}
	for _, username := range updatedRoleUsernames {
		username2IDMap[username] = idgenerator.GenerateID("u")
	}
	username2IDMap, err = user_repo.EnsureExists(tx, username2IDMap)
	if err != nil {
		return nil, nil, err
	}
	userId2NameMap := map[models.IDType]models.WikimediaUsernameType{}
	for username := range username2IDMap {
		userId := username2IDMap[username]
		userId2NameMap[userId] = username
	}
	log.Printf("Existing roles: %v", existingRoles)

	id2RoleMap := map[models.IDType]models.Role{}
	for _, existingRole := range existingRoles {
		id2RoleMap[existingRole.UserID] = existingRole
		if existingRole.DeletedAt != nil {
			// Already soft deleted, so either way, pretend it does not exist
			log.Printf("Soft %s", existingRole.UserID)
			continue
		}
		_, ok := userId2NameMap[existingRole.UserID]
		if !ok {
			// not exist in updated roles
			// remove the role by adding isAllowed = false and permission would be null

			log.Printf("%s is being removed as %s", existingRole.RoleID, Type)
			removedRoles = append(removedRoles, &existingRole)
		} else {
			// Remain unchanged
			log.Printf("%s is unchanged as %s", userId2NameMap[existingRole.UserID], Type)
		}
	}

	for userId := range userId2NameMap {
		role, ok := id2RoleMap[userId]
		if !ok || role.DeletedAt != nil {
			// Newly added
			newRole := models.Role{
				RoleID:          idgenerator.GenerateID("j"),
				Type:            Type,
				UserID:          userId,
				ProjectID:       filter.ProjectID,
				CampaignID:      filter.CampaignID,
				RoundID:         filter.RoundID,
				TargetProjectID: filter.TargetProjectID,
				Permission:      Type.GetPermission(),
			}
			if ok {
				log.Printf("%s was banned earlier. Unbanning ", role.RoleID)
				// already exisiting but was banned
				newRole.RoleID = role.RoleID
			}
			log.Printf("adding %s was banned earlier.", newRole.RoleID)
			addedRoles = append(addedRoles, newRole)
		} else {
			//remain unchanged
		}
	}
	log.Println("Add: ", addedRoles)
	log.Println("Remove ", removedRoles)
	return addedRoles, removedRoles, nil
}
func (r *RoleService) CalculateRoleDifference(tx *gorm.DB, Type models.RoleType, filter *models.RoleFilter, updatedRoleUsernames []models.WikimediaUsernameType) (addedRoles []models.Role, removedRoleIds []models.IDType, err error) {
	addedRoles, removedRoles, err := r.CalculateRoleDifferenceWithRole(tx, Type, filter, updatedRoleUsernames)
	if err != nil {
		return nil, nil, err
	}
	for _, role := range removedRoles {
		removedRoleIds = append(removedRoleIds, role.RoleID)
	}
	return addedRoles, removedRoleIds, nil
}
func (service *RoleService) FetchChangeRoles(tx *gorm.DB, roleType models.RoleType, projectID models.IDType, targetProjectID *models.IDType, campaignId *models.IDType, roundID *models.IDType, updatedRoleUsernames []models.WikimediaUsernameType, unscoped bool) (currentRoles []models.Role, removedRoles []models.IDType, err error) {
	filter := &models.RoleFilter{
		ProjectID: projectID,
		Type:      &roleType,
	}
	if targetProjectID != nil {
		filter.TargetProjectID = targetProjectID
	}
	if campaignId != nil {
		filter.CampaignID = campaignId
	}
	if roundID != nil {
		filter.RoundID = roundID
	}
	addedRoles, removedRoles, err := service.CalculateRoleDifference(tx, roleType, filter, updatedRoleUsernames)
	if err != nil {
		log.Println(err)
		return
	}
	if len(addedRoles) > 0 {
		res := tx.Save(addedRoles)
		if res.Error != nil {
			log.Println(res.Error)
			return nil, nil, res.Error
		}
	}
	if len(removedRoles) > 0 {
		for _, roleID := range removedRoles {
			tx1 := tx
			if unscoped {
				tx1 = tx.Unscoped()
			}
			res := tx1.Delete(&models.Role{RoleID: roleID})
			if res.Error != nil {
				return nil, nil, res.Error
			}
		}
	}
	currentRoles = []models.Role{}
	res := tx.Where(filter).Find(&currentRoles)
	if res.Error != nil {
		return nil, nil, res.Error
	}
	return currentRoles, removedRoles, nil
}

func (service *RoleService) ListRoles(tx *gorm.DB, filter *models.RoleFilter) ([]models.Role, error) {
	role_repo := repository.NewRoleRepository()
	return role_repo.ListAllRoles(tx, filter)
}
func (service *RoleService) CheckPermissionByUserID(targetPermission consts.Permission, against any) bool {
	sourcePermission := consts.Permission(0)
	ok := false
	if sourcePermission, ok = against.(consts.Permission); ok {
		return sourcePermission&targetPermission == targetPermission
	}
	if sourcePermissionGroup, ok := against.(consts.PermissionGroup); ok {
		return sourcePermissionGroup.HasPermission(targetPermission)
	}
	return false
}
