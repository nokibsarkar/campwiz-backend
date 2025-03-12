package models

import (
	"nokib/campwiz/consts"
	"time"
)

type User struct {
	UserID       IDType                 `json:"id" gorm:"primaryKey"`
	RegisteredAt time.Time              `json:"registeredAt"`
	Username     WikimediaUsernameType  `json:"username" gorm:"unique;not null;index"`
	Permission   consts.PermissionGroup `json:"permission" gorm:"type:bigint;default:0"`
	// The project this person is leading, because a person can lead only one project
	// This is a many to one relationship
	// it can be null because a person can be a user without leading any project
	// and for most of the users this field will be null
	LeadingProjectID *IDType `json:"projectId" gorm:"index;null"`
	// The project this person is leading
	LeadingProject *Project `json:"project" gorm:"foreignKey:LeadingProjectID;references:ProjectID;constraint:OnUpdate:CASCADE,OnDelete:Set Null"`
}
type ExtendedUserDetails struct {
	User
	Permissions   []consts.PermissionName                     `json:"permissions"`
	PermissionMap map[consts.PermissionName]consts.Permission `json:"permissionMap"`
}
