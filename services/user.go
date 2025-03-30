package services

import (
	"log"
	"nokib/campwiz/consts"
	"nokib/campwiz/models"
	"nokib/campwiz/repository"
	idgenerator "nokib/campwiz/services/idGenerator"

	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/util/sets"
)

type UserService struct{}

func NewUserService() *UserService {
	return &UserService{}
}

func (u *UserService) GetUserByID_(conn *gorm.DB, id models.IDType) (*models.User, error) {
	userFilter := &models.User{UserID: id}
	user := &models.User{}
	result := conn.First(user, userFilter)
	if result.Error != nil {
		log.Println("Error: ", result.Error)
		return nil, result.Error
	}
	return user, nil

}
func (u *UserService) GetUserByID(id models.IDType) (*models.User, error) {
	user_repo := repository.NewUserRepository()
	conn, close, err := repository.GetDB()
	if err != nil {
		return nil, err
	}
	defer close()
	return user_repo.FindByID(conn, id)
}
func (u *UserService) GetExtendedDetails(id models.IDType) (*models.ExtendedUserDetails, error) {
	user_repo := repository.NewUserRepository()
	conn, close, err := repository.GetDB()
	if err != nil {
		return nil, err
	}
	defer close()
	user, err := user_repo.FindByID(conn, id)
	if err != nil {
		return nil, err
	}
	details := &models.ExtendedUserDetails{
		User:          *user,
		PermissionMap: consts.GetPermissionMap(),
		Permissions:   user.Permission.GetPermissions(consts.GetPermissionMap()),
	}
	return details, nil
}
func (u *UserService) GetUserByUsername(conn *gorm.DB, username models.WikimediaUsernameType) (*models.User, error) {
	userFilter := &models.User{Username: username}
	user := &models.User{}
	result := conn.First(user, userFilter)
	if result.Error != nil {
		return nil, result.Error
	}
	return user, nil
}
func (u *UserService) GetOrCreateUser(conn *gorm.DB, user *models.User) (*models.User, error) {
	result := conn.FirstOrCreate(user, user)
	if result.Error != nil {
		return nil, result.Error
	}
	return user, nil
}
func (u *UserService) EnsureExists(tx *gorm.DB, usernameSet sets.Set[models.WikimediaUsernameType]) (map[models.WikimediaUsernameType]models.IDType, error) {
	user_repo := repository.NewUserRepository()
	userName2Id, err := user_repo.FetchExistingUsernames(tx, usernameSet.UnsortedList())
	if err != nil {
		return nil, err
	}
	if len(userName2Id) > 0 {
		for username := range userName2Id {
			usernameSet.Delete(username)
		}
	}
	nonExistentUsers := usernameSet.UnsortedList()
	if len(nonExistentUsers) == 0 {
		return userName2Id, nil
	}
	commons_repo := repository.NewCommonsRepository()
	users, err := commons_repo.GeUsersFromUsernames(nonExistentUsers)
	if err != nil {
		return nil, err
	}
	new_users := []models.User{}
	for _, u := range users {
		new_user := models.User{
			UserID:       idgenerator.GenerateID("user"),
			RegisteredAt: u.Registered,
			Username:     u.Name,
			Permission:   consts.PermissionGroupADMIN,
		}
		new_users = append(new_users, new_user)
		userName2Id[new_user.Username] = new_user.UserID
	}
	result := tx.Create(new_users)
	return userName2Id, result.Error
}
