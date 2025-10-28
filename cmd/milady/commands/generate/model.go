package generate

import (
	"errors"
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/moweilong/milady/pkg/replacer"
	"github.com/moweilong/milady/pkg/sql2code"
	"github.com/moweilong/milady/pkg/sql2code/parser"
)

// ModelCommand 创建并返回生成模型代码的命令对象
// 该函数定义了用于从数据库表生成Go模型代码的命令行工具，支持MySQL、MongoDB、PostgreSQL和SQLite等数据库。
// 命令允许用户通过参数指定数据库连接信息、表名、输出目录等配置。
//
// 参数:
//
//	parentName - 父命令名称，用于生成命令示例中的完整命令路径
//
// 返回值:
//
//	*cobra.Command - 配置好的命令对象，包含参数定义和执行逻辑
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

	cmd := &cobra.Command{
		Use:   "model",
		Short: "Generate model code based on sql",
		Long:  "Generate model code based on sql.",
		Example: color.HiBlackString(fmt.Sprintf(`  # Generate model code.
  milady %s model --db-driver=mysql --db-dsn=root:123456@(127.0.0.1:3306)/test --db-table=user

  # Generate model code with multiple table names.
  milady %s model --db-driver=mysql --db-dsn=root:123456@(127.0.0.1:3306)/test --db-table=t1,t2

  # Generate model code and specify the server directory, Note: code generation will be canceled when the latest generated file already exists.
  milady %s model --db-driver=mysql --db-dsn=root:123456@(127.0.0.1:3306)/test --db-table=user --out=./yourServerDir`,
			parentName, parentName, parentName)),
		SilenceErrors: true,
		SilenceUsage:  true,
		// RunE 命令执行的核心逻辑
		// 该函数处理命令行参数，遍历指定的表名，为每个表生成相应的模型代码：
		// 1. 解析表名列表（支持逗号分隔的多个表名）
		// 2. 针对每个有效的表名，配置sql2code参数
		// 3. 调用sql2code.Generate生成各类代码
		// 4. 使用modelGenerator将生成的代码写入文件系统
		// 5. 生成完成后输出帮助信息和成功消息
		//
		// 参数:
		//   cmd - 命令对象
		//   args - 命令行参数
		//
		// 返回值:
		//   error - 执行过程中的错误，如果成功则为nil
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
				if err != nil {
					return err
				}

				g := &modelGenerator{
					codes:   codes,
					outPath: outPath,
				}
				outPath, err = g.generateCode()
				if err != nil {
					return err
				}
			}

			fmt.Printf(`
using help:
  move the folder "internal" to your project code folder.

`)
			fmt.Printf("generate \"model\" code successfully, out = %s\n", outPath)
			return nil
		},
	}

	cmd.Flags().StringVarP(&sqlArgs.DBDriver, "db-driver", "k", "mysql", "database driver, support mysql, mongodb, postgresql, sqlite")
	cmd.Flags().StringVarP(&sqlArgs.DBDsn, "db-dsn", "d", "", "database content address, e.g. user:password@(host:port)/database. Note: if db-driver=sqlite, db-dsn must be a local sqlite db file, e.g. --db-dsn=/tmp/milady_sqlite.db") //nolint
	_ = cmd.MarkFlagRequired("db-dsn")
	cmd.Flags().StringVarP(&dbTables, "db-table", "t", "", "table name, multiple names separated by commas")
	_ = cmd.MarkFlagRequired("db-table")
	cmd.Flags().BoolVarP(&sqlArgs.IsEmbed, "embed", "e", false, "whether to embed gorm.model struct, invalid for mongodb")
	cmd.Flags().IntVarP(&sqlArgs.JSONNamedType, "json-name-type", "j", 0, "json tags name type, 0:snake case, 1:camel case")
	cmd.Flags().StringVarP(&outPath, "out", "o", "", "output directory, default is ./model_<time>")

	return cmd
}

type modelGenerator struct {
	codes   map[string]string
	outPath string
}

// generateCode 生成模型代码并写入文件系统
// 该方法负责将sql2code.Generate生成的各类代码写入到适当的文件中，
// 创建必要的目录结构，并处理文件命名和内容替换。
//
// 返回值:
//
//	string - 生成的代码目录路径
//	error - 生成过程中的错误，如果成功则为nil
func (g *modelGenerator) generateCode() (string, error) {
	// 标识要使用的子模板类型
	subTplName := codeNameModel
	// 获取代码替换器实例
	r := Replacers[TplNameMilady] // 在 milady 运行之初就要求执行 init 命令, 确保 Replacers 已初始化
	if r == nil {
		return "", errors.New("replacer is nil")
	}

	// specify the subdirectory and files
	subDirs := []string{}
	subFiles := []string{"internal/apiserver/model/userExample.go"}

	// 配置替换器需要处理的文件
	//  1. 子目录列表为空并且文件列表为空, 表示处理所有文件
	//  2. 其中一个不为空, 表示只处理指定的子目录和文件
	r.SetSubDirsAndFiles(subDirs, subFiles...)
	// 添加模型代码替换规则
	fields := g.addFields(r)
	// 设置替换规则
	r.SetReplacementFields(fields)
	// 设置输出目录
	_ = r.SetOutputDir(g.outPath, subTplName)
	// 保存生成的文件
	if err := r.SaveFiles(); err != nil {
		return "", err
	}

	return r.GetOutputDir(), nil
}

// addFields 添加模型代码替换规则
//
// 接收一个replacer实例作为参数
//
// 返回一个replacer.Field切片，用于定义内容替换规则
func (g *modelGenerator) addFields(r replacer.Replacer) []replacer.Field {
	var fields []replacer.Field

	// 调用 deleteFieldsMark 函数删除模板文件中特定标记之间的内容
	fields = append(fields, deleteFieldsMark(r, modelFile, startMark, endMark)...)
	fields = append(fields, []replacer.Field{
		{ // replace the contents of the model/userExample.go file
			// 将模板中的 modelFileMark 标记替换为实际生成的模型代码
			Old: modelFileMark,
			New: g.codes[parser.CodeTypeModel],
		},
		{
			// 将模板中的 "UserExample" 字符串（大小写敏感）替换为实际的表名
			Old:             "UserExample",
			New:             g.codes[parser.TableName],
			IsCaseSensitive: true,
		},
	}...)

	return fields
}
