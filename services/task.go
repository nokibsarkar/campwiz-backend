package services

import (
	"nokib/campwiz/models"
	"nokib/campwiz/repository"
)

type TaskService struct{}

func NewTaskService() *TaskService {
	return &TaskService{}
}

type TaskRequest struct {
	Type                 string `json:"type"`
	AssociatedRoundID    string `json:"roundId"`
	AssociatedCampaignID string `json:"campaignId"`
	AssociatedUserID     string `json:"userId"`
	Data                 any    `json:"data"`
	CreatedByID          string `json:"createdById"`
	HandlerFunc          func(c ...any)
}
type TaskResponse struct {
	models.Task
	handlerFunc func(c ...any)
}
type TaskFilter struct {
	Type                 string `json:"type"`
	AssociatedRoundID    string `json:"roundId"`
	AssociatedCampaignID string `json:"campaignId"`
	AssociatedUserID     string `json:"userId"`
	Status               string `json:"status"`
}

func (t *TaskResponse) Start() {
	go t.handlerFunc()
}

func (t *TaskService) CreateTask(TaskRequest, handlerFunc func(c ...any)) (*TaskResponse, error) {

	return nil, nil
}
func (t *TaskService) GetTask(taskId models.IDType) (*models.Task, error) {
	task_repo := repository.NewTaskRepository()
	conn, close, err := repository.GetDB()
	if err != nil {
		return nil, err
	}
	defer close()
	return task_repo.FindByID(conn, taskId)
}
func (t *TaskService) ListTasks(filter *TaskFilter) ([]models.Task, error) {
	return nil, nil
}
