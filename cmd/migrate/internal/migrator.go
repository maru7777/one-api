package internal

import (
	"context"
	"fmt"
	"time"

	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/model"
)

// Migrator handles database migration between different database types
type Migrator struct {
	SourceType string
	SourceDSN  string
	TargetType string
	TargetDSN  string
	DryRun     bool
	Verbose    bool
	Workers    int // Number of concurrent workers
	BatchSize  int // Batch size for processing

	sourceConn *DatabaseConnection
	targetConn *DatabaseConnection
}

// MigrationStats holds statistics about the migration process
type MigrationStats struct {
	StartTime    time.Time
	EndTime      time.Time
	TablesTotal  int
	TablesDone   int
	RecordsTotal int64
	RecordsDone  int64
	Errors       []error
}

// Migrate performs the complete migration process
func (m *Migrator) Migrate(ctx context.Context) error {
	stats := &MigrationStats{
		StartTime: time.Now(),
		Errors:    make([]error, 0),
	}

	logger.SysLog("Starting database migration process")
	logger.SysLog(fmt.Sprintf("Source: %s (%s)", m.SourceType, m.SourceDSN))
	logger.SysLog(fmt.Sprintf("Target: %s (%s)", m.TargetType, m.TargetDSN))

	if m.DryRun {
		logger.SysLog("Running in DRY RUN mode - no changes will be made")
	}

	// Step 1: Connect to databases
	if err := m.connectDatabases(); err != nil {
		return fmt.Errorf("failed to connect to databases: %w", err)
	}
	defer m.closeDatabases()

	// Step 2: Validate connections and compatibility
	if err := m.validateMigration(); err != nil {
		return fmt.Errorf("migration validation failed: %w", err)
	}

	// Step 3: Analyze source database
	if err := m.analyzeSource(stats); err != nil {
		return fmt.Errorf("source analysis failed: %w", err)
	}

	// Step 4: Prepare target database
	if !m.DryRun {
		if err := m.prepareTarget(); err != nil {
			return fmt.Errorf("target preparation failed: %w", err)
		}
	}

	// Step 5: Migrate data
	if err := m.migrateData(ctx, stats); err != nil {
		return fmt.Errorf("data migration failed: %w", err)
	}

	// Step 6: Fix PostgreSQL sequences (if target is PostgreSQL)
	if !m.DryRun && m.targetConn.Type == "postgres" {
		if err := m.fixPostgreSQLSequences(); err != nil {
			return fmt.Errorf("PostgreSQL sequence fix failed: %w", err)
		}
	}

	// Step 7: Validate migration results
	if !m.DryRun {
		if err := m.validateResults(stats); err != nil {
			return fmt.Errorf("migration validation failed: %w", err)
		}
	}

	stats.EndTime = time.Now()
	m.printStats(stats)

	return nil
}

// connectDatabases establishes connections to both source and target databases
func (m *Migrator) connectDatabases() error {
	var err error

	// Connect to source database
	m.sourceConn, err = ConnectDatabaseFromDSN(m.SourceDSN)
	if err != nil {
		return fmt.Errorf("failed to connect to source database: %w", err)
	}

	// Connect to target database
	m.targetConn, err = ConnectDatabaseFromDSN(m.TargetDSN)
	if err != nil {
		return fmt.Errorf("failed to connect to target database: %w", err)
	}

	return nil
}

// closeDatabases closes all database connections
func (m *Migrator) closeDatabases() {
	if m.sourceConn != nil {
		if err := m.sourceConn.Close(); err != nil {
			logger.SysError(fmt.Sprintf("Failed to close source database: %v", err))
		}
	}
	if m.targetConn != nil {
		if err := m.targetConn.Close(); err != nil {
			logger.SysError(fmt.Sprintf("Failed to close target database: %v", err))
		}
	}
}

// validateMigration performs pre-migration validation
func (m *Migrator) validateMigration() error {
	logger.SysLog("Validating database connections...")

	// Validate source connection
	if err := m.sourceConn.ValidateConnection(); err != nil {
		return fmt.Errorf("source database validation failed: %w", err)
	}

	// Validate target connection
	if err := m.targetConn.ValidateConnection(); err != nil {
		return fmt.Errorf("target database validation failed: %w", err)
	}

	// Check if source and target are the same
	if m.sourceConn.Type == m.targetConn.Type && m.sourceConn.DSN == m.targetConn.DSN {
		return fmt.Errorf("source and target databases cannot be the same")
	}

	logger.SysLog("Database connections validated successfully")
	return nil
}

