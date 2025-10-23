# internal/database/mysql.go 文件生成指南

## 概述

`internal/database/mysql.go` 文件是通过 `sponge` 工具的 `patch gen-db-init` 命令生成的，用于初始化 MySQL 数据库连接。本文档详细分析了该文件的生成流程、核心代码结构以及使用方法。

## 生成流程分析

### 1. 命令行入口

`gen-db-init.go` 中的 `GenerateDBInitCommand` 函数定义了生成数据库初始化代码的命令行接口：

```go
func GenerateDBInitCommand() *cobra.Command {
    var (
        moduleName string // go.mod 模块名称
        dbDriver   string // 数据库驱动，例如 mysql, mongodb, postgresql, sqlite
        outPath    string // 输出目录
        targetFile = "internal/database/init.go"
    )
    
    // 命令定义和参数配置...
}
```
<mcfile name="gen-db-init.go" path="/home/murphy/workspace/golang/src/github.com/moweilong/sponge/cmd/sponge/commands/patch/gen-db-init.go"></mcfile>

### 2. 生成器初始化

当用户执行命令后，系统会创建 `dbInitGenerator` 实例并调用其 `generateCode` 方法：

```go
g := &dbInitGenerator{
    moduleName: moduleName,
    dbDriver:   dbDriver,
    outPath:    outPath,
    serverName:     serverName,
    suitedMonoRepo: suitedMonoRepo,
}
outPath, err = g.generateCode()
```
<mcfile name="gen-db-init.go" path="/home/murphy/workspace/golang/src/github.com/moweilong/sponge/cmd/sponge/commands/patch/gen-db-init.go"></mcfile>

### 3. 文件选择逻辑

在 `generateCode` 方法中，系统会根据指定的数据库驱动类型选择要生成的文件。对于 MySQL 驱动，会通过 `SetSelectFiles` 函数选择以下文件：

```go
func SetSelectFiles(dbDriver string, selectFiles map[string][]string) error {
    dbDriver = strings.ToLower(dbDriver)
    switch dbDriver {
    case DBDriverMysql, DBDriverTidb:
        selectFiles["internal/database"] = []string{"init.go", "redis.go", "mysql.go"}
    // 其他数据库驱动的处理...
    }
    return nil
}
```
<mcfile name="common.go" path="/home/murphy/workspace/golang/src/github.com/moweilong/sponge/cmd/sponge/commands/generate/common.go"></mcfile>

### 4. 代码生成与模板替换

`dbInitGenerator.generateCode` 方法会设置模板信息并进行代码替换：

```go
func (g *dbInitGenerator) generateCode() (string, error) {
    subTplName := "init_" + g.dbDriver
    r := generate.Replacers[generate.TplNameSponge]
    if r == nil {
        return "", errors.New("replacer is nil")
    }

    subDirs := []string{}
    selectFiles := map[string][]string{}
    err := generate.SetSelectFiles(g.dbDriver, selectFiles)
    if err != nil {
        return "", err
    }
    
    r.SetSubDirsAndFiles(subDirs, getSubFiles(selectFiles)...)
    _ = r.SetOutputDir(g.outPath, subTplName)
    fields := g.addFields(r)
    r.SetReplacementFields(fields)
    if err := r.SaveFiles(); err != nil {
        return "", err
    }

    return r.GetOutputDir(), nil
}
```
<mcfile name="gen-db-init.go" path="/home/murphy/workspace/golang/src/github.com/moweilong/sponge/cmd/sponge/commands/patch/gen-db-init.go"></mcfile>

### 5. 模块名称替换

在 `addFields` 方法中，系统会将模板中的模块名称替换为用户指定的模块名称：

```go
func (g *dbInitGenerator) addFields(r replacer.Replacer) []replacer.Field {
    var fields []replacer.Field

    fields = append(fields, generate.DeleteCodeMark(r, generate.ModelInitDBFile, generate.StartMark, generate.EndMark)...)    
    fields = append(fields, []replacer.Field{
        {
            Old:             "github.com/go-dev-frame/sponge/internal",
            New:             g.moduleName + "/internal",
            IsCaseSensitive: false,
        },
        {
            Old:             "github.com/go-dev-frame/sponge/configs",
            New:             g.moduleName + "/configs",
            IsCaseSensitive: false,
        },
        // 其他替换字段...
    }...)    
    
    return fields
}
```
<mcfile name="gen-db-init.go" path="/home/murphy/workspace/golang/src/github.com/moweilong/sponge/cmd/sponge/commands/patch/gen-db-init.go"></mcfile>

## mysql.go 文件结构与功能

生成的 `mysql.go` 文件主要包含以下内容：

