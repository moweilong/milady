package sqlitex

import (
	"fmt"
	glog "log"
	"os"
	"time"

	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"github.com/moweilong/milady/pkg/log"
	"github.com/moweilong/milady/pkg/store/dbclose"
)

// SQLiteOptions defines options for sqlite database.
type SQLiteOptions struct {
	MaxIdleConnections    int
	MaxOpenConnections    int
	MaxConnectionLifeTime time.Duration

	// +optional
	IsLog             bool
	Logger            logger.Interface
	SlowThreshold     time.Duration
	DisableForeignKey bool
	EnableTrace       bool
	RequestIDKey      string
	plugins           []gorm.Plugin
}

// DSN return DSN from SQLiteOptions.
func (o *SQLiteOptions) DSN(dbFile string) string {
	dsn := fmt.Sprintf("%s?_journal=WAL&_vacuum=incremental", dbFile)
	return dsn
}

// NewSQLite create a new gorm db instance with the given options.
func NewSQLite(dbFile string, opts *SQLiteOptions) (*gorm.DB, error) {
	setSQLiteDefaults(opts)

	db, err := gorm.Open(sqlite.Open(opts.DSN(dbFile)), gormConfig(opts))
	if err != nil {
		return nil, err
	}
	db.Set("gorm:auto_increment", true)

	// register trace plugin
	if opts.EnableTrace {
		err = db.Use(otelgorm.NewPlugin())
		if err != nil {
			return nil, fmt.Errorf("using gorm opentelemetry, err: %v", err)
		}
	}

	// register plugins
	for _, plugin := range opts.plugins {
		err = db.Use(plugin)
		if err != nil {
			return nil, err
		}
	}

	return db, nil
}

// setSQLiteDefaults set available default values for some fields.
func setSQLiteDefaults(opts *SQLiteOptions) {
	if opts.MaxIdleConnections == 0 {
		opts.MaxIdleConnections = 100
	}
	if opts.MaxOpenConnections == 0 {
		opts.MaxOpenConnections = 100
	}
	if opts.MaxConnectionLifeTime == 0 {
		opts.MaxConnectionLifeTime = time.Duration(10) * time.Second
	}
	if opts.Logger == nil {
		opts.Logger = log.Default()
	}
}

// gorm setting
func gormConfig(opts *SQLiteOptions) *gorm.Config {
	config := &gorm.Config{
		// disable foreign key constraints, not recommended for production environments
		DisableForeignKeyConstraintWhenMigrating: opts.DisableForeignKey,
		// removing the plural of an epithet
		NamingStrategy: schema.NamingStrategy{SingularTable: true},
	}

	// print SQL
	if opts.IsLog {
		if opts.Logger == nil {
			config.Logger = logger.Default.LogMode(logger.Info)
		} else {
			config.Logger = log.Default().LogMode(logger.Info)
		}
	} else {
		config.Logger = logger.Default.LogMode(logger.Silent)
	}

	// print only slow queries
	if opts.SlowThreshold > 0 {
		config.Logger = logger.New(
			glog.New(os.Stdout, "\r\n", glog.LstdFlags), // use the standard output asWriter
			logger.Config{
				SlowThreshold: opts.SlowThreshold,
				Colorful:      true,
				LogLevel:      logger.Warn, // set the logging level, only above the specified level will output the slow query log
			},
		)
	}

	return config
}

// Close close gorm db
func Close(db *gorm.DB) error {
	return dbclose.Close(db)
}
