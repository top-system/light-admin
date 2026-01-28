package lib

import (
	"os"
	"path/filepath"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// DatabaseEngine represents the database engine type
type DatabaseEngine string

const (
	DatabaseEngineMySQL  DatabaseEngine = "mysql"
	DatabaseEngineSQLite DatabaseEngine = "sqlite"
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

	if config.Database.IsSQLite() {
		db, err = openSQLite(config, gormConfig, logger)
		CurrentDatabaseEngine = DatabaseEngineSQLite
	} else {
		db, err = openMySQL(config, gormConfig, logger)
		CurrentDatabaseEngine = DatabaseEngineMySQL
	}

	if err != nil {
		logger.Zap.Fatalf("Error to open database connection: %v", err)
	}

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

	db, err := gorm.Open(sqlite.Open(dbPath), gormConfig)
	if err != nil {
		return nil, err
	}

	// Enable foreign keys for SQLite
	db.Exec("PRAGMA foreign_keys = ON")

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
