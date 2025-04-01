package main

import (
	"log"
	"nokib/campwiz/consts"
	"nokib/campwiz/repository"
	"nokib/campwiz/repository/cache"
	"nokib/campwiz/routes"

	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "nokib/campwiz/docs"

	"github.com/gin-gonic/gin"
)

// preRun is a function that will be called before the main function
func preRun() {
	gin.SetMode(consts.Config.Server.Mode)

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
	if Mode == "debug" {
		r.Use(gin.Logger())
	} else if Mode == "release" {
		r.Use(gin.Recovery())

	} else {
		log.Panicf("Invalid mode %s", Mode)
	}
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	if consts.Config.Sentry.DSN != "" {

		r.Use(routes.NewSentryMiddleWare())
	}
	routes.NewRoutes(r.Group("/"))
	afterSetupRouter(testing)
	return r
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
	r := SetupRouter(false)
	r.Run("localhost:" + consts.Config.Server.Port)
	postRun()

}
