package handler

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"

	"github.com/go-dev-frame/sponge/pkg/copier"
	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
	"github.com/go-dev-frame/sponge/pkg/gin/middleware"
	"github.com/go-dev-frame/sponge/pkg/logger"

	serverNameExampleV1 "github.com/go-dev-frame/sponge/api/serverNameExample/v1"
	"github.com/go-dev-frame/sponge/internal/cache"
	"github.com/go-dev-frame/sponge/internal/dao"
	"github.com/go-dev-frame/sponge/internal/database"
	"github.com/go-dev-frame/sponge/internal/ecode"
	"github.com/go-dev-frame/sponge/internal/model"
)

var _ serverNameExampleV1.{{.TableNameCamel}}Logicer = (*{{.TableNameCamelFCL}}Handler)(nil)
var _ time.Time

type {{.TableNameCamelFCL}}Handler struct {
	{{.TableNameCamelFCL}}Dao dao.{{.TableNameCamel}}Dao
}

// New{{.TableNameCamel}}Handler create a handler
func New{{.TableNameCamel}}Handler() serverNameExampleV1.{{.TableNameCamel}}Logicer {
	return &{{.TableNameCamelFCL}}Handler{
		{{.TableNameCamelFCL}}Dao: dao.New{{.TableNameCamel}}Dao(
			database.GetDB(), // todo show db driver name here
			cache.New{{.TableNameCamel}}Cache(database.GetCacheType()),
		),
	}
}

// Create a new {{.TableNameCamelFCL}}
func (h *{{.TableNameCamelFCL}}Handler) Create(ctx context.Context, req *serverNameExampleV1.Create{{.TableNameCamel}}Request) (*serverNameExampleV1.Create{{.TableNameCamel}}Reply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), middleware.CtxRequestIDField(ctx))
		return nil, ecode.InvalidParams.Err()
	}

	{{.TableNameCamelFCL}} := &model.{{.TableNameCamel}}{}
	err = copier.Copy({{.TableNameCamelFCL}}, req)
	if err != nil {
		return nil, ecode.ErrCreate{{.TableNameCamel}}.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	err = h.{{.TableNameCamelFCL}}Dao.Create(ctx, {{.TableNameCamelFCL}})
	if err != nil {
		logger.Error("Create error", logger.Err(err), logger.Any("{{.TableNameCamelFCL}}", {{.TableNameCamelFCL}}), middleware.CtxRequestIDField(ctx))
		return nil, ecode.InternalServerError.Err()
	}

	return &serverNameExampleV1.Create{{.TableNameCamel}}Reply{ {{if .IsStandardPrimaryKey}}Id{{else}}{{.ColumnNameCamel}}{{end}}: {{.TableNameCamelFCL}}.{{.ColumnNameCamel}} }, nil
}

// DeleteBy{{.ColumnNameCamel}} delete a {{.TableNameCamelFCL}} by {{.ColumnNameCamelFCL}}
func (h *{{.TableNameCamelFCL}}Handler) DeleteBy{{.ColumnNameCamel}}(ctx context.Context, req *serverNameExampleV1.Delete{{.TableNameCamel}}By{{.ColumnNameCamel}}Request) (*serverNameExampleV1.Delete{{.TableNameCamel}}By{{.ColumnNameCamel}}Reply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), middleware.CtxRequestIDField(ctx))
		return nil, ecode.InvalidParams.Err()
	}

	err = h.{{.TableNameCamelFCL}}Dao.DeleteBy{{.ColumnNameCamel}}(ctx, req.{{if .IsStandardPrimaryKey}}Id{{else}}{{.ColumnNameCamel}}{{end}})
	if err != nil {
		logger.Warn("DeleteBy{{.ColumnNameCamel}} error", logger.Err(err), middleware.CtxRequestIDField(ctx))
		return nil, ecode.InternalServerError.Err()
	}

	return &serverNameExampleV1.Delete{{.TableNameCamel}}By{{.ColumnNameCamel}}Reply{}, nil
}

