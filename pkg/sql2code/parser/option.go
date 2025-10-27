package parser

// NullStyle null type
type NullStyle int

// nolint
const (
	NullDisable NullStyle = iota
	NullInSql
	NullInPointer
)

// Option function
type Option func(*options)

type options struct {
	DBDriver       string            // 数据库驱动，默认mysql
	FieldTypes     map[string]string // name:type
	Charset        string            // 数据库字符集
	Collation      string            // 数据库排序规则
	JSONTag        bool              // 是否添加json tag
	JSONNamedType  int               // 0: snake case, 1: PascalCase
	TablePrefix    string            // 表前缀
	ColumnPrefix   string            // 列前缀
	NoNullType     bool              // 是否禁用null类型
	NullStyle      NullStyle         // null类型，默认NullInSql
	Package        string            // 包名，默认 model
	GormType       bool              // 是否使用 gorm type tag
	ForceTableName bool              // 是否强制使用表名
	IsEmbed        bool              // is gorm.Model embedded
	IsWebProto     bool              // true: proto file include router path and swagger info, false: normal proto file without router and swagger
	IsExtendedAPI  bool              // true: extended api (9 api), false: basic api (5 api)

	IsCustomTemplate bool // true: custom extend template, false: use milady template
}

var defaultOptions = options{
	DBDriver:   "mysql",
	FieldTypes: map[string]string{},
	NullStyle:  NullInSql,
	Package:    "model",
}

// WithDBDriver set db driver
func WithDBDriver(driver string) Option {
	return func(o *options) {
		if driver != "" {
			o.DBDriver = driver
		}
	}
}

// WithFieldTypes set field types
func WithFieldTypes(fieldTypes map[string]string) Option {
	return func(o *options) {
		if fieldTypes != nil {
			o.FieldTypes = fieldTypes
		}
	}
}

// WithCharset set charset
func WithCharset(charset string) Option {
	return func(o *options) {
		o.Charset = charset
	}
}

// WithCollation set collation
func WithCollation(collation string) Option {
	return func(o *options) {
		o.Collation = collation
	}
}

// WithTablePrefix set table prefix
func WithTablePrefix(p string) Option {
	return func(o *options) {
		o.TablePrefix = p
	}
}

// WithColumnPrefix set column prefix
func WithColumnPrefix(p string) Option {
	return func(o *options) {
		o.ColumnPrefix = p
	}
}

// WithJSONTag set json tag, 0 for underscore, other values for hump
func WithJSONTag(namedType int) Option {
	return func(o *options) {
		o.JSONTag = true
		o.JSONNamedType = namedType
	}
}

// WithNoNullType set NoNullType
func WithNoNullType() Option {
	return func(o *options) {
		o.NoNullType = true
	}
}

// WithNullStyle set NullType
func WithNullStyle(s NullStyle) Option {
	return func(o *options) {
		o.NullStyle = s
	}
}

// WithPackage set package name
func WithPackage(pkg string) Option {
	return func(o *options) {
		o.Package = pkg
	}
}

// WithGormType will write type in gorm tag
func WithGormType() Option {
	return func(o *options) {
		o.GormType = true
	}
}

// WithForceTableName set forceFloats
func WithForceTableName() Option {
	return func(o *options) {
		o.ForceTableName = true
	}
}

// WithEmbed is embed gorm.Model
func WithEmbed() Option {
	return func(o *options) {
		o.IsEmbed = true
	}
}

// WithWebProto set proto file type
func WithWebProto() Option {
	return func(o *options) {
		o.IsWebProto = true
	}
}

// WithExtendedAPI set extended api
func WithExtendedAPI() Option {
	return func(o *options) {
		o.IsExtendedAPI = true
	}
}

// WithCustomTemplate set custom template
func WithCustomTemplate() Option {
	return func(o *options) {
		o.IsCustomTemplate = true
	}
}

// parseOption apply options to override default options
func parseOption(options []Option) options {
	o := defaultOptions
	for _, f := range options {
		f(&o)
	}
	if o.NoNullType {
		o.NullStyle = NullDisable
	}
	return o
}
