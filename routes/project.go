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
// @Success 200 {object} models.ResponseSingle[models.Project]
// @Router /project/ [post]
// @Param projectRequest body models.ProjectRequest true "The project request"
// @Param includeProjectLeads query bool false "Include project leads"
// @Tags Project
// @Security ApiKeyAuth
// @Error 400 {object} models.ResponseError
// @Error 403 {object} models.ResponseError
// @Error 404 {object} models.ResponseError
func CreateProject(c *gin.Context, sess *cache.Session) {
	projectRequest := &models.ProjectRequest{}
	err := c.ShouldBindJSON(projectRequest)
	if err != nil {
		c.JSON(400, models.ResponseError{Detail: "Invalid request : " + err.Error()})
		return
	}
	q := &ProjectSingleQuery{}
	err = c.ShouldBindQuery(q)
	if err != nil {
		c.JSON(400, models.ResponseError{Detail: "Invalid request : " + err.Error()})
		return
	}
	projectRequest.CreatedByID = sess.UserID
	project_service := services.NewProjectService()
	project, err := project_service.CreateProject(c, projectRequest, q.IncludeProjectLeads)
	if err != nil {
		c.JSON(400, models.ResponseError{Detail: "Error creating project : " + err.Error()})
		return
	}
	c.JSON(200, models.ResponseSingle[models.ProjectExtended]{Data: *project})
}

// UpdateProject godoc
// @Summary Update a project
// @Description Update a project
// @Produce  json
// @Success 200 {object} models.ResponseSingle[models.Project]
// @Router /project/{projectId} [post]
// @Param projectId path string true "The project ID"
// @Param projectRequest body models.ProjectRequest true "The project request"
// @Tags Project
// @Security ApiKeyAuth
// @Error 400 {object} models.ResponseError
// @Error 403 {object} models.ResponseError
// @Error 404 {object} models.ResponseError
func UpdateProject(c *gin.Context, sess *cache.Session) {
	projectId := c.Param("projectId")
	if projectId == "" {
		c.JSON(400, models.ResponseError{Detail: "Invalid request : Project ID is required"})
		return
	}
	projectRequest := &models.ProjectRequest{}
	err := c.ShouldBindBodyWithJSON(projectRequest)
	if err != nil {
		c.JSON(400, models.ResponseError{Detail: "Invalid request: " + err.Error()})
		return
	}
	projectRequest.ProjectID = models.IDType(projectId)
	// projectRequest.CreatedByID = sess.UserID
	project_service := services.NewProjectService()
	project, err := project_service.UpdateProject(c, projectRequest)
	if err != nil {
		c.JSON(400, models.ResponseError{Detail: "Error updating project : " + err.Error()})
		return
	}
	c.JSON(200, models.ResponseSingle[models.ProjectExtended]{Data: *project})
}

// ProjectSingleQuery is a query struct for getting a single project
type ProjectSingleQuery struct {
	IncludeProjectLeads bool `form:"includeProjectLeads"`
}

// GetSingleProject godoc
// @Summary Get a single project
// @Description Get a single project
// @Produce  json
// @Success 200 {object} models.ResponseSingle[models.Project]
// @Router /project/{projectId} [get]
// @Param projectId path string true "The project ID"
// @Param includeProjectLeads query bool false "Include project leads"
// @Tags Project
// @Security ApiKeyAuth
// @Error 400 {object} models.ResponseError
// @Error 403 {object} models.ResponseError
// @Error 404 {object} models.ResponseError
func GetSingleProject(c *gin.Context, sess *cache.Session) {
	projectId := c.Param("projectId")
	if projectId == "" {
		c.JSON(400, models.ResponseError{Detail: "Invalid request : Project ID is required"})
		return
	}
	q := &ProjectSingleQuery{}
	err := c.ShouldBindQuery(q)
	if err != nil {
		c.JSON(400, models.ResponseError{Detail: "Invalid request : " + err.Error()})
		return
	}
	u := GetCurrentUser(c)
	if u == nil {
		c.JSON(400, models.ResponseError{Detail: "Invalid request : User not found"})
		return
	}
	// User does not have permission for other projects
	// User does not have
	if !u.Permission.HasPermission(consts.PermissionOtherProjectAccess) && u.LeadingProjectID != nil && *u.LeadingProjectID != models.IDType(projectId) {
		c.JSON(403, models.ResponseError{Detail: "User does not have permission to access this project"})
		return
	}
	project_service := services.NewProjectService()
	project, err := project_service.GetProjectByID(c, models.IDType(projectId), q.IncludeProjectLeads)
	if err != nil {
		c.JSON(400, models.ResponseError{Detail: "Error getting project : " + err.Error()})
		return
	}
	if project == nil {
		c.JSON(404, models.ResponseError{Detail: "Project not found"})
		return
	}
	c.JSON(200, models.ResponseSingle[models.ProjectExtended]{Data: *project})
}

// ListProjects godoc
// @Summary List all projects
// @Description List all projects
// @Produce  json
// @Success 200 {object} models.ResponseList[models.ProjectExtended]
// @Router /project/ [get]
// @Tags Project
// @Security ApiKeyAuth
// @Error 400 {object} models.ResponseError
// @Error 403 {object} models.ResponseError
// @Error 404 {object} models.ResponseError
func ListProjects(c *gin.Context, sess *cache.Session) {
	q := &models.ProjectFilter{}
	err := c.ShouldBindQuery(q)
	if err != nil {
		c.JSON(400, models.ResponseError{Detail: "Invalid request : " + err.Error()})
		return
	}
	u := GetCurrentUser(c)
	if u == nil {
		c.JSON(400, models.ResponseError{Detail: "Invalid request : User not found"})
		return
	}
	if !u.Permission.HasPermission(consts.PermissionOtherProjectAccess) {
		// user does not have permission to access other projects
		// but user is trying to access other projects
		if q.IncludeOtherProjects {
			c.JSON(200, models.ResponseList[models.ProjectExtended]{Data: []models.ProjectExtended{}})
			return
		} else if u.LeadingProjectID != nil {
			// User is not an admin, but he is trying to access his own project
			q.IDs = []models.IDType{*u.LeadingProjectID}
		} else {
			// neither admin nor leading any project
			c.JSON(200, models.ResponseList[models.ProjectExtended]{Data: []models.ProjectExtended{}})
			return
		}
	}

	project_service := services.NewProjectService()
	projects, err := project_service.ListProjects(c, &u.UserID, q)
	if err != nil {
		c.JSON(400, models.ResponseError{Detail: "Error getting projects : " + err.Error()})
		return
	}

	c.JSON(200, models.ResponseList[models.ProjectExtended]{Data: projects})
}
