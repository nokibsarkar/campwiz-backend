package routes

import (
	"fmt"
	"nokib/campwiz/consts"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
)

func NewSentryMiddleWare() gin.HandlerFunc {
	isDebug := consts.Config.Server.Mode == "debug"
	// To initialize Sentry's handler, you need to initialize Sentry itself beforehand
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:         consts.Config.Sentry.DSN,
		Debug:       isDebug,
		Environment: consts.Config.Server.Mode,
		Tags: map[string]string{
			"base-url": consts.Config.Server.BaseURL,
		},
	}); err != nil {
		fmt.Printf("Sentry initialization failed: %v\n", err)
	}
	return sentrygin.New(sentrygin.Options{})
}
