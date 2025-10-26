// Package sql2code is a code generation engine that generates CRUD code for model,
// dao, handler, service, protobuf based on sql and supports database types mysql,
// mongodb, postgresql, sqlite3.
package sql2code

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/moweilong/milady/pkg/gofile"
	"github.com/moweilong/milady/pkg/sql2code/parser"
	"github.com/moweilong/milady/pkg/utils"
)

// Args generate code arguments
type Args struct {
	SQL string // DDL sql

	DDLFile string // DDL file

	DBDriver   string            // db driver name, such as mysql, mongodb, postgresql, sqlite, default is mysql
	DBDsn      string            // connecting to mysql's dsn, if DBDriver is sqlite, DBDsn is local db file
	DBTable    string            // table name
	fieldTypes map[string]string // field name:type

	Package        string // specify the package name (only valid for model types)
	GormType       bool   // whether to display the gorm type name (only valid for model type codes)
	JSONTag        bool   // does it include a json tag
	JSONNamedType  int    // json field naming type, 0: snake case such as my_field_name, 1: camel sase, such as myFieldName
	IsEmbed        bool   // is gorm.Model embedded
	IsWebProto     bool   // proto file type, true: include router path and swagger info, false: normal proto file without router and swagger
	CodeType       string // specify the different types of code to be generated, namely model (default), json, dao, handler, proto
	ForceTableName bool
	Charset        string
	Collation      string
	TablePrefix    string
	ColumnPrefix   string
	NoNullType     bool
	NullStyle      string
	IsExtendedAPI  bool // true: generate extended api (9 api), false: generate basic api (5 api)

	IsCustomTemplate bool // whether to use custom template, default is false
}

// checkValid check the code generation arguments
func (a *Args) checkValid() error {
	if a.SQL == "" && a.DDLFile == "" && (a.DBDsn == "" && a.DBTable == "") {
		return errors.New("you must specify sql or ddl file")
	}
	if a.DBTable != "" {
		tables := strings.SplitSeq(a.DBTable, ",")
		for name := range tables {
			if strings.HasSuffix(name, "_test") {
				return fmt.Errorf(`the table name (%s) suffix "_test" is not supported for code generation, please delete suffix "_test" or change it to another name. `, name)
			}
		}
	}

	switch a.DBDriver {
	case "":
		a.DBDriver = parser.DBDriverMysql
	case parser.DBDriverSqlite:
		if !gofile.IsExists(a.DBDsn) {
			return fmt.Errorf("sqlite db file %s not found in local host", a.DBDsn)
		}
	}
	if a.fieldTypes == nil {
		a.fieldTypes = make(map[string]string)
	}
	return nil
}

// getSQL get the sql string from args
//
// if args.SQL is not empty, return args.SQL
// if args.DDLFile is not empty, parse the sql file and return the sql string
// if args.DBDsn is not empty, get the sql string from database and return the sql string
//
// return the sql string, field type map, and error
// getSQL 从不同来源获取SQL语句
// 该函数根据参数配置，从以下来源之一获取SQL语句：
// 1. 直接从参数中获取SQL文本（Args.SQL）
// 2. 从指定的SQL文件中读取（Args.SQLFile）
// 3. 连接到数据库并获取表结构（Args.DBTable）
// 对于不同的数据库驱动，使用相应的方法获取表结构信息。
//
// 参数:
//
//	args - 代码生成参数配置
//
// 返回值:
//
//	string - 获取到的SQL语句
//	map[string]string - 字段类型映射，如果有特殊字段类型需要处理
//	error - 获取过程中的错误，如果成功则为nil
func getSQL(args *Args) (string, map[string]string, error) {
	// return the sql if it is not empty
	if args.SQL != "" {
		return args.SQL, nil, nil
	}

	sql := ""
	dbDriverName := strings.ToLower(args.DBDriver)
	// only mysql is supported for parsing the sql file
	if args.DDLFile != "" {
		if dbDriverName != parser.DBDriverMysql {
			return sql, nil, fmt.Errorf("not support driver %s for parsing the sql file, only mysql is supported", args.DBDriver)
		}
		b, err := os.ReadFile(args.DDLFile)
		if err != nil {
			return sql, nil, fmt.Errorf("read %s failed, %s", args.DDLFile, err)
		}
		return string(b), nil, nil
	} else if args.DBDsn != "" {
		if args.DBTable == "" {
			return sql, nil, errors.New("miss database table")
		}

		switch dbDriverName {
		case parser.DBDriverMysql, parser.DBDriverTidb:
			dsn := utils.AdaptiveMysqlDsn(args.DBDsn)
			sqlStr, err := parser.GetMysqlTableInfo(dsn, args.DBTable)
			return sqlStr, nil, err
		case parser.DBDriverPostgresql:
			dsn := utils.AdaptivePostgresqlDsn(args.DBDsn)
			fields, err := parser.GetPostgresqlTableInfo(dsn, args.DBTable)
			if err != nil {
				return "", nil, err
			}
			sqlStr, pgTypeMap := parser.ConvertToSQLByPgFields(args.DBTable, fields)
			return sqlStr, pgTypeMap, nil
		case parser.DBDriverSqlite:
			sqlStr, err := parser.GetSqliteTableInfo(args.DBDsn, args.DBTable)
			return sqlStr, nil, err
		case parser.DBDriverMongodb:
			dsn := utils.AdaptiveMongodbDsn(args.DBDsn)
			fields, err := parser.GetMongodbTableInfo(dsn, args.DBTable)
			if err != nil {
				return "", nil, err
			}
			sqlStr, mongoTypeMap := parser.ConvertToSQLByMgoFields(args.DBTable, fields)
			return sqlStr, mongoTypeMap, nil
		default:
			return "", nil, errors.New("get sql error, unsupported database driver: " + dbDriverName)
		}
	}

	return sql, nil, errors.New("no SQL input(-sql|-f|-db-dsn)")
}

