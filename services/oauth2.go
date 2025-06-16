package services

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"nokib/campwiz/consts"
	"nokib/campwiz/models"
	"nokib/campwiz/repository"
	idgenerator "nokib/campwiz/services/idGenerator"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

const META_PROFILE_URL = "https://meta.wikimedia.org/w/rest.php/oauth2/resource/profile"
const DATETIMEFORMAT = "20060102150405"

type OAuth2Service struct {
	Config       *oauth2.Config
	ctx          *gin.Context
	redirectPath string
}

func NewOAuth2Service(ctx *gin.Context, config *oauth2.Config, redirectPath string) *OAuth2Service {
	return &OAuth2Service{
		Config:       config,
		ctx:          ctx,
		redirectPath: redirectPath,
	}
}
func (o *OAuth2Service) Init(callback string) string {
	// state := url.QueryEscape(callback)
	return o.Config.AuthCodeURL(callback)
}
func (o *OAuth2Service) GetToken(code string, redirectURL string) (*oauth2.Token, error) {
	previousRedirectURL := o.Config.RedirectURL
	defer func() {
		o.Config.RedirectURL = previousRedirectURL
	}()
	o.Config.RedirectURL = redirectURL
	token, err := o.Config.Exchange(o.ctx, code)
	if err != nil {
		return nil, err
	}
	return token, nil
}

/*
sub (central user id)
username
editcount
confirmed_email
blocked
registered
groups
rights
*/
type WikipediaProfileBasic struct {
	CentralID string                       `json:"sub"`
	Name      models.WikimediaUsernameType `json:"username"`
	Rights    []string                     `json:"rights"`
	Blocked   bool                         `json:"blocked"`
	Groups    []string                     `json:"groups"`
}
type WikipediaProfile struct {
	WikipediaProfileBasic
	Registered string `json:"registered"`
}
type WikipediaProfileFull struct {
	WikipediaProfileBasic
	Registered time.Time `json:"registered"`
}

func (o *OAuth2Service) GetUser(token *oauth2.Token) (*WikipediaProfileFull, error) {
	client := o.Config.Client(context.Background(), token)
	resp, err := client.Get(META_PROFILE_URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() //nolint:errcheck
	user := &WikipediaProfile{}
	err = json.NewDecoder(resp.Body).Decode(user)
	if err != nil {
		return nil, err
	}
	registered, err := time.Parse(DATETIMEFORMAT, user.Registered)
	if err != nil {
		return nil, err
	}

	return &WikipediaProfileFull{
		WikipediaProfileBasic: user.WikipediaProfileBasic,
		Registered:            registered,
	}, nil
}
func (s *OAuth2Service) FetchTokenFromWikimediaServer() (db_user *models.User, state string, accessToken *oauth2.Token, err error) {
	query := s.ctx.Request.URL.Query()
	code := query.Get("code")
	if code == "" {
		err = errors.New("noCodeOnQuery")
		return
	}
	state = query.Get("state")
	if state == "" || strings.HasPrefix(state, "/user/login") {
		state = "/"
	}
	baseURL := consts.Config.Server.BaseURL
	baseURLRaw, ok := s.ctx.GetQuery("baseURL")
	if ok {
		baseURL = baseURLRaw
	}
	accessToken, err = s.GetToken(code, baseURL+s.redirectPath)
	if err != nil {
		return
	}
	user, err := s.GetUser(accessToken)
	if err != nil {
		return
	}
	conn, close, err := repository.GetDB(s.ctx)
	if err != nil {
		return
	}
	defer close()
	user_service := NewUserService()
	db_user, err = user_service.GetUserByUsername(conn, user.Name)
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
				err = trx.Error
				return
			}
			log.Println("User created: ", trx.RowsAffected)

		} else {
			return
		}
	}
	return
}
