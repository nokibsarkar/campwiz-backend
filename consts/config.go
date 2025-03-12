package consts

import (
	"fmt"

	"github.com/spf13/viper"
)

type DistributionConfiguration struct {
	Strategy          string `mapstructure:"Algorithm"`
	MinimumBatchSize  int    `mapstructure:"MinimumBatchSize"`
	MaximumBatchCount int    `mapstructure:"MaximumBatchCount"`
}
type MainDatabaseConfiguration struct {
	DSN     string `mapstructure:"DSN"`
	TestDSN string `mapstructure:"TestDSN"`
}
type CacheDatabaseConfiguration struct {
	DSN     string `mapstructure:"DSN"`
	TestDSN string `mapstructure:"TestDSN"`
}
type DatabaseConfiguration struct {
	Main  MainDatabaseConfiguration  `mapstructure:"Main"`
	Cache CacheDatabaseConfiguration `mapstructure:"Cache"`
}
type ServerConfiguration struct {
	Port    string `mapstructure:"Port"`
	Host    string `mapstructure:"Host"`
	BaseURL string `mapstructure:"BaseURL"`
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
		panic(err)
	}

	err = viper.Unmarshal(Config)
	if err != nil {
		panic(err)
	}
	fmt.Println(Config.Auth.OAuth2.RedirectPath)

}
