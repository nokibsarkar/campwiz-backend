package routes

import (
	"nokib/campwiz/consts"
	"nokib/campwiz/models"
	"nokib/campwiz/repository/cache"
	"nokib/campwiz/services"

	"github.com/gin-gonic/gin"
)

// CreateProject godoc
// @Summary Create a new project
// @Description Create a new project
// @Produce  json
// @Success 200 {object} ResponseSingle[models.Project]
// @Router /project/ [post]
// @Param projectRequest body models.ProjectRequest true "The project request"
// @Param includeProjectLeads query bool false "Include project leads"
// @Tags Project
// @Security ApiKeyAuth
// @Error 400 {object} ResponseError
// @Error 403 {object} ResponseError
// @Error 404 {object} ResponseError
func CreateProject(c *gin.Context, sess *cache.Session) {
	projectRequest := &models.ProjectRequest{}
	err := c.ShouldBindJSON(projectRequest)
	if err != nil {
		c.JSON(400, ResponseError{Detail: "Invalid request : " + err.Error()})
		return
	}
	q := &ProjectSingleQuery{}
	err = c.ShouldBindQuery(q)
	if err != nil {
		c.JSON(400, ResponseError{Detail: "Invalid request : " + err.Error()})
		return
	}
	projectRequest.CreatedByID = sess.UserID
	project_service := services.NewProjectService()
	project, err := project_service.CreateProject(projectRequest, q.IncludeProjectLeads)
	if err != nil {
		c.JSON(400, ResponseError{Detail: "Error creating project : " + err.Error()})
		return
	}
	c.JSON(200, ResponseSingle[models.ProjectExtended]{Data: *project})
}

// UpdateProject godoc
// @Summary Update a project
// @Description Update a project
// @Produce  json
// @Success 200 {object} ResponseSingle[models.Project]
// @Router /project/{projectId} [post]
// @Param projectId path string true "The project ID"
// @Param projectRequest body models.ProjectRequest true "The project request"
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
	projectRequest := &models.ProjectRequest{}
	err := c.ShouldBindJSON(projectRequest)
	if err != nil {
		c.JSON(400, ResponseError{Detail: "Invalid request: " + err.Error()})
		return
	}
	projectRequest.ProjectID = models.IDType(projectId)
	projectRequest.CreatedByID = sess.UserID
	project_service := services.NewProjectService()
	project, err := project_service.UpdateProject(projectRequest)
	if err != nil {
		c.JSON(400, ResponseError{Detail: "Error updating project : " + err.Error()})
		return
	}
	c.JSON(200, ResponseSingle[models.ProjectExtended]{Data: *project})
}

// ProjectSingleQuery is a query struct for getting a single project
type ProjectSingleQuery struct {
	IncludeProjectLeads bool `form:"includeProjectLeads"`
}

// GetSingleProject godoc
// @Summary Get a single project
// @Description Get a single project
// @Produce  json
// @Success 200 {object} ResponseSingle[models.Project]
// @Router /project/{projectId} [get]
// @Param projectId path string true "The project ID"
// @Param includeProjectLeads query bool false "Include project leads"
// @Tags Project
// @Security ApiKeyAuth
// @Error 400 {object} ResponseError
// @Error 403 {object} ResponseError
// @Error 404 {object} ResponseError
func GetSingleProject(c *gin.Context) {
	projectId := c.Param("projectId")
	if projectId == "" {
		c.JSON(400, ResponseError{Detail: "Invalid request : Project ID is required"})
		return
	}
	q := &ProjectSingleQuery{}
	err := c.ShouldBindQuery(q)
	if err != nil {
		c.JSON(400, ResponseError{Detail: "Invalid request : " + err.Error()})
		return
	}
	project_service := services.NewProjectService()
	project, err := project_service.GetProjectByID(models.IDType(projectId), q.IncludeProjectLeads)
	if err != nil {
		c.JSON(400, ResponseError{Detail: "Error getting project : " + err.Error()})
		return
	}
	c.JSON(200, ResponseSingle[models.ProjectExtended]{Data: *project})
}
func ListProjects(c *gin.Context, sess *cache.Session) {
	project_service := services.NewProjectService()
	projects, err := project_service.ListProjects()
	if err != nil {
		c.JSON(400, ResponseError{Detail: "Error getting projects : " + err.Error()})
		return
	}
	pj := []models.ProjectExtended{}
	for _, p := range projects {
		px := models.ProjectExtended{Project: p}
		px.Leads = []models.WikimediaUsernameType{}
		pj = append(pj, px)
	}
	c.JSON(200, ResponseList[models.ProjectExtended]{Data: pj})
}

func NewProjectRoutes(parent *gin.RouterGroup) *gin.RouterGroup {
	r := parent.Group("/project")
	r.GET("/", WithPermission(consts.PermissionOtherProjectAccess, ListProjects))
	r.POST("/", WithPermission(consts.PermissionCreateCampaign, CreateProject))
	r.POST("/:projectId", WithPermission(consts.PermissionUpdateProject, UpdateProject))
	r.GET("/:projectId", GetSingleProject)

	return r
}
