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

	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type RedirectResponse struct {
	// Redirect is the URL to redirect to
	Redirect string `json:"redirect"`
}

// HandleOAuth2Callback godoc
// @Summary Handle the OAuth2 callback
// @Description Handle the OAuth2 callback
// @Produce  json
// @Success 200 {object} models.ResponseSingle[RedirectResponse]
// @Router /user/callback [get]
// @Tags User
// @Param code query string true "The code from the OAuth2 provider"
// @Param state query string false "The state"
// @Param baseURL query string false "The base URL"
// @Error 400 {object} models.ResponseError
func HandleOAuth2Callback(c *gin.Context) {
	query := c.Request.URL.Query()
	code := query.Get("code")
	if code == "" {
		c.JSON(400, models.ResponseError{
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
		c.JSON(400, models.ResponseError{
			Detail: err.Error(),
		})
		return
	}
	user, err := oauth2_service.GetUser(accessToken)
	if err != nil {
		c.JSON(400, models.ResponseError{
			Detail: err.Error(),
		})
		return
	}
	conn, close, err := repository.GetDB(c)
	if err != nil {
		c.JSON(500, models.ResponseError{
			Detail: err.Error(),
		})
		return
	}
	defer close()
	user_service := services.NewUserService()
	db_user, err := user_service.GetUserByUsername(conn, user.Name)
	if err != nil {
		log.Println("Error: ", err)
		if err == gorm.ErrRecordNotFound {
			// Create the user
			db_user = &models.User{
				UserID:       idgenerator.GenerateID("u"),
				RegisteredAt: user.Registered,
				Username:     user.Name,
				Permission:   consts.PermissionGroupUSER,
			}
			trx := conn.Create(db_user)
			if trx.Error != nil {
				c.JSON(500, models.ResponseError{
					Detail: trx.Error.Error(),
				})
				return
			}
			log.Println("User created: ", trx.RowsAffected)

		} else {
			c.JSON(500, models.ResponseError{
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
		c.JSON(500, models.ResponseError{
			Detail: err.Error(),
		})
		return
	}
	newRefreshToken, err := auth_service.NewRefreshToken(claims)
	log.Println("Refresh Token :", newRefreshToken)
	if err != nil {
		tx.Rollback()
		c.JSON(500, models.ResponseError{
			Detail: err.Error(),
		})
		return
	}
	c.SetCookie(AuthenticationCookieName, newAccessToken, consts.Config.Auth.Expiry*60, "/", "", false, false)
	c.SetCookie(RefreshCookieName, newRefreshToken, consts.Config.Auth.Refresh*60, "/", "", false, false)
	c.JSON(200, models.ResponseSingle[RedirectResponse]{Data: RedirectResponse{Redirect: state}})
	tx.Commit()
}

func WithSession(callback func(*gin.Context, *cache.Session)) gin.HandlerFunc {

	return func(c *gin.Context) {
		span := sentrygin.GetHubFromContext(c)
		fmt.Printf("Span -1: %v", span)
		session := GetSession(c)
		if session == nil {
			c.JSON(401, models.ResponseError{
				Detail: "Internal Server Error : Session not found",
			})
			return
		}
		callback(c, session)
	}
}
func WithSessionOptional(callback func(*gin.Context, *cache.Session)) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := GetSession(c)
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
	user, err := user_service.GetUserByID(c, session.UserID)
	if err != nil {
		return nil
	}
	return user
}

// RedirectForLogin godoc
// @Summary Redirect to the OAuth2 login
// @Description Redirect to the OAuth2 login
// @Produce  json
// @Success 200 {object} models.ResponseSingle[RedirectResponse]
// @Router /user/login [get]
// @Tags User
// @Param callback query string false "The callback URL"
// @Error 400 {object} models.ResponseError
func RedirectForLogin(c *gin.Context) {
	oauth2_service := services.NewOAuth2Service()
	callback, ok := c.GetQuery("next")
	if !ok {
		callback = "/"
	}
	redirect_uri := oauth2_service.Init(callback)
	c.JSON(200, models.ResponseSingle[RedirectResponse]{Data: RedirectResponse{Redirect: redirect_uri}})
}
