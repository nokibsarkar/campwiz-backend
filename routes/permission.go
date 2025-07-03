package routes

import (
	"nokib/campwiz/consts"
	"nokib/campwiz/models"
	"nokib/campwiz/repository/cache"

	"github.com/gin-gonic/gin"
)

func WithPermission(requiredPermission consts.Permission, handler func(c *gin.Context, sess *cache.Session)) gin.HandlerFunc {
	return WithSession(func(c *gin.Context, sess *cache.Session) {
		if sess.Permission.HasPermission(requiredPermission) {
			handler(c, sess)
		} else {
			c.JSON(403, models.ResponseError{Detail: "Permission denied"})
		}
	})
}

// GetPermissionMap godoc
// @Summary Get the permission map
// @Description Get the permission map
// @Produce json
// @Success 200 {object} models.ResponseSingle[consts.PermissionMap]
// @Router /permissions [get]
// @Tags Permissions
func GetPermissionMap(c *gin.Context) {
	c.JSON(200, models.ResponseSingle[consts.PermissionMap]{
		Data: consts.GetPermissionMap(),
	})
}

func NewPermissionRoutes(parent *gin.RouterGroup) {
	r := parent.Group("/permissions")
	r.GET("/", GetPermissionMap)
}
