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
	environment := "development"
	if !isDebug {
		environment = "production"
	}
	// To initialize Sentry's handler, you need to initialize Sentry itself beforehand
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:         consts.Config.Sentry.DSN,
		Debug:       isDebug,
		Environment: environment,
		Tags: map[string]string{
			"base-url":    consts.Config.Server.BaseURL,
			"Build-Time":  consts.BuildTime,
			"Version":     consts.Version,
			"Commit-Hash": consts.CommitHash,
		},
		EnableTracing: true,
		// Or provide a custom sample rate:
		TracesSampler: sentry.TracesSampler(func(ctx sentry.SamplingContext) float64 {
			// As an example, this does not send some
			// transactions to Sentry based on their name.
			if ctx.Span.Name == "GET /health" {
				return 0.0
			}
			return 0.8
		}),
		Release:          consts.Release,
		AttachStacktrace: true,
		SampleRate:       0.8,
		SendDefaultPII:   false,
		EnableLogs:       gin.IsDebugging(),
	}); err != nil {
		fmt.Printf("Sentry initialization failed: %v\n", err)
	}
	return sentrygin.New(sentrygin.Options{})
}