// UpdateBy{{.ColumnNameCamel}} update a {{.TableNameCamelFCL}} by {{.ColumnNameCamelFCL}}
func (h *{{.TableNameCamelFCL}}Handler) UpdateBy{{.ColumnNameCamel}}(ctx context.Context, req *serverNameExampleV1.Update{{.TableNameCamel}}By{{.ColumnNameCamel}}Request) (*serverNameExampleV1.Update{{.TableNameCamel}}By{{.ColumnNameCamel}}Reply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), middleware.CtxRequestIDField(ctx))
		return nil, ecode.InvalidParams.Err()
	}

	{{.TableNameCamelFCL}} := &model.{{.TableNameCamel}}{}
	err = copier.Copy({{.TableNameCamelFCL}}, req)
	if err != nil {
		return nil, ecode.ErrUpdateBy{{.ColumnNameCamel}}{{.TableNameCamel}}.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here
	{{.TableNameCamelFCL}}.{{.ColumnNameCamel}} = req.{{if .IsStandardPrimaryKey}}Id{{else}}{{.ColumnNameCamel}}{{end}}

	err = h.{{.TableNameCamelFCL}}Dao.UpdateBy{{.ColumnNameCamel}}(ctx, {{.TableNameCamelFCL}})
	if err != nil {
		logger.Error("UpdateBy{{.ColumnNameCamel}} error", logger.Err(err), logger.Any("{{.TableNameCamelFCL}}", {{.TableNameCamelFCL}}), middleware.CtxRequestIDField(ctx))
		return nil, ecode.InternalServerError.Err()
	}

	return &serverNameExampleV1.Update{{.TableNameCamel}}By{{.ColumnNameCamel}}Reply{}, nil
}

// GetBy{{.ColumnNameCamel}} get a {{.TableNameCamelFCL}} by {{.ColumnNameCamelFCL}}
func (h *{{.TableNameCamelFCL}}Handler) GetBy{{.ColumnNameCamel}}(ctx context.Context, req *serverNameExampleV1.Get{{.TableNameCamel}}By{{.ColumnNameCamel}}Request) (*serverNameExampleV1.Get{{.TableNameCamel}}By{{.ColumnNameCamel}}Reply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), middleware.CtxRequestIDField(ctx))
		return nil, ecode.InvalidParams.Err()
	}

	record, err := h.{{.TableNameCamelFCL}}Dao.GetBy{{.ColumnNameCamel}}(ctx, req.{{if .IsStandardPrimaryKey}}Id{{else}}{{.ColumnNameCamel}}{{end}})
	if err != nil {
		if errors.Is(err, database.ErrRecordNotFound) {
			logger.Warn("GetBy{{.ColumnNameCamel}} error", logger.Err(err), logger.Any("{{.ColumnNameCamelFCL}}", req.{{if .IsStandardPrimaryKey}}Id{{else}}{{.ColumnNameCamel}}{{end}}), middleware.CtxRequestIDField(ctx))
			return nil, ecode.NotFound.Err()
		}
		logger.Error("GetBy{{.ColumnNameCamel}} error", logger.Err(err), logger.Any("{{.ColumnNameCamelFCL}}", req.{{if .IsStandardPrimaryKey}}Id{{else}}{{.ColumnNameCamel}}{{end}}), middleware.CtxRequestIDField(ctx))
		return nil, ecode.InternalServerError.Err()
	}

	data, err := convert{{.TableNameCamel}}Pb(record)
	if err != nil {
		logger.Warn("convert{{.TableNameCamel}} error", logger.Err(err), logger.Any("{{.TableNameCamelFCL}}", record), middleware.CtxRequestIDField(ctx))
		return nil, ecode.ErrGetBy{{.ColumnNameCamel}}{{.TableNameCamel}}.Err()
	}

	return &serverNameExampleV1.Get{{.TableNameCamel}}By{{.ColumnNameCamel}}Reply{
		{{.TableNameCamel}}: data,
	}, nil
}

// List get a paginated list of {{.TableNamePluralCamelFCL}} by custom conditions
func (h *{{.TableNameCamelFCL}}Handler) List(ctx context.Context, req *serverNameExampleV1.List{{.TableNameCamel}}Request) (*serverNameExampleV1.List{{.TableNameCamel}}Reply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), middleware.CtxRequestIDField(ctx))
		return nil, ecode.InvalidParams.Err()
	}

	params := &query.Params{}
	err = copier.Copy(params, req.Params)
	if err != nil {
		return nil, ecode.ErrList{{.TableNameCamel}}.Err()
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	records, total, err := h.{{.TableNameCamelFCL}}Dao.GetByColumns(ctx, params)
	if err != nil {
		if strings.Contains(err.Error(), "query params error:") {
			logger.Warn("GetByColumns error", logger.Err(err), logger.Any("params", params), middleware.CtxRequestIDField(ctx))
			return nil, ecode.InvalidParams.Err()
		}
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("params", params), middleware.CtxRequestIDField(ctx))
		return nil, ecode.InternalServerError.Err()
	}

	{{.TableNamePluralCamelFCL}} := []*serverNameExampleV1.{{.TableNameCamel}}{}
	for _, record := range records {
		data, err := convert{{.TableNameCamel}}Pb(record)
		if err != nil {
			logger.Warn("convert{{.TableNameCamel}} error", logger.Err(err), logger.Any("{{.ColumnNameCamelFCL}}", record.{{.ColumnNameCamel}}), middleware.CtxRequestIDField(ctx))
			continue
		}
		{{.TableNamePluralCamelFCL}} = append({{.TableNamePluralCamelFCL}}, data)
	}

	return &serverNameExampleV1.List{{.TableNameCamel}}Reply{
		Total:        total,
		{{.TableNamePluralCamel}}: {{.TableNamePluralCamelFCL}},
	}, nil
}

