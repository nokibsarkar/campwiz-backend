package services

import (
	"bytes"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"log"
	"nokib/campwiz/consts"
	"nokib/campwiz/models"
	"nokib/campwiz/repository/cache"
	idgenerator "nokib/campwiz/services/idGenerator"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"

	sentrygin "github.com/getsentry/sentry-go/gin"
)

var RSAPrivateKey *ecdsa.PrivateKey
var RSAPublicKey *ecdsa.PublicKey

type AuthenticationService struct {
	Config *consts.AuthenticationConfiguration
}
type SessionClaims struct {
	Permission consts.PermissionGroup       `json:"permission"`
	Name       models.WikimediaUsernameType `json:"name"`
	jwt.RegisteredClaims
}

func init() {
	// Load the RSA keys
	rsaPrivateFp, err := os.Open(consts.Config.Auth.RSAPrivateKeyPath)
	if err != nil {
		log.Panicln("Error: ", err)
	}
	rsaPublicFp, err := os.Open(consts.Config.Auth.RSAPublicKeyPath)
	if err != nil {
		log.Panicln("Error: ", err)
	}
	privateKeyBytes := bytes.NewBuffer(nil)
	publicBytes := bytes.NewBuffer(nil)
	if _, err := rsaPublicFp.WriteTo(publicBytes); err != nil {
		log.Panicln("Error writing public key: ", err)
	}
	if _, err := rsaPrivateFp.WriteTo(privateKeyBytes); err != nil {
		log.Panicln("Error writing private key: ", err)
	}
	RSAPrivateKey, err = jwt.ParseECPrivateKeyFromPEM(privateKeyBytes.Bytes())
	if err != nil {
		log.Panicln("Error parsing private key: ", err)
	}
	RSAPublicKey, err = jwt.ParseECPublicKeyFromPEM(publicBytes.Bytes())
	if err != nil {
		log.Panicln("Error parsing public key: ", err)
	}
}

