package adaptor

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/stretchr/testify/assert"
)

func TestSetupCommonRequestHeader(t *testing.T) {
	// 创建测试用的gin上下文
	c, _ := gin.CreateTestContext(nil)
	c.Request = &http.Request{
		Header: make(http.Header),
	}
	c.Request.Header.Set("Content-Type", "application/json")
	c.Request.Header.Set("Accept", "application/json")
	c.Request.Header.Set("x-oneapi-test-header", "test-value")

	// 创建测试用的http请求
	req, _ := http.NewRequest("GET", "http://example.com", nil)

	// 创建测试用的meta
	m := &meta.Meta{
		IsStream: true,
	}

	// 调用被测试函数
	SetupCommonRequestHeader(c, req, m)

	// 验证结果
	assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
	assert.Equal(t, "application/json", req.Header.Get("Accept"))
	assert.Equal(t, "test-value", req.Header.Get("test-header"))
}
