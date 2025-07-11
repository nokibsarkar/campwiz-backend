package routes

import (
	"log"
	"nokib/campwiz/consts"
	"nokib/campwiz/models"
	"nokib/campwiz/repository/cache"
	"nokib/campwiz/services"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type RedirectResponse struct {
	// Redirect is the URL to redirect to
	Redirect string `json:"redirect"`
}

// HandleOAuth2IdentityVerificationCallback godoc
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
func HandleOAuth2IdentityVerificationCallback(c *gin.Context) {
	oauth2_service := services.NewOAuth2Service(c, consts.Config.Auth.GetOAuth2IdentityVerificationOauthConfig(), consts.Config.Auth.OAuth2IdentityVerification.RedirectPath)
	db_user, state, _, err := oauth2_service.FetchTokenFromWikimediaServer()
	if err != nil {
		c.JSON(400, models.ResponseError{
			Detail: err.Error(),
		})
		return
	}
	// we can assume that the user is created
	// we can now create the session
	auth_service := services.NewAuthenticationService()
	cacheDB, close := cache.GetCacheDB(c)
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
	c.SetCookie(consts.AuthenticationCookieName, newAccessToken, consts.Config.Auth.Expiry*60, "/", "", false, false)
	c.SetCookie(consts.RefreshCookieName, newRefreshToken, consts.Config.Auth.Refresh*60, "/", "", false, false)
	c.JSON(200, models.ResponseSingle[RedirectResponse]{Data: RedirectResponse{Redirect: state}})
	tx.Commit()
}

func WithSession(callback func(*gin.Context, *cache.Session)) gin.HandlerFunc {

	return func(c *gin.Context) {
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
	sess, ok := c.Get(consts.SESSION_KEY)
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
// @Summary Redirect to the OAuth2 login ReadOnly scope
// @Description Redirect to the OAuth2 login for ReadOnly scope
// @Produce  json
// @Success 200 {object} models.ResponseSingle[RedirectResponse]
// @Router /user/login [get]
// @Tags User
// @Param callback query string false "The callback URL"
// @Error 400 {object} models.ResponseError
func RedirectForLogin(c *gin.Context) {
	oauth2_service := services.NewOAuth2Service(c, consts.Config.Auth.GetOAuth2IdentityVerificationOauthConfig(), consts.Config.Auth.OAuth2IdentityVerification.RedirectPath)
	callback, ok := c.GetQuery("next")
	if !ok {
		callback = "/"
	}
	redirect_uri := oauth2_service.Init(callback)
	c.JSON(200, models.ResponseSingle[RedirectResponse]{Data: RedirectResponse{Redirect: redirect_uri}})
}

// RedirectForLogin godoc
// @Summary Redirect to the OAuth2 login
// @Description Redirect to the OAuth2 login for ReadWrite scope.
// @Produce  json
// @Success 200 {object} models.ResponseSingle[RedirectResponse]
// @Router /user/login/write [get]
// @Tags User
// @Param callback query string false "The callback URL"
// @Error 400 {object} models.ResponseError
func RedirectForLoginWrite(c *gin.Context) {
	if consts.Config.Auth.Oauth2WriteAccess == nil {
		c.JSON(400, models.ResponseError{
			Detail: "OAuth2 ReadWrite is not configured",
		})
		return
	}
	oauth2_service := services.NewOAuth2Service(c, consts.Config.Auth.GetOAuth2ReadWriteOauthConfig(), consts.Config.Auth.Oauth2WriteAccess.RedirectPath)
	callback, ok := c.GetQuery("next")
	if !ok {
		callback = "/"
	}
	log.Printf("Redirecting to OAuth2 Write login with callback: %s", oauth2_service.Config.Endpoint.AuthURL)
	redirect_uri := oauth2_service.Init(callback)
	c.JSON(200, models.ResponseSingle[RedirectResponse]{Data: RedirectResponse{Redirect: redirect_uri}})
}

// HandleOAuth2ReadWriteCallback godoc
// @Summary Handle the OAuth2 callback for ReadWrite scope
// @Description Handle the OAuth2 callback for the ReadWrite scope. This endpoint would fetch an access token and set it as a cookie, it would not, by any means, store it on the server. Refresh Token would also be set as a cookie.
// @Produce  json
// @Success 200 {object} models.ResponseSingle[RedirectResponse]
// @Router /user/callback/write [get]
// @Tags User
// @Param code query string true "The code from the OAuth2 provider"
// @Param state query string false "The state"
// @Param baseURL query string false "The base URL"
// @Error 400 {object} models.ResponseError
func HandleOAuth2ReadWriteCallback(c *gin.Context) {
	if consts.Config.Auth.Oauth2WriteAccess == nil {
		c.JSON(400, models.ResponseError{
			Detail: "OAuth2 ReadWrite is not configured",
		})
		return
	}
	oauth_service := services.NewOAuth2Service(c, consts.Config.Auth.GetOAuth2ReadWriteOauthConfig(), consts.Config.Auth.Oauth2WriteAccess.RedirectPath)
	_, state, newAccessToken, err := oauth_service.FetchTokenFromWikimediaServer()
	if err != nil {
		c.JSON(400, models.ResponseError{
			Detail: err.Error(),
		})
		return
	}
	// we can assume that the user is created
	expiresIn := int(newAccessToken.Expiry.UTC().Unix() - time.Now().UTC().Unix())
	c.SetCookie(consts.ReadWriteAuthenticationCookieName, newAccessToken.AccessToken, expiresIn, "/", "", false, false)
	// we can also set the refresh token, expires in 7 days
	c.SetCookie(consts.ReadWriteRefreshCookieName, newAccessToken.RefreshToken, expiresIn+7*24*3600, "/", "", false, false)
	c.JSON(200, models.ResponseSingle[RedirectResponse]{Data: RedirectResponse{Redirect: state}})
}
