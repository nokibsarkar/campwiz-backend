package models

// This is a representative category model for the submission category.
// It has many to one relationship with the submission model.
type Category struct {
	CategoryID     IDType `json:"categoryId" gorm:"primaryKey"`
	Name           string `json:"name" gorm:"not null;uniqueIndex:idx_submission_category"`
	SubmissionName IDType `json:"submissionName" gorm:"not null;uniqueIndex:idx_submission_category;index"`
}
