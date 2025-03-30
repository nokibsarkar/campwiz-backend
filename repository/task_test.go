package repository_test

import (
	"fmt"
	"nokib/campwiz/models"
	"nokib/campwiz/repository"
	"testing"
)

func TestTaskFindByIDError(t *testing.T) {
	// Initialize the database connection
	db, mock, close := repository.GetTestDB()
	defer close()
	// Create a new user repository
	taskRepo := repository.NewTaskRepository()

	// Mock the database behavior
	testTask := &models.Task{
		TaskID: "testtask",
	}
	// mock.ExpectBegin()
	mock.ExpectQuery("SELECT").
		WithArgs(testTask.TaskID, 1).
		WillReturnError(fmt.Errorf("record not found"))
	// mock.ExpectRollback()
	_, err := taskRepo.FindByID(db, testTask.TaskID)
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}
