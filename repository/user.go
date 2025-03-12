package repository

import (
	"log"
	"nokib/campwiz/consts"
	"nokib/campwiz/models"

	"gorm.io/gorm"
)

type UserRepository struct{}

func NewUserRepository() *UserRepository {
	return &UserRepository{}
}
func (u *UserRepository) FetchExistingUsernames(conn *gorm.DB, usernames []models.WikimediaUsernameType) (map[models.WikimediaUsernameType]models.IDType, error) {
	type APIUser struct {
		Username models.WikimediaUsernameType
		UserID   models.IDType
	}
	exists := []APIUser{}

	res := conn.Model(&models.User{}).Limit(len(usernames)).Find(&APIUser{}, "username IN (?)", usernames).Find(&exists)
	if res.Error != nil {
		return nil, res.Error
	}
	userName2IDMap := map[models.WikimediaUsernameType]models.IDType{}
	for _, u := range exists {
		userName2IDMap[u.Username] = u.UserID
	}
	return userName2IDMap, nil

}
func (u *UserRepository) EnsureExists(tx *gorm.DB, usernameToRandomIdMap map[models.WikimediaUsernameType]models.IDType) (map[models.WikimediaUsernameType]models.IDType, error) {
	usernames := []models.WikimediaUsernameType{}
	if len(usernameToRandomIdMap) == 0 {
		return usernameToRandomIdMap, nil
	}
	for username := range usernameToRandomIdMap {
		usernames = append(usernames, username)
	}
	userName2Id, err := u.FetchExistingUsernames(tx, usernames)
	if err != nil {
		return nil, err
	}
	if len(userName2Id) > 0 {
		for username := range userName2Id {
			delete(usernameToRandomIdMap, username)
		}
	}
	if len(usernameToRandomIdMap) == 0 {
		return userName2Id, nil
	}
	nonExistentUsers := make([]models.WikimediaUsernameType, 0, len(usernameToRandomIdMap))
	for nonExistingUsername := range usernameToRandomIdMap {
		nonExistentUsers = append(nonExistentUsers, nonExistingUsername)
	}
	commons_repo := NewCommonsRepository()
	users, err := commons_repo.GeUsersFromUsernames(nonExistentUsers)
	if err != nil {
		log.Printf("Error fetching users: %v\n", err)
		return nil, err
	}
	log.Println("Fetched users: ", users)
	new_users := []models.User{}
	for _, u := range users {
		new_user := models.User{
			UserID:       usernameToRandomIdMap[u.Name],
			RegisteredAt: u.Registered,
			Username:     u.Name,
			Permission:   consts.PermissionGroupUSER,
		}
		new_users = append(new_users, new_user)
		userName2Id[new_user.Username] = new_user.UserID
	}
	if len(new_users) == 0 {
		return userName2Id, nil
	}
	result := tx.Create(new_users)
	return userName2Id, result.Error
}
func (u *UserRepository) FindByID(tx *gorm.DB, userID models.IDType) (*models.User, error) {
	user := &models.User{}
	result := tx.Limit(1).Find(user, &models.User{UserID: userID})
	return user, result.Error
}
func (u *UserRepository) FindProjectLeads(tx *gorm.DB, projectID *models.IDType) ([]models.User, error) {
	users := []models.User{}
	result := tx.Find(&users, &models.User{LeadingProjectID: projectID})
	return users, result.Error
}
func (u *UserRepository) FetchRoles(tx *gorm.DB, filter *models.RoleFilter) ([]models.Role, error) {
	roles := []models.Role{}
	result := tx.Find(&roles, filter)
	if result.Error != nil {
		return nil, result.Error
	}
	return roles, nil
}