```go
package database

import (
    "time"

    "github.com/go-dev-frame/sponge/pkg/logger"
    "github.com/go-dev-frame/sponge/pkg/sgorm"
    "github.com/go-dev-frame/sponge/pkg/sgorm/mysql"
    "github.com/go-dev-frame/sponge/pkg/utils"

    "yourModuleName/internal/config"
)

// InitMysql 连接 MySQL 数据库
func InitMysql() *sgorm.DB {
    mysqlCfg := config.Get().Database.Mysql
    opts := []mysql.Option{
        mysql.WithMaxIdleConns(mysqlCfg.MaxIdleConns),
        mysql.WithMaxOpenConns(mysqlCfg.MaxOpenConns),
        mysql.WithConnMaxLifetime(time.Duration(mysqlCfg.ConnMaxLifetime) * time.Minute),
    }
    if mysqlCfg.EnableLog {
        opts = append(opts,
            mysql.WithLogging(logger.Get()),
            mysql.WithLogRequestIDKey("request_id"),
        )
    }

    if config.Get().App.EnableTrace {
        opts = append(opts, mysql.WithEnableTrace())
    }

    // 可以根据需要启用读写分离
    // opts = append(opts, mysql.WithRWSeparation(
    //     mysqlCfg.SlavesDsn,
    //     mysqlCfg.MastersDsn...,
    // ))

    dsn := utils.AdaptiveMysqlDsn(mysqlCfg.Dsn)
    db, err := mysql.Init(dsn, opts...)
    if err != nil {
        panic("init mysql error: " + err.Error())
    }
    return db
}
```
<mcfile name="mysql.go" path="/home/murphy/workspace/golang/src/github.com/moweilong/sponge/internal/database/mysql.go"></mcfile>

### 主要功能

1. **数据库连接初始化**：通过 `InitMysql` 函数初始化 MySQL 数据库连接
2. **连接池配置**：设置最大空闲连接数、最大打开连接数和连接最大生命周期
3. **日志配置**：可选启用数据库操作日志记录
4. **分布式追踪**：可选启用数据库操作的分布式追踪
5. **读写分离支持**：预留了读写分离配置接口（默认注释掉）

## 使用方法

### 生成 mysql.go 文件

可以通过以下命令生成 `mysql.go` 文件：

```bash
# 指定模块名称和数据库驱动
sponge patch gen-db-init --module-name=yourModuleName --db-driver=mysql

# 或者指定输出目录（会自动检测模块名称）
sponge patch gen-db-init --db-driver=mysql --out=./yourServerDir
```

### 配置数据库连接

生成文件后，需要在配置文件中设置数据库连接信息：

```yaml
# database setting
database:
  driver: "mysql"           # 数据库驱动
  # mysql settings
  mysql:
    # dsn 格式: <username>:<password>@(<hostname>:<port>)/<db>?[k=v& ......]
    dsn: "root:123456@(localhost:3306)/yourdb?parseTime=true&loc=Local&charset=utf8,utf8mb4"
    enableLog: true         # 是否开启日志
    maxIdleConns: 10        # 空闲连接池最大连接数
    maxOpenConns: 100       # 最大打开连接数
    connMaxLifetime: 30     # 连接最大生命周期（分钟）
```

### 在代码中使用

生成的 `mysql.go` 文件会被 `init.go` 中的 `InitDB` 函数调用：

```go
// InitDB 连接数据库
func InitDB() {
    dbDriver := config.Get().Database.Driver
    switch strings.ToLower(dbDriver) {
    case sgorm.DBDriverMysql, sgorm.DBDriverTidb:
        gdb = InitMysql()
    default:
        panic("InitDB error, please modify the correct 'database' configuration at yaml file.")
    }
}
```
<mcfile name="init.go" path="/home/murphy/workspace/golang/src/github.com/moweilong/sponge/internal/database/init.go"></mcfile>

在应用程序中，可以通过 `database.GetDB()` 获取数据库连接：

```go
db := database.GetDB()
// 使用 db 进行数据库操作...
```

## 代码优化建议

1. **错误处理改进**：当前代码在初始化失败时直接 `panic`，可以考虑返回错误让上层处理：

```go
func InitMysql() (*sgorm.DB, error) {
    // ... 现有代码 ...
    dsn := utils.AdaptiveMysqlDsn(mysqlCfg.Dsn)
    db, err := mysql.Init(dsn, opts...)
    if err != nil {
        return nil, fmt.Errorf("init mysql error: %w", err)
    }
    return db, nil
}
```

2. **连接健康检查**：添加数据库连接健康检查机制：

```go
func InitMysql() *sgorm.DB {
    // ... 现有代码 ...
    db, err := mysql.Init(dsn, opts...)
    if err != nil {
        panic("init mysql error: " + err.Error())
    }
    
    // 添加连接健康检查
    if err := db.Ping(); err != nil {
        panic("mysql connection health check failed: " + err.Error())
    }
    
    return db
}
```

3. **配置验证**：在初始化前添加配置验证：

```go
func InitMysql() *sgorm.DB {
    mysqlCfg := config.Get().Database.Mysql
    
    // 验证必要配置
    if mysqlCfg.Dsn == "" {
        panic("mysql dsn is not configured")
    }
    
    // ... 现有代码 ...
}
```

## 总结

`internal/database/mysql.go` 文件是通过 `sponge patch gen-db-init` 命令生成的，用于初始化 MySQL 数据库连接。该文件提供了灵活的配置选项，支持连接池管理、日志记录和分布式追踪等功能。生成过程中，系统会根据用户指定的模块名称替换相应的导入路径，确保生成的代码可以直接在用户的项目中使用。