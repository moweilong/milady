// nolint: err113
package options

import (
	"errors"
	"time"

	genericoptions "github.com/moweilong/milady/pkg/options"
	"github.com/spf13/pflag"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"

	"github.com/moweilong/milady/internal/apiserver"
)

// ServerOptions contains the configuration options for the server.
type ServerOptions struct {
	// JWTKey 定义 JWT 密钥.
	JWTKey string `json:"jwt-key" mapstructure:"jwt-key"`
	// Expiration 定义 JWT Token 的过期时间.
	Expiration time.Duration `json:"expiration" mapstructure:"expiration"`
	// TLSOptions contains the TLS configuration options.
	TLSOptions *genericoptions.TLSOptions `json:"tls" mapstructure:"tls"`
	// HTTPOptions contains the HTTP configuration options.
	HTTPOptions *genericoptions.HTTPOptions `json:"http" mapstructure:"http"`
	// MySQLOptions contains the MySQL configuration options.
	MySQLOptions *genericoptions.MySQLOptions `json:"mysql" mapstructure:"mysql"`
	// LogOptions used to specify the log options.
	LogOptions *genericoptions.LogOptions
}

// NewServerOptions creates a ServerOptions instance with default values.
func NewServerOptions() *ServerOptions {
	opts := &ServerOptions{
		JWTKey:      "",
		Expiration:  2 * time.Hour,
		TLSOptions:  genericoptions.NewTLSOptions(),
		HTTPOptions: genericoptions.NewHTTPOptions(),

		MySQLOptions: genericoptions.NewMySQLOptions(),
		LogOptions:   genericoptions.NewLogOptions(),
	}
	opts.HTTPOptions.Addr = ":5555"

	return opts
}

// AddFlags binds the options in ServerOptions to command-line flags.
func (o *ServerOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.JWTKey, "jwt-key", o.JWTKey, "JWT signing key. Must be at least 6 characters long.")
	// 绑定 JWT Token 的过期时间选项到命令行标志。
	// 参数名称为 `--expiration`，默认值为 o.Expiration
	fs.DurationVar(&o.Expiration, "expiration", o.Expiration, "The expiration duration of JWT tokens.")
	// Add command-line flags for sub-options.
	o.TLSOptions.AddFlags(fs)
	o.HTTPOptions.AddFlags(fs)
	o.MySQLOptions.AddFlags(fs)
	o.LogOptions.AddFlags(fs)
}

// Complete completes all the required options.
func (o *ServerOptions) Complete() error {
	// TODO: Add the completion logic if needed.
	return nil
}

// Validate checks whether the options in ServerOptions are valid.
func (o *ServerOptions) Validate() error {
	errs := []error{}
	// 校验 JWTKey 长度
	if len(o.JWTKey) < 6 {
		errs = append(errs, errors.New("JWTKey must be at least 6 characters long"))
	}

	// Validate sub-options.
	errs = append(errs, o.TLSOptions.Validate()...)
	errs = append(errs, o.HTTPOptions.Validate()...)

	errs = append(errs, o.MySQLOptions.Validate()...)
	errs = append(errs, o.LogOptions.Validate()...)

	// Aggregate all errors and return them.
	return utilerrors.NewAggregate(errs)
}

// Config builds an apiserver.Config based on ServerOptions.
func (o *ServerOptions) Config() (*apiserver.Config, error) {
	return &apiserver.Config{
		JWTKey:       o.JWTKey,
		Expiration:   o.Expiration,
		TLSOptions:   o.TLSOptions,
		HTTPOptions:  o.HTTPOptions,
		MySQLOptions: o.MySQLOptions,
	}, nil
}