// analyzeSource analyzes the source database structure and data
func (m *Migrator) analyzeSource(stats *MigrationStats) error {
	logger.SysLog("Analyzing source database...")

	// Get all tables
	tables, err := m.sourceConn.GetTableNames()
	if err != nil {
		return fmt.Errorf("failed to get source table names: %w", err)
	}

	stats.TablesTotal = len(tables)
	logger.SysLog(fmt.Sprintf("Found %d tables in source database", len(tables)))

	// Count total records
	var totalRecords int64
	for _, table := range tables {
		count, err := m.sourceConn.GetRowCount(table)
		if err != nil {
			logger.SysWarn(fmt.Sprintf("Failed to count rows in table %s: %v", table, err))
			continue
		}
		totalRecords += count
		if m.Verbose {
			logger.SysLog(fmt.Sprintf("Table %s: %d records", table, count))
		}
	}

	stats.RecordsTotal = totalRecords
	logger.SysLog(fmt.Sprintf("Total records to migrate: %d", totalRecords))

	return nil
}

// prepareTarget prepares the target database for migration
func (m *Migrator) prepareTarget() error {
	logger.SysLog("Preparing target database...")

	// Run GORM auto-migration to create tables
	if err := m.runAutoMigration(); err != nil {
		return fmt.Errorf("failed to run auto-migration: %w", err)
	}

	logger.SysLog("Target database prepared successfully")
	return nil
}

// runAutoMigration runs GORM's AutoMigrate on the target database
func (m *Migrator) runAutoMigration() error {
	logger.SysLog("Running GORM auto-migration on target database...")

	// Set the global DB to target connection for migration
	originalDB := model.DB
	model.DB = m.targetConn.DB
	defer func() {
		model.DB = originalDB
	}()

	// Run migrations for all models
	if err := model.DB.AutoMigrate(&model.Channel{}); err != nil {
		return fmt.Errorf("failed to migrate Channel: %w", err)
	}
	if err := model.DB.AutoMigrate(&model.Token{}); err != nil {
		return fmt.Errorf("failed to migrate Token: %w", err)
	}
	if err := model.DB.AutoMigrate(&model.User{}); err != nil {
		return fmt.Errorf("failed to migrate User: %w", err)
	}
	if err := model.DB.AutoMigrate(&model.Option{}); err != nil {
		return fmt.Errorf("failed to migrate Option: %w", err)
	}
	if err := model.DB.AutoMigrate(&model.Redemption{}); err != nil {
		return fmt.Errorf("failed to migrate Redemption: %w", err)
	}
	if err := model.DB.AutoMigrate(&model.Ability{}); err != nil {
		return fmt.Errorf("failed to migrate Ability: %w", err)
	}
	if err := model.DB.AutoMigrate(&model.Log{}); err != nil {
		return fmt.Errorf("failed to migrate Log: %w", err)
	}
	if err := model.DB.AutoMigrate(&model.UserRequestCost{}); err != nil {
		return fmt.Errorf("failed to migrate UserRequestCost: %w", err)
	}

	logger.SysLog("GORM auto-migration completed successfully")
	return nil
}

// printStats prints migration statistics
func (m *Migrator) printStats(stats *MigrationStats) {
	duration := stats.EndTime.Sub(stats.StartTime)

	logger.SysLog("=== Migration Statistics ===")
	logger.SysLog(fmt.Sprintf("Duration: %v", duration))
	logger.SysLog(fmt.Sprintf("Tables processed: %d/%d", stats.TablesDone, stats.TablesTotal))
	logger.SysLog(fmt.Sprintf("Records processed: %d/%d", stats.RecordsDone, stats.RecordsTotal))

	if len(stats.Errors) > 0 {
		logger.SysLog(fmt.Sprintf("Errors encountered: %d", len(stats.Errors)))
		for i, err := range stats.Errors {
			logger.SysError(fmt.Sprintf("Error %d: %v", i+1, err))
		}
	} else {
		logger.SysLog("No errors encountered")
	}
}

