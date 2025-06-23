package repository

import (
	"fmt"
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

	// Use case-sensitive comparison - remove BINARY as it's causing collation issues
	res := conn.Model(&models.User{}).Limit(len(usernames)).Where("username IN (?)", usernames).Find(&exists)
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
			fmt.Printf("User `%s` already exists with ID %s\n", username, userName2Id[username])
			// Remove from the map of users to create (exact case-sensitive match)
			delete(usernameToRandomIdMap, username)
		}
	}

	log.Printf("Remaining users to be created: %v\n", usernameToRandomIdMap)
	if len(usernameToRandomIdMap) == 0 {
		return userName2Id, nil
	}

	// Only try to create users that are truly non-existent (case-sensitive)
	nonExistentUsers := make([]models.WikimediaUsernameType, 0, len(usernameToRandomIdMap))
	for nonExistingUsername := range usernameToRandomIdMap {
		// Final check: make sure this username doesn't already exist (case-sensitive)
		if _, exists := userName2Id[nonExistingUsername]; !exists {
			nonExistentUsers = append(nonExistentUsers, nonExistingUsername)
		} else {
			fmt.Printf("Skipping creation of `%s` as it already exists (exact match)\n", nonExistingUsername)
		}
	}
	fmt.Printf("Fetching non-existent users: %v\n", nonExistentUsers)
	commons_repo := NewCommonsRepository(nil)
	users, err := commons_repo.GeUsersFromUsernames(nonExistentUsers)
	if err != nil {
		log.Printf("Error fetching users: %v\n", err)
		return nil, err
	}
	log.Println("Fetched users: ", users)
	new_users := []models.User{}
	for _, u := range users {
		// Check if this username (from API) already exists in database
		if existingID, exists := userName2Id[u.Name]; exists {
			fmt.Printf("API returned user `%s` but it already exists with ID %s, skipping creation\n", u.Name, existingID)
			continue
		}

		// Find the corresponding random ID for this user
		var randomID models.IDType
		found := false
		for requestedUsername, randomUserID := range usernameToRandomIdMap {
			if requestedUsername == u.Name {
				randomID = randomUserID
				found = true
				break
			}
		}

		if !found {
			fmt.Printf("Warning: API returned user `%s` that we didn't request, skipping\n", u.Name)
			continue
		}

		new_user := models.User{
			UserID:       randomID,
			RegisteredAt: u.Registered,
			Username:     u.Name,
			Permission:   consts.PermissionGroupUSER,
		}
		new_users = append(new_users, new_user)
		userName2Id[new_user.Username] = new_user.UserID
		fmt.Printf("Prepared user `%s` for creation with ID %s\n", new_user.Username, new_user.UserID)
	}
	if len(new_users) == 0 {
		log.Printf("No new users to create after filtering\n")
		return userName2Id, nil
	}

	fmt.Printf("About to create %d users in database:\n", len(new_users))
	for i, user := range new_users {
		fmt.Printf("  %d. Username: '%s', UserID: '%s'\n", i+1, user.Username, user.UserID)
	}

	result := tx.Create(new_users)
	if result.Error != nil {
		log.Printf("Database error creating users: %v\n", result.Error)
	} else {
		log.Printf("Successfully created %d users\n", len(new_users))
	}
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
