package routes

import (
	"fmt"
	"log"
	"nokib/campwiz/consts"
	"nokib/campwiz/models"
	"nokib/campwiz/repository"
	"nokib/campwiz/repository/cache"
	"nokib/campwiz/services"
	idgenerator "nokib/campwiz/services/idGenerator"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

func HandleOAuth2Callback(c *gin.Context) {
	query := c.Request.URL.Query()
	code := query.Get("code")
	if code == "" {
		c.JSON(400, ResponseError{
			Detail: "No code found in the query",
		})
		return
	}
	state := query.Get("state")
	if state == "" || strings.HasPrefix(state, "/user/login") {
		state = "/"
	}
	baseURL := consts.Config.Server.BaseURL
	baseURLRaw, ok := c.GetQuery("baseURL")
	if ok {
		baseURL = baseURLRaw
	}
	oauth2_service := services.NewOAuth2Service()
	accessToken, err := oauth2_service.GetToken(code, baseURL+consts.Config.Auth.OAuth2.RedirectPath)
	if err != nil {
		c.JSON(400, ResponseError{
			Detail: err.Error(),
		})
		return
	}
	user, err := oauth2_service.GetUser(accessToken)
	if err != nil {
		c.JSON(400, ResponseError{
			Detail: err.Error(),
		})
		return
	}
	conn, close := repository.GetDB()
	defer close()
	user_service := services.NewUserService()
	db_user, err := user_service.GetUserByUsername(conn, user.Name)
	if err != nil {
		fmt.Println("Error: ", err)
		if err == gorm.ErrRecordNotFound {
			// Create the user
			db_user = &models.User{
				UserID:       idgenerator.GenerateID("usr"),
				RegisteredAt: user.Registered,
				Username:     user.Name,
				Permission:   consts.PermissionGroupADMIN,
			}
			trx := conn.Create(db_user)
			if trx.Error != nil {
				c.JSON(500, ResponseError{
					Detail: trx.Error.Error(),
				})
				return
			}
			log.Println("User created: ", trx.RowsAffected)

		} else {
			c.JSON(500, ResponseError{
				Detail: err.Error(),
			})
			return
		}
	}
	// we can assume that the user is created
	// we can now create the session
	auth_service := services.NewAuthenticationService()
	cacheDB, close := cache.GetCacheDB()
	defer close()
	nextExpiry := time.Now().UTC().Add(time.Minute * time.Duration(consts.Config.Auth.Expiry))
	log.Println("Session expire at : ", nextExpiry, "Now :", time.Now().UTC())
	claims := &services.SessionClaims{
		Permission: consts.PermissionGroup(db_user.Permission),
		Name:       db_user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			Audience:  jwt.ClaimStrings{"campwiz"},
			Subject:   string(db_user.UserID),
			Issuer:    consts.Config.Auth.Issuer,
			ExpiresAt: jwt.NewNumericDate(nextExpiry),
		},
	}
	tx := cacheDB.Begin()
	newAccessToken, _, err := auth_service.NewSession(tx, claims)
	log.Println("New Access token ", newAccessToken)
	if err != nil {
		tx.Rollback()
		c.JSON(500, ResponseError{
			Detail: err.Error(),
		})
		return
	}
	newRefreshToken, err := auth_service.NewRefreshToken(claims)
	log.Println("Refresh Token :", newRefreshToken)
	if err != nil {
		tx.Rollback()
		c.JSON(500, ResponseError{
			Detail: err.Error(),
		})
		return
	}
	c.SetCookie(AuthenticationCookieName, newAccessToken, consts.Config.Auth.Expiry*60, "/", "", false, false)
	c.SetCookie(RefreshCookieName, newRefreshToken, consts.Config.Auth.Refresh*60, "/", "", false, false)
	c.Redirect(302, state)
	tx.Commit()
}
func RedirectForLogin(c *gin.Context) {
	oauth2_service := services.NewOAuth2Service()
	callback, ok := c.GetQuery("callback")
	if !ok {
		callback = "/"
	}
	redirect_uri := oauth2_service.Init(callback)
	c.Redirect(302, redirect_uri)
}

func NewUserAuthenticationRoutes(parent *gin.RouterGroup) {
	user := parent.Group("/")
	user.GET("/user/login", RedirectForLogin)
	user.GET("/user/callback", HandleOAuth2Callback)
}
