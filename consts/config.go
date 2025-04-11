package consts

import (
	"flag"
	"log"
	"strconv"

	"github.com/spf13/viper"
)

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
	Port    string `mapstructure:"Port"`
	Host    string `mapstructure:"Host"`
	BaseURL string `mapstructure:"BaseURL"`
	Mode    string `mapstructure:"Mode"`
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
	Secret            string              `mapstructure:"Secret"`
	Expiry            int                 `mapstructure:"Expiry"`
	Refresh           int                 `mapstructure:"Refresh"`
	Issuer            string              `mapstructure:"Issuer"`
	OAuth2            OAuth2Configuration `mapstructure:"OAuth2"`
	AccessToken       string              `mapstructure:"AccessToken"`
	RSAPrivateKeyPath string              `mapstructure:"RSAPrivateKeyPath"`
	RSAPublicKeyPath  string              `mapstructure:"RSAPublicKeyPath"`
}

type ApplicationConfiguration struct {
	Server       ServerConfiguration         `mapstructure:"Server"`
	Database     DatabaseConfiguration       `mapstructure:"Database"`
	Auth         AuthenticationConfiguration `mapstructure:"Authentication"`
	Distribution DistributionConfiguration   `mapstructure:"DistributionStrategy"`
	Sentry       SentryConfig                `mapstructure:"Sentry"`
}

var Config *ApplicationConfiguration

func init() {
	if Config != nil {
		return
	}
	Config = &ApplicationConfiguration{}
	viper.SetConfigName(".env")
	viper.AddConfigPath(".")
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		log.Printf("Error reading config file: %s", err)
		return
	}

	err = viper.Unmarshal(Config)
	if err != nil {
		log.Printf("Error unmarshalling config file: %s", err)
	}
	flagPort := flag.Int("port", 8081, "Port to run the server on")
	flag.Parse()
	if flagPort != nil {
		Config.Server.Port = strconv.Itoa(*flagPort)
		log.Printf("Using port from commandline: %s", Config.Server.Port)
	}

}