// DeleteBy{{.ColumnNamePluralCamel}} batch delete {{.TableNamePluralCamelFCL}} by {{.ColumnNamePluralCamelFCL}}
func (h *{{.TableNameCamelFCL}}Handler) DeleteBy{{.ColumnNamePluralCamel}}(ctx context.Context, req *serverNameExampleV1.Delete{{.TableNameCamel}}By{{.ColumnNamePluralCamel}}Request) (*serverNameExampleV1.Delete{{.TableNameCamel}}By{{.ColumnNamePluralCamel}}Reply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), middleware.CtxRequestIDField(ctx))
		return nil, ecode.InvalidParams.Err()
	}

	err = h.{{.TableNameCamelFCL}}Dao.DeleteBy{{.ColumnNamePluralCamel}}(ctx, req.{{if .IsStandardPrimaryKey}}Id{{else}}{{.ColumnNameCamel}}{{end}}s)
	if err != nil {
		logger.Warn("DeleteBy{{.ColumnNamePluralCamel}} error", logger.Err(err), middleware.CtxRequestIDField(ctx))
		return nil, ecode.InternalServerError.Err()
	}

	return &serverNameExampleV1.Delete{{.TableNameCamel}}By{{.ColumnNamePluralCamel}}Reply{}, nil
}

// GetByCondition get a {{.TableNameCamelFCL}} by custom condition
func (h *{{.TableNameCamelFCL}}Handler) GetByCondition(ctx context.Context, req *serverNameExampleV1.Get{{.TableNameCamel}}ByConditionRequest) (*serverNameExampleV1.Get{{.TableNameCamel}}ByConditionReply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), middleware.CtxRequestIDField(ctx))
		return nil, ecode.InvalidParams.Err()
	}

	conditions := &query.Conditions{}
	for _, v := range req.Conditions.GetColumns() {
		column := query.Column{}
		_ = copier.Copy(&column, v)
		conditions.Columns = append(conditions.Columns, column)
	}
	err = conditions.CheckValid()
	if err != nil {
		logger.Warn("Parameters error", logger.Err(err), logger.Any("conditions", conditions), middleware.CtxRequestIDField(ctx))
		return nil, ecode.InvalidParams.Err()
	}

	record, err := h.{{.TableNameCamelFCL}}Dao.GetByCondition(ctx, conditions)
	if err != nil {
		if errors.Is(err, database.ErrRecordNotFound) {
			logger.Warn("GetBy{{.ColumnNameCamel}} error", logger.Err(err), logger.Any("req", req), middleware.CtxRequestIDField(ctx))
			return nil, ecode.NotFound.Err()
		}
		logger.Error("GetBy{{.ColumnNameCamel}} error", logger.Err(err), logger.Any("req", req), middleware.CtxRequestIDField(ctx))
		return nil, ecode.InternalServerError.Err()
	}

	data, err := convert{{.TableNameCamel}}Pb(record)
	if err != nil {
		logger.Warn("convert{{.TableNameCamel}} error", logger.Err(err), logger.Any("{{.TableNameCamelFCL}}", record), middleware.CtxRequestIDField(ctx))
		return nil, ecode.ErrGetBy{{.ColumnNameCamel}}{{.TableNameCamel}}.Err()
	}

	return &serverNameExampleV1.Get{{.TableNameCamel}}ByConditionReply{
		{{.TableNameCamel}}: data,
	}, nil
}

