# MySQL SQL 代码生成流程分析

## 概述

本文档详细分析 Milady 项目中 MySQL SQL 语句如何通过 `sql2code` 包转换为 Go 代码的完整流程。这一流程从命令行界面开始，经过 SQL 解析、数据结构映射，最终生成模型、JSON、DAO、Handler 等多种类型的代码。

## 核心流程概览

整体代码生成流程如下：

1. **命令行交互与参数解析**：用户通过 `milady generate model` 命令提供数据库连接信息和表名
2. **SQL 获取**：从数据库中获取表的 CREATE TABLE DDL 语句
3. **SQL 解析**：使用 sqlparser 库解析 SQL 语句，提取表结构信息
4. **数据结构映射**：将表字段映射为 Go 结构体字段，处理类型转换、标签等
5. **代码生成**：基于模板生成各种类型的 Go 代码

## 详细流程分析

### 1. 命令行入口与参数处理

命令行功能在 <mcfile name="model.go" path="/Users/moweilong/Workspace/go/src/github.com/moweilong/milady/cmd/milady/commands/generate/model.go"></mcfile> 中实现：

```go
// ModelCommand generate model code
func ModelCommand(parentName string) *cobra.Command {
    var (
        outPath  string // output directory
        dbTables string // table names

        sqlArgs = sql2code.Args{
            Package:  "model",
            JSONTag:  true,
            GormType: true,
        }
    )
    // ... 命令定义与参数设置 ...
}
```

关键参数说明：
- `--db-driver`：数据库驱动，默认为 mysql
- `--db-dsn`：数据库连接字符串，如 `root:123456@(127.0.0.1:3306)/test`
- `--db-table`：表名，多个表用逗号分隔
- `--out`：输出目录

### 2. SQL 获取流程

当用户执行命令后，核心处理逻辑位于 RunE 函数中：

```go
RunE: func(cmd *cobra.Command, args []string) error {
    tableNames := strings.SplitSeq(dbTables, ",")
    for tableName := range tableNames {
        if tableName == "" {
            continue
        }

        if sqlArgs.DBDriver == DBDriverMongodb {
            sqlArgs.IsEmbed = false
        }
        sqlArgs.DBTable = tableName
        codes, err := sql2code.Generate(&sqlArgs)
        // ... 代码生成与输出 ...
    }
}
```

在 <mcfile name="sql2code.go" path="/Users/moweilong/Workspace/go/src/github.com/moweilong/milady/pkg/sql2code/sql2code.go"></mcfile> 中，`Generate` 函数是核心入口：

```go
// Generate model, json, dao, handler, proto codes
func Generate(args *Args) (map[string]string, error) {
    if err := args.checkValid(); err != nil {
        return nil, err
    }

    sql, fieldTypes, err := getSQL(args)
    if err != nil {
        return nil, err
    }
    if fieldTypes != nil {
        args.fieldTypes = fieldTypes
    }
    if sql == "" {
        return nil, fmt.Errorf("get sql from %s error, maybe the table %s doesn't exist", args.DBDriver, args.DBTable)
    }

    opt := setOptions(args)

    return parser.ParseSQL(sql, opt...)
}
```

对于 MySQL 数据库，`getSQL` 函数会调用 `parser.GetMysqlTableInfo` 获取表的 DDL：

```go
func getSQL(args *Args) (string, map[string]string, error) {
    // ...
    switch dbDriverName {
    case parser.DBDriverMysql, parser.DBDriverTidb:
        dsn := utils.AdaptiveMysqlDsn(args.DBDsn)
        sqlStr, err := parser.GetMysqlTableInfo(dsn, args.DBTable)
        return sqlStr, nil, err
    // ... 其他数据库类型 ...
    }
    // ...
}
```

### 3. MySQL 表信息获取

在 <mcfile name="mysql.go" path="/Users/moweilong/Workspace/go/src/github.com/moweilong/milady/pkg/sql2code/parser/mysql.go"></mcfile> 中，`GetMysqlTableInfo` 函数通过 MySQL 的 `SHOW CREATE TABLE` 命令获取表结构：

