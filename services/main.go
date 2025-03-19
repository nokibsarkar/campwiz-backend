package services

import (
	"log"
	"nokib/campwiz/consts"
	"nokib/campwiz/models"
	"nokib/campwiz/query"

	"gorm.io/gorm"
)

const AdminPermission = consts.Permission(consts.PermissionGroupADMIN)

func checkPermissionWithSuperAdminStatus(targetPermission consts.Permission, matchAgainst consts.Permission) bool {
	if matchAgainst&AdminPermission == AdminPermission {
		return true
	}
	return matchAgainst&targetPermission == targetPermission
}

// This function would check whether the given entity has the sufficient permission
// first check if we are matching against a permission, if so, voila!
// if not, check if the entity is a permissionGroup, if so, voila!
// if not, check if the entity is a role, if so, voila!
// if not, check if the entity is a user, if so, voila!
// if not, check if the entity is a project, if so, voila!
// if not, check if the entity is a campaign, if so, voila!
// if not, check if the entity is a round, if so, voila!
func CheckAccess(targetPermission consts.Permission, matchAgainst any, userIdJustInCase *models.IDType, tx *gorm.DB) bool {
	asPermission, ok := matchAgainst.(consts.Permission)
	if ok {
		return checkPermissionWithSuperAdminStatus(targetPermission, asPermission)
	}
	asPermissionPointer, ok := matchAgainst.(*consts.Permission)
	if ok {
		return checkPermissionWithSuperAdminStatus(targetPermission, *asPermissionPointer)
	}
	asPermissionGroup, ok := matchAgainst.(consts.PermissionGroup)
	if ok {
		return checkPermissionWithSuperAdminStatus(targetPermission, consts.Permission(asPermissionGroup))
	}
	asPermissionGroupPointer, ok := matchAgainst.(*consts.PermissionGroup)
	if ok {
		return checkPermissionWithSuperAdminStatus(targetPermission, consts.Permission(*asPermissionGroupPointer))
	}
	asUser, ok := matchAgainst.(models.User)
	if ok {
		// check if the given user is an admin
		if asUser.Permission.HasPermission(consts.Permission(consts.PermissionGroupADMIN)) {
			return true
		}
		return checkPermissionWithSuperAdminStatus(targetPermission, consts.Permission(asUser.Permission))
	}
	asUserPointer, ok := matchAgainst.(*models.User)
	if ok {
		return checkPermissionWithSuperAdminStatus(targetPermission, consts.Permission(asUserPointer.Permission))
	}
	// indirect permission check because this is not a system permission
	asRole, ok := matchAgainst.(models.Role)
	if ok {
		return checkPermissionWithSuperAdminStatus(targetPermission, consts.Permission(asRole.Permission))
	}
	asRolePointer, ok := matchAgainst.(*models.Role)
	if ok {
		return checkPermissionWithSuperAdminStatus(targetPermission, consts.Permission(asRolePointer.Permission))
	}
	// Now we are getting into the more complex checks
	// we have to fetch the relevant entity from the database
	// and then check the permission
	// so check if userIdJustInCase is provided and tx is provided
	// if not, return false
	if userIdJustInCase == nil || tx == nil {
		return false
	}
	q := query.Use(tx)

	asProject, ok := matchAgainst.(models.Project)
	if ok {
		log.Println("Checkin agains a project")
		// Select the person who is lead of the project by checking his project ID
		u, err := q.User.Select(q.User.Permission).Where(q.User.LeadingProjectID.Eq(asProject.ProjectID.String())).First()
		if err != nil {
			return false
		}
		return checkPermissionWithSuperAdminStatus(targetPermission, consts.Permission(u.Permission))
	}
	asProjectPointer, ok := matchAgainst.(*models.Project)
	if ok {
		log.Println("Checkin agains a project pointer")
		return CheckAccess(targetPermission, *asProjectPointer, userIdJustInCase, tx)
	}
	stmt := q.Role.Select(q.Role.Permission)
	asCampaign, ok := matchAgainst.(models.Campaign)
	if ok {
		log.Println("Checkin agains a campaign")
		role, err := stmt.Where(q.Role.CampaignID.Eq(asCampaign.CampaignID.String())).First()
		if err != nil {
			return false
		}
		return checkPermissionWithSuperAdminStatus(targetPermission, consts.Permission(role.Permission))
	}
	asCampaignPointer, ok := matchAgainst.(*models.Campaign)
	if ok {
		log.Println("Checkin agains a campaign pointer")
		return CheckAccess(targetPermission, *asCampaignPointer, userIdJustInCase, tx)
	}
	asRound, ok := matchAgainst.(models.Round)
	if ok {
		log.Println("Checkin agains a round")
		role, err := stmt.Where(q.Role.RoundID.Eq(asRound.RoundID.String())).First()
		if err != nil {
			return false
		}
		return checkPermissionWithSuperAdminStatus(targetPermission, consts.Permission(role.Permission))
	}
	asRoundPointer, ok := matchAgainst.(*models.Round)
	if ok {
		log.Println("Checkin agains a round pointer")
		return CheckAccess(targetPermission, *asRoundPointer, userIdJustInCase, tx)
	}
	asUserID, ok := matchAgainst.(models.IDType)
	if ok {
		log.Println("Checkin agains a user ID")
		u, err := q.User.Select(q.User.Permission).Where(q.User.UserID.Eq(asUserID.String())).First()
		if err != nil {
			return false
		}
		return checkPermissionWithSuperAdminStatus(targetPermission, consts.Permission(u.Permission))
	}
	asUserIDPointer, ok := matchAgainst.(*models.IDType)
	if ok {
		log.Println("Checkin agains a user ID pointer")
		return CheckAccess(targetPermission, *asUserIDPointer, userIdJustInCase, tx)
	}
	log.Println("All checks failed")
	return false
}
