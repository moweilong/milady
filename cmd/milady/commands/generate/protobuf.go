package generate

import (
	"errors"
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/moweilong/milady/pkg/gofile"
	"github.com/moweilong/milady/pkg/replacer"
	"github.com/moweilong/milady/pkg/sql2code"
	"github.com/moweilong/milady/pkg/sql2code/parser"
)

// ProtobufCommand generate protobuf code
func ProtobufCommand() *cobra.Command {
	var (
		moduleName string // module name for go.mod
		serverName string // server name
		outPath    string // output directory
		dbTables   string // table names

		sqlArgs = sql2code.Args{
			JSONTag: true,
		}
	)

	cmd := &cobra.Command{
		Use:   "protobuf",
		Short: "Generate protobuf code based on sql",
		Long:  "Generate protobuf code based on sql.",
		Example: color.HiBlackString(`  # Generate protobuf code.
  sponge micro protobuf --module-name=yourModuleName --server-name=yourServerName --db-driver=mysql --db-dsn=root:123456@(192.168.3.37:3306)/test --db-table=user

  # Generate protobuf code with multiple table names.
  sponge micro protobuf --module-name=yourModuleName --server-name=yourServerName --db-driver=mysql --db-dsn=root:123456@(192.168.3.37:3306)/test --db-table=t1,t2

  # Generate protobuf code with extended api.
  sponge micro protobuf --module-name=yourModuleName --server-name=yourServerName --db-driver=mysql --db-dsn=root:123456@(192.168.3.37:3306)/test --db-table=user --extended-api=true

  # Generate protobuf code that include router path and swagger info.
  sponge micro protobuf --module-name=yourModuleName --server-name=yourServerName --db-driver=mysql --db-dsn=root:123456@(192.168.3.37:3306)/test --db-table=user --web-type=true

  # Generate protobuf code and specify the server directory, Note: code generation will be canceled when the latest generated file already exists.
  sponge micro protobuf --db-driver=mysql --db-dsn=root:123456@(192.168.3.37:3306)/test --db-table=user --out=./yourServerDir`),
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			mdName, srvName, _ := getNamesFromOutDir(outPath)
			if mdName != "" {
				moduleName = mdName
			} else if moduleName == "" {
				return errors.New(`required flag(s) "module-name" not set, use "sponge micro protobuf -h" for help`)
			}
			if srvName != "" {
				serverName = srvName
			} else if serverName == "" {
				return errors.New(`required flag(s) "server-name" not set, use "sponge micro protobuf -h" for help`)
			}

			serverName = convertServerName(serverName)
			if sqlArgs.DBDriver == DBDriverMongodb {
				sqlArgs.IsEmbed = false
			}

			tableNames := strings.Split(dbTables, ",")
			for _, tableName := range tableNames {
				if tableName == "" {
					continue
				}

				sqlArgs.DBTable = tableName
				codes, err := sql2code.Generate(&sqlArgs)
				if err != nil {
					return err
				}

				g := &protobufGenerator{
					moduleName: moduleName,
					serverName: serverName,
					codes:      codes,
					outPath:    outPath,
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
			fmt.Printf("generate \"protobuf\" code successfully, out = %s\n", outPath)

			return nil
		},
	}

	cmd.Flags().StringVarP(&moduleName, "module-name", "m", "", "module-name is the name of the module in the go.mod file")
	//_ = cmd.MarkFlagRequired("module-name")
	cmd.Flags().StringVarP(&serverName, "server-name", "s", "", "server name")
	//_ = cmd.MarkFlagRequired("server-name")
	cmd.Flags().StringVarP(&sqlArgs.DBDriver, "db-driver", "k", "mysql", "database driver, support mysql, mongodb, postgresql, sqlite")
	cmd.Flags().StringVarP(&sqlArgs.DBDsn, "db-dsn", "d", "", "database content address, e.g. user:password@(host:port)/database. Note: if db-driver=sqlite, db-dsn must be a local sqlite db file, e.g. --db-dsn=/tmp/sponge_sqlite.db") //nolint
	_ = cmd.MarkFlagRequired("db-dsn")
	cmd.Flags().StringVarP(&dbTables, "db-table", "t", "", "table name, multiple names separated by commas")
	_ = cmd.MarkFlagRequired("db-table")
	cmd.Flags().IntVarP(&sqlArgs.JSONNamedType, "json-name-type", "j", 1, "json tags name type, 0:snake case, 1:camel case")
	cmd.Flags().BoolVarP(&sqlArgs.IsWebProto, "web-type", "w", false, "if true, the proto file include router path and swagger info")
	cmd.Flags().BoolVarP(&sqlArgs.IsExtendedAPI, "extended-api", "a", false, "whether to generate extended crud api, additional includes: DeleteByIDs, GetByCondition, ListByIDs, ListByLatestID")
	cmd.Flags().StringVarP(&outPath, "out", "o", "", "output directory, default is ./protobuf_<time>, "+flagTip("module-name", "server-name"))

	return cmd
}

type protobufGenerator struct {
	moduleName string
	serverName string
	codes      map[string]string
	outPath    string
}

func (g *protobufGenerator) generateCode() (string, error) {
	subTplName := codeNameProtobuf
	r := Replacers[TplNameSponge]
	if r == nil {
		return "", errors.New("replacer is nil")
	}

	if g.serverName == "" {
		g.serverName = g.moduleName
	}

	// specify the subdirectory and files
	subDirs := []string{}
	subFiles := []string{"api/serverNameExample/v1/userExample.proto"}

	r.SetSubDirsAndFiles(subDirs, subFiles...)
	fields := g.addFields(r)
	r.SetReplacementFields(fields)
	_ = r.SetOutputDir(g.outPath, subTplName)
	if err := r.SaveFiles(); err != nil {
		return "", err
	}

	return r.GetOutputDir(), nil
}

func (g *protobufGenerator) addFields(r replacer.Replacer) []replacer.Field {
	var fields []replacer.Field

	fields = append(fields, deleteFieldsMark(r, protoFile, startMark, endMark)...)
	fields = append(fields, []replacer.Field{
		{ // replace the contents of the v1/userExample.proto file
			Old: protoFileMark,
			New: g.codes[parser.CodeTypeProto],
		},
		{
			Old: "github.com/go-dev-frame/sponge",
			New: g.moduleName,
		},
		// replace directory name
		{
			Old: strings.Join([]string{"api", "serverNameExample", "v1"}, gofile.GetPathDelimiter()),
			New: strings.Join([]string{"api", g.serverName, "v1"}, gofile.GetPathDelimiter()),
		},
		{
			Old: "api/serverNameExample/v1",
			New: fmt.Sprintf("api/%s/v1", g.serverName),
		},
		// Note: protobuf package no "-" signs allowed
		{
			Old: "api.serverNameExample.v1",
			New: fmt.Sprintf("api.%s.v1", g.serverName),
		},
		{
			Old: "serverNameExample",
			New: g.serverName,
		},
		{
			Old:             "UserExample",
			New:             g.codes[parser.TableName],
			IsCaseSensitive: true,
		},
	}...)

	return fields
}
