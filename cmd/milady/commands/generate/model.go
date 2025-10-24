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
		RunE: func(cmd *cobra.Command, args []string) error {
			tableNames := strings.Split(dbTables, ",")
			for _, tableName := range tableNames {
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
  move the folder "internal/apiserver" to your project code folder.

`)
			fmt.Printf("generate \"model\" code successfully, out = %s\n", outPath)
			return nil
		},
	}

	cmd.Flags().StringVarP(&sqlArgs.DBDriver, "db-driver", "k", "mysql", "database driver, support mysql, mongodb, postgresql, sqlite")
	cmd.Flags().StringVarP(&sqlArgs.DBDsn, "db-dsn", "d", "", "database content address, e.g. user:password@(host:port)/database. Note: if db-driver=sqlite, db-dsn must be a local sqlite db file, e.g. --db-dsn=/tmp/sponge_sqlite.db") //nolint
	_ = cmd.MarkFlagRequired("db-dsn")
	cmd.Flags().StringVarP(&dbTables, "db-table", "t", "", "table name, multiple names separated by commas")
	_ = cmd.MarkFlagRequired("db-table")
	cmd.Flags().BoolVarP(&sqlArgs.IsEmbed, "embed", "e", false, "whether to embed gorm.model struct")
	cmd.Flags().IntVarP(&sqlArgs.JSONNamedType, "json-name-type", "j", 1, "json tags name type, 0:snake case, 1:camel case")
	cmd.Flags().StringVarP(&outPath, "out", "o", "", "output directory, default is ./model_<time>")

	return cmd
}

type modelGenerator struct {
	codes   map[string]string
	outPath string
}

// generateCode 生成模型代码
//
// 返回生成的代码目录路径和可能的错误
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
	// 指定要处理的模板文件, 这是一个模型文件的示例模板
	subFiles := []string{"internal/apiserver/model/userExample.go"}

	// 配置替换器 ：设置子目录和文件
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
