package mysqlx

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInitMysql(t *testing.T) {
	dsn := "root:123456@(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True&loc=Local"
	opts := &MySQLOptions{
		Dsn:         dsn,
		EnableTrace: true,
	}
	setMySQLDefaults(opts)
	db, err := NewMySQL(opts)
	if err != nil {
		// ignore test error about not being able to connect to real mysql
		t.Logf("connect to mysql failed, err=%v, dsn=%s", err, dsn)
		return
	}
	defer Close(db)

	t.Logf("%+v", db.Name())
}

func Test_gormConfig(t *testing.T) {
	opts := &MySQLOptions{
		SlowThreshold:         time.Millisecond * 100,
		EnableTrace:           true,
		MaxIdleConnections:    5,
		MaxOpenConnections:    50,
		MaxConnectionLifeTime: time.Minute * 3,
		DisableForeignKey:     true,
	}
	setMySQLDefaults(opts)

	c := gormConfig(opts)
	assert.NotNil(t, c)

	err := rwSeparationPlugin(opts)
	assert.NotNil(t, err)
}
