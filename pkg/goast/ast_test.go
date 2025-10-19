package goast

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseGoFile(t *testing.T) {
	astInfos, err := ParseFile("ast.go")
	assert.NoError(t, err)
	assert.Greater(t, len(astInfos), 10)
}

func TestParseGoCode(t *testing.T) {
	var src = `
package main

import "fmt"
import "strings"
import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

const (
	pi = 3.14
	language = "Go"
)

var (
	version = "v1.0.0"
	repo  = "sponge"
)

type User struct {
	Name string
	Age  int
}

func (u *User) SayHello() {
	fmt.Println("Hello, my name is", u.Name)
}

func main() {
	fmt.Println(pi)
	fmt.Println(language)
	fmt.Println(version)

	user:=&User{Name:"Tom",Age:20}
	fmt.Println(user.Name)
	fmt.Println(user.Age)
	user.SayHello()
}
`

	astInfos, err := ParseGoCode("", []byte(src))
	assert.NoError(t, err)
	for _, info := range astInfos {
		fmt.Printf("    %-20s: %s\n", "name", info.Names)
		fmt.Printf("    %-20s: %s\n", "comment", info.Comment)
		fmt.Printf("    %-20s: %s\n\n\n", "body", info.Body)
	}
}

func TestParseImportGroup(t *testing.T) {
	body := `
import (
	"fmt"
	"github.com/gin-gonic/gin"
	//"github.com/spf13/viper"
	apiV1   "yourModuleName/api/v1"
	// api v2
	apiV2   "yourModuleName/api/v2"
)
`

	importInfos, err := ParseImportGroup(body)
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, ii := range importInfos {
		fmt.Printf("    %-20s: %s\n", "name", ii.Path)
		fmt.Printf("    %-20s: %s\n", "alias", ii.Alias)
		fmt.Printf("    %-20s: %s\n", "comment", ii.Comment)
		fmt.Printf("    %-20s: %s\n\n\n", "body", ii.Body)
	}
}

func TestParseConstGroup(t *testing.T) {
	body := `
// pi constant
const pi = 3.14

const (
	// Version number
	version = "v1.0.0"
)

const (
	// Development language
	language = "Go"

	// database type
	dbDriver = "mysql"
)
`

	constInfos, err := ParseConstGroup(body)
	if err != nil {
		assert.NotNil(t, err)
		return
	}
	for _, ci := range constInfos {
		fmt.Printf("    %-20s: %s\n", "name", ci.Name)
		fmt.Printf("    %-20s: %s\n", "value", ci.Value)
		fmt.Printf("    %-20s: %s\n", "comment", ci.Comment)
		fmt.Printf("    %-20s: %s\n\n\n", "body", ci.Body)
	}
}

func TestParseVarGroup(t *testing.T) {
	body := `
var (
	// Version number
	version = "v1.0.0"

	// Author
	author  = "name"

	// Repository
	repo  = "sponge"

	// Function variable
	f1 = func() {
		fmt.Println("hello")
	}
)
`

	varInfos, err := ParseVarGroup(body)
	if err != nil {
		assert.NotNil(t, err)
		return
	}
	for _, vi := range varInfos {
		fmt.Printf("    %-20s: %s\n", "name", vi.Name)
		fmt.Printf("    %-20s: %s\n", "value", vi.Value)
		fmt.Printf("    %-20s: %s\n", "comment", vi.Comment)
		fmt.Printf("    %-20s: %s\n\n\n", "body", vi.Body)
	}
}

func TestParseTypeGroup(t *testing.T) {
	body := `
type (
	// Struct type
	ts struct {
		name string
	}

	// Function type
	tfn func(name string) bool

	// Interface type
	iFace interface {}

	// Channel type
	ch chan int

	// Map type
	m map[string]bool

	// Slice type
	slice []int
)
`

	typeInfos, err := ParseTypeGroup(body)
	if err != nil {
		assert.NotNil(t, err)
		return
	}

	for _, ti := range typeInfos {
		fmt.Printf("    %-20s: %s\n", "type", ti.Type)
		fmt.Printf("    %-20s: %s\n", "name", ti.Name)
		fmt.Printf("    %-20s: %s\n", "comment", ti.Comment)
		fmt.Printf("    %-20s: %s\n\n\n", "body", ti.Body)
	}
}

