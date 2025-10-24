# replacer 包

## 概述

replacer 是一个强大的文件内容替换库，支持替换本地目录和通过 embed 嵌入的目录中的文件。该包提供了灵活的文件替换功能，包括文本内容替换、文件名替换、目录名替换以及模板渲染功能。

## 核心功能

- 支持本地文件系统和 embed.FS 文件系统
- 灵活的文本内容替换，支持大小写敏感选项
- 精确的文件和目录过滤机制
- 自定义输出目录
- 文件内容和文件名支持 Go 模板渲染
- 跨平台路径分隔符处理
- 文件安全检查，避免覆盖源文件

## 接口与结构体

### Replacer 接口

```go
// Replacer interface. 一个文本替换器接口，用于替换文件中的文本。
type Replacer interface {
    SetReplacementFields(fields []Field)
    SetSubDirsAndFiles(subDirs []string, subFiles ...string)
    SetIgnoreSubDirs(dirs ...string)
    SetIgnoreSubFiles(filenames ...string)
    SetOutputDir(absDir string, name ...string) error
    GetOutputDir() string
    GetSourcePath() string
    SaveFiles() error
    ReadFile(filename string) ([]byte, error)
    GetFiles() []string
    SaveTemplateFiles(m map[string]interface{}, parentDir ...string) error
}
```

### Field 结构体

```go
// Field replace field information
type Field struct {
    Old             string // 旧字段
    New             string // 新字段
    IsCaseSensitive bool   // 首字母是否大小写敏感
}
```

## 使用方法

### 创建替换器

#### 1. 从本地目录创建

```go
replacer, err := New("path/to/template/dir")
if err != nil {
    // 处理错误
}
```

#### 2. 从 embed.FS 创建

```go
//go:embed template_dir
var fs embed.FS

replacer, err := NewFS("template_dir", fs)
if err != nil {
    // 处理错误
}
```

### 设置替换字段

```go
fields := []Field{
    {
        Old: "old_text",
        New: "new_text",
    },
    {
        Old:             "ServiceName",
        New:             "UserService",
        IsCaseSensitive: true, // 会自动生成首字母大小写的替换规则
    },
}
replacer.SetReplacementFields(fields)
```

### 设置文件过滤

```go
// 设置要处理的子目录和文件
replacer.SetSubDirsAndFiles([]string{"src/api", "src/model"}, "README.md")

// 设置要忽略的子目录
replacer.SetIgnoreSubDirs("vendor", "node_modules")

// 设置要忽略的文件
replacer.SetIgnoreSubFiles("*.tmp", "backup.txt")
```

### 设置输出目录

```go
// 指定绝对路径
err := replacer.SetOutputDir("/path/to/output")

// 或使用当前目录，自动生成带时间戳的目录名
err := replacer.SetOutputDir("", "generated", "code")
```

### 保存文件

#### 1. 普通文本替换

```go
err := replacer.SaveFiles()
if err != nil {
    // 处理错误
}
```

#### 2. 模板渲染

```go
data := map[string]interface{}{
    "ServiceName": "UserService",
    "Version":    "1.0.0",
    "Port":       8080,
}

// 可选指定父目录
err := replacer.SaveTemplateFiles(data, "output", "v1")
if err != nil {
    // 处理错误
}
```

### 其他方法

```go
// 获取源路径
fmt.Println(replacer.GetSourcePath())

// 获取输出路径
fmt.Println(replacer.GetOutputDir())

// 获取所有文件列表
files := replacer.GetFiles()

// 读取特定文件内容
content, err := replacer.ReadFile("config.yaml")
```

## 工作原理

### 1. 初始化阶段

- 扫描源目录，收集所有文件路径
- 根据源类型（本地文件系统或 embed.FS）设置相应的处理模式

### 2. 配置阶段

- 设置替换字段，对于大小写敏感的字段，会自动生成首字母大小写的变体
- 配置文件和目录过滤规则
- 设置输出目录

### 3. 执行阶段

#### 文本替换
- 读取源文件内容
- 应用替换规则到文件内容
- 替换文件名和目录名中的匹配文本
- 检查目标文件是否存在，避免覆盖
- 将替换后的内容写入到输出目录

#### 模板渲染
- 读取源文件内容
- 使用 Go 模板引擎渲染文件内容
- 渲染文件名中的模板变量
- 将渲染后的内容写入到输出目录

## 模板使用

replacer 支持在文件内容和文件名中使用 Go 模板语法：

### 模板文件示例

文件内容示例（config.yaml.tmpl）：
```yaml
server:
  name: {{.ServiceName}}
  version: {{.Version}}
  port: {{.Port}}
```

文件名示例：
```
{{.ServiceName}}.go.tmpl
```

## 注意事项

1. **路径分隔符**：包会自动处理不同操作系统的路径分隔符差异

2. **文件覆盖**：默认不会覆盖已存在的文件，会在 SaveFiles 或 SaveTemplateFiles 时返回错误

3. **大小写敏感**：当设置 IsCaseSensitive=true 时，会自动生成首字母大写和小写的替换规则

4. **模板文件扩展名**：SaveTemplateFiles 会自动移除 .tmpl、.tpl 和 .template 扩展名

5. **源文件保护**：不允许将替换后的文件写入源目录，以避免意外覆盖

## 示例

### 基本文本替换

```go
package main

import (
    "fmt"
    "log"
    "time"

    "github.com/moweilong/milady/pkg/replacer"
)

func main() {
    // 创建替换器
    r, err := replacer.New("templates")
    if err != nil {
        log.Fatal(err)
    }

    // 设置替换字段
    fields := []replacer.Field{
        {Old: "APP_NAME", New: "MyAwesomeApp"},
        {Old: "VERSION", New: "1.0.0"},
    }
    r.SetReplacementFields(fields)

    // 设置输出目录
    outDir := fmt.Sprintf("./output_%s", time.Now().Format("20060102"))
    err = r.SetOutputDir(outDir)
    if err != nil {
        log.Fatal(err)
    }

    // 保存文件
    err = r.SaveFiles()
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("替换完成！输出目录: %s\n", r.GetOutputDir())
}
```

### 模板渲染

```go
package main

import (
    "embed"
    "log"

    "github.com/moweilong/milady/pkg/replacer"
)

//go:embed templates
var fs embed.FS

func main() {
    // 从 embed.FS 创建替换器
    r, err := replacer.NewFS("templates", fs)
    if err != nil {
        log.Fatal(err)
    }

    // 设置输出目录
    err = r.SetOutputDir("./generated")
    if err != nil {
        log.Fatal(err)
    }

    // 模板数据
    data := map[string]interface{}{
        "App": map[string]interface{}{
            "Name":    "UserService",
            "Version": "2.0.0",
        },
        "Database": map[string]interface{}{
            "Host": "localhost",
            "Port": 3306,
        },
    }

    // 渲染并保存模板
    err = r.SaveTemplateFiles(data)
    if err != nil {
        log.Fatal(err)
    }

    log.Println("模板渲染完成！")
}
```

## 总结

replacer 包提供了强大而灵活的文件替换和模板渲染功能，适用于代码生成、配置文件生成、模板化项目创建等场景。通过简单的 API，用户可以轻松实现复杂的文件内容和名称替换需求。