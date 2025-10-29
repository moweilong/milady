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

// ModelCommand create and return a command object for generating model code based on sql.
func ModelCommand(parentName string) *cobra.Command {
	var (
		// outPath code output directory, default is current directory.
		outPath string
		// dbTables database table names, multiple table names are separated by commas.
		dbTables string

		// sqlArgs sql2code arguments. default package is "model", JSONTag and GormType are enabled.
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

// modelGenerator model code generator.
type modelGenerator struct {
	// sql2code generate codes.
	codes map[string]string
	// outPath can be "" to use default value.
	outPath string
}

// generateCode replace codes to template and write to file system.
func (g *modelGenerator) generateCode() (string, error) {
	subTplName := codeNameModel
	// r get global replacer instance milady.
	r := Replacers[TplNameMilady]
	if r == nil {
		return "", errors.New("replacer is nil")
	}

	// subDirs subdirectories of milady, if not empty, will process all files in subDirs.
	// subFiles example files, the content will be replaced by the generated model code.
	subDirs := []string{}
	subFiles := []string{"internal/model/userExample.go"}

	// replacer need to process the specified subdirectories and files.
	r.SetSubDirsAndFiles(subDirs, subFiles...)
	// generate replacer rules.
	fields := g.addFields(r)
	// set replacer rules.
	r.SetReplacementFields(fields)
	_ = r.SetOutputDir(g.outPath, subTplName)
	if err := r.SaveFiles(); err != nil {
		return "", err
	}

	return r.GetOutputDir(), nil
}

// addFields add model code replace rules. clear the contents of the model/userExample.go file between the startMark and endMark.
// and replace the "UserExample" string (case-sensitive) with the actual table name.
func (g *modelGenerator) addFields(r replacer.Replacer) []replacer.Field {
	// fields replace rules.
	var fields []replacer.Field

	// rule: delete the contents of the model/userExample.go file between the startMark and endMark.
	fields = append(fields, deleteFieldsMark(r, modelFile, startMark, endMark)...)
	// rule: replace the contents of the model/userExample.go file with the generated model code.
	// rule: replace the "UserExample" string (case-sensitive) with the actual table name.
	fields = append(fields, []replacer.Field{
		{
			Old: modelFileMark,
			New: g.codes[parser.CodeTypeModel],
		},
		{
			Old:             "UserExample",
			New:             g.codes[parser.TableName],
			IsCaseSensitive: true,
		},
	}...)

	return fields
}
