//go:build readonly

package routes

import (
	idgenerator "nokib/campwiz/services/idGenerator"

	"github.com/gin-gonic/gin"
)

var serverInstanceId = idgenerator.GenerateID("ReadOnlyServer")

func NewRoutes(nonAPIParent *gin.RouterGroup) *gin.RouterGroup {
	return NewReadOnlyRoutes(nonAPIParent)
}
