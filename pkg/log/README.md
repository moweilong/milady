# 日志模块

log 是基于高性能日志库 zap 封装的日志组件，提供更便捷的日志管理功能。

## 配置说明

在 milady 创建的服务中，日志组件默认配置为：

日志功能：默认启用，日志级别为 debug
输出目标：标准终端输出
日志格式：console（可配置为 json 格式）
支持功能：日志文件存储、日志切割、日志保留时间设置，默认关闭

## 配置示例

在 configs 目录下的 yaml 配置文件中设置 log 字段，如下所示：

```yaml
# log 设置
# 日志配置
log:
  # 指定日志级别，可选值：debug, info, warn, error, dpanic, panic, fatal
  # 生产环境建议设置为 info
  level: debug
  # 指定日志显示格式，可选值：console, json
  # 生产环境建议设置为 json
  format: json
  # 是否开启颜色输出，默认值 true, 当 format 为 json 时, 该选项无效
  # 生产环境建议设置为 false
  enable-color: false
  # 是否开启 caller，如果开启会在日志中显示调用日志所在的文件和行号
  # 生产环境建议设置为 true
  disable-caller: true
  # 是否禁止在 panic 及以上级别打印堆栈信息
  # 生产环境建议设置为 true
  disable-stacktrace: true
  # 指定日志输出位置，多个输出，用 `逗号 + 空格` 分开。stdout：标准输出
  output-paths: [stdout]
  # 是否开启文件存储，默认值 true
  # 容器环境建议设置为 false
  enable-file-storage: true
  # 文件存储配置, 当 enable-file-storage 为 true 时有效
  file-config:
    filename: "out.log"    # 文件名称，默认值 out.log
    max-size: 10            # 最大文件大小(MB)，默认值10MB
    max-backups: 100         # 保留旧文件的最大个数，默认值100个
    max-age: 30             # 保留旧文件的最大天数，默认值30天
    is-compression: false    # 是否压缩/归档旧文件，默认值 false
    local-time: true         # 是否使用本地时间，默认值 true
```