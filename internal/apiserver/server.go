package apiserver

import (
	"context"
	"time"

	"github.com/moweilong/milady/pkg/authz"
	genericoptions "github.com/moweilong/milady/pkg/options"
	"github.com/moweilong/milady/pkg/server"
	"github.com/moweilong/milady/pkg/store/registry"
	"github.com/moweilong/milady/pkg/store/where"
	"github.com/moweilong/milady/pkg/token"
	"gorm.io/gorm"
	"k8s.io/klog/v2"

	"github.com/moweilong/milady/internal/apiserver/biz"
	"github.com/moweilong/milady/internal/apiserver/model"
	"github.com/moweilong/milady/internal/apiserver/pkg/validation"
	"github.com/moweilong/milady/internal/apiserver/store"
	"github.com/moweilong/milady/internal/pkg/contextx"
	"github.com/moweilong/milady/internal/pkg/known"
	mw "github.com/moweilong/milady/internal/pkg/middleware/gin"
)

// Config contains application-related configurations.
type Config struct {
	JWTKey       string
	Expiration   time.Duration
	TLSOptions   *genericoptions.TLSOptions
	HTTPOptions  *genericoptions.HTTPOptions
	MySQLOptions *genericoptions.MySQLOptions
}

// Server represents the web server.
type Server struct {
	cfg *ServerConfig
	srv server.Server
}

// ServerConfig contains the core dependencies and configurations of the server.
type ServerConfig struct {
	*Config
	biz       biz.IBiz
	val       *validation.Validator
	retriever mw.UserRetriever
	authz     *authz.Authz
}

// NewServer initializes and returns a new Server instance.
func (cfg *Config) NewServer(ctx context.Context) (*Server, error) {
	where.RegisterTenant("userID", func(ctx context.Context) string {
		return contextx.UserID(ctx)
	})

	// 初始化 token 包的签名密钥、认证 Key 及 Token 默认过期时间
	token.Init(cfg.JWTKey, known.XUserID, cfg.Expiration)
	// Create the core server instance.
	return NewServer(cfg)
}

// Run starts the server and listens for termination signals.
// It gracefully shuts down the server upon receiving a termination signal.
func (s *Server) Run(ctx context.Context) error {
	// Start serving in background.
	go s.srv.RunOrDie()

	// Block until the context is canceled or terminated.
	// The following code is used to perform some cleanup tasks when the server shuts down.
	<-ctx.Done()
	klog.InfoS("Shutting down server...")

	// Graceful stop server with timeout derived from ctx.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	s.srv.GracefulStop(ctx)

	klog.InfoS("Server exited successfully.")

	return nil
}

// NewDB creates and returns a *gorm.DB instance for MySQL.
func (cfg *Config) NewDB() (*gorm.DB, error) {
	klog.InfoS("Initializing database connection", "type", "mariadb")
	db, err := cfg.MySQLOptions.NewDB()
	if err != nil {
		klog.ErrorS(err, "Failed to create database connection")
		return nil, err
	}

	// Automatically migrate database schema
	if err := registry.Migrate(db); err != nil {
		klog.ErrorS(err, "Failed to migrate database schema")
		return nil, err
	}

	return db, nil
}

// UserRetriever 定义一个用户数据获取器. 用来获取用户信息.
type UserRetriever struct {
	store store.IStore
}

// GetUser 根据用户 ID 获取用户信息.
func (r *UserRetriever) GetUser(ctx context.Context, userID string) (*model.UserM, error) {
	return r.store.User().Get(ctx, where.F("userID", userID))
}

// ProvideDB provides a database instance based on the configuration.
func ProvideDB(cfg *Config) (*gorm.DB, error) {
	return cfg.NewDB()
}

func NewWebServer(serverConfig *ServerConfig) (server.Server, error) {
	return serverConfig.NewGinServer()
}
