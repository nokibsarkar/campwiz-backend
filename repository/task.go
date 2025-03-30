// This would be used for running background tasks
package repository

import (
	"nokib/campwiz/models"

	"gorm.io/gorm"
)

type TaskRepository struct{}

func NewTaskRepository() *TaskRepository {
	return &TaskRepository{}
}

func (r *TaskRepository) Create(tx *gorm.DB, task *models.Task) (*models.Task, error) {
	err := tx.Create(task).Error
	if err != nil {
		return nil, err
	}
	return task, nil
}
func (r *TaskRepository) FindByID(tx *gorm.DB, taskId models.IDType) (*models.Task, error) {
	task := &models.Task{}
	err := tx.Limit(1).Find(task, &models.Task{TaskID: taskId}).First(task).Error
	if err != nil {
		return nil, err
	}
	return task, nil
}
func (r *TaskRepository) Update(tx *gorm.DB, task *models.Task) (*models.Task, error) {
	err := tx.Save(task).Error
	if err != nil {
		return nil, err
	}
	return task, nil
}
