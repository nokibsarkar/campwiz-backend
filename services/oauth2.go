package services

import (
	"context"
	"encoding/json"
	"nokib/campwiz/consts"
	"nokib/campwiz/models"
	"time"

	"golang.org/x/oauth2"
)

const META_OAUTH_AUTHORIZE_URL = "https://meta.wikimedia.org/w/rest.php/oauth2/authorize"
const META_OAUTH_ACCESS_TOKEN_URL = "https://meta.wikimedia.org/w/rest.php/oauth2/access_token"
const META_PROFILE_URL = "https://meta.wikimedia.org/w/rest.php/oauth2/resource/profile"
const DATETIMEFORMAT = "20060102150405"

var OAuth2IdentityConfig = &oauth2.Config{
	ClientID:     consts.Config.Auth.OAuth2IdentityVerification.ClientID,
	ClientSecret: consts.Config.Auth.OAuth2IdentityVerification.ClientSecret,
	RedirectURL:  consts.Config.Server.BaseURL + consts.Config.Auth.OAuth2IdentityVerification.RedirectPath,
	Endpoint: oauth2.Endpoint{
		AuthURL:  META_OAUTH_AUTHORIZE_URL,
		TokenURL: META_OAUTH_ACCESS_TOKEN_URL,
	},
}

type OAuth2Service struct {
	Config *oauth2.Config
	ctx    context.Context
}

func NewOAuth2Service(ctx context.Context, config *oauth2.Config) *OAuth2Service {
	return &OAuth2Service{
		Config: config,
		ctx:    ctx,
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
