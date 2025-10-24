package log_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/moweilong/milady/pkg/log"
)

func TestDefaultLogger(t *testing.T) {
	log.Infow("测试日志", "key", "value")
	err := fmt.Errorf("这是测试错误信息")
	log.Errorf("测试错误日志: %v", err)
}

// TestFileStorage 测试文件存储功能
func TestFileStorage(t *testing.T) {
	// 创建临时文件路径
	tempDir, err := os.MkdirTemp("", "log-test")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	logFile := filepath.Join(tempDir, "test.log")

	// 配置文件存储
	opts := &log.Options{
		Level:             "info",
		Format:            "json",
		DisableCaller:     false,
		DisableStacktrace: false,
		OutputPaths:       []string{"stdout"},
		EnableFileStorage: true,
		FileConfig: &log.FileOptions{
			Filename:   logFile,
			MaxSize:    1, // 1MB
			MaxBackups: 5,
			MaxAge:     1,
			Compress:   false,
			LocalTime:  true,
		},
	}

	// 初始化日志
	logger := log.NewLogger(opts)

	// 写入一些日志
	logger.Infow("文件存储测试", "test_key", "test_value")
	logger.Errorw(errors.New("test error"), "错误日志测试")

	// 刷新日志确保写入文件
	logger.Sync()

	// 验证文件创建和内容
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Errorf("日志文件未创建: %s", logFile)
	} else {
		content, err := os.ReadFile(logFile)
		if err != nil {
			t.Errorf("读取日志文件失败: %v", err)
		} else if !strings.Contains(string(content), "文件存储测试") {
			t.Errorf("日志文件内容不包含预期文本")
		}
	}
}

// TestLogFormats 测试不同的日志格式
func TestLogFormats(t *testing.T) {
	// 测试 JSON 格式
	jsonOpts := &log.Options{
		Level:       "info",
		Format:      "json",
		OutputPaths: []string{"stdout"},
	}

	jsonLogger := log.NewLogger(jsonOpts)
	jsonLogger.Infow("JSON格式测试", "key", "value")

	// 测试 Console 格式
	consoleOpts := &log.Options{
		Level:       "info",
		Format:      "console",
		OutputPaths: []string{"stdout"},
	}

	consoleLogger := log.NewLogger(consoleOpts)
	consoleLogger.Infow("Console格式测试", "key", "value")
}

// TestLogLevels 测试日志级别过滤
func TestLogLevels(t *testing.T) {
	// 测试不同级别的日志配置
	levels := []string{"debug", "info", "warn", "error"}

	for _, level := range levels {
		t.Run("Level_"+level, func(t *testing.T) {
			opts := &log.Options{
				Level:       level,
				Format:      "json",
				OutputPaths: []string{"stdout"},
			}

			logger := log.NewLogger(opts)

			// 写入各种级别的日志
			logger.Debugw("这是调试日志", "level", level)
			logger.Infow("这是信息日志", "level", level)
			logger.Warnw("这是警告日志", "level", level)
			logger.Errorw(errors.New("test error"), "这是错误日志", "level", level)
		})
	}
}

// TestMultiOutputPaths 测试多输出路径
func TestMultiOutputPaths(t *testing.T) {
	// 创建临时文件路径
	tempDir, err := os.MkdirTemp("", "log-multi-test")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	logFile1 := filepath.Join(tempDir, "log1.log")
	logFile2 := filepath.Join(tempDir, "log2.log")

	// 配置多输出路径
	opts := &log.Options{
		Level:       "info",
		Format:      "json",
		OutputPaths: []string{"stdout", logFile1, logFile2},
	}

	// 初始化日志
	logger := log.NewLogger(opts)

	// 写入日志
	logger.Infow("多输出路径测试", "test_key", "test_value")
	logger.Sync()

	// 验证两个文件都有日志内容
	for _, file := range []string{logFile1, logFile2} {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("日志文件未创建: %s", file)
		} else {
			content, err := os.ReadFile(file)
			if err != nil {
				t.Errorf("读取日志文件失败: %v", err)
			} else if !strings.Contains(string(content), "多输出路径测试") {
				t.Errorf("日志文件 %s 内容不包含预期文本", file)
			}
		}
	}
}

