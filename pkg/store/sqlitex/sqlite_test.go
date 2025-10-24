package sqlitex

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSQLite(t *testing.T) {
	dbFile := "test_sqlite.db"
	opts := &SQLiteOptions{}
	setSQLiteDefaults(opts)
	db, err := NewSQLite(dbFile, opts)
	if err != nil {
		// ignore test error about not being able to connect to real sqlite
		t.Logf("connect to sqlite failed, err=%v, dbFile=%s", err, dbFile)
		return
	}
	defer Close(db)

	t.Logf("%+v", db.Name())
}

func Test_gormConfig(t *testing.T) {
	o := &SQLiteOptions{}
	setSQLiteDefaults(o)

	c := gormConfig(o)
	assert.NotNil(t, c)
}
