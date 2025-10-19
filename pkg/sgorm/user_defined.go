package sgorm

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
)

// Bool is a custom type for MySQL bit(1) type and PostgreSQL boolean type.
type Bool bool
type BitBool = Bool

// Scan implement Scanner interface to read values from database
func (b *Bool) Scan(value interface{}) error {
	if value == nil {
		*b = false
		return nil
	}

	switch v := value.(type) {
	case []byte:
		*b = len(v) == 1 && v[0] == 1
	case bool:
		*b = Bool(v)
	default:
		return fmt.Errorf("unsupported type: %T for Bool", value)
	}
	return nil
}

// Value implement Valuer interface to write values to database
func (b Bool) Value() (driver.Value, error) {
	switch currentDriver {
	case "postgresql", "postgres":
		return bool(b), nil
	default: // default MySQL processing
		if b {
			return []byte{1}, nil
		}
		return []byte{0}, nil
	}
}

var currentDriver string

// SetDriver sets the name of the current database driver, such as "postgres"
// if you use postgres, you need to call SetDriver("postgres") after initializing gorm
func SetDriver(driverName string) {
	currentDriver = driverName
}

// --------------------------------------------------------------------------------------

// TinyBool is a custom type for MySQL tinyint(1) type
type TinyBool bool

// Scan implement Scanner interface to read values from database
func (b *TinyBool) Scan(value interface{}) error {
	var nb sql.NullBool
	if err := nb.Scan(value); err != nil {
		return fmt.Errorf("failed to scan TinyBool: %w", err)
	}
	*b = TinyBool(nb.Bool)
	return nil
}

// Value implement Valuer interface to write values to database
func (b TinyBool) Value() (driver.Value, error) {
	if b {
		return int64(1), nil
	}
	return int64(0), nil
}
