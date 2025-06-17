package routes

import (
	"log"
	"nokib/campwiz/consts"
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
	user, err := user_services.GetExtendedDetails(c, userID)
	if err != nil {
		c.JSON(403, models.ResponseError{
			Detail: err.Error(),
		})
		return
	}
	c.Header("Cache-Control", "force-cache")
	c.JSON(200, models.ResponseSingle[*models.ExtendedUserDetails]{Data: user})
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
		err := auth_service.Logout(c, session)
		if err != nil {
			c.JSON(500, models.ResponseError{
				Detail: err.Error(),
			})
			return
		}
	}
	c.SetCookie(consts.AuthenticationCookieName, "", -1, "/", "", false, true)
	c.SetCookie(consts.RefreshCookieName, "", -1, "/", "", false, true)
	log.Println("Logged out")
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Location", redirect)
	c.JSON(200, models.ResponseSingle[RedirectResponse]{
		Data: RedirectResponse{Redirect: redirect},
	})
}