// TestCallerAndStacktrace 测试调用者信息和堆栈跟踪
func TestCallerAndStacktrace(t *testing.T) {
	// 测试启用调用者信息
	enableCallerOpts := &log.Options{
		Level:             "info",
		Format:            "json",
		DisableCaller:     false,
		DisableStacktrace: false,
		OutputPaths:       []string{"stdout"},
	}

	enableCallerLogger := log.NewLogger(enableCallerOpts)
	enableCallerLogger.Infow("启用调用者信息测试")

	// 测试禁用调用者信息
	disableCallerOpts := &log.Options{
		Level:             "info",
		Format:            "json",
		DisableCaller:     true,
		DisableStacktrace: false,
		OutputPaths:       []string{"stdout"},
	}

	disableCallerLogger := log.NewLogger(disableCallerOpts)
	disableCallerLogger.Infow("禁用调用者信息测试")
}

// TestContextExtractors 测试上下文提取器
func TestContextExtractors(t *testing.T) {
	// 创建上下文提取器
	type userIDKey struct{}
	type requestIDKey struct{}
	extractors := log.ContextExtractors{
		"userID": func(ctx context.Context) string {
			if v, ok := ctx.Value(userIDKey{}).(string); ok {
				return v
			}
			return ""
		},
		"requestID": func(ctx context.Context) string {
			if v, ok := ctx.Value(requestIDKey{}).(string); ok {
				return v
			}
			return ""
		},
	}

	// 配置日志
	opts := &log.Options{
		Level:       "info",
		Format:      "json",
		OutputPaths: []string{"stdout"},
	}

	// 创建带上下文提取器的日志记录器
	logger := log.NewLogger(opts, log.WithContextExtractor(extractors))

	// 创建上下文
	ctx := context.WithValue(context.Background(), userIDKey{}, "test-user-123")
	ctx = context.WithValue(ctx, requestIDKey{}, "req-456")

	// 使用带上下文的日志记录器
	logger.W(ctx).Infow("上下文提取器测试")

	// 测试嵌套的上下文日志
	subLogger := logger.W(ctx)
	subLogger.Infow("嵌套上下文日志测试")
}

// TestLogFileRotation 测试日志文件轮转
func TestLogFileRotation(t *testing.T) {
	// 创建临时文件路径
	tempDir, err := os.MkdirTemp("/Users/moweilong/Workspace/go/src/github.com/moweilong/milady/pkg/log", "log-rotation-test")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir) // 注释以避免在测试完成后删除临时目录, 方便查看日志文件

	logFile := filepath.Join(tempDir, "rotation.log")

	// 配置小文件大小以触发轮转
	opts := &log.Options{
		Level:             "debug",
		Format:            "json",
		DisableCaller:     false,
		DisableStacktrace: false,
		OutputPaths:       []string{"stdout"},
		EnableFileStorage: true,
		FileConfig: &log.FileOptions{
			Filename:   logFile,
			MaxSize:    1, // 1MB
			MaxBackups: 5,
			MaxAge:     1,
			Compress:   false,
			LocalTime:  true,
		},
	}

	// 初始化日志
	logger := log.NewLogger(opts)

	// 写入大量日志以触发轮转
	for i := 0; i < 10000; i++ {
		// 每个日志条目大约1KB，这样写100条大约100MB
		largeData := strings.Repeat("x", 1000) // 创建一个1KB的字符串
		logger.Infow("轮转测试日志", "index", i, "data", largeData)
	}

	logger.Sync()

	// 验证日志文件创建
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Errorf("日志文件未创建: %s", logFile)
	}
}
