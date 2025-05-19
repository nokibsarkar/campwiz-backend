package routes

import (
	"log"
	"nokib/campwiz/models"

	"github.com/gin-gonic/gin"
)

func ReadOnlyMode(c *gin.Context) {
	c.JSON(500, models.ResponseError{
		Detail: "Internal Server Error: Sorry, the server is in read-only mode. Please try again later.",
	})
}

func NewReadOnlyRoutes(nonAPIParent *gin.RouterGroup) *gin.RouterGroup {
	log.Println("Creating Routes for ReadOnly Mode")
	r := nonAPIParent.Group("/api/v2")
	r.Use(ServerInfoHeaderMiddleware)
	authenticatorService := NewAuthenticationService()
	// NewUserAuthenticationRoutes(r)
	r.Use(authenticatorService.Authenticate2)
	NewPermissionRoutes(nonAPIParent)
	NewReadOnlyCampaignRoutes(r)
	NewReadOnlyRoundRoutes(r)
	NewReadOnlySubmissionRoutes(r)
	NewReadOnlyUserRoutes(r)
	NewTaskRoutes(r)
	NewReadOnlyEvaluationRoutes(r)
	NewReadOnlyProjectRoutes(r)
	return r
}

func NewReadOnlyRoundRoutes(parent *gin.RouterGroup) {
	r := parent.Group("/round")
	r.GET("/", WithSession(ListAllRounds))
	r.GET("/:roundId", GetRound)
	r.DELETE("/:roundId", ReadOnlyMode)
	r.POST("/:roundId/addMyselfAsPublicJury", ReadOnlyMode)
	r.GET("/:roundId/next/public", WithSession(NextPublicSubmission))
	r.GET("/:roundId/results/summary", WithSession(GetResultSummary))
	r.GET("/:roundId/results/:format", WithSession(GetResults))
	r.POST("/:roundId/status", ReadOnlyMode)
	r.POST("/", ReadOnlyMode)
	r.POST("/:roundId", ReadOnlyMode)
	r.POST("/import/:roundId/commons", ReadOnlyMode)
	r.POST("/import/:roundId/previous", ReadOnlyMode)
	r.POST("/distribute/:roundId", ReadOnlyMode)
}

func NewReadOnlySubmissionRoutes(parent *gin.RouterGroup) {
	r := parent.Group("/submission")
	r.GET("/", ListAllSubmissions)
	r.POST("/draft", ReadOnlyMode)
	r.POST("/draft/late", ReadOnlyMode)
	r.GET("/draft/:id", GetDraftSubmission)

	r.DELETE("/:id", ReadOnlyMode)

	r.GET("/:submissionId", GetSubmission)
	// r.GET("/:submissionId/judge", GetEvaluation)
	r.POST("/", ReadOnlyMode)
	r.POST("/late", ReadOnlyMode)
	r.POST("/:submissionId/judge", ReadOnlyMode)
}

func NewReadOnlyUserAuthenticationRoutes(parent *gin.RouterGroup) {
	user := parent.Group("/")
	user.GET("/user/login", RedirectForLogin)
	user.GET("/user/callback", HandleOAuth2Callback)
}

func NewReadOnlyUserRoutes(parent *gin.RouterGroup) {
	r := parent.Group("/user")
	r.GET("/", ListUsers)
	r.GET("/me", WithSession(GetMe))
	r.GET("/:id", GetUser)
	r.POST("/:id", ReadOnlyMode)
	r.GET("/translation/:language", GetTranslation)
	r.POST("/translation/:lang", ReadOnlyMode)
	r.POST("/logout", ReadOnlyMode)
}

/*
NewCampaignRoutes will create all the routes for the /campaign endpoint
*/
func NewReadOnlyCampaignRoutes(parent *gin.RouterGroup) {
	defer HandleError("/campaign")
	r := parent.Group("/campaign")
	r.GET("/", WithSessionOptional(ListAllCampaigns))
	r.GET("/timeline2", GetAllCampaignTimeLine)
	r.GET("/:campaignId/result", GetCampaignResultSummary)
	r.GET("/jury", ListAllJury)
	r.POST("/", ReadOnlyMode)
	r.POST("/:campaignId", ReadOnlyMode)
	r.GET("/:campaignId/submissions", GetCampaignSubmissions)
	r.GET("/:campaignId/next", GetNextSubmission)
	r.POST("/:campaignId/status", ReadOnlyMode)
	r.POST("/:campaignId/fountain", ReadOnlyMode)
	r.GET("/:campaignId", WithSession(GetSingleCampaign))
}

func NewReadOnlyEvaluationRoutes(r *gin.RouterGroup) {
	route := r.Group("/evaluation")
	route.GET("/", WithSession(ListEvaluations))
	route.POST("/", ReadOnlyMode)
	route.POST("/:evaluationId", ReadOnlyMode)
	route.POST("/public/:roundId/:submissionId", ReadOnlyMode)
	route.POST("/public/:roundId", ReadOnlyMode)
}

func NewReadOnlyProjectRoutes(parent *gin.RouterGroup) *gin.RouterGroup {
	r := parent.Group("/project")
	r.GET("/", WithSession(ListProjects))
	// Only super admin can create a project
	r.POST("/", ReadOnlyMode)
	// Only super admin can update a project
	r.POST("/:projectId", ReadOnlyMode)
	r.GET("/:projectId", WithSession(GetSingleProject))

	return r
}
