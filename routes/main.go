package routes

import (
	"nokib/campwiz/models"
	"nokib/campwiz/repository/cache"
	"nokib/campwiz/services"

	"github.com/gin-gonic/gin"
)

func WithSession(callback func(*gin.Context, *cache.Session)) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := GetSession(c)
		if session == nil {
			c.JSON(401, ResponseError{
				Detail: "Internal Server Error : Session not found",
			})
			return
		}
		callback(c, session)
	}
}
func GetSession(c *gin.Context) *cache.Session {
	sess, ok := c.Get(SESSION_KEY)
	if !ok {
		return nil
	}
	session, ok := sess.(*cache.Session)
	if !ok {
		return nil
	}
	return session
}
func GetCurrentUser(c *gin.Context) *models.User {
	session := GetSession(c)
	if session == nil {
		return nil
	}
	user_service := services.NewUserService()
	user, err := user_service.GetUserByID(session.UserID)
	if err != nil {
		return nil
	}
	return user
}
func NewRoutes(nonAPIParent *gin.RouterGroup) *gin.RouterGroup {
	r := nonAPIParent.Group("/api/v2")
	authenticatorService := NewAuthenticationService()
	r.Use(authenticatorService.Authenticate)
	NewPermissionRoutes(nonAPIParent)
	NewUserAuthenticationRoutes(nonAPIParent)
	NewCampaignRoutes(r)
	NewRoundRoutes(r)
	NewSubmissionRoutes(r)
	NewUserRoutes(r)
	NewTaskRoutes(r)
	NewEvaluationRoutes(r)
	NewProjectRoutes(r)
	return r
}