// ListBy{{.ColumnNamePluralCamel}} batch get {{.TableNamePluralCamelFCL}} by {{.ColumnNamePluralCamelFCL}}
func (h *{{.TableNameCamelFCL}}Handler) ListBy{{.ColumnNamePluralCamel}}(ctx context.Context, req *serverNameExampleV1.List{{.TableNameCamel}}By{{.ColumnNamePluralCamel}}Request) (*serverNameExampleV1.List{{.TableNameCamel}}By{{.ColumnNamePluralCamel}}Reply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), middleware.CtxRequestIDField(ctx))
		return nil, ecode.InvalidParams.Err()
	}

	{{.TableNameCamelFCL}}Map, err := h.{{.TableNameCamelFCL}}Dao.GetBy{{.ColumnNamePluralCamel}}(ctx, req.{{if .IsStandardPrimaryKey}}Id{{else}}{{.ColumnNameCamel}}{{end}}s)
	if err != nil {
		logger.Error("GetBy{{.ColumnNamePluralCamel}} error", logger.Err(err), logger.Any("{{.ColumnNamePluralCamelFCL}}", req.{{if .IsStandardPrimaryKey}}Id{{else}}{{.ColumnNameCamel}}{{end}}s), middleware.CtxRequestIDField(ctx))
		return nil, ecode.InternalServerError.Err()
	}

	{{.TableNamePluralCamelFCL}} := []*serverNameExampleV1.{{.TableNameCamel}}{}
	for _, {{.ColumnNameCamelFCL}} := range req.{{if .IsStandardPrimaryKey}}Id{{else}}{{.ColumnNameCamel}}{{end}}s {
		if v, ok := {{.TableNameCamelFCL}}Map[{{.ColumnNameCamelFCL}}]; ok {
			record, err := convert{{.TableNameCamel}}Pb(v)
			if err != nil {
				logger.Warn("convert{{.TableNameCamel}} error", logger.Err(err), logger.Any("{{.TableNameCamelFCL}}", v), middleware.CtxRequestIDField(ctx))
				return nil, ecode.InternalServerError.Err()
			}
			{{.TableNamePluralCamelFCL}} = append({{.TableNamePluralCamelFCL}}, record)
		}
	}

	return &serverNameExampleV1.List{{.TableNameCamel}}By{{.ColumnNamePluralCamel}}Reply{
		{{.TableNamePluralCamel}}: {{.TableNamePluralCamelFCL}},
	}, nil
}

// ListByLast{{.ColumnNameCamel}} get a paginated list of {{.TableNamePluralCamelFCL}} by last {{.ColumnNameCamelFCL}}
func (h *{{.TableNameCamelFCL}}Handler) ListByLast{{.ColumnNameCamel}}(ctx context.Context, req *serverNameExampleV1.List{{.TableNameCamel}}ByLast{{.ColumnNameCamel}}Request) (*serverNameExampleV1.List{{.TableNameCamel}}ByLast{{.ColumnNameCamel}}Reply, error) {
	err := req.Validate()
	if err != nil {
		logger.Warn("req.Validate error", logger.Err(err), logger.Any("req", req), middleware.CtxRequestIDField(ctx))
		return nil, ecode.InvalidParams.Err()
	}
	{{if .IsStringType}}if req.Last{{.ColumnNameCamel}} == "" {
		req.Last{{.ColumnNameCamel}} = "zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz"
	}{{else}}
	if req.Last{{.ColumnNameCamel}} == 0 {
		req.Last{{.ColumnNameCamel}} = math.MaxInt32
	}{{end}}

	records, err := h.{{.TableNameCamelFCL}}Dao.GetByLast{{.ColumnNameCamel}}(ctx, req.Last{{.ColumnNameCamel}}, int(req.Limit), req.Sort)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("req", req), middleware.CtxRequestIDField(ctx))
		return nil, ecode.InternalServerError.Err()
	}

	{{.TableNamePluralCamelFCL}} := []*serverNameExampleV1.{{.TableNameCamel}}{}
	for _, record := range records {
		data, err := convert{{.TableNameCamel}}Pb(record)
		if err != nil {
			logger.Warn("convert{{.TableNameCamel}} error", logger.Err(err), logger.Any("{{.ColumnNameCamelFCL}}", record.{{.ColumnNameCamel}}), middleware.CtxRequestIDField(ctx))
			continue
		}
		{{.TableNamePluralCamelFCL}} = append({{.TableNamePluralCamelFCL}}, data)
	}

	return &serverNameExampleV1.List{{.TableNameCamel}}ByLast{{.ColumnNameCamel}}Reply{
		{{.TableNamePluralCamel}}: {{.TableNamePluralCamelFCL}},
	}, nil
}

func convert{{.TableNameCamel}}Pb(record *model.{{.TableNameCamel}}) (*serverNameExampleV1.{{.TableNameCamel}}, error) {
	value := &serverNameExampleV1.{{.TableNameCamel}}{}
	err := copier.Copy(value, record)
	if err != nil {
		return nil, err
	}
	// Note: if copier.Copy cannot assign a value to a field, add it here

	return value, nil
}
