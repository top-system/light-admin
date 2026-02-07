package lib

import (
	"os"
	"path/filepath"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// DatabaseEngine represents the database engine type
type DatabaseEngine string

const (
	DatabaseEngineMySQL    DatabaseEngine = "mysql"
	DatabaseEngineSQLite   DatabaseEngine = "sqlite"
	DatabaseEnginePostgres DatabaseEngine = "postgres"
)

// CurrentDatabaseEngine holds the current database engine type
var CurrentDatabaseEngine DatabaseEngine = DatabaseEngineMySQL

type Database struct {
	ORM *gorm.DB
}

// NewDatabase creates a new database instance
func NewDatabase(config Config, logger Logger) Database {
	var db *gorm.DB
	var err error

	gormConfig := &gorm.Config{
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
		SkipDefaultTransaction:                   true,
		DisableForeignKeyConstraintWhenMigrating: true,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
			TablePrefix:   config.Database.TablePrefix + "_",
		},
		QueryFields: true,
	}

	switch {
	case config.Database.IsSQLite():
		db, err = openSQLite(config, gormConfig, logger)
		CurrentDatabaseEngine = DatabaseEngineSQLite
	case config.Database.IsPostgreSQL():
		db, err = openPostgreSQL(config, gormConfig, logger)
		CurrentDatabaseEngine = DatabaseEnginePostgres
	default:
		db, err = openMySQL(config, gormConfig, logger)
		CurrentDatabaseEngine = DatabaseEngineMySQL
	}

	if err != nil {
		logger.Zap.Fatalf("Error to open database connection: %v", err)
	}

	// Apply connection pool settings
	sqlDB, err := db.DB()
	if err != nil {
		logger.Zap.Fatalf("Error to get underlying sql.DB: %v", err)
	}

	sqlDB.SetMaxIdleConns(config.Database.MaxIdleConns)
	sqlDB.SetMaxOpenConns(config.Database.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(config.Database.MaxLifetime) * time.Second)

	if config.Log.Level == "debug" {
		db = db.Debug()
	}

	logger.Zap.Infof("Database connection established (engine: %s)", CurrentDatabaseEngine)
	return Database{
		ORM: db,
	}
}

// openMySQL opens a MySQL database connection
func openMySQL(config Config, gormConfig *gorm.Config, logger Logger) (*gorm.DB, error) {
	mc := mysql.Config{
		DSN:                       config.Database.DSN(),
		DefaultStringSize:         191,
		SkipInitializeWithVersion: false,
		DisableDatetimePrecision:  true,
		DontSupportRenameIndex:    true,
		DontSupportRenameColumn:   true,
	}

	db, err := gorm.Open(mysql.New(mc), gormConfig)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// openSQLite opens a SQLite database connection
func openSQLite(config Config, gormConfig *gorm.Config, logger Logger) (*gorm.DB, error) {
	dbPath := config.Database.Name
	if dbPath == "" {
		dbPath = "./data/app.db"
	}

	// Ensure the directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	// Add SQLite connection parameters for better concurrency
	// _journal_mode=WAL: Enable Write-Ahead Logging for better concurrent access
	// _busy_timeout=5000: Wait up to 5 seconds when database is locked
	// _synchronous=NORMAL: Balance between safety and speed
	dsn := dbPath + "?_journal_mode=WAL&_busy_timeout=5000&_synchronous=NORMAL"

	db, err := gorm.Open(sqlite.Open(dsn), gormConfig)
	if err != nil {
		return nil, err
	}

	// Enable foreign keys for SQLite
	db.Exec("PRAGMA foreign_keys = ON")

	return db, nil
}

// openPostgreSQL opens a PostgreSQL database connection
func openPostgreSQL(config Config, gormConfig *gorm.Config, logger Logger) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(config.Database.DSN()), gormConfig)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// IsSQLite returns true if current database is SQLite
func IsSQLite() bool {
	return CurrentDatabaseEngine == DatabaseEngineSQLite
}

// IsMySQL returns true if current database is MySQL
func IsMySQL() bool {
	return CurrentDatabaseEngine == DatabaseEngineMySQL
}

// IsPostgreSQL returns true if current database is PostgreSQL
func IsPostgreSQL() bool {
	return CurrentDatabaseEngine == DatabaseEnginePostgres
}
