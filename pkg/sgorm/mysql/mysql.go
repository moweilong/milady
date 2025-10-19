// Package mysql provides a gorm driver for mysql.
package mysql

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	mysqlDriver "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"gorm.io/plugin/dbresolver"

	"github.com/moweilong/milady/pkg/sgorm/dbclose"
	"github.com/moweilong/milady/pkg/sgorm/glog"
	"github.com/moweilong/milady/pkg/utils"
)

// Init mysql
func Init(dsn string, opts ...Option) (*gorm.DB, error) {
	o := defaultOptions()
	o.apply(opts...)

	sqlDB, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(o.maxIdleConns)       // set the maximum number of connections in the idle connection pool
	sqlDB.SetMaxOpenConns(o.maxOpenConns)       // set the maximum number of open database connections
	sqlDB.SetConnMaxLifetime(o.connMaxLifetime) // set the maximum time a connection can be reused

	db, err := gorm.Open(mysqlDriver.New(mysqlDriver.Config{Conn: sqlDB}), gormConfig(o))
	if err != nil {
		return nil, err
	}
	db.Set("gorm:table_options", "CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci") // automatic appending of table suffixes when creating tables

	// register trace plugin
	if o.enableTrace {
		err = db.Use(otelgorm.NewPlugin())
		if err != nil {
			return nil, fmt.Errorf("using gorm opentelemetry, err: %v", err)
		}
	}

	// register read-write separation plugin
	if len(o.slavesDsn) > 0 {
		err = db.Use(rwSeparationPlugin(o))
		if err != nil {
			return nil, err
		}
	}

	// register plugins
	for _, plugin := range o.plugins {
		err = db.Use(plugin)
		if err != nil {
			return nil, err
		}
	}

	return db, nil
}

// InitTidb init tidb
func InitTidb(dsn string, opts ...Option) (*gorm.DB, error) {
	return Init(dsn, opts...)
}

// gorm setting
func gormConfig(o *options) *gorm.Config {
	config := &gorm.Config{
		// disable foreign key constraints, not recommended for production environments
		DisableForeignKeyConstraintWhenMigrating: o.disableForeignKey,
		// removing the plural of an epithet
		NamingStrategy: schema.NamingStrategy{SingularTable: true},
	}

	// print SQL
	if o.isLog {
		if o.gLog == nil {
			config.Logger = logger.Default.LogMode(o.logLevel)
		} else {
			config.Logger = glog.NewCustomGormLogger(o.gLog, o.requestIDKey, o.logLevel)
		}
	} else {
		config.Logger = logger.Default.LogMode(logger.Silent)
	}

	// print only slow queries
	if o.slowThreshold > 0 {
		config.Logger = logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), // use the standard output asWriter
			logger.Config{
				SlowThreshold: o.slowThreshold,
				Colorful:      true,
				LogLevel:      logger.Warn, // set the logging level, only above the specified level will output the slow query log
			},
		)
	}

	return config
}

func rwSeparationPlugin(o *options) gorm.Plugin {
	slaves := []gorm.Dialector{}
	for _, dsn := range o.slavesDsn {
		slaves = append(slaves, mysqlDriver.New(mysqlDriver.Config{
			DSN: utils.AdaptiveMysqlDsn(dsn),
		}))
	}

	masters := []gorm.Dialector{}
	for _, dsn := range o.mastersDsn {
		masters = append(masters, mysqlDriver.New(mysqlDriver.Config{
			DSN: utils.AdaptiveMysqlDsn(dsn),
		}))
	}

	return dbresolver.Register(dbresolver.Config{
		Sources:  masters,
		Replicas: slaves,
		Policy:   dbresolver.RandomPolicy{},
	})
}

// Close close gorm db
func Close(db *gorm.DB) error {
	return dbclose.Close(db)
}
