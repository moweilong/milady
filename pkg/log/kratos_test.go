package log

import (
	"testing"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/assert"
)

// 测试zapLogger是否正确实现了KratosLogger接口
func TestKratosLoggerImplementation(t *testing.T) {
	// 使用默认logger实例进行接口实现验证
	logger := Default()

	// 验证logger是否实现了kratos log.Logger接口
	var _ log.Logger = logger
}

// 测试Log方法对不同日志级别的处理
func TestLogMethodLevelHandling(t *testing.T) {
	// 创建基本logger
	logger := Default()

	// 测试各种日志级别都能正常调用而不报错
	levels := []log.Level{
		log.LevelDebug,
		log.LevelInfo,
		log.LevelWarn,
		log.LevelError,
		// 注意：不测试LevelFatal，因为它会终止进程
	}

	for _, level := range levels {
		err := logger.Log(level, "test_key", "test_value")
		assert.Nil(t, err, "日志级别 %v 不应返回错误", level)
	}
}

// 测试Log方法对键值对参数的处理
func TestLogMethodKeyValuePairs(t *testing.T) {
	logger := Default()

	// 测试正常的键值对参数
	err := logger.Log(log.LevelInfo, "key1", "value1", "key2", 123)
	assert.Nil(t, err, "正确的键值对格式不应返回错误")

	// 测试奇数个参数 - 应该发出警告但不返回错误
	err = logger.Log(log.LevelInfo, "key1", "value1", "key2")
	assert.Nil(t, err, "奇数个参数不应返回错误，但应发出警告")

	// 测试空参数 - 应该发出警告但不返回错误
	err = logger.Log(log.LevelInfo)
	assert.Nil(t, err, "空参数不应返回错误，但应发出警告")
}

// 测试在各种场景下的日志记录
func TestLogMethodVariousScenarios(t *testing.T) {
	logger := Default()

	// 测试不同类型的值
	testCases := []struct {
		name      string
		level     log.Level
		keyvals   []interface{}
		expectErr bool
	}{
		{
			name:      "字符串键值对",
			level:     log.LevelInfo,
			keyvals:   []interface{}{"message", "测试消息", "user", "admin"},
			expectErr: false,
		},
		{
			name:      "数字值",
			level:     log.LevelDebug,
			keyvals:   []interface{}{"count", 100, "success_rate", 0.95},
			expectErr: false,
		},
		{
			name:      "布尔值",
			level:     log.LevelWarn,
			keyvals:   []interface{}{"critical", true, "enabled", false},
			expectErr: false,
		},
		{
			name:      "错误日志",
			level:     log.LevelError,
			keyvals:   []interface{}{"error", "处理失败", "code", 500},
			expectErr: false,
		},
		{
			name:      "奇数参数",
			level:     log.LevelInfo,
			keyvals:   []interface{}{"key1", "value1", "key2"},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := logger.Log(tc.level, tc.keyvals...)
			if tc.expectErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

// 测试日志接口在实际应用场景中的集成使用
func TestKratosLoggerPracticalUsage(t *testing.T) {
	logger := Default()

	// 模拟HTTP请求日志
	httpLog := func() {
		err := logger.Log(log.LevelInfo,
			"method", "GET",
			"path", "/api/users",
			"status", 200,
			"latency", "123ms",
			"client_ip", "192.168.1.1",
		)
		assert.Nil(t, err)
	}

	// 模拟数据库操作日志
	dbLog := func() {
		err := logger.Log(log.LevelDebug,
			"operation", "SELECT",
			"table", "users",
			"conditions", "id = 123",
			"rows_affected", 1,
			"duration", "5ms",
		)
		assert.Nil(t, err)
	}

	// 模拟错误处理日志
	errorLog := func() {
		err := logger.Log(log.LevelError,
			"component", "auth",
			"error", "authentication failed",
			"user_id", "12345",
			"attempts", 3,
		)
		assert.Nil(t, err)
	}

	// 执行各种日志记录场景
	httpLog()
	dbLog()
	errorLog()
}
