package goast

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// This is a demo function for testing, default panic message is "implement me"
func demoFn1() {
	panic("implement me")
}

// This is a demo function for testing, default panic message is "ai todo"
func demoFn2() {
	panic("implement me")
}

// This is a demo function for testing, default panic message is "foobar"
func demoFn3() {
	panic("implement me")
}

func TestFilterFuncCodeByFile(t *testing.T) {
	code, infos, err := FilterFuncCodeByFile("filter_test.go")
	assert.NoError(t, err)
	assert.NotNil(t, code)
	assert.Equal(t, 3, len(infos))
	assert.Equal(t, "demoFn1", infos[0].Name)
	assert.Contains(t, infos[0].ExtractComment(), `"implement me"`)
	//t.Log(code)

	code, infos, err = FilterFuncCodeByFile("filter_test.go", "ai todo", "foobar")
	assert.NoError(t, err)
	assert.NotNil(t, code)
	assert.Equal(t, 3, len(infos))
	assert.Equal(t, "demoFn3", infos[2].Name)
	assert.Contains(t, infos[2].ExtractComment(), `"foobar"`)
	//t.Log(code)
}
