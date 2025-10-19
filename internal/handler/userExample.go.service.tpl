package handler

import (
	"context"

	serverNameExampleV1 "github.com/go-dev-frame/sponge/api/serverNameExample/v1"
	"github.com/go-dev-frame/sponge/internal/service"
)

var _ serverNameExampleV1.{{.TableNameCamel}}Logicer = (*{{.TableNameCamelFCL}}Handler)(nil)

type {{.TableNameCamelFCL}}Handler struct {
	server serverNameExampleV1.{{.TableNameCamel}}Server
}

// New{{.TableNameCamel}}Handler create a handler
func New{{.TableNameCamel}}Handler() serverNameExampleV1.{{.TableNameCamel}}Logicer {
	return &{{.TableNameCamelFCL}}Handler{
		server: service.New{{.TableNameCamel}}Server(),
	}
}

// Create a new {{.TableNameCamelFCL}}
func (h *{{.TableNameCamelFCL}}Handler) Create(ctx context.Context, req *serverNameExampleV1.Create{{.TableNameCamel}}Request) (*serverNameExampleV1.Create{{.TableNameCamel}}Reply, error) {
	return h.server.Create(ctx, req)
}

// DeleteBy{{.ColumnNameCamel}} delete a {{.TableNameCamelFCL}} by {{.ColumnNameCamelFCL}}
func (h *{{.TableNameCamelFCL}}Handler) DeleteBy{{.ColumnNameCamel}}(ctx context.Context, req *serverNameExampleV1.Delete{{.TableNameCamel}}By{{.ColumnNameCamel}}Request) (*serverNameExampleV1.Delete{{.TableNameCamel}}By{{.ColumnNameCamel}}Reply, error) {
	return h.server.DeleteBy{{.ColumnNameCamel}}(ctx, req)
}

// UpdateBy{{.ColumnNameCamel}} update a {{.TableNameCamelFCL}} by {{.ColumnNameCamelFCL}}
func (h *{{.TableNameCamelFCL}}Handler) UpdateBy{{.ColumnNameCamel}}(ctx context.Context, req *serverNameExampleV1.Update{{.TableNameCamel}}By{{.ColumnNameCamel}}Request) (*serverNameExampleV1.Update{{.TableNameCamel}}By{{.ColumnNameCamel}}Reply, error) {
	return h.server.UpdateBy{{.ColumnNameCamel}}(ctx, req)
}

// GetBy{{.ColumnNameCamel}} get a {{.TableNameCamelFCL}} by {{.ColumnNameCamelFCL}}
func (h *{{.TableNameCamelFCL}}Handler) GetBy{{.ColumnNameCamel}}(ctx context.Context, req *serverNameExampleV1.Get{{.TableNameCamel}}By{{.ColumnNameCamel}}Request) (*serverNameExampleV1.Get{{.TableNameCamel}}By{{.ColumnNameCamel}}Reply, error) {
	return h.server.GetBy{{.ColumnNameCamel}}(ctx, req)
}

// List get a paginated list of {{.TableNamePluralCamelFCL}} by custom conditions
func (h *{{.TableNameCamelFCL}}Handler) List(ctx context.Context, req *serverNameExampleV1.List{{.TableNameCamel}}Request) (*serverNameExampleV1.List{{.TableNameCamel}}Reply, error) {
	return h.server.List(ctx, req)
}
