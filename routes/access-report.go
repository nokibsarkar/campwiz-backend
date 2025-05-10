package routes

import (
	"nokib/campwiz/repository/cache"
	"os"

	"github.com/gin-gonic/gin"
)

func GetAccessReport(c *gin.Context, sess *cache.Session) {
	if sess == nil {
		c.String(401, "Session not found")
		return
	}
	// Implementation for getting access report
	const ACCESS_REPORT = "access-report.html"
	fp, err := os.Open(ACCESS_REPORT)
	if err != nil {
		c.String(404, "Error opening file: %v", err)
		return
	}
	defer fp.Close()
	c.Header("Content-Type", "text/html")
	fp.Seek(0, 0)
	_, err = fp.WriteTo(c.Writer)
	if err != nil {
		c.String(500, "Error writing file: %v", err)
		return
	}
}
func AccessReportRoutes(parent *gin.RouterGroup) {
	r := parent.Group("/access-report")
	r.GET("/", WithSession(GetAccessReport))
}
