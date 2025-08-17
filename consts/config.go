package consts

import (
	"log"
	"os"

	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

var Version string
var BuildTime string
var CommitHash string
var Release string

const MAX_CSV_FILE_SIZE = 10 * 1024 * 1024 // 10 MB
type SentryConfig struct {
	DSN         string            `mapstructure:"DSN"`
	Environment string            `mapstructure:"Environment"`
	Debug       bool              `mapstructure:"Debug"`
	Release     string            `mapstructure:"Release"`
	Tags        map[string]string `mapstructure:"Tags"`
}
type DistributionConfiguration struct {
	Strategy          string `mapstructure:"Algorithm"`
	MinimumBatchSize  int    `mapstructure:"MinimumBatchSize"`
	MaximumBatchCount int    `mapstructure:"MaximumBatchCount"`
}
type MainDatabaseConfiguration struct {
	DSN     string `mapstructure:"DSN"`
	TestDSN string `mapstructure:"TestDSN"`
	Debug   bool   `mapstructure:"Debug"`
}
type CacheDatabaseConfiguration struct {
	DSN     string `mapstructure:"DSN"`
	TestDSN string `mapstructure:"TestDSN"`
	Debug   bool   `mapstructure:"Debug"`
}
type TaskCacheDatabaseConfiguration struct {
	DSN   string `mapstructure:"DSN"`
	Debug bool   `mapstructure:"Debug"`
}
type CommonsReplicaDatabaseConfiguration struct {
	DSN string `mapstructure:"DSN"`
}
type DatabaseConfiguration struct {
	Main    MainDatabaseConfiguration           `mapstructure:"Main"`
	Cache   CacheDatabaseConfiguration          `mapstructure:"Cache"`
	Task    TaskCacheDatabaseConfiguration      `mapstructure:"Task"`
	Commons CommonsReplicaDatabaseConfiguration `mapstructure:"Commons"`
}
type ServerConfiguration struct {
	Port        string `mapstructure:"Port"`
	Host        string `mapstructure:"Host"`
	BaseURL     string `mapstructure:"BaseURL"`
	Mode        string `mapstructure:"Mode"`
	Environment string `mapstructure:"Environment"`
}
type OAuth2Configuration struct {
	ClientID     string `mapstructure:"ClientID"`
	ClientSecret string `mapstructure:"ClientSecret"`
	RedirectPath string `mapstructure:"RedirectURL"`
	AuthURL      string `mapstructure:"AuthURL"`
	TokenURL     string `mapstructure:"TokenURL"`
	APIURL       string `mapstructure:"APIURL"`
}
type AuthenticationConfiguration struct {
	Secret                     string               `mapstructure:"Secret"`
	Expiry                     int                  `mapstructure:"Expiry"`
	Refresh                    int                  `mapstructure:"Refresh"`
	Issuer                     string               `mapstructure:"Issuer"`
	OAuth2IdentityVerification OAuth2Configuration  `mapstructure:"OAuth2"`
	Oauth2WriteAccess          *OAuth2Configuration `mapstructure:"OAuth2WriteAccess"`
	AccessToken                string               `mapstructure:"AccessToken"`
	RSAPrivateKeyPath          string               `mapstructure:"RSAPrivateKeyPath"`
	RSAPublicKeyPath           string               `mapstructure:"RSAPublicKeyPath"`
}
type TaskManagerConfiguration struct {
	Host string `mapstructure:"Host"`
	Port string `mapstructure:"Port"`
}
type ApplicationConfiguration struct {
	Server       ServerConfiguration         `mapstructure:"Server"`
	Database     DatabaseConfiguration       `mapstructure:"Database"`
	Auth         AuthenticationConfiguration `mapstructure:"Authentication"`
	Distribution DistributionConfiguration   `mapstructure:"DistributionStrategy"`
	Sentry       SentryConfig                `mapstructure:"Sentry"`
	TaskManager  TaskManagerConfiguration    `mapstructure:"TaskManager"`
}

var Config *ApplicationConfiguration

func LoadConfig() {
	if Config != nil {
		return
	}
	Config = &ApplicationConfiguration{}

	// viper.AddConfigPath(".")
	viper.AddConfigPath(os.Getenv("TOOL_DATA_DIR"))
	viper.SetConfigName(".env")
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err == nil {
		err = viper.Unmarshal(Config)
		if err != nil {
			log.Printf("Error unmarshalling config file: %s", err)
		}
	} else {
		log.Printf("Error reading config file: %s", err)
	}

}

const META_OAUTH_AUTHORIZE_URL = "https://meta.wikimedia.org/w/rest.php/oauth2/authorize"
const META_OAUTH_ACCESS_TOKEN_URL = "https://meta.wikimedia.org/w/rest.php/oauth2/access_token"
const AuthenticationCookieName = "c-auth"
const ReadWriteAuthenticationCookieName = "c-auth-rw"   // This is used for read-write access
const ReadWriteRefreshCookieName = "X-Refresh-Token-RW" // This is used for read-write access to refresh token
const RefreshCookieName = "X-Refresh-Token"
const SESSION_KEY = "session"

func init() {
	// Load the config file
	LoadConfig()
	// Set the release version
	Release = Version
}
func (authConfig *AuthenticationConfiguration) GetOAuth2IdentityVerificationOauthConfig() *oauth2.Config {
	if authConfig.OAuth2IdentityVerification.ClientID == "" {
		return nil
	}
	return &oauth2.Config{
		ClientID:     authConfig.OAuth2IdentityVerification.ClientID,
		ClientSecret: authConfig.OAuth2IdentityVerification.ClientSecret,
		RedirectURL:  Config.Server.BaseURL + authConfig.OAuth2IdentityVerification.RedirectPath,
		Endpoint: oauth2.Endpoint{
			AuthURL:  META_OAUTH_AUTHORIZE_URL,
			TokenURL: META_OAUTH_ACCESS_TOKEN_URL,
		},
		// APIURL: authConfig.OAuth2IdentityVerification.APIURL,
	}
}
func (authConfig *AuthenticationConfiguration) GetOAuth2ReadWriteOauthConfig() *oauth2.Config {
	if authConfig.Oauth2WriteAccess == nil {
		return nil
	}
	return &oauth2.Config{
		ClientID:     authConfig.Oauth2WriteAccess.ClientID,
		ClientSecret: authConfig.Oauth2WriteAccess.ClientSecret,
		RedirectURL:  Config.Server.BaseURL + authConfig.Oauth2WriteAccess.RedirectPath,
		Endpoint: oauth2.Endpoint{
			AuthURL:  META_OAUTH_AUTHORIZE_URL,
			TokenURL: META_OAUTH_ACCESS_TOKEN_URL,
		},
	}
}
