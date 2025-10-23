package log

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// Example 展示如何使用文件存储功能
func TestExample(t *testing.T) {
	// 基本文件存储配置示例
	basicFileExample()

	// 高级文件存储配置示例
	highLevelFileExample()

	// 同时输出到控制台和文件的示例
	consoleAndFileExample()
}

// basicFileExample 基本文件存储配置示例
func basicFileExample() {
	// 创建默认的文件配置
	fileConfig := NewFileOptions()

	// 配置日志选项，启用文件存储
	opts := NewOptions()
	opts.EnableFileStorage = true // 启用文件存储
	opts.FileConfig = fileConfig

	// 初始化日志器
	Init(opts)

	// 记录日志，这些日志将写入到默认的 "out.log" 文件中
	Infof("这是一个基本的文件存储日志示例: %s", time.Now())
	Debugw("debug信息", "key1", "value1", "key2", 123)
	Warnw("警告信息", "reason", "测试")
	Errorw(fmt.Errorf("测试错误"), "错误信息", "code", 500)

	// 示例：禁用文件存储（即使配置了FileConfig，日志也不会写入文件）
	disableFileExample()
}

// highLevelFileExample 高级文件存储配置示例
func highLevelFileExample() {
	// 创建自定义的文件配置
	fileConfig := NewFileOptions()
	WithFileName("app.log")(fileConfig) // 自定义文件名
	WithFileMaxSize(50)(fileConfig)     // 文件最大50MB
	WithFileMaxBackups(50)(fileConfig)  // 保留最多50个旧文件
	WithFileMaxAge(7)(fileConfig)       // 文件保留7天
	WithFileCompress(true)(fileConfig)  // 压缩旧文件
	WithFileLocalTime(true)(fileConfig) // 使用本地时间

	// 配置日志选项 - 使用函数式选项设置
	opts := NewOptions()
	opts.Level = "debug"              // 设置日志级别为debug
	opts.Format = "json"              // 设置日志格式为json
	opts.FileConfig = fileConfig      // 设置文件配置
	WithEnableFileStorage(true)(opts) // 启用文件存储

	// 初始化日志器
	Init(opts)

	// 记录各类日志
	Debugf("这是一个JSON格式的调试日志: %d", 123)
	Infow("用户登录", "username", "admin", "ip", "127.0.0.1")
	Warnf("配置文件缺少某些可选字段: %s", "timeout")
	Errorw(fmt.Errorf("数据库连接失败"), "数据库错误", "database", "users", "retry", 3)
}

// consoleAndFileExample 同时输出到控制台和文件的示例
func consoleAndFileExample() {
	// 创建文件配置
	fileConfig := NewFileOptions()
	WithFileName("mixed.log")(fileConfig)
	WithFileMaxSize(20)(fileConfig)

	// 配置日志选项
	opts := NewOptions()
	opts.Level = "info"
	opts.Format = "console"
	opts.EnableColor = true               // 控制台启用彩色输出
	opts.OutputPaths = []string{"stdout"} // 输出到控制台
	opts.EnableFileStorage = true         // 启用文件存储
	opts.FileConfig = fileConfig          // 设置文件配置

	// 初始化日志器
	Init(opts)

	// 创建一个带有上下文的日志记录
	ctx := context.Background()

	// 使用上下文记录日志
	logWithCtx := W(ctx)
	logWithCtx.Infow("带上下文的日志", "action", "process", "result", "success")

	// 记录一些示例日志，它们将同时显示在控制台和写入文件
	Infof("这条日志将同时显示在控制台和写入到文件: %s", time.Now().Format("2006-01-02 15:04:05"))
	Warnw("系统警告", "component", "scheduler", "status", "degraded")
	Errorw(fmt.Errorf("网络超时"), "操作失败", "operation", "fetch_data", "url", "http://example.com")

	// 确保所有日志都被刷新到文件
	Sync()
}

// disableFileExample 展示如何禁用文件存储功能
func disableFileExample() {
	// 创建文件配置（即使配置了文件选项，如果EnableFileStorage为false，日志也不会写入文件）
	fileConfig := NewFileOptions()
	WithFileName("disabled.log")(fileConfig)

	// 配置日志选项，但禁用文件存储
	opts := NewOptions()
	opts.EnableFileStorage = false // 禁用文件存储
	opts.FileConfig = fileConfig   // 配置了文件选项，但不会生效

	// 初始化日志器
	Init(opts)

	// 这些日志只会输出到控制台，不会写入文件
	Infof("这条日志只会输出到控制台，不会写入文件: %s", time.Now().Format("2006-01-02 15:04:05"))
	Debugw("这是一个调试信息", "key", "value")
}
