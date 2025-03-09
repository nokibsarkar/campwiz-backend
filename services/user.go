package services

import (
	"fmt"
	"nokib/campwiz/consts"
	"nokib/campwiz/database"
	idgenerator "nokib/campwiz/services/idGenerator"

	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/util/sets"
)

type UserService struct{}

func NewUserService() *UserService {
	return &UserService{}
}

func (u *UserService) GetUserByID_(conn *gorm.DB, id database.IDType) (*database.User, error) {
	userFilter := &database.User{UserID: id}
	user := &database.User{}
	result := conn.First(user, userFilter)
	if result.Error != nil {
		fmt.Println("Error: ", result.Error)
		return nil, result.Error
	}
	return user, nil

}
func (u *UserService) GetUserByID(id database.IDType) (*database.User, error) {
	user_repo := database.NewUserRepository()
	conn, close := database.GetDB()
	defer close()
	return user_repo.FindByID(conn, id)
}
func (u *UserService) GetExtendedDetails(id database.IDType) (*database.ExtendedUserDetails, error) {
	user_repo := database.NewUserRepository()
	conn, close := database.GetDB()
	defer close()
	user, err := user_repo.FindByID(conn, id)
	if err != nil {
		return nil, err
	}
	details := &database.ExtendedUserDetails{
		User:          *user,
		PermissionMap: consts.GetPermissionMap(),
		Permissions:   user.Permission.GetPermissions(consts.GetPermissionMap()),
	}
	return details, nil
}
func (u *UserService) GetUserByUsername(conn *gorm.DB, username database.WikimediaUsernameType) (*database.User, error) {
	userFilter := &database.User{Username: username}
	user := &database.User{}
	result := conn.First(user, userFilter)
	if result.Error != nil {
		return nil, result.Error
	}
	return user, nil
}
func (u *UserService) GetOrCreateUser(conn *gorm.DB, user *database.User) (*database.User, error) {
	result := conn.FirstOrCreate(user, user)
	if result.Error != nil {
		return nil, result.Error
	}
	return user, nil
}
func (u *UserService) EnsureExists(tx *gorm.DB, usernameSet sets.Set[database.WikimediaUsernameType]) (map[database.WikimediaUsernameType]database.IDType, error) {
	user_repo := database.NewUserRepository()
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
	commons_repo := database.NewCommonsRepository()
	users, err := commons_repo.GeUsersFromUsernames(nonExistentUsers)
	if err != nil {
		return nil, err
	}
	new_users := []database.User{}
	for _, u := range users {
		new_user := database.User{
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
