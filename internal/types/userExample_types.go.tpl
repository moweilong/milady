package types

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.

// todo generate the request and response struct to here
// delete the templates code start

// CreateUserExampleRequest request params
type CreateUserExampleRequest struct {
	Name     string `json:"name" binding:"min=2"`         // username
	Email    string `json:"email" binding:"email"`        // email
	Password string `json:"password" binding:"md5"`       // password
	Phone    string `json:"phone" binding:"e164"`         // phone number, e164 rules, e.g. +8612345678901
	Avatar   string `json:"avatar" binding:"min=5"`       // avatar
	Age      int    `json:"age" binding:"gt=0,lt=120"`    // age
	Gender   int    `json:"gender" binding:"gte=0,lte=2"` // gender, 1:Male, 2:Female, other values:unknown
}

// UpdateUserExampleByIDRequest request params
type UpdateUserExampleByIDRequest struct {
	ID       uint64 `json:"id" binding:"-"`      // id
	Name     string `json:"name" binding:""`     // username
	Email    string `json:"email" binding:""`    // email
	Password string `json:"password" binding:""` // password
	Phone    string `json:"phone" binding:""`    // phone number
	Avatar   string `json:"avatar" binding:""`   // avatar
	Age      int    `json:"age" binding:""`      // age
	Gender   int    `json:"gender" binding:""`   // gender, 1:Male, 2:Female, other values:unknown
}

// UserExampleObjDetail detail
type UserExampleObjDetail struct {
	ID        uint64    `json:"id"`        // id
	Name      string    `json:"name"`      // username
	Email     string    `json:"email"`     // email
	Phone     string    `json:"phone"`     // phone number
	Avatar    string    `json:"avatar"`    // avatar
	Age       int       `json:"age"`       // age
	Gender    int       `json:"gender"`    // gender, 1:Male, 2:Female, other values:unknown
	Status    int       `json:"status"`    // account status, 1:inactive, 2:activated, 3:blocked
	LoginAt   int64     `json:"loginAt"`   // login timestamp
	CreatedAt time.Time `json:"createdAt"` // create time
	UpdatedAt time.Time `json:"updatedAt"` // update time
}

// delete the templates code end

// Create{{.TableNameCamel}}Reply only for api docs
type Create{{.TableNameCamel}}Reply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		{{.ColumnNameCamel}} {{.GoType}} `json:"{{.ColumnNameCamelFCL}}"`
	} `json:"data"` // return data
}

// Delete{{.TableNameCamel}}By{{.ColumnNameCamel}}Reply only for api docs
type Delete{{.TableNameCamel}}By{{.ColumnNameCamel}}Reply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// Update{{.TableNameCamel}}By{{.ColumnNameCamel}}Reply only for api docs
type Update{{.TableNameCamel}}By{{.ColumnNameCamel}}Reply struct {
	Code int      `json:"code"` // return code
	Msg  string   `json:"msg"`  // return information description
	Data struct{} `json:"data"` // return data
}

// Get{{.TableNameCamel}}By{{.ColumnNameCamel}}Reply only for api docs
type Get{{.TableNameCamel}}By{{.ColumnNameCamel}}Reply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		{{.TableNameCamel}} {{.TableNameCamel}}ObjDetail `json:"{{.TableNameCamelFCL}}"`
	} `json:"data"` // return data
}

// List{{.TableNamePluralCamel}}Request request params
type List{{.TableNamePluralCamel}}Request struct {
	query.Params
}

// List{{.TableNamePluralCamel}}Reply only for api docs
type List{{.TableNamePluralCamel}}Reply struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		{{.TableNamePluralCamel}} []{{.TableNameCamel}}ObjDetail `json:"{{.TableNamePluralCamelFCL}}"`
	} `json:"data"` // return data
}
