package generate

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/moweilong/milady/pkg/gofile"
	"github.com/moweilong/milady/pkg/replacer"
	"github.com/moweilong/milady/pkg/sql2code"
	"github.com/moweilong/milady/pkg/sql2code/parser"
)

// HandlerPbCommand generate handler and protobuf code
func HandlerPbCommand() *cobra.Command {
	var (
		moduleName string // module name for go.mod
		serverName string // server name
		outPath    string // output directory
		dbTables   string // table names

		sqlArgs = sql2code.Args{
			Package:    "model",
			JSONTag:    true,
			GormType:   true,
			IsWebProto: true,
		}

		suitedMonoRepo bool // whether the generated code is suitable for mono-repo
	)

	cmd := &cobra.Command{
		Use:   "handler-pb",
		Short: "Generate handler and protobuf CRUD code based on sql",
		Long:  "Generate handler and protobuf CRUD code based on sql.",
		Example: color.HiBlackString(`  # Generate handler and protobuf code.
  sponge web handler-pb --module-name=yourModuleName --server-name=yourServerName --db-driver=mysql --db-dsn=root:123456@(192.168.3.37:3306)/test --db-table=user

  # Generate handler and protobuf code with multiple table names.
  sponge web handler-pb --module-name=yourModuleName --server-name=yourServerName --db-driver=mysql --db-dsn=root:123456@(192.168.3.37:3306)/test --db-table=t1,t2

  # Generate handler and protobuf code with extended api.
  sponge web handler-pb --module-name=yourModuleName --server-name=yourServerName --db-driver=mysql --db-dsn=root:123456@(192.168.3.37:3306)/test --db-table=user --extended-api=true

  # Generate handler and protobuf code and specify the server directory, Note: code generation will be canceled when the latest generated file already exists.
  sponge web handler-pb --db-driver=mysql --db-dsn=root:123456@(192.168.3.37:3306)/test --db-table=user --out=./yourServerDir

  # If you want the generated code to suited to mono-repo, you need to set the parameter --suited-mono-repo=true`),
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			mdName, srvName, smr := getNamesFromOutDir(outPath)
			if mdName != "" {
				moduleName = mdName
				suitedMonoRepo = smr
			} else if moduleName == "" {
				return errors.New(`required flag(s) "module-name" not set, use "sponge web handler-pb -h" for help`)
			}
			if srvName != "" {
				serverName = srvName
			} else if serverName == "" {
				return errors.New(`required flag(s) "server-name" not set, use "sponge web handler-pb -h" for help`)
			}

			serverName = convertServerName(serverName)
			if suitedMonoRepo {
				outPath = changeOutPath(outPath, serverName)
			}
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

				g := &handlerPbGenerator{
					moduleName:     moduleName,
					serverName:     serverName,
					dbDriver:       sqlArgs.DBDriver,
					isEmbed:        sqlArgs.IsEmbed,
					isExtendedAPI:  sqlArgs.IsExtendedAPI,
					codes:          codes,
					outPath:        outPath,
					suitedMonoRepo: suitedMonoRepo,
				}
				outPath, err = g.generateCode()
				if err != nil {
					return err
				}
			}

			fmt.Printf(`
using help:
  1. move the folders "api" and "internal" to your project code folder.
  2. open a terminal and execute the command: make proto
  3. compile and run server: make run
  4. access http://localhost:8080/apis/swagger/index.html in your browser, and test the http CRUD api.

`)
			fmt.Printf("generate \"handler-pb\" code successfully, out = %s\n", outPath)
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
	cmd.Flags().BoolVarP(&sqlArgs.IsEmbed, "embed", "e", false, "whether to embed gorm.model struct")
	cmd.Flags().BoolVarP(&sqlArgs.IsExtendedAPI, "extended-api", "a", false, "whether to generate extended crud api, additional includes: DeleteByIDs, GetByCondition, ListByIDs, ListByLatestID")
	cmd.Flags().BoolVarP(&suitedMonoRepo, "suited-mono-repo", "l", false, "whether the generated code is suitable for mono-repo")
	cmd.Flags().IntVarP(&sqlArgs.JSONNamedType, "json-name-type", "j", 1, "json tags name type, 0:snake case, 1:camel case")
	cmd.Flags().StringVarP(&outPath, "out", "o", "", "output directory, default is ./handler-pb_<time>, "+flagTip("module-name", "server-name"))

	return cmd
}

type handlerPbGenerator struct {
	moduleName     string
	serverName     string
	dbDriver       string
	isEmbed        bool
	isExtendedAPI  bool
	codes          map[string]string
	outPath        string
	suitedMonoRepo bool

	fields []replacer.Field
}

