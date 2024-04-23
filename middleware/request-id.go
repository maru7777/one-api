package middleware

import (
	"context"

	"github.com/Laisky/one-api/common/ctxkey"
	"github.com/Laisky/one-api/common/helper"
	"github.com/gin-gonic/gin"
)

func RequestId() func(c *gin.Context) {
	return func(c *gin.Context) {
		id := helper.GenRequestID()
		c.Set(ctxkey.RequestId, id)
		ctx := context.WithValue(c.Request.Context(), ctxkey.RequestId, id)
		c.Request = c.Request.WithContext(ctx)
		c.Header(ctxkey.RequestId, id)
		c.Next()
	}
}