// setOptions set the parser options
func setOptions(args *Args) []parser.Option {
	var opts []parser.Option

	if args.DBDriver != "" {
		opts = append(opts, parser.WithDBDriver(args.DBDriver))
	}
	if args.fieldTypes != nil {
		opts = append(opts, parser.WithFieldTypes(args.fieldTypes))
	}

	if args.Charset != "" {
		opts = append(opts, parser.WithCharset(args.Charset))
	}
	if args.Collation != "" {
		opts = append(opts, parser.WithCollation(args.Collation))
	}
	if args.JSONTag {
		opts = append(opts, parser.WithJSONTag(args.JSONNamedType))
	}
	if args.TablePrefix != "" {
		opts = append(opts, parser.WithTablePrefix(args.TablePrefix))
	}
	if args.ColumnPrefix != "" {
		opts = append(opts, parser.WithColumnPrefix(args.ColumnPrefix))
	}
	if args.NoNullType {
		opts = append(opts, parser.WithNoNullType())
	}
	if args.IsEmbed {
		opts = append(opts, parser.WithEmbed())
	}
	if args.IsWebProto {
		opts = append(opts, parser.WithWebProto())
	}

	if args.NullStyle != "" {
		switch args.NullStyle {
		case "sql":
			opts = append(opts, parser.WithNullStyle(parser.NullInSql))
		case "ptr":
			opts = append(opts, parser.WithNullStyle(parser.NullInPointer))
		default:
			fmt.Printf("invalid null style: %s\n", args.NullStyle)
			return nil
		}
	} else {
		opts = append(opts, parser.WithNullStyle(parser.NullDisable))
	}
	if args.Package != "" {
		opts = append(opts, parser.WithPackage(args.Package))
	}
	if args.GormType {
		opts = append(opts, parser.WithGormType())
	}
	if args.ForceTableName {
		opts = append(opts, parser.WithForceTableName())
	}
	if args.IsExtendedAPI {
		opts = append(opts, parser.WithExtendedAPI())
	}
	if args.IsCustomTemplate {
		opts = append(opts, parser.WithCustomTemplate())
	}

	return opts
}

// GenerateOne generate gorm code from sql, which can be obtained from parameters, files and db, with priority from highest to lowest
func GenerateOne(args *Args) (string, error) {
	codes, err := Generate(args)
	if err != nil {
		return "", err
	}

	if args.CodeType == "" {
		args.CodeType = parser.CodeTypeModel // default is model code
	}
	out, ok := codes[args.CodeType]
	if !ok {
		return "", fmt.Errorf("unknown code type %s", args.CodeType)
	}

	return out, nil
}

// Generate 生成模型、JSON、DAO、Handler、Proto等多种类型的代码
// 该函数是sql2code包的主入口，负责协调整个代码生成流程：
// 1. 验证输入参数的有效性
// 2. 从数据库或SQL文件获取表结构信息
// 3. 设置代码生成选项
// 4. 调用解析器解析SQL并生成代码
//
// 参数:
//
//	args - 代码生成参数配置，包含数据库连接信息、表名、包名等配置项
//
// 返回值:
//
//	map[string]string - 生成的各类代码映射，键为代码类型，值为代码内容
//	error - 代码生成过程中的错误，如果成功则为nil
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
	fmt.Println("生成代码的配置选项:")
	fmt.Println("--------------------------------------------------")
	fmt.Printf("参数总数: %d\n", len(opt))
	fmt.Printf("数据库驱动: %s\n", args.DBDriver)
	fmt.Printf("数据库表名: %s\n", args.DBTable)
	if args.Package != "" {
		fmt.Printf("包名: %s\n", args.Package)
	}
	if args.JSONTag {
		fmt.Printf("JSON标签: 已启用 (命名类型: %d)\n", args.JSONNamedType)
	}
	if args.TablePrefix != "" {
		fmt.Printf("表前缀: %s\n", args.TablePrefix)
	}
	if args.ColumnPrefix != "" {
		fmt.Printf("列前缀: %s\n", args.ColumnPrefix)
	}
	if args.NullStyle != "" {
		fmt.Printf("空值处理: %s\n", args.NullStyle)
	}
	fmt.Printf("无空类型: %v\n", args.NoNullType)
	fmt.Printf("嵌入模式: %v\n", args.IsEmbed)
	fmt.Printf("Web Proto: %v\n", args.IsWebProto)
	fmt.Printf("GORM类型: %v\n", args.GormType)
	fmt.Printf("强制表名: %v\n", args.ForceTableName)
	fmt.Printf("扩展API: %v\n", args.IsExtendedAPI)
	fmt.Printf("自定义模板: %v\n", args.IsCustomTemplate)
	fmt.Println("--------------------------------------------------")
	return parser.ParseSQL(sql, opt...)
}
