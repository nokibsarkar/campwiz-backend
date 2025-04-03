package routes

import (
	"log"
	"nokib/campwiz/models"
	"nokib/campwiz/repository/cache"
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
	user, err := user_services.GetExtendedDetails(userID)
	if err != nil {
		c.JSON(403, ResponseError{
			Detail: err.Error(),
		})
		return
	}
	c.Header("Cache-Control", "force-cache")
	c.JSON(200, ResponseSingle[*models.ExtendedUserDetails]{user})
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
	log.Println("Logged out")
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Location", redirect)
	c.JSON(302, ResponseSingle[RedirectResponse]{
		Data: RedirectResponse{Redirect: redirect},
	})
}