func TestParseInterface(t *testing.T) {
	body := `
type GreeterDao interface {
	// get by id
	Create(ctx context.Context, table *model.Greeter) error
	// delete by id
	DeleteByID(ctx context.Context, id uint64) error
	// update by id
	UpdateByID(ctx context.Context, table *model.Greeter) error
	UserExampleDao
}

type UserExampleDao interface {
	// get by id
	Create(ctx context.Context, table *model.UserExample) error
	// update by id
	UpdateByID(ctx context.Context, table *model.UserExample) error
}
`

	interfaceInfos, err := ParseInterface(body)
	if err != nil {
		assert.NotNil(t, err)
		return
	}
	for _, info := range interfaceInfos {
		fmt.Printf("%-20s     : %s\n", "name", info.Name)
		fmt.Printf("%-20s     : %s\n", "comment", info.Comment)
		for _, mi := range info.MethodInfos {
			fmt.Printf("    %-20s: %s\n", "method name", mi.Name)
			fmt.Printf("    %-20s: %s\n", "comment", mi.Comment)
			fmt.Printf("    %-20s: %s\n", "body", mi.Body)
			fmt.Printf("    %-20v: %t\n\n\n", "embedded", mi.IsIdent)
		}
	}
}

func TestParseStructMethods(t *testing.T) {
	src := `
package demo

type userHandler struct {
	server userV1.UserServer
}

// Create a record
func (h *userHandler) Create(ctx context.Context, req *userV1.CreateUserRequest) (*userV1.CreateUserReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a record by id
func (h *userHandler) DeleteByID(ctx context.Context, req *userV1.DeleteUserByIDRequest) (*userV1.DeleteUserByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}

// UpdateByID update a record by id
func (h *userHandler) UpdateByID(ctx context.Context, req *userV1.UpdateUserByIDRequest) (*userV1.UpdateUserByIDReply, error) {
	return h.server.UpdateByID(ctx, req)
}

type greeterHandler struct {
	server greeterV1.GreeterServer
}

// Create a record
func (h *greeterHandler) Create(ctx context.Context, req *greeterV1.CreateGreeterRequest) (*greeterV1.CreateGreeterReply, error) {
	return h.server.Create(ctx, req)
}

// DeleteByID delete a record by id
func (h *greeterHandler) DeleteByID(ctx context.Context, req *greeterV1.DeleteGreeterByIDRequest) (*greeterV1.DeleteGreeterByIDReply, error) {
	return h.server.DeleteByID(ctx, req)
}
`
	astInfos, err := ParseGoCode("", []byte(src))
	if err != nil {
		assert.NotNil(t, err)
		return
	}

	methods := ParseStructMethods(astInfos)
	for structName, methodInfos := range methods {
		fmt.Printf("%-20s     : %s\n", "name", structName)
		for _, mi := range methodInfos {
			fmt.Printf("    %-20s: %s\n", "method name", mi.Name)
			fmt.Printf("    %-20s: %s\n", "comment", mi.Comment)
			fmt.Printf("    %-20s: %s\n\n\n", "body", mi.Body)
		}
	}
}

func TestParseStruct(t *testing.T) {
	body := `
package goast

// AstInfo is the information of a code block.
type AstInfo struct {
	Kind string

	// Names is the name of the code block, such as "func Name", "type Names", "const Names", "var Names", "import Paths".
	// If Type is "func", a standalone function without a receiver has a single name.
	// If the function is a method belonging to a struct, it has two names: the first
	// represents the function name, and the second represents the struct name.
	Names []string // todo add name

	// User information
	User struct {
		Name string
		Age string
	}

	// embedded struct
	*Address
	// embedded struct2
	Address

	reader interface {}
	writer any
	sayMap map[string]string
	ch1 chan int
	ch2 chan *Address
}

// Address is address
type Address struct {
	State    string
	// Addr is address
	Addr string
}
`
	structInfos, err := ParseStruct(body)
	if err != nil {
		t.Error(err)
		return
	}
	for name, structInfo := range structInfos {
		fmt.Printf("%-20s     : %s\n", "struct name", name)
		fmt.Printf("%-20s     : %s\n", "struct comment", structInfo.Comment)
		for _, field := range structInfo.Fields {
			fmt.Printf("    %-20s: %s\n", "name", field.Name)
			fmt.Printf("    %-20s: %s\n", "type", field.Type)
			fmt.Printf("    %-20s: %s\n", "comment", field.Comment)
			fmt.Printf("    %-20s: %s\n\n", "body", field.Body)
		}
		fmt.Printf("\n\n\n")
	}
}