```go
// GetMysqlTableInfo get create table ddl from mysql
func GetMysqlTableInfo(dsn, tableName string) (string, error) {
    db, err := sql.Open("mysql", dsn)
    if err != nil {
        return "", fmt.Errorf("GetMysqlTableInfo error, %v", err)
    }
    defer db.Close()

    rows, err := db.Query("SHOW CREATE TABLE `" + tableName + "`")
    if err != nil {
        return "", fmt.Errorf("query show create table error, %v", err)
    }

    defer rows.Close()
    if !rows.Next() {
        return "", fmt.Errorf("not found found table '%s'", tableName)
    }

    var table string
    var info string
    err = rows.Scan(&table, &info)
    if err != nil {
        return "", err
    }

    return info, nil
}
```

### 4. SQL 解析与代码生成

回到 <mcfile name="parser.go" path="/Users/moweilong/Workspace/go/src/github.com/moweilong/milady/pkg/sql2code/parser/parser.go"></mcfile>，`ParseSQL` 函数使用 `github.com/zhufuyi/sqlparser` 库解析 SQL 语句：

```go
// ParseSQL generate different usage codes based on sql
func ParseSQL(sql string, options ...Option) (map[string]string, error) {
    initTemplate()
    initCommonTemplate()
    opt := parseOption(options)

    stmts, err := parser.New().Parse(sql, opt.Charset, opt.Collation)
    if err != nil {
        return nil, err
    }
    // ... 初始化各种代码集合 ...
    for _, stmt := range stmts {
        if ct, ok := stmt.(*ast.CreateTableStmt); ok {
            code, err2 := makeCode(ct, opt)
            if err2 != nil {
                return nil, err2
            }
            // ... 收集生成的代码 ...
        }
    }
    // ... 生成最终代码并返回 ...
}
```

### 5. 核心代码生成：makeCode 函数

`makeCode` 函数是代码生成的核心，它将解析后的表结构转换为各种代码：

```go
func makeCode(stmt *ast.CreateTableStmt, opt options) (*codeText, error) {
    importPath := make([]string, 0, 1)
    data := tmplData{
        TableNamePrefix: opt.TablePrefix,
        RawTableName:    stmt.Table.Name.String(),
        DBDriver:        opt.DBDriver,
    }

    // 处理表名，转换为驼峰命名
    tablePrefix := data.TableNamePrefix
    if tablePrefix != "" && strings.HasPrefix(data.RawTableName, tablePrefix) {
        data.NameFunc = true
        data.TableName = toCamel(data.RawTableName[len(tablePrefix):])
    } else {
        data.TableName = toCamel(data.RawTableName)
    }
    data.TName = firstLetterToLower(data.TableName)

    // 获取表注释
    for _, o := range stmt.Options {
        if o.Tp == ast.TableOptionComment {
            data.Comment = replaceCommentNewline(o.StrValue)
            break
        }
    }

    // 识别主键
    isPrimaryKey := make(map[string]bool)
    for _, con := range stmt.Constraints {
        if con.Tp == ast.ConstraintPrimaryKey {
            isPrimaryKey[con.Keys[0].Column.String()] = true
        }
        // 外键处理（待实现）
    }

    // 处理列信息
    columnPrefix := opt.ColumnPrefix
    for _, col := range stmt.Cols {
        colName := col.Name.Name.String()
        goFieldName := colName
        // 处理列前缀
        if columnPrefix != "" && strings.HasPrefix(goFieldName, columnPrefix) {
            goFieldName = goFieldName[len(columnPrefix):]
        }
        // 处理JSON命名格式
        jsonName := colName
        if opt.JSONNamedType == 0 { // snake case
            jsonName = customToSnake(jsonName)
        } else {
            jsonName = customToCamel(jsonName) // camel case (default)
        }
        // 创建字段信息
        field := tmplField{
            Name:     toCamel(goFieldName),
            ColName:  colName,
            JSONName: jsonName,
            // ... 更多字段设置 ...
        }
        // ... 处理字段类型、标签等 ...
    }
    // ... 生成各种代码 ...
}
```

### 6. 字段类型映射

在代码生成过程中，MySQL 数据类型会被映射为对应的 Go 数据类型。这一映射基于字段的类型、是否为主键、是否可空等特性。

系统支持多种数据类型转换，包括：
- 整数类型 (INT, BIGINT 等) → int, int64 等
- 字符串类型 (VARCHAR, TEXT 等) → string
- 浮点数类型 (FLOAT, DOUBLE 等) → float32, float64
- 时间类型 (DATETIME, TIMESTAMP) → time.Time
- JSON类型 → datatypes.JSON
- DECIMAL类型 → decimal.Decimal
- 布尔类型 → bool 或自定义 sgorm.Bool

