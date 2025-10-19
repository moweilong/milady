// Package query is a library of custom condition queries, support for complex conditional paging queries.
package query

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	// Eq equal
	Eq = "eq"
	// Neq not equal
	Neq = "neq"
	// Gt greater than
	Gt = "gt"
	// Gte greater than or equal
	Gte = "gte"
	// Lt less than
	Lt = "lt"
	// Lte less than or equal
	Lte = "lte"
	// Like fuzzy lookup
	Like = "like"
	// In include
	In = "in"
	// NotIN not include
	NotIN = "notin"
	// IsNull is null
	IsNull = "isnull"
	// IsNotNull is not null
	IsNotNull = "isnotnull"

	// AND logic and
	AND string = "and"
	// OR logic or
	OR string = "or"
)

var expMap = map[string]string{
	Eq:        " = ",
	Neq:       " <> ",
	Gt:        " > ",
	Gte:       " >= ",
	Lt:        " < ",
	Lte:       " <= ",
	Like:      " LIKE ",
	In:        " IN ",
	NotIN:     " NOT IN ",
	IsNull:    " IS NULL ",
	IsNotNull: " IS NOT NULL ",

	"=":           " = ",
	"!=":          " <> ",
	">":           " > ",
	">=":          " >= ",
	"<":           " < ",
	"<=":          " <= ",
	"not in":      " NOT IN ",
	"is null":     " IS NULL ",
	"is not null": " IS NOT NULL ",
}

var logicMap = map[string]string{
	AND: " AND ",
	OR:  " OR ",

	"&":   " AND ",
	"&&":  " AND ",
	"|":   " OR ",
	"||":  " OR ",
	"AND": " AND ",
	"OR":  " OR ",

	"and:(": " AND ",
	"and:)": " AND ",
	"or:(":  " OR ",
	"or:)":  " OR ",
}

// ---------------------------------------------------------------------------

type rulerOptions struct {
	whitelistNames map[string]bool
	validateFn     func(columns []Column) error
}

// RulerOption set the parameters of ruler options
type RulerOption func(*rulerOptions)

func (o *rulerOptions) apply(opts ...RulerOption) {
	for _, opt := range opts {
		opt(o)
	}
}

// WithWhitelistNames set white list names of columns
func WithWhitelistNames(whitelistNames map[string]bool) RulerOption {
	return func(o *rulerOptions) {
		o.whitelistNames = whitelistNames
	}
}

// WithValidateFn set validate function of columns
func WithValidateFn(fn func(columns []Column) error) RulerOption {
	return func(o *rulerOptions) {
		o.validateFn = fn
	}
}

// -----------------------------------------------------------------------------

// Params query parameters
type Params struct {
	Page  int    `json:"page" form:"page" binding:"gte=0"`
	Limit int    `json:"limit" form:"limit" binding:"gte=1"`
	Sort  string `json:"sort,omitempty" form:"sort" binding:""`

	Columns []Column `json:"columns,omitempty" form:"columns"` // not required

	// Deprecated: use Limit instead in sponge version v1.8.6, will remove in the future
	Size int `json:"size" form:"size"`
}

// Column query info
type Column struct {
	Name  string      `json:"name" form:"name"`   // column name
	Exp   string      `json:"exp" form:"exp"`     // expressions, default value is "=", support =, !=, >, >=, <, <=, like, in, notin, isnull, isnotnull
	Value interface{} `json:"value" form:"value"` // column value
	Logic string      `json:"logic" form:"logic"` // logical type, defaults to and when the value is null, with &(and), ||(or)
}

// converting ExpType to sql expressions and LogicType to sql using characters
func (c *Column) checkExp() (string, error) {
	symbol := "?"
	if c.Exp == "" {
		c.Exp = Eq
	}
	if v, ok := expMap[strings.ToLower(c.Exp)]; ok { //nolint
		c.Exp = v
		switch c.Exp {
		case " LIKE ":
			val, ok1 := c.Value.(string)
			if !ok1 {
				return symbol, fmt.Errorf("invalid value type '%s'", c.Value)
			}
			// Use rune-safe slicing to preserve multi-byte characters
			r := []rune(val)
			if len(r) > 2 {
				middle := string(r[1 : len(r)-1])
				middle = strings.ReplaceAll(middle, "%", "\\%")
				middle = strings.ReplaceAll(middle, "_", "\\_")
				val = string(r[0]) + middle + string(r[len(r)-1])
			}
			if strings.HasPrefix(val, "%") ||
				strings.HasPrefix(val, "_") ||
				strings.HasSuffix(val, "%") ||
				strings.HasSuffix(val, "_") {
				c.Value = val
			} else {
				c.Value = "%" + val + "%"
			}
		case " IN ", " NOT IN ":
			val, ok1 := c.Value.(string)
			if ok1 {
				values := []interface{}{}
				ss := strings.Split(val, ",")
				for _, s := range ss {
					s = strings.TrimSpace(s)
					if strings.HasPrefix(s, "\"") {
						values = append(values, strings.Trim(s, "\""))
						continue
					} else if strings.HasPrefix(s, "'") {
						values = append(values, strings.Trim(s, "'"))
						continue
					}
					value, err := strconv.Atoi(s)
					if err == nil {
						values = append(values, value)
					} else {
						values = append(values, s)
					}
				}
				c.Value = values
			}
			symbol = "(?)"
		case " IS NULL ", " IS NOT NULL ":
			c.Value = nil
			symbol = ""
		}
	} else {
		return symbol, fmt.Errorf("unsupported exp type '%s'", c.Exp)
	}

	if c.Logic == "" {
		c.Logic = AND
	} else {
		logic := strings.ToLower(c.Logic)
		if _, ok := logicMap[logic]; ok { //nolint
			c.Logic = logic
		} else {
			return symbol, fmt.Errorf("unsupported logic type '%s'", c.Logic)
		}
	}

	return symbol, nil
}

