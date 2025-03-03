package routes

import (
	"nokib/campwiz/database"
	"nokib/campwiz/database/cache"
	"nokib/campwiz/services"

	"github.com/gin-gonic/gin"
)

func ListUsers(c *gin.Context) {
	// ...
}
func GetMe(c *gin.Context, session *cache.Session) {
	// ...
	userID := session.UserID
	user_services := services.NewUserService()
	user, err := user_services.GetUserByID(userID)
	if err != nil {
		c.JSON(403, ResponseError{
			Detail: err.Error(),
		})
		return
	}
	c.JSON(200, ResponseSingle[*database.User]{user})
}
func GetUser(c *gin.Context) {
	// ...
}
func UpdateUser(c *gin.Context) {
	// ...
}
func GetTranslationPath(c *gin.Context) {
	// ...
}
func GetTranslation(c *gin.Context) {
	// ...
}
func UpdateTranslation(c *gin.Context) {
	// ...
}
func Logout(c *gin.Context) {
	redirect := c.Query("next")
	if redirect == "" {
		redirect = "/"
	}
	session := GetSession(c)
	if session != nil {
		auth_service := services.NewAuthenticationService()
		err := auth_service.Logout(session)
		if err != nil {
			c.JSON(500, ResponseError{
				Detail: err.Error(),
			})
			return
		}
	}
	c.SetCookie(AuthenticationCookieName, "", -1, "/", "", false, true)
	c.SetCookie(RefreshCookieName, "", -1, "/", "", false, true)
	c.Redirect(302, redirect)
}

func NewUserRoutes(parent *gin.RouterGroup) {
	r := parent.Group("/user")
	r.GET("/", ListUsers)
	r.GET("/me", WithSession(GetMe))
	r.GET("/:id", GetUser)
	r.POST("/:id", UpdateUser)
	r.GET("/translation/:language", GetTranslation)
	r.POST("/translation/:lang", UpdateTranslation)
	r.GET("/logout", Logout)
}
