package dao

import (
	"context"
	"errors"
	"time"

	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"

	"github.com/go-dev-frame/sponge/pkg/logger"
	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
	"github.com/go-dev-frame/sponge/pkg/utils"

	"github.com/go-dev-frame/sponge/internal/cache"
	"github.com/go-dev-frame/sponge/internal/database"
	"github.com/go-dev-frame/sponge/internal/model"
)

var _ {{.TableNameCamel}}Dao = (*{{.TableNameCamelFCL}}Dao)(nil)

// {{.TableNameCamel}}Dao defining the dao interface
type {{.TableNameCamel}}Dao interface {
	Create(ctx context.Context, table *model.{{.TableNameCamel}}) error
	DeleteBy{{.ColumnNameCamel}}(ctx context.Context, {{.ColumnNameCamelFCL}} {{.GoType}}) error
	UpdateBy{{.ColumnNameCamel}}(ctx context.Context, table *model.{{.TableNameCamel}}) error
	GetBy{{.ColumnNameCamel}}(ctx context.Context, {{.ColumnNameCamelFCL}} {{.GoType}}) (*model.{{.TableNameCamel}}, error)
	GetByColumns(ctx context.Context, params *query.Params) ([]*model.{{.TableNameCamel}}, int64, error)

	DeleteBy{{.ColumnNamePluralCamel}}(ctx context.Context, {{.ColumnNamePluralCamelFCL}} []{{.GoType}}) error
	GetByCondition(ctx context.Context, condition *query.Conditions) (*model.{{.TableNameCamel}}, error)
	GetBy{{.ColumnNamePluralCamel}}(ctx context.Context, {{.ColumnNamePluralCamelFCL}} []{{.GoType}}) (map[{{.GoType}}]*model.{{.TableNameCamel}}, error)
	GetByLast{{.ColumnNameCamel}}(ctx context.Context, last{{.ColumnNameCamel}} {{.GoType}}, limit int, sort string) ([]*model.{{.TableNameCamel}}, error)

	CreateByTx(ctx context.Context, tx *gorm.DB, table *model.{{.TableNameCamel}}) ({{.GoType}}, error)
	DeleteByTx(ctx context.Context, tx *gorm.DB, {{.ColumnNameCamelFCL}} {{.GoType}}) error
	UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.{{.TableNameCamel}}) error
}

type {{.TableNameCamelFCL}}Dao struct {
	db    *gorm.DB
	cache cache.{{.TableNameCamel}}Cache // if nil, the cache is not used.
	sfg   *singleflight.Group    // if cache is nil, the sfg is not used.
}

// New{{.TableNameCamel}}Dao creating the dao interface
func New{{.TableNameCamel}}Dao(db *gorm.DB, xCache cache.{{.TableNameCamel}}Cache) {{.TableNameCamel}}Dao {
	if xCache == nil {
		return &{{.TableNameCamelFCL}}Dao{db: db}
	}
	return &{{.TableNameCamelFCL}}Dao{
		db:    db,
		cache: xCache,
		sfg:   new(singleflight.Group),
	}
}

func (d *{{.TableNameCamelFCL}}Dao) deleteCache(ctx context.Context, {{.ColumnNameCamelFCL}} {{.GoType}}) error {
	if d.cache != nil {
		return d.cache.Del(ctx, {{.ColumnNameCamelFCL}})
	}
	return nil
}

// Create a new {{.TableNameCamelFCL}}, insert the record and the {{.ColumnNameCamelFCL}} value is written back to the table
func (d *{{.TableNameCamelFCL}}Dao) Create(ctx context.Context, table *model.{{.TableNameCamel}}) error {
	return d.db.WithContext(ctx).Create(table).Error
}

