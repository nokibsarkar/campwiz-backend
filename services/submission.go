package services

import (
	"context"
	"nokib/campwiz/models"
	"nokib/campwiz/models/types"
	"nokib/campwiz/repository"
)

type SubmissionService struct{}

func NewSubmissionService() *SubmissionService {
	return &SubmissionService{}
}

func (s *SubmissionService) ListAllSubmissions(ctx context.Context, filter *models.SubmissionListFilter) ([]models.Submission, error) {
	conn, close, err := repository.GetDB(ctx)
	if err != nil {
		return nil, err
	}
	defer close()
	submission_repo := repository.NewSubmissionRepository()
	submissions, err := submission_repo.ListAllSubmissions(conn, filter)
	if err != nil {
		return nil, err
	}
	return submissions, nil
}
func (s *SubmissionService) GetSubmission(ctx context.Context, submissionID types.SubmissionIDType) (*models.Submission, error) {
	conn, close, err := repository.GetDB(ctx)
	if err != nil {
		return nil, err
	}
	defer close()
	submission_repo := repository.NewSubmissionRepository()
	submission, err := submission_repo.FindSubmissionByID(conn, submissionID)
	if err != nil {
		return nil, err
	}
	return submission, nil
}
