package repository_test

import (
	"fmt"
	"nokib/campwiz/models"
	"nokib/campwiz/repository"
	"testing"
	"time"
)

// Create a new user
var user = &models.User{
	UserID:       "testuser",
	Username:     "testuser",
	RegisteredAt: time.Now(),
	Permission:   1,
}

func init() {
	// // Initialize the database connection
	// _, _, close := repository.GetTestDB()
	// defer close()
	// // Auto-migrate the models
	// // db.AutoMigrate(&models.User{})
}
func TestUserFindByIDError(t *testing.T) {
	// Initialize the database connection
	db, mock, close := repository.GetTestDB()
	defer close()
	// Create a new user repository
	userRepo := repository.NewUserRepository()

	// Mock the database behavior

	// mock.ExpectBegin()
	mock.ExpectQuery("SELECT").
		WithArgs(user.UserID, 1).
		WillReturnError(fmt.Errorf("record not found"))
	// mock.ExpectRollback()
	_, err := userRepo.FindByID(db, user.UserID)
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}
