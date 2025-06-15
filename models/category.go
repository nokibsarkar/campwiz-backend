package models

import (
	"time"

	"gorm.io/gorm"
)

// This is a representative category model for the submission category.
// It has many to one relationship with the submission model.
type CategoryWithWriteableFields struct {
	CategoryName string  `json:"categoryName" gorm:"not null;primaryKey;uniqueIndex:idx_submission_category;index"`
	SubmissionID IDType  `json:"submissionId" gorm:"not null;primaryKey;uniqueIndex:idx_submission_category"` // The submission this category belongs to
	AddedByID    *IDType `json:"addedById" gorm:"null;index;uniqueIndex:idx_submission_category"`             // The user who added this category
}
type Category struct {
	CategoryWithWriteableFields
	CreatedAt  *time.Time  `json:"createdAt" gorm:"<-:create"`                                                    // The time the category was created
	AddedBy    *User       `json:"-" gorm:"foreignKey:AddedByID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`   // The user who added this category
	Submission *Submission `json:"-" gorm:"foreignKey:SubmissionID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"` // The submission this category belongs to
	gorm.DeletedAt
}
type CategoryResponse struct {
	Added   []string `json:"added"`
	Removed []string `json:"removed"`
}
