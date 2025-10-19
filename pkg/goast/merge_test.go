package goast

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	srcFile = "data/src.go.code"
	genFile = "data/gen.go.code"
)

func TestMergeGoFile(t *testing.T) {
	t.Run("without cover same func", func(t *testing.T) {
		codeAst, err := MergeGoFile(srcFile, genFile)
		if err != nil {
			t.Error(err)
			return
		}
		assert.Equal(t, codeAst.changeCodeFlag, true)
		fmt.Println(codeAst.Code)
	})

	t.Run("cover same func", func(t *testing.T) {
		codeAst, err := MergeGoFile(srcFile, genFile, WithCoverSameFunc())
		if err != nil {
			t.Error(err)
			return
		}
		assert.Equal(t, codeAst.changeCodeFlag, true)
		fmt.Println(codeAst.Code)
	})

	t.Run("test same file", func(t *testing.T) {
		codeAst, err := MergeGoFile(srcFile, srcFile)
		if err != nil {
			t.Error(err)
			return
		}
		assert.Equal(t, codeAst.changeCodeFlag, false)
	})
}

func TestMergeGoCode(t *testing.T) {
	t.Run("without cover same func", func(t *testing.T) {
		srcData, genData, _ := getGoCode()
		codeAst, err := MergeGoCode(srcData, genData)
		if err != nil {
			t.Error(err)
			return
		}
		assert.Equal(t, codeAst.changeCodeFlag, true)
		fmt.Println(codeAst.Code)
	})

	t.Run("cover same func", func(t *testing.T) {
		srcData, genData, _ := getGoCode()
		codeAst, err := MergeGoCode(srcData, genData,
			WithCoverSameFunc(),
			WithIgnoreMergeFunc("GetByID", "Hi"))
		if err != nil {
			t.Error(err)
			return
		}
		assert.Equal(t, codeAst.changeCodeFlag, true)
		fmt.Println(codeAst.Code)
	})

	t.Run("test same file", func(t *testing.T) {
		srcData, _, _ := getGoCode()
		codeAst, err := MergeGoCode(srcData, srcData)
		if err != nil {
			t.Error(err)
			return
		}
		assert.Equal(t, codeAst.changeCodeFlag, false)
	})
}

func TestNewCodeAstFromData(t *testing.T) {
	srcData, _, err := getGoCode()
	if err != nil {
		t.Error(err)
		return
	}

	codeAst, err := NewCodeAstFromData(srcData)
	if err != nil {
		t.Error(err)
		return
	}
	codeAst.FilePath = srcFile
	assert.Equal(t, true, len(codeAst.AstInfos) > 0)
}

func getGoCode() ([]byte, []byte, error) {
	srcData, err := os.ReadFile(srcFile)
	if err != nil {
		return nil, nil, err

	}
	genData, err := os.ReadFile(genFile)
	if err != nil {
		return nil, nil, err
	}
	return srcData, genData, nil
}
