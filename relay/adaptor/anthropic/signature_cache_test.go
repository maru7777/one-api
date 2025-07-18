package anthropic

import (
	"testing"
	"time"

	"github.com/songquanpeng/one-api/relay/model"
)

func TestThinkingSignatureCache(t *testing.T) {
	// Test basic cache operations
	cache := NewThinkingSignatureCache(time.Hour)

	// Test Store and Get
	key := "test_key"
	signature := "test_signature_value"

	cache.Store(key, signature)

	result := cache.Get(key)
	if result == nil {
		t.Fatal("Expected signature to be found, got nil")
	}
	if *result != signature {
		t.Fatalf("Expected signature %s, got %s", signature, *result)
	}

	// Test non-existent key
	nonExistentResult := cache.Get("non_existent_key")
	if nonExistentResult != nil {
		t.Fatal("Expected nil for non-existent key")
	}

	// Test Delete
	cache.Delete(key)
	deletedResult := cache.Get(key)
	if deletedResult != nil {
		t.Fatal("Expected nil after deletion")
	}
}

func TestSignatureCacheTTL(t *testing.T) {
	// Test TTL functionality
	cache := NewThinkingSignatureCache(100 * time.Millisecond)

	key := "ttl_test_key"
	signature := "ttl_test_signature"

	cache.Store(key, signature)

	// Should be available immediately
	result := cache.Get(key)
	if result == nil || *result != signature {
		t.Fatal("Signature should be available immediately after storage")
	}

	// Wait for TTL to expire
	time.Sleep(150 * time.Millisecond)

	// Should be expired now
	expiredResult := cache.Get(key)
	if expiredResult != nil {
		t.Fatal("Signature should be expired after TTL")
	}
}

func TestGenerateSignatureKey(t *testing.T) {
	tokenID := "token_123"
	conversationID := "conv_abc"
	messageIndex := 5
	thinkingIndex := 2

	expected := "thinking_sig:token_123:conv_abc:5:2"
	result := generateSignatureKey(tokenID, conversationID, messageIndex, thinkingIndex)

	if result != expected {
		t.Fatalf("Expected key %s, got %s", expected, result)
	}
}

func TestGenerateConversationID(t *testing.T) {
	// Test with empty messages
	emptyMessages := []model.Message{}
	emptyResult := generateConversationID(emptyMessages)
	if emptyResult == "" {
		t.Fatal("Conversation ID should not be empty for empty messages")
	}

	// Test with user messages
	messages := []model.Message{
		{Role: "user", Content: "Hello, how are you?"},
		{Role: "assistant", Content: "I'm doing well, thank you!"},
		{Role: "user", Content: "What's the weather like?"},
	}

	result1 := generateConversationID(messages)
	if result1 == "" {
		t.Fatal("Conversation ID should not be empty")
	}

	// Same messages should produce same ID
	result2 := generateConversationID(messages)
	if result1 != result2 {
		t.Fatalf("Same messages should produce same conversation ID: %s != %s", result1, result2)
	}

	// Different messages should produce different ID
	differentMessages := []model.Message{
		{Role: "user", Content: "Different message"},
	}
	result3 := generateConversationID(differentMessages)
	if result1 == result3 {
		t.Fatal("Different messages should produce different conversation IDs")
	}
}

func TestTruncateForHash(t *testing.T) {
	// Test normal case
	content := "This is a test message"
	result := truncateForHash(content, 10)
	expected := "This is a "
	if result != expected {
		t.Fatalf("Expected %s, got %s", expected, result)
	}

	// Test case where content is shorter than maxLen
	shortContent := "Short"
	result2 := truncateForHash(shortContent, 10)
	if result2 != shortContent {
		t.Fatalf("Expected %s, got %s", shortContent, result2)
	}

	// Test empty content
	emptyResult := truncateForHash("", 10)
	if emptyResult != "" {
		t.Fatal("Expected empty string for empty input")
	}
}

func TestGetTokenIDFromRequest(t *testing.T) {
	tokenID := 12345
	expected := "token_12345"
	result := getTokenIDFromRequest(tokenID)

	if result != expected {
		t.Fatalf("Expected %s, got %s", expected, result)
	}
}

func TestCacheSize(t *testing.T) {
	cache := NewThinkingSignatureCache(time.Hour)

	// Initially empty
	if cache.Size() != 0 {
		t.Fatal("Cache should be empty initially")
	}

	// Add some entries
	cache.Store("key1", "sig1")
	cache.Store("key2", "sig2")
	cache.Store("key3", "sig3")

	if cache.Size() != 3 {
		t.Fatalf("Expected cache size 3, got %d", cache.Size())
	}

	// Delete one entry
	cache.Delete("key2")

	if cache.Size() != 2 {
		t.Fatalf("Expected cache size 2 after deletion, got %d", cache.Size())
	}
}

func TestCacheCleanup(t *testing.T) {
	// Test that cleanup removes expired entries
	cache := NewThinkingSignatureCache(50 * time.Millisecond)

	// Add some entries
	cache.Store("key1", "sig1")
	cache.Store("key2", "sig2")

	// Verify they exist
	if cache.Size() != 2 {
		t.Fatal("Expected 2 entries before expiration")
	}

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Manually trigger cleanup
	cache.cleanupExpired()

	// Should be empty now
	if cache.Size() != 0 {
		t.Fatalf("Expected 0 entries after cleanup, got %d", cache.Size())
	}
}

// Mock backend for testing
type mockBackend struct {
	data map[string]string
	fail bool
}

func (m *mockBackend) Store(key, signature string, ttl time.Duration) error {
	if m.fail {
		return &mockError{}
	}
	if m.data == nil {
		m.data = make(map[string]string)
	}
	m.data[key] = signature
	return nil
}

func (m *mockBackend) Get(key string) (*string, error) {
	if m.fail {
		return nil, &mockError{}
	}
	if value, exists := m.data[key]; exists {
		return &value, nil
	}
	return nil, nil
}

func (m *mockBackend) Delete(key string) error {
	if m.fail {
		return &mockError{}
	}
	delete(m.data, key)
	return nil
}

type mockError struct{}

func (e *mockError) Error() string {
	return "mock error"
}

func TestCacheWithBackend(t *testing.T) {
	cache := NewThinkingSignatureCache(time.Hour)
	backend := &mockBackend{}
	cache.SetBackend(backend)

	// Test Store and Get with backend
	key := "backend_test_key"
	signature := "backend_test_signature"

	cache.Store(key, signature)

	// Should retrieve from backend
	result := cache.Get(key)
	if result == nil || *result != signature {
		t.Fatal("Should retrieve signature from backend")
	}

	// Verify it's actually in the backend
	if backendValue, exists := backend.data[key]; !exists || backendValue != signature {
		t.Fatal("Signature should be stored in backend")
	}
}

func TestCacheBackendFallback(t *testing.T) {
	cache := NewThinkingSignatureCache(time.Hour)
	backend := &mockBackend{fail: true} // Backend that always fails
	cache.SetBackend(backend)

	// Test Store fallback to in-memory
	key := "fallback_test_key"
	signature := "fallback_test_signature"

	cache.Store(key, signature)

	// Should fallback to in-memory cache
	result := cache.Get(key)
	if result == nil || *result != signature {
		t.Fatal("Should fallback to in-memory cache when backend fails")
	}
}