// ValidateOnly performs validation without migration
func (m *Migrator) ValidateOnly(ctx context.Context) error {
	logger.SysLog("Running validation-only mode")

	// Connect to databases
	if err := m.connectDatabases(); err != nil {
		return fmt.Errorf("failed to connect to databases: %w", err)
	}
	defer m.closeDatabases()

	// Validate connections
	if err := m.validateMigration(); err != nil {
		return fmt.Errorf("migration validation failed: %w", err)
	}

	// Analyze source
	stats := &MigrationStats{
		StartTime: time.Now(),
		Errors:    make([]error, 0),
	}

	if err := m.analyzeSource(stats); err != nil {
		return fmt.Errorf("source analysis failed: %w", err)
	}

	logger.SysLog("Validation completed successfully")
	return nil
}

// GetMigrationPlan returns a plan of what will be migrated
func (m *Migrator) GetMigrationPlan() (*MigrationPlan, error) {
	plan := &MigrationPlan{
		SourceType: m.SourceType,
		TargetType: m.TargetType,
		Tables:     make([]TablePlan, 0),
	}

	// Connect to source database
	sourceConn, err := ConnectDatabase(m.SourceType, m.SourceDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to source database: %w", err)
	}
	defer sourceConn.Close()

	// Analyze each table
	for _, tableInfo := range TableMigrationOrder {
		exists, err := sourceConn.TableExists(tableInfo.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to check table %s: %w", tableInfo.Name, err)
		}

		if !exists {
			continue
		}

		count, err := sourceConn.GetRowCount(tableInfo.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to get row count for %s: %w", tableInfo.Name, err)
		}

		plan.Tables = append(plan.Tables, TablePlan{
			Name:        tableInfo.Name,
			RecordCount: count,
			Exists:      exists,
		})
		plan.TotalRecords += count
	}

	return plan, nil
}

// MigrationPlan represents a migration plan
type MigrationPlan struct {
	SourceType   string      `json:"source_type"`
	TargetType   string      `json:"target_type"`
	Tables       []TablePlan `json:"tables"`
	TotalRecords int64       `json:"total_records"`
}

// TablePlan represents a table migration plan
type TablePlan struct {
	Name        string `json:"name"`
	RecordCount int64  `json:"record_count"`
	Exists      bool   `json:"exists"`
}

// fixPostgreSQLSequences updates PostgreSQL sequences to match the maximum ID values
// This is necessary after migrating data from other databases to ensure new records
// get correct auto-increment IDs
func (m *Migrator) fixPostgreSQLSequences() error {
	logger.SysLog("Fixing PostgreSQL sequences after data migration...")

	// Define tables that have auto-increment ID columns
	tablesWithSequences := []string{
		"users",
		"tokens",
		"channels",
		"options",
		"redemptions",
		"abilities",
		"logs",
		"user_request_costs",
	}

	for _, tableName := range tablesWithSequences {
		if err := m.fixTableSequence(tableName); err != nil {
			logger.SysWarn(fmt.Sprintf("Failed to fix sequence for table %s: %v", tableName, err))
			// Continue with other tables instead of failing completely
			continue
		}
		logger.SysLog(fmt.Sprintf("Fixed sequence for table: %s", tableName))
	}

	logger.SysLog("PostgreSQL sequence fixing completed")
	return nil
}

// fixTableSequence fixes the sequence for a specific table
func (m *Migrator) fixTableSequence(tableName string) error {
	// First check if the table exists and has records
	var count int64
	if err := m.targetConn.DB.Table(tableName).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to count records in table %s: %w", tableName, err)
	}

	if count == 0 {
		logger.SysLog(fmt.Sprintf("Table %s is empty, skipping sequence fix", tableName))
		return nil
	}

	// Get the maximum ID value from the table
	var maxID int64
	if err := m.targetConn.DB.Table(tableName).Select("COALESCE(MAX(id), 0)").Scan(&maxID).Error; err != nil {
		return fmt.Errorf("failed to get max ID from table %s: %w", tableName, err)
	}

	if maxID == 0 {
		logger.SysLog(fmt.Sprintf("Table %s has no valid IDs, skipping sequence fix", tableName))
		return nil
	}

	// Update the sequence to start from maxID + 1
	sequenceName := tableName + "_id_seq"
	sql := fmt.Sprintf("SELECT setval('%s', %d, true)", sequenceName, maxID)

	if err := m.targetConn.DB.Exec(sql).Error; err != nil {
		return fmt.Errorf("failed to update sequence %s: %w", sequenceName, err)
	}

	logger.SysLog(fmt.Sprintf("Updated sequence %s to start from %d", sequenceName, maxID+1))
	return nil
}
