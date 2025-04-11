//go:build !readonly

package routes

import (
	"nokib/campwiz/consts"
	idgenerator "nokib/campwiz/services/idGenerator"

	"github.com/gin-gonic/gin"
)

var serverInstanceId = idgenerator.GenerateID("Server-")

func NewRoutes(nonAPIParent *gin.RouterGroup) *gin.RouterGroup {
	r := nonAPIParent.Group("/api/v2")
	r.Use(ServerInfoHeaderMiddleware)
	authenticatorService := NewAuthenticationService()
	NewUserAuthenticationRoutes(r)
	r.Use(authenticatorService.Authenticate2)
	NewPermissionRoutes(nonAPIParent)
	NewCampaignRoutes(r)
	NewRoundRoutes(r)
	NewSubmissionRoutes(r)
	NewUserRoutes(r)
	NewTaskRoutes(r)
	NewEvaluationRoutes(r)
	NewProjectRoutes(r)
	return r
}
func NewRoundRoutes(parent *gin.RouterGroup) {
	r := parent.Group("/round")
	r.GET("/", WithSession(ListAllRounds))
	r.GET("/:roundId", GetRound)
	r.DELETE("/:roundId", WithSession(DeleteRound))
	r.POST("/:roundId/addMyselfAsPublicJury", WithSession(addMySelfAsJury))
	r.GET("/:roundId/next/public", WithSession(NextPublicSubmission))
	r.GET("/:roundId/results/summary", WithSession(GetResultSummary))
	r.GET("/:roundId/results/:format", WithSession(GetResults))
	r.POST("/:roundId/status", WithSession(UpdateStatus))
	r.POST("/", WithSession(CreateRound))
	r.POST("/:roundId", WithSession(UpdateRoundDetails))
	r.POST("/import/:roundId/commons", WithSession(ImportFromCommons))
	r.POST("/import/:roundId/previous", WithSession(ImportFromPreviousRound))
	r.POST("/distribute/:roundId", WithSession(DistributeEvaluations))

}
func NewSubmissionRoutes(parent *gin.RouterGroup) {
	r := parent.Group("/submission")
	r.GET("/", ListAllSubmissions)
	r.POST("/draft", CreateDraftSubmission)
	r.POST("/draft/late", CreateLateDraftSubmission)
	r.GET("/draft/:id", GetDraftSubmission)

	r.DELETE("/:id", DeleteSubmission)

	r.GET("/:submissionId", GetSubmission)
	// r.GET("/:submissionId/judge", GetEvaluation)
	r.POST("/", CreateSubmission)
	r.POST("/late", CreateLateSubmission)
	r.POST("/:submissionId/judge", EvaluateSubmission)
}
func NewUserAuthenticationRoutes(parent *gin.RouterGroup) {
	user := parent.Group("/")
	user.GET("/user/login", RedirectForLogin)
	user.GET("/user/callback", HandleOAuth2Callback)
}
func NewUserRoutes(parent *gin.RouterGroup) {
	r := parent.Group("/user")
	r.GET("/", ListUsers)
	r.GET("/me", WithSession(GetMe))
	r.GET("/:id", GetUser)
	r.POST("/:id", UpdateUser)
	r.GET("/translation/:language", GetTranslation)
	r.POST("/translation/:lang", UpdateTranslation)
	r.POST("/logout", Logout)
}

/*
NewCampaignRoutes will create all the routes for the /campaign endpoint
*/
func NewCampaignRoutes(parent *gin.RouterGroup) {
	defer HandleError("/campaign")
	r := parent.Group("/campaign")
	r.GET("/", WithSessionOptional(ListAllCampaigns))
	r.GET("/timeline2", GetAllCampaignTimeLine)
	r.GET("/:campaignId/result", GetCampaignResultSummary)
	r.GET("/jury", ListAllJury)
	r.POST("/", WithSession(CreateCampaign))
	r.POST("/:campaignId", WithSession(UpdateCampaign))
	r.GET("/:campaignId/submissions", GetCampaignSubmissions)
	r.GET("/:campaignId/next", GetNextSubmission)
	r.POST("/:campaignId/status", WithSession(UpdateCampaignStatus))
	r.POST("/:campaignId/fountain", ImportEvaluationFromFountain)
	r.GET("/:campaignId", WithSession(GetSingleCampaign))
}
func NewEvaluationRoutes(r *gin.RouterGroup) {
	route := r.Group("/evaluation")
	route.GET("/", WithSession(ListEvaluations))
	route.POST("/", WithSession(BulkEvaluate))
	route.GET("/:evaluationId", WithSession(GetEvaluation))
	route.POST("/:evaluationId", WithSession(UpdateEvaluation))
	route.POST("/public/:roundId/:submissionId", WithSession(SubmitNewPublicEvaluation))
	route.POST("/public/:roundId", WithSession(SubmitNewBulkPublicEvaluation))
}

func NewProjectRoutes(parent *gin.RouterGroup) *gin.RouterGroup {
	r := parent.Group("/project")
	r.GET("/", WithSession(ListProjects))
	// Only super admin can create a project
	r.POST("/", WithPermission(consts.PermissionCreateCampaign, CreateProject))
	// Only super admin can update a project
	r.POST("/:projectId", WithPermission(consts.PermissionUpdateProject, UpdateProject))
	r.GET("/:projectId", WithSession(GetSingleProject))

	return r
}
