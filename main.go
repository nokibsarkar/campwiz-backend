package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"nokib/campwiz/consts"
	"nokib/campwiz/models"
	"nokib/campwiz/repository"
	"nokib/campwiz/repository/cache"
	"nokib/campwiz/routes"
	"time"

	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "nokib/campwiz/docs"

	"github.com/gin-gonic/gin"
)

// preRun is a function that will be called before the main function
func preRun() {
	fmt.Printf("Release: %s\n", consts.Release)
	fmt.Printf("Build Time: %s\n", consts.BuildTime)
	fmt.Printf("Version: %s\n", consts.Version)
	gin.SetMode(consts.Config.Server.Mode)
	log.Println(models.Date2Int(time.Now()))
}
func postRun() {
}
func beforeSetupRouter(ctx context.Context, testing bool) {
	repository.InitDB(ctx, testing)
	cache.InitCacheDB(ctx, testing)

}
func afterSetupRouter(ctx context.Context, testing bool) {
}
func SetupRouter(ctx context.Context, testing bool, readOnly bool) *gin.Engine {
	beforeSetupRouter(ctx, testing)
	r := gin.Default()
	r.Use(routes.NewSentryMiddleWare())
	if consts.Config.Sentry.DSN != "" {
		log.Printf("Sentry DSN is set, enabling Sentry middleware")

	}
	Mode := consts.Config.Server.Mode
	switch Mode {
	case "debug":
		r.Use(gin.Logger())
	case "release":
		r.Use(gin.Recovery())
	case "test":
		r.Use(gin.Recovery())
	default:
		log.Panicf("Invalid mode %s", Mode)
	}
	r.GET("/api/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	if readOnly {
		log.Println("Creating Routes for ReadOnly Mode")
		routes.NewReadOnlyRoutes(r.Group("/"))
	} else {
		log.Println("Creating Routes for ReadWrite Mode")
		routes.NewRoutes(r.Group("/"))
	}
	afterSetupRouter(ctx, testing)
	return r
}
func init() {
	consts.LoadConfig()
}

// @title Campwiz API
// @version 1
// @description This is the API documentation for Campwiz
// @BasePath /api/v2
// @produce json
// @consumes json
// @securitydefinitions.oauth2 implicit
// @type oauth2
// @authorizationurl https://meta.wikimedia.org/w/rest.php/oauth2/authorize
// @tokenurl https://meta.wikimedia.org/w/rest.php/oauth2/access_token
// @securityDefinitions.apikey ApiKeyAuth
// @in cookie
// @name c-auth
// @description Authentication cookie for the API. It would be set by the server when the user logs in.
// @security cookieAuth
// @license.name GPL-3.0
// @license.url https://www.gnu.org/licenses/gpl-3.0.html
// @contact.name Nokib Sarkar
// @contact.email nokibsarkar@gmail.com
// @contact.url https://github.com/nokibsarkar
// @query.collection.format multi
func main() {
	ctx := context.Background()
	preRun()
	portVal := 8081 // default port fallback
	if _, err := fmt.Sscanf(consts.Config.Server.Port, "%d", &portVal); err != nil {
		log.Printf("Failed to parse port from config: %s", err.Error())
	}
	port := flag.Int("port", portVal, "Port to run the server on")
	readOnly := flag.Bool("readonly", false, "Run the server in read-only mode")
	flag.Parse()
	r := SetupRouter(ctx, false, *readOnly)
	fmt.Printf("Port: %d\n", *port)
	gin.SetMode(consts.Release)
	if err := r.Run(fmt.Sprintf("0.0.0.0:%d", *port)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	postRun()

}
