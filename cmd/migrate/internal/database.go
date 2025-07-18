package internal

import (
	"fmt"
	"strings"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/songquanpeng/one-api/common"
	oneapilogger "github.com/songquanpeng/one-api/common/logger"
)

// DatabaseConnection represents a database connection with metadata
type DatabaseConnection struct {
	DB     *gorm.DB
	Type   string
	DSN    string
	Driver string
}

// ConnectDatabase establishes a connection to the specified database
func ConnectDatabase(dbType, dsn string) (*DatabaseConnection, error) {
	var db *gorm.DB
	var err error
	var driver string

	// Clean up DSN by removing scheme prefix for actual connection
	cleanDSN := dsn
	if strings.HasPrefix(dsn, "sqlite://") {
		cleanDSN = strings.TrimPrefix(dsn, "sqlite://")
	} else if strings.HasPrefix(dsn, "mysql://") {
		// Convert mysql://user:pass@host:port/db to user:pass@tcp(host:port)/db
		cleanDSN = strings.TrimPrefix(dsn, "mysql://")
		if strings.Contains(cleanDSN, "@") && strings.Contains(cleanDSN, "/") {
			parts := strings.Split(cleanDSN, "@")
			if len(parts) == 2 {
				userPass := parts[0]
				hostDb := parts[1]
				if strings.Contains(hostDb, "/") {
					hostParts := strings.Split(hostDb, "/")
					host := hostParts[0]
					db := strings.Join(hostParts[1:], "/")
					cleanDSN = fmt.Sprintf("%s@tcp(%s)/%s", userPass, host, db)
				}
			}
		}
	}
	// postgres:// DSN can be used directly

	switch strings.ToLower(dbType) {
	case "sqlite":
		driver = "sqlite"
		oneapilogger.SysLog("Connecting to SQLite database")
		// Add busy timeout for SQLite
		if !strings.Contains(cleanDSN, "?") {
			cleanDSN += fmt.Sprintf("?_busy_timeout=%d", common.SQLiteBusyTimeout)
		} else if !strings.Contains(cleanDSN, "_busy_timeout") {
			cleanDSN += fmt.Sprintf("&_busy_timeout=%d", common.SQLiteBusyTimeout)
		}
		db, err = gorm.Open(sqlite.Open(cleanDSN), &gorm.Config{
			PrepareStmt: true,
			Logger:      logger.Default.LogMode(logger.Silent),
		})
	case "mysql":
		driver = "mysql"
		oneapilogger.SysLog("Connecting to MySQL database")
		db, err = gorm.Open(mysql.Open(cleanDSN), &gorm.Config{
			PrepareStmt: true,
			Logger:      logger.Default.LogMode(logger.Silent),
		})
	case "postgres", "postgresql":
		driver = "postgres"
		oneapilogger.SysLog("Connecting to PostgreSQL database")
		db, err = gorm.Open(postgres.New(postgres.Config{
			DSN:                  cleanDSN,
			PreferSimpleProtocol: true,
		}), &gorm.Config{
			PrepareStmt: true,
			Logger:      logger.Default.LogMode(logger.Silent),
		})
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s database: %w", dbType, err)
	}

	// Test the connection
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping %s database: %w", dbType, err)
	}

	oneapilogger.SysLog(fmt.Sprintf("Successfully connected to %s database", dbType))

	return &DatabaseConnection{
		DB:     db,
		Type:   strings.ToLower(dbType),
		DSN:    dsn,
		Driver: driver,
	}, nil
}

// ConnectDatabaseFromDSN establishes a connection by extracting the database type from the DSN
func ConnectDatabaseFromDSN(dsn string) (*DatabaseConnection, error) {
	dbType, err := ExtractDatabaseTypeFromDSN(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to determine database type from DSN: %w", err)
	}

	return ConnectDatabase(dbType, dsn)
}

// Close closes the database connection
func (dc *DatabaseConnection) Close() error {
	if dc.DB == nil {
		return nil
	}

	sqlDB, err := dc.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close %s database connection: %w", dc.Type, err)
	}

	oneapilogger.SysLog(fmt.Sprintf("Closed %s database connection", dc.Type))
	return nil
}

// GetTableNames returns all table names in the database
func (dc *DatabaseConnection) GetTableNames() ([]string, error) {
	var tables []string
	var err error

	switch dc.Type {
	case "sqlite":
		err = dc.DB.Raw("SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name").Scan(&tables).Error
	case "mysql":
		err = dc.DB.Raw("SHOW TABLES").Scan(&tables).Error
	case "postgres":
		err = dc.DB.Raw("SELECT tablename FROM pg_tables WHERE schemaname = 'public' ORDER BY tablename").Scan(&tables).Error
	default:
		return nil, fmt.Errorf("unsupported database type for table listing: %s", dc.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get table names from %s database: %w", dc.Type, err)
	}

	return tables, nil
}

// GetRowCount returns the number of rows in a table
func (dc *DatabaseConnection) GetRowCount(tableName string) (int64, error) {
	var count int64
	err := dc.DB.Table(tableName).Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count rows in table %s: %w", tableName, err)
	}
	return count, nil
}

// TableExists checks if a table exists in the database
func (dc *DatabaseConnection) TableExists(tableName string) (bool, error) {
	var exists bool
	var err error

	switch dc.Type {
	case "sqlite":
		err = dc.DB.Raw("SELECT COUNT(*) > 0 FROM sqlite_master WHERE type='table' AND name=?", tableName).Scan(&exists).Error
	case "mysql":
		err = dc.DB.Raw("SELECT COUNT(*) > 0 FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?", tableName).Scan(&exists).Error
	case "postgres":
		err = dc.DB.Raw("SELECT COUNT(*) > 0 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = ?", tableName).Scan(&exists).Error
	default:
		return false, fmt.Errorf("unsupported database type for table existence check: %s", dc.Type)
	}

	if err != nil {
		return false, fmt.Errorf("failed to check if table %s exists: %w", tableName, err)
	}

	return exists, nil
}

// ValidateConnection performs basic validation on the database connection
func (dc *DatabaseConnection) ValidateConnection() error {
	// Test basic query
	var result int
	if err := dc.DB.Raw("SELECT 1").Scan(&result).Error; err != nil {
		return fmt.Errorf("failed to execute test query: %w", err)
	}

	if result != 1 {
		return fmt.Errorf("unexpected result from test query: got %d, expected 1", result)
	}

	oneapilogger.SysLog(fmt.Sprintf("%s database connection validated successfully", dc.Type))
	return nil
}
