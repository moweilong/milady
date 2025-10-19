package routers

import (
	"context"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/moweilong/milady/pkg/utils"

	"github.com/moweilong/milady/configs"
	"github.com/moweilong/milady/internal/config"
)

func TestNewRouter(t *testing.T) {
	err := config.Init(configs.Path("serverNameExample.yml"))
	if err != nil {
		t.Fatal(err)
	}

	config.Get().App.EnableMetrics = false
	config.Get().App.EnableTrace = true
	config.Get().App.EnableHTTPProfile = true
	config.Get().App.EnableLimit = true
	config.Get().App.EnableCircuitBreaker = true

	utils.SafeRunWithTimeout(time.Second*2, func(cancel context.CancelFunc) {
		gin.SetMode(gin.ReleaseMode)
		r := NewRouter()
		assert.NotNil(t, r)
		cancel()
	})
}

func TestNewRouter2(t *testing.T) {
	err := config.Init(configs.Path("serverNameExample.yml"))
	if err != nil {
		t.Fatal(err)
	}

	config.Get().App.EnableMetrics = true

	utils.SafeRunWithTimeout(time.Second*2, func(cancel context.CancelFunc) {
		gin.SetMode(gin.ReleaseMode)
		r := NewRouter()
		assert.NotNil(t, r)
		cancel()
	})
}

type mock struct{}

func (u mock) Create(c *gin.Context)     { return }
func (u mock) DeleteByID(c *gin.Context) { return }
func (u mock) UpdateByID(c *gin.Context) { return }
func (u mock) GetByID(c *gin.Context)    { return }
func (u mock) List(c *gin.Context)       { return }

func Test_userExampleRouter(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	userExampleRouter(r.Group("/"), &mock{})
}
