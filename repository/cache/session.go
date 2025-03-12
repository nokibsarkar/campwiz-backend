package cache

import (
	"nokib/campwiz/consts"
	"nokib/campwiz/models"
	"time"
)

type Session struct {
	ID         models.IDType                `json:"id" gorm:"primaryKey"`
	UserID     models.IDType                `json:"userId"`
	Username   models.WikimediaUsernameType `json:"username"`
	Permission consts.PermissionGroup       `json:"permission"`
	ExpiresAt  time.Time                    `json:"expiresAt"`
}
