package routes

import (
	"errors"
	"fmt"
	"log"
	"nokib/campwiz/consts"
	"nokib/campwiz/models"
	"nokib/campwiz/services"
	"strings"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
)

/*
This is the authentication service. It is used to authenticate users.
It would usually access the cache database to check if the user is authenticated.
It would access the JWT
*/
type AuthenticationMiddleWare struct {
	Config *consts.AuthenticationConfiguration
}

const AuthenticationCookieName = "c-auth"
const RefreshCookieName = "X-Refresh-Token"
const SESSION_KEY = "session"

func NewAuthenticationService() *AuthenticationMiddleWare {
	return &AuthenticationMiddleWare{
		Config: &consts.Config.Auth,
	}
}

// This function extracts the access token from the cookies or headers
func (a *AuthenticationMiddleWare) extractAccessToken(c *gin.Context) (string, error) {
	token, _ := c.Cookie(AuthenticationCookieName)
	if token != "" {
		return token, nil
	}
	// Check if the token is in the headers
	token = c.GetHeader("Authorization")
	if token != "" {
		if token[:7] == "Bearer " {
			return token[7:], nil
		}
	}
	return "", errors.New("no-access-token")
}

func (a *AuthenticationMiddleWare) checkIfUnauthenticatedAllowed(c *gin.Context) bool {
	fmt.Println("Checking if unauthenticated allowed", c.Request.Method, c.Request.URL.Path)
	if c.Request.Method != "GET" {
		return false
	}
	path := c.Request.URL.Path
	// return true
	var UnRestrictedPaths = []string{
		"/user/login",
		"/user/callback",
		"/api/v2/campaign/",
		// "/api/v2/user/logout",
	}
	for _, p := range UnRestrictedPaths {
		if strings.HasSuffix(p, "*") {
			rest := p[:len(p)-1]
			if strings.HasPrefix(path, rest) {
				return true
			}
		} else {
			if p == path {
				log.Println("Unrestricted path")
				return true
			}
		}
	}
	return false
}

/*
This is the authenticator middleware. It is used to authenticate users.
It would usually access the cache database to check if the user is authenticated.
It would access the JWT
*/
func (a *AuthenticationMiddleWare) Authenticate(c *gin.Context) {
	if !a.checkIfUnauthenticatedAllowed(c) {
		token, err := a.extractAccessToken(c)
		if err != nil {
			log.Println("Error", err)
			c.Set("error", err)
			c.AbortWithStatusJSON(401, models.ResponseError{Detail: "Unauthorized : No token found"})
			return
		} else {
			auth_service := services.NewAuthenticationService()
			accessToken, session, err, setCookie := auth_service.Authenticate(token)
			if err != nil {
				log.Println("Error", err)
				c.Set("error", err)
				c.AbortWithStatusJSON(401, models.ResponseError{Detail: "Unauthorized : Invalid token"})
				return
			} else {
				if setCookie {
					log.Println("Setting Authentication Cookie")
					c.SetCookie(AuthenticationCookieName, accessToken, consts.Config.Auth.Expiry, "/", "", false, true)
				}
				c.Set(SESSION_KEY, session)
			}
		}
		c.Next()
	}
}
func (a *AuthenticationMiddleWare) Authenticate2(c *gin.Context) {
	token, err := a.extractAccessToken(c)
	if err != nil {
		if !a.checkIfUnauthenticatedAllowed(c) {
			log.Println("Error", err)
			c.Set("error", err)
			c.AbortWithStatusJSON(401, models.ResponseError{Detail: "Unauthorized : No token found"})
			return
		}
	} else {

		auth_service := services.NewAuthenticationService()
		accessToken, session, err, setCookie := auth_service.Authenticate(token)
		if err != nil {
			if !a.checkIfUnauthenticatedAllowed(c) {
				log.Println("Error", err)
				c.Set("error", err)
				c.AbortWithStatusJSON(401, models.ResponseError{Detail: "Unauthorized : Invalid token"})
				return
			}
		} else {
			if setCookie {
				log.Println("Setting Authentication Cookie")
				c.SetCookie(AuthenticationCookieName, accessToken, consts.Config.Auth.Expiry, "/", "", false, true)
			}
			c.Set(SESSION_KEY, session)
		}
		if hub := sentrygin.GetHubFromContext(c); hub != nil && session != nil {
			hub.Scope().SetUser(sentry.User{
				ID:       session.UserID.String(),
				Username: string(session.Username),
				// IPAddress: c.ClientIP(),
				Name: string(session.Username),
				Data: map[string]string{
					"sessionId":  session.ID.String(),
					"permission": fmt.Sprintf("%d", session.Permission),
					"expiresAt":  session.ExpiresAt.String(),
				},
			})
		}
	}
	c.Next()
}