func NewAuthenticationService() *AuthenticationService {
	return &AuthenticationService{
		Config: &consts.Config.Auth,
	}
}
func (a *AuthenticationService) VerifyToken(cacheDB *gorm.DB, tokenMap *SessionClaims) (*cache.Session, error) {
	// Check if the token is in the cache
	sessionIDString := tokenMap.ID
	if sessionIDString == "" {
		return nil, fmt.Errorf("no session ID found")
	}
	session := &cache.Session{
		ID:     models.IDType(sessionIDString),
		UserID: models.IDType(tokenMap.Subject),
	}
	result := cacheDB.First(session)
	if result.Error != nil {
		log.Println("Error: ", result.Error)
		return nil, result.Error
	}
	return session, nil
}
func (a *AuthenticationService) NewSession(tx *gorm.DB, tokenMap *SessionClaims) (string, *cache.Session, error) {
	session := &cache.Session{
		ID:         idgenerator.GenerateID("ses"),
		UserID:     models.IDType(tokenMap.Subject),
		Username:   tokenMap.Name,
		Permission: tokenMap.Permission,
		ExpiresAt:  tokenMap.ExpiresAt.Time,
	}
	result := tx.Create(session)
	if result.Error != nil {
		return "", nil, result.Error
	}
	tokenMap.ID = string(models.IDType(session.ID))
	token := jwt.NewWithClaims(jwt.SigningMethodES256, tokenMap)
	accessToken, err := token.SignedString(RSAPrivateKey)
	if err != nil {
		log.Println("Error: ", err)
		return "", nil, err
	}
	return accessToken, session, nil
}
func (a *AuthenticationService) NewRefreshToken(tokenMap *SessionClaims) (string, error) {
	refreshClaims := &SessionClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Audience:  jwt.ClaimStrings{"campwiz"},
			Subject:   tokenMap.Subject,
			Issuer:    a.Config.Issuer,
			ExpiresAt: jwt.NewNumericDate(tokenMap.ExpiresAt.Add(time.Second * time.Duration(a.Config.Refresh))),
		},
		Permission: tokenMap.Permission,
		Name:       tokenMap.Name,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodES256, refreshClaims)
	refreshToken, err := token.SignedString(RSAPrivateKey)
	if err != nil {
		log.Println("Error: ", err)
		return "", err
	}
	return refreshToken, nil
}
func (a *AuthenticationService) RefreshSession(cacheDB *gorm.DB, tokenMap *SessionClaims) (accessToken string, session *cache.Session, err error) {
	log.Println("Refreshing session")
	sessionIDString := tokenMap.ID
	if sessionIDString == "" {
		return "", nil, fmt.Errorf("no session ID found")
	}
	session = &cache.Session{
		ID:         models.IDType(sessionIDString),
		UserID:     models.IDType(tokenMap.Subject),
		Username:   tokenMap.Name,
		Permission: tokenMap.Permission,
		ExpiresAt:  tokenMap.ExpiresAt.Time,
	}
	tx := cacheDB.Begin()
	result := tx.First(session, &cache.Session{ID: models.IDType(sessionIDString)})
	if result.Error != nil {
		log.Println("Error: ", result.Error)
		tx.Rollback()
		return "", nil, result.Error
	}
	session.ExpiresAt = time.Now().UTC().Add(time.Second * time.Duration(a.Config.Expiry))
	log.Println("Session expires at: ", session.ExpiresAt)
	result = tx.Save(session)
	if result.Error != nil {
		log.Println("Error: ", result.Error)
		tx.Rollback()
		return "", nil, result.Error
	}

	accessToken, session, err = a.NewSession(tx, tokenMap)
	if err != nil {
		tx.Rollback()
		return "", nil, err
	}
	tx.Commit()
	return accessToken, session, nil
}
func (a *AuthenticationService) RemoveSession(cacheDB *gorm.DB, ID models.IDType) error {
	session := &cache.Session{ID: ID}
	result := cacheDB.Delete(session)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
func (a *AuthenticationService) Logout(session *cache.Session) error {
	conn, close := cache.GetCacheDB()
	defer close()
	// Remove the session
	return a.RemoveSession(conn, session.ID)
}
func (a *AuthenticationService) decodeToken(tokenString string) (*SessionClaims, error) {
	claims := &SessionClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		claims := token.Claims
		iss, ok := claims.GetIssuer()
		if ok != nil {
			return nil, errors.New("issuer not found")
		}
		if iss != a.Config.Issuer {
			return nil, errors.New("invalid issuer")
		}

		return RSAPublicKey, nil
	})
	if err != nil {
		return claims, err
	}
	if !token.Valid {
		return claims, errors.New("invalid-token")
	}
	return claims, nil
}
func (auth_service *AuthenticationService) Authenticate(ctx *gin.Context, token string) (string, *cache.Session, error, bool) {
	tokenMap, err := auth_service.decodeToken(token)
	cache_db, close := cache.GetCacheDB()
	hub := sentrygin.GetHubFromContext(ctx)
	if hub != nil {
		hub.Scope().SetExtra("AuthTokenSessionID", tokenMap.ID)
		hub.Scope().SetExtra("AuthTokenUserID", tokenMap.Subject)
		hub.Scope().SetExtra("AuthTokenPermission", tokenMap.Permission)
		hub.Scope().SetExtra("AuthTokenName", tokenMap.Name)
		hub.Scope().SetExtra("AuthTokenExpiresAt", tokenMap.ExpiresAt)
	}
	defer close()
	if err != nil {
		if hub != nil {
			hub.Scope().SetExtra("AuthorizationTokenCache", "MISS")
			hub.Scope().SetExtra("AuthorizationTokenMissError", err.Error())
		}
		if strings.Contains(err.Error(), "token is expired") {
			// Token is expired
			newAccessToken, session, err := auth_service.RefreshSession(cache_db, tokenMap)
			if err != nil {
				if hub != nil {
					hub.Scope().SetExtra("AuthorizationTokenRefreshError", err.Error())
				}
				return "", nil, errors.New("token expired and could not be refreshed"), false
			} else {
				if hub != nil {
					hub.Scope().SetExtra("AuthorizationTokenRefreshNewSessionID", session.ID)
				}
				return newAccessToken, session, nil, true
			}
		} else {
			return "", nil, errors.New("token could not be decoded"), false
		}
	} else {
		session, err := auth_service.VerifyToken(cache_db, tokenMap)
		if err != nil {
			if hub != nil {
				hub.Scope().SetExtra("AuthorizationTokenVerifyError", err.Error())
			}
			return "", nil, errors.New("invalid token"), false
		}
		if hub != nil {
			hub.Scope().SetExtra("AuthorizationTokenCache", "HIT")
			hub.Scope().SetExtra("AuthorizationTokenHitSessionID", session.ID)
		}
		return token, session, nil, false
	}
}