// DeleteBy{{.ColumnNameCamel}} delete a {{.TableNameCamelFCL}} by {{.ColumnNameCamelFCL}}
func (d *{{.TableNameCamelFCL}}Dao) DeleteBy{{.ColumnNameCamel}}(ctx context.Context, {{.ColumnNameCamelFCL}} {{.GoType}}) error {
	err := d.db.WithContext(ctx).Where("{{.ColumnName}} = ?", {{.ColumnNameCamelFCL}}).Delete(&model.{{.TableNameCamel}}{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, {{.ColumnNameCamelFCL}})

	return nil
}

// UpdateBy{{.ColumnNameCamel}} update a {{.TableNameCamelFCL}} by {{.ColumnNameCamelFCL}}
func (d *{{.TableNameCamelFCL}}Dao) UpdateBy{{.ColumnNameCamel}}(ctx context.Context, table *model.{{.TableNameCamel}}) error {
	err := d.updateDataBy{{.ColumnNameCamel}}(ctx, d.db, table)

	// delete cache
	_ = d.deleteCache(ctx, table.{{.ColumnNameCamel}})

	return err
}

func (d *{{.TableNameCamelFCL}}Dao) updateDataBy{{.ColumnNameCamel}}(ctx context.Context, db *gorm.DB, table *model.{{.TableNameCamel}}) error {
	{{if .IsStringType}}if table.{{.ColumnNameCamel}} == "" {
		return errors.New("{{.ColumnNameCamelFCL}} cannot be empty")
	}
{{else}}	if table.{{.ColumnNameCamel}} < 1 {
		return errors.New("{{.ColumnNameCamelFCL}} cannot be 0")
	}
{{end}}

	update := map[string]interface{}{}
	// todo generate the update fields code to here
	// delete the templates code start
	if table.Name != "" {
		update["name"] = table.Name
	}
	if table.Password != "" {
		update["password"] = table.Password
	}
	if table.Email != "" {
		update["email"] = table.Email
	}
	if table.Phone != "" {
		update["phone"] = table.Phone
	}
	if table.Avatar != "" {
		update["avatar"] = table.Avatar
	}
	if table.Age > 0 {
		update["age"] = table.Age
	}
	if table.Gender > 0 {
		update["gender"] = table.Gender
	}
	if table.LoginAt > 0 {
		update["login_at"] = table.LoginAt
	}
	// delete the templates code end

	return db.WithContext(ctx).Model(table).Updates(update).Error
}

// GetBy{{.ColumnNameCamel}} get a {{.TableNameCamelFCL}} by {{.ColumnNameCamelFCL}}
func (d *{{.TableNameCamelFCL}}Dao) GetBy{{.ColumnNameCamel}}(ctx context.Context, {{.ColumnNameCamelFCL}} {{.GoType}}) (*model.{{.TableNameCamel}}, error) {
	// no cache
	if d.cache == nil {
		record := &model.{{.TableNameCamel}}{}
		err := d.db.WithContext(ctx).Where("{{.ColumnName}} = ?", {{.ColumnNameCamelFCL}}).First(record).Error
		return record, err
	}

	// get from cache
	record, err := d.cache.Get(ctx, {{.ColumnNameCamelFCL}})
	if err == nil {
		return record, nil
	}

	// get from database
	if errors.Is(err, database.ErrCacheNotFound) {
		// for the same {{.ColumnNameCamelFCL}}, prevent high concurrent simultaneous access to database
		{{if .IsStringType}}val, err, _ := d.sfg.Do({{.ColumnNameCamelFCL}}, func() (interface{}, error) {
{{else}}		val, err, _ := d.sfg.Do(utils.{{.GoTypeFCU}}ToStr({{.ColumnNameCamelFCL}}), func() (interface{}, error) {
{{end}}
			table := &model.{{.TableNameCamel}}{}
			err = d.db.WithContext(ctx).Where("{{.ColumnName}} = ?", {{.ColumnNameCamelFCL}}).First(table).Error
			if err != nil {
				// set placeholder cache to prevent cache penetration, default expiration time 10 minutes
				if errors.Is(err, database.ErrRecordNotFound) {
					if err = d.cache.SetPlaceholder(ctx, {{.ColumnNameCamelFCL}}); err != nil {
						logger.Warn("cache.SetPlaceholder error", logger.Err(err), logger.Any("{{.ColumnNameCamelFCL}}", {{.ColumnNameCamelFCL}}))
					}
					return nil, database.ErrRecordNotFound
				}
				return nil, err
			}
			// set cache
			if err = d.cache.Set(ctx, {{.ColumnNameCamelFCL}}, table, cache.{{.TableNameCamel}}ExpireTime); err != nil {
				logger.Warn("cache.Set error", logger.Err(err), logger.Any("{{.ColumnNameCamelFCL}}", {{.ColumnNameCamelFCL}}))
			}
			return table, nil
		})
		if err != nil {
			return nil, err
		}
		table, ok := val.(*model.{{.TableNameCamel}})
		if !ok {
			return nil, database.ErrRecordNotFound
		}
		return table, nil
	}

	if d.cache.IsPlaceholderErr(err) {
		return nil, database.ErrRecordNotFound
	}

	return nil, err
}

// GetByColumns get a paginated list of {{.TableNamePluralCamelFCL}} by custom conditions.
// For more details, please refer to https://go-sponge.com/component/data/custom-page-query.html
func (d *{{.TableNameCamelFCL}}Dao) GetByColumns(ctx context.Context, params *query.Params) ([]*model.{{.TableNameCamel}}, int64, error) {
	if params.Sort == "" {
		params.Sort = "-{{.ColumnName}}"
	}
	queryStr, args, err := params.ConvertToGormConditions(query.WithWhitelistNames(model.{{.TableNameCamel}}ColumnNames))
	if err != nil {
		return nil, 0, errors.New("query params error: " + err.Error())
	}

	var total int64
	if params.Sort != "ignore count" { // determine if count is required
		err = d.db.WithContext(ctx).Model(&model.{{.TableNameCamel}}{}).Where(queryStr, args...).Count(&total).Error
		if err != nil {
			return nil, 0, err
		}
		if total == 0 {
			return nil, total, nil
		}
	}

	records := []*model.{{.TableNameCamel}}{}
	order, limit, offset := params.ConvertToPage()
	err = d.db.WithContext(ctx).Order(order).Limit(limit).Offset(offset).Where(queryStr, args...).Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, err
}

// DeleteBy{{.ColumnNamePluralCamel}} batch delete {{.TableNamePluralCamelFCL}} by {{.ColumnNamePluralCamelFCL}}
func (d *{{.TableNameCamelFCL}}Dao) DeleteBy{{.ColumnNamePluralCamel}}(ctx context.Context, {{.ColumnNamePluralCamelFCL}} []{{.GoType}}) error {
	err := d.db.WithContext(ctx).Where("{{.ColumnName}} IN (?)", {{.ColumnNamePluralCamelFCL}}).Delete(&model.{{.TableNameCamel}}{}).Error
	if err != nil {
		return err
	}

	// delete cache
	for _, {{.ColumnNameCamelFCL}} := range {{.ColumnNamePluralCamelFCL}} {
		_ = d.deleteCache(ctx, {{.ColumnNameCamelFCL}})
	}

	return nil
}

// GetByCondition get a {{.TableNameCamelFCL}} by custom condition
// For more details, please refer to https://go-sponge.com/component/data/custom-page-query.html#_2-condition-parameters-optional
func (d *{{.TableNameCamelFCL}}Dao) GetByCondition(ctx context.Context, c *query.Conditions) (*model.{{.TableNameCamel}}, error) {
	queryStr, args, err := c.ConvertToGorm(query.WithWhitelistNames(model.{{.TableNameCamel}}ColumnNames))
	if err != nil {
		return nil, err
	}

	table := &model.{{.TableNameCamel}}{}
	err = d.db.WithContext(ctx).Where(queryStr, args...).First(table).Error
	if err != nil {
		return nil, err
	}

	return table, nil
}

// GetBy{{.ColumnNamePluralCamel}} batch get {{.TableNamePluralCamelFCL}} by {{.ColumnNamePluralCamelFCL}}
func (d *{{.TableNameCamelFCL}}Dao) GetBy{{.ColumnNamePluralCamel}}(ctx context.Context, {{.ColumnNamePluralCamelFCL}} []{{.GoType}}) (map[{{.GoType}}]*model.{{.TableNameCamel}}, error) {
	// no cache
	if d.cache == nil {
		var records []*model.{{.TableNameCamel}}
		err := d.db.WithContext(ctx).Where("{{.ColumnName}} IN (?)", {{.ColumnNamePluralCamelFCL}}).Find(&records).Error
		if err != nil {
			return nil, err
		}
		itemMap := make(map[{{.GoType}}]*model.{{.TableNameCamel}})
		for _, record := range records {
			itemMap[record.{{.ColumnNameCamel}}] = record
		}
		return itemMap, nil
	}

	// get form cache
	itemMap, err := d.cache.MultiGet(ctx, {{.ColumnNamePluralCamelFCL}})
	if err != nil {
		return nil, err
	}

	var missed{{.ColumnNamePluralCamel}} []{{.GoType}}
	for _, {{.ColumnNameCamelFCL}} := range {{.ColumnNamePluralCamelFCL}} {
		if _, ok := itemMap[{{.ColumnNameCamelFCL}}]; !ok {
			missed{{.ColumnNamePluralCamel}} = append(missed{{.ColumnNamePluralCamel}}, {{.ColumnNameCamelFCL}})
		}
	}

	// get missed data
	if len(missed{{.ColumnNamePluralCamel}}) > 0 {
		// find the {{.ColumnNameCamelFCL}} of an active placeholder, i.e. an {{.ColumnNameCamelFCL}} that does not exist in database
		var realMissed{{.ColumnNamePluralCamel}} []{{.GoType}}
		for _, {{.ColumnNameCamelFCL}} := range missed{{.ColumnNamePluralCamel}} {
			_, err = d.cache.Get(ctx, {{.ColumnNameCamelFCL}})
			if d.cache.IsPlaceholderErr(err) {
				continue
			}
			realMissed{{.ColumnNamePluralCamel}} = append(realMissed{{.ColumnNamePluralCamel}}, {{.ColumnNameCamelFCL}})
		}

		if len(realMissed{{.ColumnNamePluralCamel}}) > 0 {
			var records []*model.{{.TableNameCamel}}
			var record{{.ColumnNameCamel}}Map = make(map[{{.GoType}}]struct{})
			err = d.db.WithContext(ctx).Where("{{.ColumnName}} IN (?)", realMissed{{.ColumnNamePluralCamel}}).Find(&records).Error
			if err != nil {
				return nil, err
			}

			if len(records) > 0 {
				for _, record := range records {
					itemMap[record.{{.ColumnNameCamel}}] = record
					record{{.ColumnNameCamel}}Map[record.{{.ColumnNameCamel}}] = struct{}{}
				}
				err = d.cache.MultiSet(ctx, records, cache.{{.TableNameCamel}}ExpireTime)
				if err != nil {
					logger.Warn("cache.MultiSet error", logger.Err(err), logger.Any("{{.ColumnNamePluralCamelFCL}}", records))
				}
				if len(records) == len(realMissed{{.ColumnNamePluralCamel}}) {
					return itemMap, nil
				}
			}
			for _, {{.ColumnNameCamelFCL}} := range realMissed{{.ColumnNamePluralCamel}} {
				if _, ok := record{{.ColumnNameCamel}}Map[{{.ColumnNameCamelFCL}}]; !ok {
					if err = d.cache.SetPlaceholder(ctx, {{.ColumnNameCamelFCL}}); err != nil {
						logger.Warn("cache.SetPlaceholder error", logger.Err(err), logger.Any("{{.ColumnNameCamelFCL}}", {{.ColumnNameCamelFCL}}))
					}
				}
			}
		}
	}

	return itemMap, nil
}

// GetByLast{{.ColumnNameCamel}} get a paginated list of {{.TableNamePluralCamelFCL}} by last {{.ColumnNameCamelFCL}}
func (d *{{.TableNameCamelFCL}}Dao) GetByLast{{.ColumnNameCamel}}(ctx context.Context, last{{.ColumnNameCamel}} {{.GoType}}, limit int, sort string) ([]*model.{{.TableNameCamel}}, error) {
	if sort == "" {
		sort = "-{{.ColumnName}}"
	}
	page := query.NewPage(0, limit, sort)

	records := []*model.{{.TableNameCamel}}{}
	err := d.db.WithContext(ctx).Order(page.Sort()).Limit(page.Limit()).Where("{{.ColumnName}} < ?", last{{.ColumnNameCamel}}).Find(&records).Error
	if err != nil {
		return nil, err
	}
	return records, nil
}

// CreateByTx create a record in the database using the provided transaction
func (d *{{.TableNameCamelFCL}}Dao) CreateByTx(ctx context.Context, tx *gorm.DB, table *model.{{.TableNameCamel}}) ({{.GoType}}, error) {
	err := tx.WithContext(ctx).Create(table).Error
	return table.{{.ColumnNameCamel}}, err
}

// DeleteByTx delete a record by {{.ColumnNameCamelFCL}} in the database using the provided transaction
func (d *{{.TableNameCamelFCL}}Dao) DeleteByTx(ctx context.Context, tx *gorm.DB, {{.ColumnNameCamelFCL}} {{.GoType}}) error {
	update := map[string]interface{}{
		"deleted_at": time.Now(),
	}
	err := tx.WithContext(ctx).Model(&model.{{.TableNameCamel}}{}).Where("{{.ColumnName}} = ?", {{.ColumnNameCamelFCL}}).Updates(update).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, {{.ColumnNameCamelFCL}})

	return nil
}

// UpdateByTx update a record by {{.ColumnNameCamelFCL}} in the database using the provided transaction
func (d *{{.TableNameCamelFCL}}Dao) UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.{{.TableNameCamel}}) error {
	err := d.updateDataBy{{.ColumnNameCamel}}(ctx, tx, table)

	// delete cache
	_ = d.deleteCache(ctx, table.{{.ColumnNameCamel}})

	return err
}
