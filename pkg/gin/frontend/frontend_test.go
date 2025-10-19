package frontend

import (
	"bytes"
	"embed"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestFrontEnd_LocalFile(t *testing.T) {
	r := gin.New()
	gin.SetMode(gin.ReleaseMode)
	err := New("dist").SetRouter(r)
	if err != nil {
		t.Error(err)
	}

	r = gin.New()
	gin.SetMode(gin.ReleaseMode)
	err = New("dist", With404ToHome()).SetRouter(r)
	if err != nil {
		t.Error(err)
	}
}

//go:embed README.md
var embedFS embed.FS

func TestFrontEnd_EmbedFS(t *testing.T) {
	var (
		configFile     = "config.js"
		modifyConfigFn = func(content []byte) []byte {
			return bytes.ReplaceAll(content, []byte("localhost"), []byte("192.168.3.37"))
		}
	)

	r := gin.New()
	gin.SetMode(gin.ReleaseMode)
	err := New("dist", WithEmbedFS(embedFS)).SetRouter(r)
	if err != nil {
		t.Error(err)
	}
	time.Sleep(time.Millisecond * 500)

	r = gin.New()
	gin.SetMode(gin.ReleaseMode)
	err = New("dist",
		WithEmbedFS(embedFS),
		WithHandleContent(modifyConfigFn, configFile),
		With404ToHome(),
	).SetRouter(r)
	if err != nil {
		t.Error(err)
	}
}
