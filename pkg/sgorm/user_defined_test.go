package sgorm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBool(t *testing.T) {
	var v1 Bool
	assert.NoError(t, v1.Scan(nil))
	assert.Equal(t, false, bool(v1))

	// mysql
	assert.NoError(t, v1.Scan([]byte{0}))
	assert.Equal(t, false, bool(v1))
	assert.NoError(t, v1.Scan([]byte{1}))
	assert.Equal(t, true, bool(v1))

	// postgres
	assert.NoError(t, v1.Scan(false))
	assert.Equal(t, false, bool(v1))
	assert.NoError(t, v1.Scan(true))
	assert.Equal(t, true, bool(v1))

	// mysql
	v2, err := Bool(true).Value()
	assert.NoError(t, err)
	assert.Equal(t, []byte{1}, v2)
	v2, err = Bool(false).Value()
	assert.NoError(t, err)
	assert.Equal(t, []byte{0}, v2)

	SetDriver("postgres")

	// postgres
	v3, err := Bool(true).Value()
	assert.NoError(t, err)
	assert.Equal(t, true, v3)
	v3, err = Bool(false).Value()
	assert.NoError(t, err)
	assert.Equal(t, false, v3)
}

func TestTinyBool(t *testing.T) {
	var v1 TinyBool
	assert.NoError(t, v1.Scan(nil))
	assert.Equal(t, false, bool(v1))
	assert.NoError(t, v1.Scan(0))
	assert.Equal(t, false, bool(v1))
	assert.NoError(t, v1.Scan(1))
	assert.Equal(t, true, bool(v1))

	v2, err := TinyBool(true).Value()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), v2)

	v2, err = TinyBool(false).Value()
	assert.NoError(t, err)
	assert.Equal(t, int64(0), v2)
}