func (g *handlerPbGenerator) generateCode() (string, error) {
	subTplName := codeNameHandlerPb
	r, _ := replacer.New(SpongeDir)
	if r == nil {
		return "", errors.New("replacer is nil")
	}

	if g.serverName == "" {
		g.serverName = g.moduleName
	}

	// specify the subdirectory and files
	subDirs := []string{}
	subFiles := []string{}

	selectFiles := map[string][]string{
		"api/serverNameExample/v1": {
			"userExample.proto",
		},
		"internal/cache": {
			"userExample.go", "userExample_test.go",
		},
		"internal/dao": {
			"userExample.go", "userExample_test.go",
		},
		"internal/ecode": {
			"userExample_http.go",
		},
		"internal/handler": {
			"userExample_logic.go", "userExample_logic_test.go",
		},
		"internal/model": {
			"userExample.go",
		},
	}

	info := g.codes[parser.CodeTypeCrudInfo]
	crudInfo, _ := unmarshalCrudInfo(info)
	if crudInfo.CheckCommonType() {
		selectFiles = map[string][]string{
			"api/serverNameExample/v1": {
				"userExample.proto",
			},
			"internal/cache": {
				"userExample.go.tpl",
			},
			"internal/dao": {
				"userExample.go.tpl",
			},
			"internal/ecode": {
				"userExample_http.go.tpl",
			},
			"internal/handler": {
				"userExample_logic.go.tpl",
			},
			"internal/model": {
				"userExample.go",
			},
		}
		var fields []replacer.Field
		if g.isExtendedAPI {
			selectFiles["internal/dao"] = []string{"userExample.go.exp.tpl"}
			selectFiles["internal/ecode"] = []string{"userExample_http.go.exp.tpl"}
			selectFiles["internal/handler"] = []string{"userExample_logic.go.exp.tpl"}
			fields = commonHandlerPbExtendedFields(r)
		} else {
			fields = commonHandlerPbFields(r)
		}
		contentFields, err := replaceFilesContent(r, getTemplateFiles(selectFiles), crudInfo)
		if err != nil {
			return "", err
		}
		g.fields = append(g.fields, contentFields...)
		g.fields = append(g.fields, fields...)
	}

	replaceFiles := make(map[string][]string)
	switch strings.ToLower(g.dbDriver) {
	case DBDriverMysql, DBDriverPostgresql, DBDriverTidb, DBDriverSqlite:
		g.fields = append(g.fields, getExpectedSQLForDeletionField(g.isEmbed)...)
		if g.isExtendedAPI {
			var fields []replacer.Field
			if !crudInfo.CheckCommonType() {
				replaceFiles, fields = handlerPbExtendedAPI(r)
			}
			g.fields = append(g.fields, fields...)
		}
	case DBDriverMongodb:
		if g.isExtendedAPI {
			var fields []replacer.Field
			replaceFiles, fields = handlerPbMongoDBExtendedAPI(r)
			g.fields = append(g.fields, fields...)
		} else {
			replaceFiles = map[string][]string{
				"internal/cache": {
					"userExample.go.mgo",
				},
				"internal/dao": {
					"userExample.go.mgo",
				},
				"internal/handler": {
					"userExample_logic.go.mgo",
				},
			}
			g.fields = append(g.fields, deleteFieldsMark(r, handlerLogicFile+mgoSuffix, startMark, endMark)...)
		}

	default:
		return "", dbDriverErr(g.dbDriver)
	}

	subFiles = append(subFiles, getSubFiles(selectFiles, replaceFiles)...)

	r.SetSubDirsAndFiles(subDirs, subFiles...)
	_ = r.SetOutputDir(g.outPath, subTplName)
	fields := g.addFields(r)
	r.SetReplacementFields(fields)
	if err := r.SaveFiles(); err != nil {
		return "", err
	}

	if g.suitedMonoRepo {
		if err := moveProtoFileToAPIDir(g.moduleName, g.serverName, g.suitedMonoRepo, r.GetOutputDir()); err != nil {
			return "", err
		}
	}

	return r.GetOutputDir(), nil
}