// ConvertToPage converted to page
func (p *Params) ConvertToPage() (order string, limit int, offset int) { //nolint
	page := NewPage(p.Page, p.Limit, p.Sort)
	order = page.sort
	limit = page.limit
	offset = page.page * page.limit
	return //nolint
}

// ConvertToGormConditions conversion to gorm-compliant parameters based on the Columns parameter
// ignore the logical type of the last column, whether it is a one-column or multi-column query
func (p *Params) ConvertToGormConditions(opts ...RulerOption) (string, []interface{}, error) { //nolint
	str := ""
	args := []interface{}{}
	l := len(p.Columns)
	if l == 0 {
		return "", nil, nil
	}

	isUseIN := true
	if l == 1 {
		isUseIN = false
	}
	field := p.Columns[0].Name

	o := rulerOptions{}
	o.apply(opts...)
	if o.validateFn != nil {
		err := o.validateFn(p.Columns)
		if err != nil {
			return "", nil, err
		}
	}

	for i, column := range p.Columns {
		// check name
		if column.Name == "" || (o.whitelistNames != nil && !o.whitelistNames[column.Name]) {
			return "", nil, fmt.Errorf("field name '%s' is not allowed", column.Name)
		}

		// check value
		if column.Value == nil {
			v := expMap[strings.ToLower(column.Exp)]
			if v != " IS NULL " && v != " IS NOT NULL " {
				return "", nil, fmt.Errorf("field 'value' cannot be nil")
			}
		} else {
			column.Value = convertValue(column.Value)
		}

		// check exp
		symbol, err := column.checkExp()
		if err != nil {
			return "", nil, err
		}

		if i == l-1 { // ignore the logical type of the last column
			switch column.Logic {
			case "or:)", "and:)":
				str += column.Name + column.Exp + symbol + " ) "
			default:
				str += column.Name + column.Exp + symbol
			}
		} else {
			switch column.Logic {
			case "or:(", "and:(":
				str += " ( " + column.Name + column.Exp + symbol + logicMap[column.Logic]
			case "or:)", "and:)":
				str += column.Name + column.Exp + symbol + " ) " + logicMap[column.Logic]
			default:
				str += column.Name + column.Exp + symbol + logicMap[column.Logic]
			}
		}
		if column.Value != nil {
			args = append(args, column.Value)
		}
		// when multiple columns are the same, determine whether the use of IN
		if isUseIN {
			if field != column.Name {
				isUseIN = false
				continue
			}
			if column.Exp != expMap[Eq] {
				isUseIN = false
			}
		}
	}

	if isUseIN {
		str = field + " IN (?)"
		args = []interface{}{args}
	}

	return str, args, nil
}

// if the value is a string or an integer, if true means it is a string, otherwise it is an integer
func convertValue(v interface{}) interface{} {
	s, ok := v.(string)
	if !ok {
		return v
	}

	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "\"") {
		s2 := strings.Trim(s, "\"")
		if _, err := strconv.Atoi(s2); err == nil {
			return s2
		}
		return s
	}
	intVal, err := strconv.Atoi(s)
	if err == nil {
		return intVal
	}
	boolVal, err := strconv.ParseBool(s)
	if err == nil {
		return boolVal
	}
	floatVal, err := strconv.ParseFloat(s, 64)
	if err == nil {
		return floatVal
	}

	// try to parse as RFC3339
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t
	}
	// support other formats
	layouts := []string{
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05Z0700",
		"2006-01-02T15:04:05.999999999Z0700",
		"2006-01-02T15:04:05.999999999Z07:00",
		"2006-01-02",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05 -07:00",
		"2006-01-02 15:04:05.999999999 -07:00",
	}
	for _, layout := range layouts {
		t, err := time.Parse(layout, s)
		if err == nil {
			return t
		}
	}
	return v
}

// -------------------------------------------------------------------------------------------

// Conditions query conditions
type Conditions struct {
	Columns []Column `json:"columns" form:"columns" binding:"min=1"` // columns info
}

// ConvertToGorm conversion to gorm-compliant parameters based on the Columns parameter
// ignore the logical type of the last column, whether it is a one-column or multi-column query
func (c *Conditions) ConvertToGorm(opts ...RulerOption) (string, []interface{}, error) {
	p := &Params{Columns: c.Columns}
	return p.ConvertToGormConditions(opts...)
}

// CheckValid check valid
func (c *Conditions) CheckValid() error {
	if len(c.Columns) == 0 {
		return fmt.Errorf("field 'columns' cannot be empty")
	}

	return nil
}
