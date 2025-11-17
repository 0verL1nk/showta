package model

import (
	"fmt"
	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	syslog "log"
	"os"
	"path/filepath"
	"overlink.top/app/lib/util"
	"overlink.top/app/system/conf"
	"time"
)

var db *gorm.DB

func InitDb(cfg conf.Database) {
	newLogger := logger.New(
		syslog.New(os.Stdout, "\r\n", syslog.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second,   // Slow SQL threshold
			LogLevel:                  logger.Silent, // Log level
			IgnoreRecordNotFoundError: true,          // Ignore ErrRecordNotFound error for logger
			ParameterizedQueries:      true,          // Don't include params in the SQL log
			Colorful:                  false,         // Disable color
		},
	)

	var err error

	// Determine database type
	dbType := cfg.Type
	if dbType == "" {
		// Auto-detect based on dbname extension
		if filepath.Ext(cfg.Dbname) == ".db" {
			dbType = "sqlite"
		} else {
			dbType = "sqlite" // default to sqlite for backward compatibility
		}
	}

	switch dbType {
	case "mysql":
		// Create MySQL connection string
		dsn := cfg.User + ":" + cfg.Password + "@tcp(" + cfg.Host + ":" + fmt.Sprintf("%d", cfg.Port) + ")/" + cfg.Dbname + "?charset=utf8mb4&parseTime=True&loc=Local"

		// Add TLS options if enabled
		if cfg.TLS {
			if cfg.TLSSkipVerify {
				dsn += "&tls=skip-verify"
			} else if cfg.TLSCAFile != "" || cfg.TLSCertFile != "" || cfg.TLSKeyFile != "" {
				// Custom TLS configuration would be needed here
				dsn += "&tls=true"
			} else {
				dsn += "&tls=true"
			}
		}

		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: newLogger})
		if err != nil {
			panic("failed to connect to MySQL database: " + err.Error())
		}

		// Configure MySQL connection pooling
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.SetMaxIdleConns(10)
			sqlDB.SetMaxOpenConns(100)
			sqlDB.SetConnMaxLifetime(time.Hour)
		}

	case "sqlite":
		fallthrough
	default:
		fullDbName := conf.AbsPath(cfg.Dbname)
		checkDbDir(fullDbName)

		db, err = gorm.Open(sqlite.Open(fullDbName), &gorm.Config{Logger: newLogger})
		if err != nil {
			panic("failed to connect to SQLite database: " + err.Error())
		}

		db.Exec("PRAGMA journal_mode=WAL;")
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.SetMaxIdleConns(10)
			sqlDB.SetMaxOpenConns(100)
			sqlDB.SetConnMaxLifetime(time.Hour)
		}
	}

	// Migrate the schema
	db.AutoMigrate(&User{}, &Storage{}, &FolderSetting{}, &Preference{})
}

func checkDbDir(pathStr string) {
	dirName, _ := filepath.Split(pathStr)
	if dirName == "" {
		return
	}

	ok, _ := util.PathExist(dirName)
	if ok {
		return
	}

	os.MkdirAll(dirName, os.ModePerm)
}
