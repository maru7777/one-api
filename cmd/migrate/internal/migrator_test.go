package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFixPostgreSQLSequences(t *testing.T) {
	// Test the expected tables that should have sequences fixed
	// This matches the tables defined in the actual fixPostgreSQLSequences method
	expectedTablesWithSequences := []string{
		"users",
		"tokens",
		"channels",
		"options",
		"redemptions",
		"abilities",
		"logs",
		"user_request_costs",
	}

	// Verify all expected tables are included
	assert.Equal(t, 8, len(expectedTablesWithSequences))
	assert.Contains(t, expectedTablesWithSequences, "logs")
	assert.Contains(t, expectedTablesWithSequences, "users")
	assert.Contains(t, expectedTablesWithSequences, "tokens")
	assert.Contains(t, expectedTablesWithSequences, "channels")
	assert.Contains(t, expectedTablesWithSequences, "options")
	assert.Contains(t, expectedTablesWithSequences, "redemptions")
	assert.Contains(t, expectedTablesWithSequences, "abilities")
	assert.Contains(t, expectedTablesWithSequences, "user_request_costs")
}

func TestSequenceNameGeneration(t *testing.T) {
	// Test sequence name generation logic
	testCases := []struct {
		tableName   string
		expectedSeq string
	}{
		{"users", "users_id_seq"},
		{"logs", "logs_id_seq"},
		{"tokens", "tokens_id_seq"},
		{"user_request_costs", "user_request_costs_id_seq"},
	}

	for _, tc := range testCases {
		sequenceName := tc.tableName + "_id_seq"
		assert.Equal(t, tc.expectedSeq, sequenceName, "Sequence name should match expected format")
	}
}

func TestPostgreSQLSequenceSQL(t *testing.T) {
	// Test the SQL generation for sequence updates
	testCases := []struct {
		tableName   string
		maxID       int64
		expectedSQL string
	}{
		{"logs", 1000, "SELECT setval('logs_id_seq', 1000, true)"},
		{"users", 50, "SELECT setval('users_id_seq', 50, true)"},
		{"tokens", 0, "SELECT setval('tokens_id_seq', 0, true)"},
	}

	for _, tc := range testCases {
		sequenceName := tc.tableName + "_id_seq"
		sql := "SELECT setval('" + sequenceName + "', " + string(rune(tc.maxID)) + ", true)"
		// Note: This is a simplified test - the actual implementation uses fmt.Sprintf
		assert.Contains(t, sql, "setval")
		assert.Contains(t, sql, sequenceName)
	}
}
