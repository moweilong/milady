## sql2code

`sql2code` is a code generation engine that generates CRUD code for model, dao, handler, service, protobuf based on sql and supports database types mysql, mongodb, postgresql, sqlite3.

<br>

### Example of use

Generate code based on database table.

```go
    import "github.com/go-dev-frame/sponge/pkg/sql2code"

    // generate model, dao, handler, service and protobuf code, supports database type: mysql, mongodb, postgres, sqlite3
    codes, err := sql2code.Generate(&sql2code.Args{
      DBDriver: "mysql",
      DBDsn: "root:123456@(127.0.0.1:3306)/account"
      DBTable "user"
      GormType: true,
      JSONTag: true,
      IsEmbed: true,
      IsExtendedAPI: false
    })

    // write code to file
```

Generate table information based on database table, used for customized code generation.

```go
    import "github.com/go-dev-frame/sponge/pkg/sql2code"

    // generate table information based on database table, supports database type: mysql, mongodb, postgres, sqlite3
    codes, err := sql2code.Generate(&sql2code.Args{
      DBDriver: "mysql",
      DBDsn: "root:123456@(127.0.0.1:3306)/account"
      DBTable "user"
      GormType: true,
      JSONTag: true,
      IsEmbed: true,
      IsExtendedAPI: true
    })

    // generate customized code to file
```