package api

import (
	"taskboard-api-go/service"

	"github.com/gin-gonic/gin"
)

// GetPathParameter gets path parameter. If value doesn't exist, returns path parameter error.
func GetPathParameter(c *gin.Context, key string) (string, error) {
	value := c.Param(key)
	if value == "" {
		return "", service.NewPathParameterError(key)
	}
	return value, nil
}
