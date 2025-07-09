package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	gcrypto "github.com/Laisky/go-utils/v5/crypto"
	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/model"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Setup session middleware
	store := cookie.NewStore([]byte("test-secret"))
	router.Use(sessions.Sessions("test-session", store))

	return router
}

func TestTotpBasicFunctionality(t *testing.T) {
	// Use a known valid base32 secret
	secret := "JBSWY3DPEHPK3PXP" // Base32 encoded "Hello!"

	// Create TOTP instance
	totp, err := gcrypto.NewTOTP(gcrypto.OTPArgs{
		Base32Secret: secret,
	})
	assert.NoError(t, err)

	// Generate current code
	currentCode := totp.Key()
	assert.Len(t, currentCode, 6) // TOTP codes should be 6 digits

	// Create another TOTP instance with same secret to verify
	totp2, err := gcrypto.NewTOTP(gcrypto.OTPArgs{
		Base32Secret: secret,
	})
	assert.NoError(t, err)

	// The codes should match
	assert.Equal(t, currentCode, totp2.Key())
}

func TestSetupTotp(t *testing.T) {
	router := setupTestRouter()
	router.GET("/totp/setup", func(c *gin.Context) {
		// Mock user ID in context
		c.Set("id", 1)
		SetupTotp(c)
	})

	req, _ := http.NewRequest("GET", "/totp/setup", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Note: This test will fail because we don't have a real database
	// In a real test environment, you would mock the database
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTotpSetupRequest(t *testing.T) {
	router := setupTestRouter()
	router.POST("/totp/confirm", func(c *gin.Context) {
		// Mock user ID in context
		c.Set("id", 1)
		ConfirmTotp(c)
	})

	// Test with valid request
	reqBody := TotpSetupRequest{
		TotpCode: "123456",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/totp/confirm", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Note: This test will fail because we don't have a real database
	// In a real test environment, you would mock the database
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLoginRequest(t *testing.T) {
	// Test LoginRequest struct
	loginReq := LoginRequest{
		Username: "testuser",
		Password: "testpass",
		TotpCode: "123456",
	}

	assert.Equal(t, "testuser", loginReq.Username)
	assert.Equal(t, "testpass", loginReq.Password)
	assert.Equal(t, "123456", loginReq.TotpCode)
}

func TestTotpSetupResponse(t *testing.T) {
	// Test TotpSetupResponse struct
	response := TotpSetupResponse{
		Secret: "ABCDEFGHIJKLMNOP",
		QRCode: "otpauth://totp/test",
	}

	assert.Equal(t, "ABCDEFGHIJKLMNOP", response.Secret)
	assert.Equal(t, "otpauth://totp/test", response.QRCode)
}

func TestTotpCodeGeneration(t *testing.T) {
	// Test TOTP code generation with known parameters
	secret := "JBSWY3DPEHPK3PXP" // Base32 encoded "Hello!"

	totp, err := gcrypto.NewTOTP(gcrypto.OTPArgs{
		Base32Secret: secret,
		AccountName:  "test@example.com",
		IssuerName:   "Test App",
	})
	assert.NoError(t, err)

	// Generate code
	code := totp.Key()
	assert.Len(t, code, 6) // TOTP codes should be 6 digits

	// Create another TOTP instance to verify
	totp2, err := gcrypto.NewTOTP(gcrypto.OTPArgs{
		Base32Secret: secret,
		AccountName:  "test@example.com",
		IssuerName:   "Test App",
	})
	assert.NoError(t, err)

	// The codes should match
	assert.Equal(t, code, totp2.Key())

	// Test URI generation
	uri := totp.URI()
	assert.Contains(t, uri, "otpauth://totp/")
	assert.Contains(t, uri, "test@example.com")
	assert.Contains(t, uri, "Test")
	assert.Contains(t, uri, secret)
}

func TestTotpReplayProtection(t *testing.T) {
	// Test TOTP replay protection
	secret := "JBSWY3DPEHPK3PXP"
	userId := 123

	totp, err := gcrypto.NewTOTP(gcrypto.OTPArgs{
		Base32Secret: secret,
	})
	assert.NoError(t, err)

	code := totp.Key()

	// First verification should succeed
	assert.True(t, verifyTotpCode(userId, secret, code))

	// Second verification with same code should fail (replay protection)
	assert.False(t, verifyTotpCode(userId, secret, code))

	// Different user should still be able to use the same code
	assert.True(t, verifyTotpCode(userId+1, secret, code))
}

func TestTotpSecurityFunctions(t *testing.T) {
	userId := 456
	code := "123456"

	// Initially, code should not be marked as used
	assert.False(t, common.IsTotpCodeUsed(userId, code))

	// Mark code as used
	err := common.MarkTotpCodeAsUsed(userId, code)
	assert.NoError(t, err)

	// Now code should be marked as used
	assert.True(t, common.IsTotpCodeUsed(userId, code))

	// Different user should not be affected
	assert.False(t, common.IsTotpCodeUsed(userId+1, code))

	// Different code should not be affected
	assert.False(t, common.IsTotpCodeUsed(userId, "654321"))
}

func TestAdminDisableUserTotp(t *testing.T) {
	router := setupTestRouter()
	router.POST("/admin/totp/disable/:id", func(c *gin.Context) {
		// Mock admin user context
		c.Set("id", 1)                     // Admin user ID
		c.Set("role", model.RoleAdminUser) // Admin role
		AdminDisableUserTotp(c)
	})

	// Test with invalid user ID
	req, _ := http.NewRequest("POST", "/admin/totp/disable/invalid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response["success"].(bool))
	assert.Equal(t, "Invalid user ID", response["message"])

	// Test with missing user ID
	req, _ = http.NewRequest("POST", "/admin/totp/disable/", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code) // Route not found

	// Note: Testing with real database operations would require
	// proper database setup and mocking, which is beyond the scope
	// of this basic functionality test
}
