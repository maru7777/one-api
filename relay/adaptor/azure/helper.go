package azure

import (
	"github.com/Laisky/one-api/common/ctxkey"
	"github.com/gin-gonic/gin"
)

func GetAPIVersion(c *gin.Context) string {
	query := c.Request.URL.Query()
	apiVersion := query.Get("api-version")
	if apiVersion == "" {
		apiVersion = c.GetString(ctxkey.ConfigAPIVersion)
	}
	return apiVersion
}