func (g *handlerPbGenerator) addFields(r replacer.Replacer) []replacer.Field {
	var fields []replacer.Field
	fields = append(fields, g.fields...)
	fields = append(fields, deleteFieldsMark(r, modelFile, startMark, endMark)...)
	fields = append(fields, deleteFieldsMark(r, daoFile, startMark, endMark)...)
	fields = append(fields, deleteFieldsMark(r, daoMgoFile, startMark, endMark)...)
	fields = append(fields, deleteFieldsMark(r, daoTestFile, startMark, endMark)...)
	fields = append(fields, deleteFieldsMark(r, handlerLogicFile, startMark, endMark)...)
	fields = append(fields, deleteFieldsMark(r, handlerPbTestFile, startMark, endMark)...)
	fields = append(fields, deleteFieldsMark(r, protoFile, startMark, endMark)...)
	fields = append(fields, []replacer.Field{
		{ // replace the contents of the model/userExample.go file
			Old: modelFileMark,
			New: g.codes[parser.CodeTypeModel],
		},
		{ // replace the contents of the dao/userExample.go file
			Old: daoFileMark,
			New: g.codes[parser.CodeTypeDAO],
		},
		{ // replace the contents of the handler/userExample_logic.go file
			Old: embedTimeMark,
			New: getEmbedTimeCode(g.isEmbed),
		},
		{ // replace the contents of the v1/userExample.proto file
			Old: protoFileMark,
			New: g.codes[parser.CodeTypeProto],
		},
		{
			Old: selfPackageName + "/" + r.GetSourcePath(),
			New: g.moduleName,
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
			Old: "userExampleNO       = 1",
			New: fmt.Sprintf("userExampleNO = %d", rand.Intn(99)+1),
		},
		{
			Old: g.moduleName + pkgPathSuffix,
			New: "github.com/go-dev-frame/sponge/pkg",
		},
		{
			Old: "userExample_logic.go.mgo",
			New: "userExample.go",
		},
		{
			Old: "userExample_logic.go",
			New: "userExample.go",
		},
		{
			Old: "userExample_logic_test.go",
			New: "userExample_test.go",
		},
		{
			Old: "userExample.go.mgo",
			New: "userExample.go",
		},
		{
			Old: showDbNameMark,
			New: CurrentDbDriver(g.dbDriver),
		},
		{
			Old:             "UserExamplePb",
			New:             "UserExample",
			IsCaseSensitive: true,
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

	if g.suitedMonoRepo {
		fs := SubServerCodeFields(g.moduleName, g.serverName)
		fields = append(fields, fs...)
	}

	return fields
}

func handlerPbExtendedAPI(r replacer.Replacer) (map[string][]string, []replacer.Field) {
	replaceFiles := map[string][]string{
		"internal/dao": {
			"userExample.go.exp", "userExample_test.go.exp",
		},
		"internal/ecode": {
			"userExample_http.go.exp",
		},
		"internal/handler": {
			"userExample_logic.go.exp", "userExample_logic_test.go.exp",
		},
	}

	var fields []replacer.Field

	fields = append(fields, deleteFieldsMark(r, daoFile+expSuffix, startMark, endMark)...)
	fields = append(fields, deleteFieldsMark(r, daoTestFile+expSuffix, startMark, endMark)...)
	fields = append(fields, deleteFieldsMark(r, handlerLogicFile+expSuffix, startMark, endMark)...)

	fields = append(fields, []replacer.Field{
		{
			Old: "userExample_http.go.exp",
			New: "userExample_http.go",
		},
		{
			Old: "userExample_logic_test.go.exp",
			New: "userExample_test.go",
		},
		{
			Old: "userExample_test.go.exp",
			New: "userExample_test.go",
		},
		{
			Old: "userExample_logic.go.exp",
			New: "userExample.go",
		},
		{
			Old: "userExample.go.exp",
			New: "userExample.go",
		},
	}...)

	return replaceFiles, fields
}

func handlerPbMongoDBExtendedAPI(r replacer.Replacer) (map[string][]string, []replacer.Field) {
	replaceFiles := map[string][]string{
		"internal/cache": {
			"userExample.go.mgo",
		},
		"internal/dao": {
			"userExample.go.mgo.exp",
		},
		"internal/ecode": {
			"userExample_http.go.exp",
		},
		"internal/handler": {
			"userExample_logic.go.mgo.exp",
		},
	}

	var fields []replacer.Field

	fields = append(fields, deleteFieldsMark(r, daoMgoFile+expSuffix, startMark, endMark)...)
	fields = append(fields, deleteFieldsMark(r, handlerLogicFile+".mgo.exp", startMark, endMark)...)

	fields = append(fields, []replacer.Field{
		{
			Old: "userExample_http.go.exp",
			New: "userExample_http.go",
		},
		{
			Old: "userExample.go.mgo.exp",
			New: "userExample.go",
		},
		{
			Old: "userExample_logic.go.mgo.exp",
			New: "userExample.go",
		},
		{
			Old: "userExample.go.mgo",
			New: "userExample.go",
		},
	}...)

	return replaceFiles, fields
}

func commonHandlerPbFields(r replacer.Replacer) []replacer.Field {
	var fields []replacer.Field

	fields = append(fields, deleteFieldsMark(r, daoFile+tplSuffix, startMark, endMark)...)
	fields = append(fields, deleteFieldsMark(r, handlerPbFile+tplSuffix, startMark, endMark)...)
	fields = append(fields, []replacer.Field{
		{
			Old: "userExample_http.go.tpl",
			New: "userExample_http.go",
		},
		{
			Old: "userExample_logic.go.tpl",
			New: "userExample.go",
		},
		{
			Old: "userExample.go.tpl",
			New: "userExample.go",
		},
	}...)

	return fields
}

func commonHandlerPbExtendedFields(r replacer.Replacer) []replacer.Field {
	var fields []replacer.Field

	fields = append(fields, deleteFieldsMark(r, daoFile+expSuffix+tplSuffix, startMark, endMark)...)
	fields = append(fields, deleteFieldsMark(r, handlerPbFile+expSuffix+tplSuffix, startMark, endMark)...)

	fields = append(fields, []replacer.Field{
		{
			Old: "userExample_http.go.exp.tpl",
			New: "userExample_http.go",
		},
		{
			Old: "userExample_logic.go.exp.tpl",
			New: "userExample.go",
		},
		{
			Old: "userExample.go.tpl",
			New: "userExample.go",
		},
		{
			Old: "userExample.go.exp.tpl",
			New: "userExample.go",
		},
	}...)

	return fields
}
