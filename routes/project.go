package routes

import (
	"nokib/campwiz/consts"
	"nokib/campwiz/database"
	"nokib/campwiz/database/cache"
	"nokib/campwiz/services"

	"github.com/gin-gonic/gin"
)

// CreateProject godoc
// @Summary Create a new project
// @Description Create a new project
// @Produce  json
// @Success 200 {object} ResponseSingle[database.Project]
// @Router /project/ [post]
// @Param projectRequest body database.ProjectRequest true "The project request"
// @Tags Project
// @Security ApiKeyAuth
// @Error 400 {object} ResponseError
// @Error 403 {object} ResponseError
// @Error 404 {object} ResponseError
func CreateProject(c *gin.Context, sess *cache.Session) {
	projectRequest := &database.ProjectRequest{}
	err := c.ShouldBindJSON(projectRequest)
	if err != nil {
		c.JSON(400, ResponseError{Detail: "Invalid request : " + err.Error()})
		return
	}
	projectRequest.CreatedByID = sess.UserID
	project_service := services.NewProjectService()
	project, err := project_service.CreateProject(projectRequest)
	if err != nil {
		c.JSON(400, ResponseError{Detail: "Error creating project : " + err.Error()})
		return
	}
	c.JSON(200, ResponseSingle[database.Project]{Data: *project})
}

// UpdateProject godoc
// @Summary Update a project
// @Description Update a project
// @Produce  json
// @Success 200 {object} ResponseSingle[database.Project]
// @Router /project/{projectId} [post]
// @Param projectId path string true "The project ID"
// @Param projectRequest body database.ProjectRequest true "The project request"
// @Tags Project
// @Security ApiKeyAuth
// @Error 400 {object} ResponseError
// @Error 403 {object} ResponseError
// @Error 404 {object} ResponseError
func UpdateProject(c *gin.Context, sess *cache.Session) {
	projectId := c.Param("projectId")
	if projectId == "" {
		c.JSON(400, ResponseError{Detail: "Invalid request : Project ID is required"})
		return
	}
	projectRequest := &database.ProjectRequest{}
	err := c.ShouldBindJSON(projectRequest)
	if err != nil {
		c.JSON(400, ResponseError{Detail: "Invalid request : " + err.Error()})
		return
	}
	projectRequest.ProjectID = database.IDType(projectId)
	projectRequest.CreatedByID = sess.UserID
	project_service := services.NewProjectService()
	project, err := project_service.UpdateProject(projectRequest)
	if err != nil {
		c.JSON(400, ResponseError{Detail: "Error creating project : " + err.Error()})
		return
	}
	c.JSON(200, ResponseSingle[database.Project]{Data: *project})
}

func NewProjectRoutes(parent *gin.RouterGroup) *gin.RouterGroup {
	r := parent.Group("/project")
	r.POST("/", WithPermission(consts.PermissionCreateCampaign, CreateProject))
	r.POST("/:projectId", WithPermission(consts.PermissionCreateCampaign, UpdateProject))
	return r
}
