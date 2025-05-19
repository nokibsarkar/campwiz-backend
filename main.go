package main

import (
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
func beforeSetupRouter(testing bool) {
	repository.InitDB(testing)
	cache.InitCacheDB(testing)

}
func afterSetupRouter(testing bool) {
}
func SetupRouter(testing bool) *gin.Engine {
	beforeSetupRouter(testing)
	r := gin.Default()
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
	// if consts.Config.Sentry.DSN != "" {
	// 	r.Use(routes.NewSentryMiddleWare())
	// }
	routes.NewRoutes(r.Group("/"))
	afterSetupRouter(testing)
	return r
}
func init() {
	consts.LoadConfig()
}

// @title Campwiz API
// @version 1
// @description This is the API documentation for Campwiz
// @host localhost:8081
// @BasePath /api/v2
// @schemes http https
// @produce json
// @consumes json
// @securitydefinitions.oauth2 implicit
// @type oauth2
// @authorizationurl https://meta.wikimedia.org/w/rest.php/oauth2/authorize
// @tokenurl https://meta.wikimedia.org/w/rest.php/oauth2/access_token

// @license.name GPL-3.0
// @license.url https://www.gnu.org/licenses/gpl-3.0.html
// @contact.name Nokib Sarkar
// @contact.email nokibsarkar@gmail.com
// @contact.url https://github.com/nokibsarkar
// @query.collection.format multi
func main() {
	preRun()
	portVal := 8081 // default port fallback
	fmt.Sscanf(consts.Config.Server.Port, "%d", &portVal)
	port := flag.Int("port", portVal, "Port to run the server on")
	flag.Parse()
	r := SetupRouter(false)
	if err := r.Run(fmt.Sprintf("0.0.0.0:%d", *port)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	postRun()

}
