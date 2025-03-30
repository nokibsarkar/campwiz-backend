package repository_test

import (
	"fmt"
	"nokib/campwiz/models"
	"nokib/campwiz/repository"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
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
func TestTaskCreateSuccess(t *testing.T) {
	// Initialize the database connection
	db, mock, close := repository.GetTestDB()
	defer close()
	// Create a new user repository
	taskRepo := repository.NewTaskRepository()

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
	_, err := taskRepo.Create(db, testTask)
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}
