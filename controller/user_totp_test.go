package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	gcrypto "github.com/Laisky/go-utils/v5/crypto"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/model"
)

func setupTestDB(t *testing.T) *gorm.DB {
	// Create in-memory SQLite database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate the tables
	err = db.AutoMigrate(&model.User{}, &model.Channel{}, &model.Token{}, &model.Option{}, &model.Redemption{}, &model.Ability{}, &model.Log{}, &model.UserRequestCost{})
	require.NoError(t, err)

	return db
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Setup session middleware
	store := cookie.NewStore([]byte("test-secret"))
	router.Use(sessions.Sessions("test-session", store))

	return router
}

func setupTestEnvironment(t *testing.T) (*gorm.DB, func()) {
	// Setup test database
	testDB := setupTestDB(t)

	// Store original DB and replace with test DB
	originalDB := model.DB
	originalLogDB := model.LOG_DB
	model.DB = testDB
	model.LOG_DB = testDB // Use same DB for logging in tests

	// Set SQLite flag for proper query handling
	originalUsingSQLite := common.UsingSQLite
	common.UsingSQLite = true

	// Disable Redis for testing to use memory-based rate limiting
	originalRedisEnabled := common.RedisEnabled
	common.RedisEnabled = false

	// Create a test user for TOTP tests
	testUser := &model.User{
		Id:          1,
		Username:    "testuser",
		Password:    "hashedpassword",
		Role:        model.RoleCommonUser,
		Status:      model.UserStatusEnabled,
		DisplayName: "Test User",
		Email:       "test@example.com",
		AccessToken: "test-access-token-1",
		AffCode:     "TEST1",
		TotpSecret:  "",
	}
	err := testDB.Create(testUser).Error
	require.NoError(t, err)

	// Return cleanup function
	cleanup := func() {
		model.DB = originalDB
		model.LOG_DB = originalLogDB
		common.UsingSQLite = originalUsingSQLite
		common.RedisEnabled = originalRedisEnabled
	}

	return testDB, cleanup
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
	_, cleanup := setupTestEnvironment(t)
	defer cleanup()

	router := setupTestRouter()
	router.GET("/totp/setup", func(c *gin.Context) {
		// Mock user ID in context
		c.Set("id", 1)
		SetupTotp(c)
	})

	req, _ := http.NewRequest("GET", "/totp/setup", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Parse response to verify TOTP setup data
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response["success"].(bool))

	data := response["data"].(map[string]interface{})
	assert.NotEmpty(t, data["secret"])
	assert.NotEmpty(t, data["qr_code"])
}

func TestTotpSetupRequest(t *testing.T) {
	_, cleanup := setupTestEnvironment(t)
	defer cleanup()

	router := setupTestRouter()

	// First, set up TOTP to get a secret in session
	router.GET("/totp/setup", func(c *gin.Context) {
		c.Set("id", 1)
		SetupTotp(c)
	})

	router.POST("/totp/confirm", func(c *gin.Context) {
		c.Set("id", 1)
		ConfirmTotp(c)
	})

	// Step 1: Call setup to get secret and store it in session
	setupReq, _ := http.NewRequest("GET", "/totp/setup", nil)
	setupW := httptest.NewRecorder()
	router.ServeHTTP(setupW, setupReq)

	assert.Equal(t, http.StatusOK, setupW.Code)

	// Parse setup response to get the secret
	var setupResponse map[string]interface{}
	err := json.Unmarshal(setupW.Body.Bytes(), &setupResponse)
	assert.NoError(t, err)
	assert.True(t, setupResponse["success"].(bool))

	data := setupResponse["data"].(map[string]interface{})
	secret := data["secret"].(string)

	// Step 2: Generate a valid TOTP code using the secret from setup
	totp, err := gcrypto.NewTOTP(gcrypto.OTPArgs{
		Base32Secret: secret,
	})
	require.NoError(t, err)
	validCode := totp.Key()

	// Step 3: Confirm TOTP with the valid code
	// We need to use the same session, so we'll extract cookies from setup response
	reqBody := TotpSetupRequest{
		TotpCode: validCode,
	}
	jsonBody, _ := json.Marshal(reqBody)

	confirmReq, _ := http.NewRequest("POST", "/totp/confirm", bytes.NewBuffer(jsonBody))
	confirmReq.Header.Set("Content-Type", "application/json")

	// Copy cookies from setup request to maintain session
	for _, cookie := range setupW.Result().Cookies() {
		confirmReq.AddCookie(cookie)
	}

	confirmW := httptest.NewRecorder()
	router.ServeHTTP(confirmW, confirmReq)

	assert.Equal(t, http.StatusOK, confirmW.Code)

	// Parse response to verify success
	var confirmResponse map[string]interface{}
	err = json.Unmarshal(confirmW.Body.Bytes(), &confirmResponse)
	assert.NoError(t, err)
	assert.True(t, confirmResponse["success"].(bool))
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
	_, cleanup := setupTestEnvironment(t)
	defer cleanup()

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
	_, cleanup := setupTestEnvironment(t)
	defer cleanup()

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
	testDB, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create an admin user
	adminUser := &model.User{
		Id:          2,
		Username:    "admin",
		Password:    "hashedpassword",
		Role:        model.RoleAdminUser,
		Status:      model.UserStatusEnabled,
		DisplayName: "Admin User",
		Email:       "admin@example.com",
		AccessToken: "test-access-token-2",
		AffCode:     "ADMIN",
	}
	err := testDB.Create(adminUser).Error
	require.NoError(t, err)

	// Create a target user with TOTP enabled
	targetUser := &model.User{
		Id:          3,
		Username:    "target",
		Password:    "hashedpassword",
		Role:        model.RoleCommonUser,
		Status:      model.UserStatusEnabled,
		DisplayName: "Target User",
		Email:       "target@example.com",
		AccessToken: "test-access-token-3",
		AffCode:     "TARG3",
		TotpSecret:  "JBSWY3DPEHPK3PXP",
	}
	err = testDB.Create(targetUser).Error
	require.NoError(t, err)

	router := setupTestRouter()
	router.POST("/admin/totp/disable/:id", func(c *gin.Context) {
		// Mock admin user context
		c.Set("id", 2)                     // Admin user ID
		c.Set("role", model.RoleAdminUser) // Admin role
		AdminDisableUserTotp(c)
	})

	// Test with valid user ID
	req, _ := http.NewRequest("POST", "/admin/totp/disable/3", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Parse response to verify success
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response["success"].(bool))

	// Verify that TOTP secret was cleared from database
	var updatedUser model.User
	err = testDB.First(&updatedUser, 3).Error
	assert.NoError(t, err)
	assert.Empty(t, updatedUser.TotpSecret)

	// Test with invalid user ID
	req, _ = http.NewRequest("POST", "/admin/totp/disable/invalid", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response["success"].(bool))
	assert.Equal(t, "Invalid user ID", response["message"])
}
