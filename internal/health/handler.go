package health

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetHandler returns a status which indicates
// that the service is currently running
func GetHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Status(http.StatusOK)
		c.Done()
	}
}

// TODO: add more advanced health check which checks
// health of DB and other external service dependencies