### 7. 代码模板与输出

系统使用 Go 的 `text/template` 包来生成各种类型的代码。根据用户配置，可以生成以下类型的代码：

- **Model**：数据模型结构体代码，包含 GORM 标签和 JSON 标签
- **JSON**：JSON 序列化相关代码
- **DAO**：数据访问对象代码，包含更新字段等
- **Handler**：HTTP 处理器代码，包含请求和响应结构体
- **Proto**：Protocol Buffers 定义文件
- **Service**：gRPC 服务代码

## 代码生成配置选项

系统支持多种配置选项来控制代码生成行为：

| 配置项 | 描述 | 默认值 |
|-------|------|--------|
| Package | 生成代码的包名 | model |
| GormType | 是否显示 GORM 类型名称 | true |
| JSONTag | 是否包含 JSON 标签 | true |
| JSONNamedType | JSON 字段命名类型 (0:snake case, 1:camel case) | 1 (camel case) |
| IsEmbed | 是否嵌入 gorm.Model 结构体 | false |
| TablePrefix | 表名前缀 | "" |
| ColumnPrefix | 列名前缀 | "" |
| NoNullType | 是否不使用可空类型 | false |
| NullStyle | 可空类型样式 (sql, ptr) | 禁用 |

## 特殊处理逻辑

### 1. 表名和列名转换

系统会自动将数据库命名风格（通常是下划线分隔）转换为 Go 的驼峰命名风格：
- 表名：`user_info` → `UserInfo`
- 列名：`user_name` → `UserName`

### 2. 主键处理

系统会识别表的主键，并在生成代码时特殊处理：
- 在结构体中添加相应的标签
- 在生成 CRUD 代码时优先使用主键

### 3. 注释提取

系统会从表和列的注释中提取信息，并添加到生成的代码中作为文档注释。

## 代码生成输出

生成的代码最终会输出到指定的目录中，默认是当前目录下的 `model_<时间戳>` 文件夹。输出的文件结构根据生成的代码类型不同而不同。

## 输入输出示例

#### 输入输出示例

输入：MySQL 表结构
```sql
CREATE TABLE `user` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '用户ID',
  `username` varchar(50) NOT NULL COMMENT '用户名',
  `email` varchar(100) DEFAULT NULL COMMENT '邮箱',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表';
```

输出：生成的 Go 模型代码
```go
package model

import (
	"time"

	"gorm.io/gorm"
)

// User 用户表
type User struct {
	ID        int64          `gorm:"primaryKey;autoIncrement;column:id" json:"id"`                  // 用户ID
	Username  string         `gorm:"not null;size:50;column:username" json:"username"`              // 用户名
	Email     string         `gorm:"size:100;column:email" json:"email"`                           // 邮箱
	CreatedAt time.Time      `gorm:"autoCreateTime;column:created_at" json:"createdAt"`             // 创建时间
	UpdatedAt time.Time      `gorm:"autoUpdateTime;column:updated_at" json:"updatedAt"`             // 更新时间
	DeletedAt gorm.DeletedAt `gorm:"index;column:deleted_at" json:"deletedAt,omitempty"`
}

// TableName get sql table name.
func (User) TableName() string {
	return "user"
}
```

## 代码优化建议

1. **外键支持完善**：当前代码中有 TODO 注释表明外键支持尚未实现，可以考虑添加外键关系的解析和处理。

2. **错误处理增强**：在一些错误处理路径中，可以提供更详细的错误信息，特别是在类型映射失败时。

3. **并发处理优化**：当处理多个表时，可以考虑使用并发方式提高代码生成效率。

4. **模板缓存机制**：对于频繁使用的模板，可以实现缓存机制避免重复初始化。

5. **配置验证增强**：增加对配置参数的更严格验证，避免生成无效代码。

## 总结

Milady 项目的 SQL 代码生成功能通过一系列精心设计的组件，实现了从 MySQL 表结构到多种类型 Go 代码的自动转换。这一功能大大提高了开发效率，减少了手写重复代码的工作量，同时保证了代码的一致性和规范性。

整个流程涵盖了命令行交互、数据库连接、SQL 解析、数据结构映射和代码生成等多个环节，形成了一个完整的代码生成流水线。用户可以通过丰富的配置选项来定制生成的代码，满足不同的业务需求。