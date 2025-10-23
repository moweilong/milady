package log_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/moweilong/milady/pkg/log"
)

func TestLogger(t *testing.T) {
	// 自定义日志配置
	opts := &log.Options{
		Level:             "debug",            // 设置日志级别为 debug
		Format:            "json",             // 设置日志格式为 JSON
		DisableCaller:     false,              // 显示调用日志的文件和行号
		DisableStacktrace: false,              // 允许打印堆栈信息
		OutputPaths:       []string{"stdout"}, // 将日志输出到标准输出
	}

	// 初始化全局日志对象
	log.Init(opts)

	// 测试不同级别的日志输出
	err := errors.New("something went wrong")
	userID := "user123"
	timestamp := time.Now()

	// =================== 测试结构化日志 (w系列方法) ===================
	log.Debugw("用户登录调试信息", "userID", userID, "ip", "192.168.1.1", "attempt", 1)
	log.Infow("用户登录成功", "userID", userID, "timestamp", timestamp)
	log.Warnw("登录失败警告", "userID", userID, "reason", "密码错误", "attempt", 3)
	log.Errorw(err, "处理用户请求失败", "userID", userID, "path", "/api/user/profile")

	// =================== 测试格式化日志 (f系列方法) ===================
	log.Infow("============================================")
	log.Debugf("调试信息: 用户 %s 访问了页面 %s", userID, "/home")
	log.Infof("信息: 处理请求耗时 %d ms", 128)
	log.Warnf("警告: 用户 %s 尝试访问未授权资源", userID)
	log.Errorf("错误: %s, 错误详情: %v", "数据库连接失败", err)

	// =================== 测试带上下文的日志 ===================
	log.Infow("============================================")
	ctx := context.WithValue(context.Background(), "requestID", "req-12345")
	// 添加上下文提取器
	log.Init(opts, log.WithContextExtractor(log.ContextExtractors{
		"requestID": func(ctx context.Context) string {
			if v, ok := ctx.Value("requestID").(string); ok {
				return v
			}
			return ""
		},
	}))
	
	// 使用带上下文的日志记录器
	log.W(ctx).Infow("带上下文的日志", "userID", userID)
	log.W(ctx).Infof("请求 %s 已处理完成", userID)

	// =================== 测试调用层级调整 ===================
	log.Infow("============================================")
	// 添加调用层级跳过，适用于在封装的日志函数中使用
	log.AddCallerSkip(1).Infow("调整后的调用位置", "key", "value")

	// 注意：Panicw 和 Fatalw 会中断程序运行，因此在测试中应小心使用。
	// 可以注释掉以下两行进行测试，或者在单独的环境中运行。
	// log.Panicw("这是一个恐慌消息", "reason", "意外情况")
	// log.Fatalw("这是一个致命消息", "reason", "严重故障")

	// 确保日志缓冲区被刷新
	log.Sync()
}
