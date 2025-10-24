package mysqlx

import (
	"database/sql"
	"fmt"
	glog "log"
	"os"
	"time"

	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"gorm.io/plugin/dbresolver"

	"github.com/moweilong/milady/pkg/log"
	"github.com/moweilong/milady/pkg/store/dbclose"
	"github.com/moweilong/milady/pkg/utils"
)

// MySQLOptions defines options for mysql database.
type MySQLOptions struct {
	MaxIdleConnections    int
	MaxOpenConnections    int
	MaxConnectionLifeTime time.Duration
	SlowThreshold         time.Duration
	DisableForeignKey     bool
	EnableTrace           bool
	Dsn                   string
	SlavesDsn             []string
	MastersDsn            []string
	Plugins               []gorm.Plugin
	// +optional
	IsLog  bool
	Logger logger.Interface
}

// NewMySQL create a new gorm db instance with the given options.
func NewMySQL(opts *MySQLOptions) (*gorm.DB, error) {
	// Set default values to ensure all fields in opts are available.
	setMySQLDefaults(opts)

	db, err := gorm.Open(mysql.Open(opts.Dsn), gormConfig(opts))
	if err != nil {
		return nil, err
	}

	db.Set("gorm:table_options", "CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci") // automatic appending of table suffixes when creating tables

	// register trace plugin
	if opts.EnableTrace {
		err = db.Use(otelgorm.NewPlugin())
		if err != nil {
			return nil, fmt.Errorf("using gorm opentelemetry, err: %v", err)
		}
	}

	// register read-write separation plugin
	if len(opts.SlavesDsn) > 0 {
		err = db.Use(rwSeparationPlugin(opts))
		if err != nil {
			return nil, err
		}
	}

	// register plugins
	for _, plugin := range opts.Plugins {
		err = db.Use(plugin)
		if err != nil {
			return nil, err
		}
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// SetMaxOpenConns sets the maximum number of open connections to the database.
	sqlDB.SetMaxOpenConns(opts.MaxOpenConnections)

	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	sqlDB.SetConnMaxLifetime(opts.MaxConnectionLifeTime)

	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	sqlDB.SetMaxIdleConns(opts.MaxIdleConnections)

	return db, nil
}

// setMySQLDefaults set available default values for some fields.
func setMySQLDefaults(opts *MySQLOptions) {
	if opts.Dsn == "" {
		opts.Dsn = "root:123456@tcp(127.0.0.1:3306)/test?charset=utf8&parseTime=%t&loc=%s"
	}
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
		opts.Logger = logger.Default
	}
}

// gorm setting
func gormConfig(o *MySQLOptions) *gorm.Config {
	config := &gorm.Config{
		// disable foreign key constraints, not recommended for production environments
		DisableForeignKeyConstraintWhenMigrating: o.DisableForeignKey,
		// removing the plural of an epithet
		NamingStrategy: schema.NamingStrategy{SingularTable: true},
	}

	// print SQL
	if o.IsLog {
		if o.Logger == nil {
			config.Logger = logger.Default.LogMode(logger.Info)
		} else {
			config.Logger = log.Default().LogMode(logger.Info)
		}
	} else {
		config.Logger = logger.Default.LogMode(logger.Silent)
	}

	// print only slow queries
	if o.SlowThreshold > 0 {
		config.Logger = logger.New(
			glog.New(os.Stdout, "\r\n", glog.LstdFlags), // use the standard output asWriter
			logger.Config{
				SlowThreshold: o.SlowThreshold,
				Colorful:      true,
				LogLevel:      logger.Warn, // set the logging level, only above the specified level will output the slow query log
			},
		)
	}

	return config
}

func MustRawDB(db *gorm.DB) *sql.DB {
	raw, err := db.DB()
	if err != nil {
		panic(err)
	}
	return raw
}

func rwSeparationPlugin(o *MySQLOptions) gorm.Plugin {
	slaves := []gorm.Dialector{}
	for _, dsn := range o.SlavesDsn {
		slaves = append(slaves, mysql.New(mysql.Config{
			DSN: utils.AdaptiveMysqlDsn(dsn),
		}))
	}

	masters := []gorm.Dialector{}
	for _, dsn := range o.MastersDsn {
		masters = append(masters, mysql.New(mysql.Config{
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
