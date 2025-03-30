package repository_test

import (
	"fmt"
	"nokib/campwiz/models"
	"nokib/campwiz/repository"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/gorm"
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
func taskCreate(db *gorm.DB, mock sqlmock.Sqlmock, testTask *models.Task) (*models.Task, error) {
	taskRepo := repository.NewTaskRepository()
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `tasks`").
		//(`task_id`,`type`,`status`,`associated_round_id`,`associated_campaign_id`,`associated_user_id`,`data`,`created_at`,`updated_at`,`success_count`,`failed_count`,`failed_ids`,`remaining_count`,`created_by_id`)
		WithArgs(testTask.TaskID, testTask.Type, testTask.Status,
			testTask.AssociatedRoundID, testTask.AssociatedCampaignID, testTask.AssociatedUserID,
			testTask.Data, testTask.CreatedAt,
			testTask.UpdatedAt,
			testTask.SuccessCount, testTask.FailedCount,
			nil, testTask.RemainingCount, testTask.CreatedByID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	createdTask, err := taskRepo.Create(db, testTask)
	return createdTask, err
}
func TestTaskCreateSuccess(t *testing.T) {
	// Initialize the database connection
	db, mock, close := repository.GetTestDB()
	defer close()
	// Mock the database behavior
	testTask := &models.Task{
		TaskID:               "testtask",
		Type:                 models.TaskTypeDistributeEvaluations,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
		AssociatedRoundID:    nil,
		AssociatedCampaignID: nil,
		AssociatedUserID:     nil,
		Data:                 nil,
		SuccessCount:         0,
	}
	createdTask, err := taskCreate(db, mock, testTask)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if createdTask.TaskID != testTask.TaskID {
		t.Errorf("expected task ID %s, got %s", testTask.TaskID, createdTask.TaskID)
	}
	if createdTask.Type != testTask.Type {
		t.Errorf("expected task type %s, got %s", testTask.Type, createdTask.Type)
	}
	if createdTask.Status != testTask.Status {
		t.Errorf("expected task status %s, got %s", testTask.Status, createdTask.Status)
	}

}
func TestTaskCreateError(t *testing.T) {
	// Initialize the database connection
	db, mock, close := repository.GetTestDB()
	defer close()
	// Mock the database behavior
	testTask := &models.Task{
		TaskID:               "",
		Type:                 models.TaskTypeDistributeEvaluations,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
		AssociatedRoundID:    nil,
		AssociatedCampaignID: nil,
		AssociatedUserID:     nil,
		Data:                 nil,
	}
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `tasks`").
		WithArgs(testTask.TaskID, testTask.Type, testTask.Status,
			testTask.AssociatedRoundID, testTask.AssociatedCampaignID, testTask.AssociatedUserID,
			testTask.Data, testTask.CreatedAt,
			testTask.UpdatedAt,
			testTask.SuccessCount, testTask.FailedCount,
			nil, testTask.RemainingCount, testTask.CreatedByID).
		WillReturnError(fmt.Errorf("error inserting task"))
	mock.ExpectRollback()
	taskRepo := repository.NewTaskRepository()
	_, err := taskRepo.Create(db, testTask)
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}
func TestTaskFindByIDSuccess(t *testing.T) {
	// Initialize the database connection
	db, mock, close := repository.GetTestDB()
	defer close()
	// Mock the database behavior
	testTask := &models.Task{
		TaskID:               "testtask",
		Type:                 models.TaskTypeDistributeEvaluations,
		Status:               models.TaskStatusPending,
		AssociatedRoundID:    nil,
		AssociatedCampaignID: nil,
		AssociatedUserID:     nil,
		Data:                 nil,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
		SuccessCount:         0,
		FailedCount:          0,
	}
	createdTask, err := taskCreate(db, mock, testTask)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if createdTask.TaskID != testTask.TaskID {
		t.Errorf("expected task ID %s, got %s", testTask.TaskID, createdTask.TaskID)
	}
	mock.ExpectQuery("SELECT").
		WithArgs(testTask.TaskID, 1).
		WillReturnRows(sqlmock.NewRows([]string{"task_id", "type", "status", "associated_round_id", "associated_campaign_id", "associated_user_id", "data"}).
			AddRow(testTask.TaskID, testTask.Type, testTask.Status, testTask.AssociatedRoundID, testTask.AssociatedCampaignID, testTask.AssociatedUserID, testTask.Data))
	mock.ExpectQuery("SELECT").
		WithArgs(testTask.TaskID, 1).
		WillReturnRows(sqlmock.NewRows([]string{"task_id", "type", "status", "associated_round_id", "associated_campaign_id", "associated_user_id", "data"}).
			AddRow(testTask.TaskID, testTask.Type, testTask.Status, testTask.AssociatedRoundID, testTask.AssociatedCampaignID, testTask.AssociatedUserID, testTask.Data))

	taskRepo := repository.NewTaskRepository()
	foundTask, err := taskRepo.FindByID(db, testTask.TaskID)
	if err != nil {
		t.Errorf("expected1 no error, got %v", err)
		return
	}
	if foundTask == nil {
		t.Errorf("expected task, got nil")
		return
	}
	if foundTask.TaskID != testTask.TaskID {
		t.Errorf("expected task ID %s, got %s", testTask.TaskID, foundTask.TaskID)
	}
}
