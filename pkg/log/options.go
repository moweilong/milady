package log

import (
	"github.com/spf13/pflag"
	"go.uber.org/zap/zapcore"
)

// FileOptions 定义了日志文件存储的配置选项
type FileOptions struct {
	// Filename 日志文件名
	Filename string `json:"filename,omitempty" mapstructure:"filename"`
	// MaxSize 日志文件最大大小（MB）
	MaxSize int `json:"max-size,omitempty" mapstructure:"max-size"`
	// MaxBackups 保留的最大旧文件数量
	MaxBackups int `json:"max-backups,omitempty" mapstructure:"max-backups"`
	// MaxAge 保留的最大天数
	MaxAge int `json:"max-age,omitempty" mapstructure:"max-age"`
	// Compress 是否压缩旧文件
	Compress bool `json:"is-compression,omitempty" mapstructure:"is-compression"`
	// LocalTime 是否使用本地时间
	LocalTime bool `json:"local-time,omitempty" mapstructure:"local-time"`
}

// Options contains configuration options for logging.
type Options struct {
	// Level 日志级别. 默认值: debug. 可选值: debug, info, warn, error, dpanic, panic, and fatal.
	Level string `json:"level,omitempty" mapstructure:"level"`
	// Format specifies the log output format. Valid values are: console and json.
	Format string `json:"format,omitempty" mapstructure:"format"`
	// EnableColor 是否启用颜色输出, 当 Format 为 json 时, 该选项无效
	EnableColor bool `json:"enable-color"       mapstructure:"enable-color"`
	// DisableCaller specifies whether to include caller information in the log.
	DisableCaller bool `json:"disable-caller,omitempty" mapstructure:"disable-caller"`
	// DisableStacktrace specifies whether to record a stack trace for all messages at or above panic level.
	DisableStacktrace bool `json:"disable-stacktrace,omitempty" mapstructure:"disable-stacktrace"`
	// EnableFileStorage specifies whether to enable file storage.
	EnableFileStorage bool `json:"enable-file-storage,omitempty" mapstructure:"enable-file-storage"`
	// FileConfig 文件存储配置
	FileConfig *FileOptions `json:"file-config,omitempty" mapstructure:"file-config"`
	// OutputPaths specifies the output paths for the logs.
	OutputPaths []string `json:"output-paths,omitempty" mapstructure:"output-paths"`
}

// NewOptions creates a new Options object with default values.
func NewOptions() *Options {
	return &Options{
		Level:             zapcore.DebugLevel.String(),
		Format:            "console",
		EnableColor:       true,
		OutputPaths:       []string{"stdout"},
		EnableFileStorage: false,
		FileConfig:        nil,
	}
}

// Validate verifies flags passed to LogsOptions.
func (o *Options) Validate() []error {
	errs := []error{}

	return errs
}

// AddFlags adds command line flags for the configuration.
func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.Level, "log.level", o.Level, "Minimum log output `LEVEL`.")
	fs.BoolVar(&o.DisableCaller, "log.disable-caller", o.DisableCaller, "Disable output of caller information in the log.")
	fs.BoolVar(&o.DisableStacktrace, "log.disable-stacktrace", o.DisableStacktrace, ""+
		"Disable the log to record a stack trace for all messages at or above panic level.")
	fs.BoolVar(&o.EnableColor, "log.enable-color", o.EnableColor, "Enable output ansi colors in plain format logs.")
	fs.StringVar(&o.Format, "log.format", o.Format, "Log output `FORMAT`, support plain or json format.")
	fs.StringSliceVar(&o.OutputPaths, "log.output-paths", o.OutputPaths, "Output paths of log.")
	fs.BoolVar(&o.EnableFileStorage, "log.enable-file-storage", o.EnableFileStorage, "Enable log file storage.")

	// 文件存储相关配置
	if o.FileConfig != nil {
		fs.StringVar(&o.FileConfig.Filename, "log.file-config.filename", o.FileConfig.Filename, "Log file name.")
		fs.IntVar(&o.FileConfig.MaxSize, "log.file-config.max-size", o.FileConfig.MaxSize, "Maximum log file size in MB.")
		fs.IntVar(&o.FileConfig.MaxBackups, "log.file-config.max-backups", o.FileConfig.MaxBackups, "Maximum number of old log files to retain.")
		fs.IntVar(&o.FileConfig.MaxAge, "log.file-config.max-age", o.FileConfig.MaxAge, "Maximum number of days to retain old log files.")
		fs.BoolVar(&o.FileConfig.Compress, "log.file-config.is-compression", o.FileConfig.Compress, "Whether to compress old log files.")
		fs.BoolVar(&o.FileConfig.LocalTime, "log.file-config.local-time", o.FileConfig.LocalTime, "Whether to use local time for log file rotation.")
	}

}
