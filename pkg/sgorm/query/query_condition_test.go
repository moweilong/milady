package query

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPage(t *testing.T) {
	page := DefaultPage(-1)
	t.Log(page.Page(), page.Limit(), page.Sort(), page.Offset())

	SetMaxSize(1)

	page = NewPage(-1, 100, "id")
	t.Log(page.Page(), page.Limit(), page.Sort(), page.Offset())
}

func TestParams_ConvertToPage(t *testing.T) {
	p := &Params{
		Page:  1,
		Limit: 50,
		Sort:  "age,-name",
	}
	order, limit, offset := p.ConvertToPage()
	t.Logf("order=%s, limit=%d, offset=%d", order, limit, offset)

}

func TestParams_ConvertToGormConditions(t *testing.T) {
	type args struct {
		columns []Column
	}
	tests := []struct {
		name    string
		args    args
		want    string
		want1   []interface{}
		wantErr bool
	}{
		// --------------------------- only 1 column query ------------------------------
		{
			name: "1 column eq",
			args: args{
				columns: []Column{
					{
						Name:  "name",
						Value: "ZhangSan",
					},
				},
			},
			want:    "name = ?",
			want1:   []interface{}{"ZhangSan"},
			wantErr: false,
		},
		{
			name: "1 column neq",
			args: args{
				columns: []Column{
					{
						Name:  "name",
						Value: "ZhangSan",
						Exp:   "!=",
					},
				},
			},
			want:    "name <> ?",
			want1:   []interface{}{"ZhangSan"},
			wantErr: false,
		},
		{
			name: "1 column gt",
			args: args{
				columns: []Column{
					{
						Name:  "age",
						Value: 20,
						Exp:   ">",
					},
				},
			},
			want:    "age > ?",
			want1:   []interface{}{20},
			wantErr: false,
		},
		{
			name: "1 column gte",
			args: args{
				columns: []Column{
					{
						Name:  "age",
						Value: 20,
						Exp:   ">=",
					},
				},
			},
			want:    "age >= ?",
			want1:   []interface{}{20},
			wantErr: false,
		},
		{
			name: "1 column lt",
			args: args{
				columns: []Column{
					{
						Name:  "age",
						Value: 20,
						Exp:   "<",
					},
				},
			},
			want:    "age < ?",
			want1:   []interface{}{20},
			wantErr: false,
		},
		{
			name: "1 column lte",
			args: args{
				columns: []Column{
					{
						Name:  "age",
						Value: 20,
						Exp:   "<=",
					},
				},
			},
			want:    "age <= ?",
			want1:   []interface{}{20},
			wantErr: false,
		},
		{
			name: "1 column lte (int)",
			args: args{
				columns: []Column{
					{
						Name:  "age",
						Value: "20",
						Exp:   "<=",
					},
				},
			},
			want:    "age <= ?",
			want1:   []interface{}{20},
			wantErr: false,
		},
		{
			name: "1 column lte (string)",
			args: args{
				columns: []Column{
					{
						Name:  "age",
						Value: "\"20\"",
						Exp:   "<=",
					},
				},
			},
			want:    "age <= ?",
			want1:   []interface{}{"20"},
			wantErr: false,
		},
		{
			name: "1 column like",
			args: args{
				columns: []Column{
					{
						Name:  "name",
						Value: "foo",
						Exp:   Like,
					},
				},
			},
			want:    "name LIKE ?",
			want1:   []interface{}{"%foo%"},
			wantErr: false,
		},
		{
			name: "1 column like prefix",
			args: args{
				columns: []Column{
					{
						Name:  "name",
						Value: "%foo",
						Exp:   Like,
					},
				},
			},
			want:    "name LIKE ?",
			want1:   []interface{}{"%foo"},
			wantErr: false,
		},
		{
			name: "1 column like suffix",
			args: args{
				columns: []Column{
					{
						Name:  "name",
						Value: "foo%",
						Exp:   Like,
					},
				},
			},
			want:    "name LIKE ?",
			want1:   []interface{}{"foo%"},
			wantErr: false,
		},
		{
			name: "1 column like with %",
			args: args{
				columns: []Column{
					{
						Name:  "name",
						Value: "f%o_o",
						Exp:   Like,
					},
				},
			},
			want:    "name LIKE ?",
			want1:   []interface{}{"%f\\%o\\_o%"},
			wantErr: false,
		},
		{
			name: "1 column IN (string)",
			args: args{
				columns: []Column{
					{
						Name:  "name",
						Value: "ab,cd,ef",
						Exp:   In,
					},
				},
			},
			want:    "name IN (?)",
			want1:   []interface{}{[]interface{}{"ab", "cd", "ef"}},
			wantErr: false,
		},
		{
			name: "1 column IN (int)",
			args: args{
				columns: []Column{
					{
						Name:  "age",
						Value: "10,20,30",
						Exp:   In,
					},
				},
			},
			want:    "age IN (?)",
			want1:   []interface{}{[]interface{}{10, 20, 30}},
			wantErr: false,
		},
		{
			name: "1 column IN (string)",
			args: args{
				columns: []Column{
					{
						Name:  "age",
						Exp:   In,
						Value: "'10', '20', \"30\"",
					},
				},
			},
			want:    "age IN (?)",
			want1:   []interface{}{[]interface{}{"10", "20", "30"}},
			wantErr: false,
		},
		{
			name: "1 column IN ([]interface{})",
			args: args{
				columns: []Column{
					{
						Name:  "age",
						Exp:   In,
						Value: []interface{}{10, 20, 30},
					},
				},
			},
			want:    "age IN (?)",
			want1:   []interface{}{[]interface{}{10, 20, 30}},
			wantErr: false,
		},
		{
			name: "1 column NOT IN",
			args: args{
				columns: []Column{
					{
						Name:  "name",
						Value: "ab,cd,ef",
						Exp:   NotIN,
					},
				},
			},
			want:    "name NOT IN (?)",
			want1:   []interface{}{[]interface{}{"ab", "cd", "ef"}},
			wantErr: false,
		},
		{
			name: "1 column IS NULL",
			args: args{
				columns: []Column{
					{
						Name: "name",
						Exp:  IsNull,
					},
				},
			},
			want:    "name IS NULL ",
			want1:   []interface{}{},
			wantErr: false,
		},
		{
			name: "1 column IS NOT NULL",
			args: args{
				columns: []Column{
					{
						Name: "name",
						Exp:  IsNotNull,
					},
				},
			},
			want:    "name IS NOT NULL ",
			want1:   []interface{}{},
			wantErr: false,
		},

		// --------------------------- query 2 columns  ------------------------------
		{
			name: "2 columns eq and",
			args: args{
				columns: []Column{
					{
						Name:  "name",
						Value: "ZhangSan",
					},
					{
						Name:  "gender",
						Value: "male",
					},
				},
			},
			want:    "name = ? AND gender = ?",
			want1:   []interface{}{"ZhangSan", "male"},
			wantErr: false,
		},
		{
			name: "2 columns neq and",
			args: args{
				columns: []Column{
					{
						Name:  "name",
						Value: "ZhangSan",
						//Exp:   Neq,
						Exp: "!=",
					},
					{
						Name:  "name",
						Value: "LiSi",
						//Exp:   Neq,
						Exp: "!=",
					},
				},
			},
			want:    "name <> ? AND name <> ?",
			want1:   []interface{}{"ZhangSan", "LiSi"},
			wantErr: false,
		},
		{
			name: "2 columns gt and",
			args: args{
				columns: []Column{
					{
						Name:  "gender",
						Value: "male",
					},
					{
						Name:  "age",
						Value: 20,
						//Exp:   Gt,
						Exp: ">",
					},
				},
			},
			want:    "gender = ? AND age > ?",
			want1:   []interface{}{"male", 20},
			wantErr: false,
		},
		{
			name: "2 columns gte and",
			args: args{
				columns: []Column{
					{
						Name:  "gender",
						Value: "male",
					},
					{
						Name:  "age",
						Value: 20,
						//Exp:   Gte,
						Exp: ">=",
					},
				},
			},
			want:    "gender = ? AND age >= ?",
			want1:   []interface{}{"male", 20},
			wantErr: false,
		},
		{
			name: "2 columns lt and",
			args: args{
				columns: []Column{
					{
						Name:  "gender",
						Value: "female",
					},
					{
						Name:  "age",
						Value: 20,
						//Exp:   Lt,
						Exp: "<",
					},
				},
			},
			want:    "gender = ? AND age < ?",
			want1:   []interface{}{"female", 20},
			wantErr: false,
		},
		{
			name: "2 columns lte and",
			args: args{
				columns: []Column{
					{
						Name:  "gender",
						Value: "female",
					},
					{
						Name:  "age",
						Value: 20,
						//Exp:   Lte,
						Exp: "<=",
					},
				},
			},
			want:    "gender = ? AND age <= ?",
			want1:   []interface{}{"female", 20},
			wantErr: false,
		},
		{
			name: "2 columns range and",
			args: args{
				columns: []Column{
					{
						Name:  "age",
						Value: 10,
						//Exp:   Gte,
						Exp: ">=",
					},
					{
						Name:  "age",
						Value: 20,
						//Exp:   Lte,
						Exp: "<=",
					},
				},
			},
			want:    "age >= ? AND age <= ?",
			want1:   []interface{}{10, 20},
			wantErr: false,
		},
		{
			name: "2 columns eq or",
			args: args{
				columns: []Column{
					{
						Name:  "name",
						Value: "LiSi",
						//Logic: OR,
						Logic: "||",
					},
					{
						Name:  "gender",
						Value: "female",
					},
				},
			},
			want:    "name = ? OR gender = ?",
			want1:   []interface{}{"LiSi", "female"},
			wantErr: false,
		},
		{
			name: "2 columns neq or",
			args: args{
				columns: []Column{
					{
						Name:  "name",
						Value: "LiSi",
						//Logic: OR,
						Logic: "||",
					},
					{
						Name:  "gender",
						Value: "male",
						//Exp:   Neq,
						Exp: "!=",
					},
				},
			},
			want:    "name = ? OR gender <> ?",
			want1:   []interface{}{"LiSi", "male"},
			wantErr: false,
		},
		{
			name: "2 columns eq and in",
			args: args{
				columns: []Column{
					{
						Name:  "gender",
						Value: "male",
						//Logic: "&",
					},
					{
						Name:  "name",
						Value: "LiSi,ZhangSan,WangWu",
						Exp:   In,
					},
				},
			},
			want:    "gender = ? AND name IN (?)",
			want1:   []interface{}{"male", []interface{}{"LiSi", "ZhangSan", "WangWu"}},
			wantErr: false,
		},

		// ------------------------------ IN -------------------------------------------------
		{
			name: "3 columns eq and",
			args: args{
				columns: []Column{
					{
						Name:  "name",
						Value: "LiSi",
					},
					{
						Name:  "name",
						Value: "ZhangSan",
					},
					{
						Name:  "name",
						Value: "WangWu",
					},
				},
			},
			want:    "name IN (?)",
			want1:   []interface{}{[]interface{}{"LiSi", "ZhangSan", "WangWu"}},
			wantErr: false,
		},

		// ------------------------------ group -------------------------------------------------
		{
			name: "4 columns group",
			args: args{
				columns: []Column{
					{
						Name:  "created_at",
						Exp:   ">=",
						Value: "2021-01-01",
						Logic: "and",
					},
					{
						Name:  "created_at",
						Exp:   "<",
						Value: "2021-01-02",
						Logic: "and",
					},
					{
						Name:  "username",
						Exp:   "like",
						Value: "Li%",
						Logic: "or:(",
					},
					{
						Name:  "nickname",
						Exp:   "like",
						Value: "%Si",
						Logic: "or:)",
					},
				},
			},
			want:    "created_at >= ? AND created_at < ? AND  ( username LIKE ? OR nickname LIKE ? ) ",
			want1:   []interface{}{time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2021, 1, 2, 0, 0, 0, 0, time.UTC), "Li%", "%Si"},
			wantErr: false,
		},
		{
			name: "4 columns group",
			args: args{
				columns: []Column{
					{
						Name:  "username",
						Exp:   "like",
						Value: "Li%",
						Logic: "or:(",
					},
					{
						Name:  "nickname",
						Exp:   "like",
						Value: "%Si",
						Logic: "and:)",
					},
					{
						Name:  "created_at",
						Exp:   ">=",
						Value: "2021-01-01",
						Logic: "and",
					},
					{
						Name:  "created_at",
						Exp:   "<",
						Value: "2021-01-02",
					},
				},
			},
			want:    " ( username LIKE ? OR nickname LIKE ? )  AND created_at >= ? AND created_at < ?",
			want1:   []interface{}{"Li%", "%Si", time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2021, 1, 2, 0, 0, 0, 0, time.UTC)},
			wantErr: false,
		},

		// --------------------------- datetime condition  ------------------------------

		{
			name: "datetime condition",
			args: args{
				columns: []Column{
					{
						Name:  "created_at",
						Exp:   ">=",
						Value: "2021-01-01 00:00:00",
					},
					{
						Name:  "created_at",
						Exp:   "<",
						Value: "2021-01-02 00:00:00",
					},
				},
			},
			want:    "created_at >= ? AND created_at < ?",
			want1:   []interface{}{time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2021, 1, 2, 0, 0, 0, 0, time.UTC)},
			wantErr: false,
		},
		{
			name: "datetime condition with timezone",
			args: args{
				columns: []Column{
					{
						Name:  "created_at",
						Exp:   ">=",
						Value: "2021-01-01T00:00:00+06:00",
					},
					{
						Name:  "created_at",
						Exp:   "<",
						Value: "2021-01-02T00:00:00+06:00",
					},
				},
			},
			want:    "created_at >= ? AND created_at < ?",
			want1:   []interface{}{time.Date(2021, 1, 1, 0, 0, 0, 0, time.FixedZone("", 6*3600)), time.Date(2021, 1, 2, 0, 0, 0, 0, time.FixedZone("", 6*3600))},
			wantErr: false,
		},
		{
			name: "datetime condition with timezone 2",
			args: args{
				columns: []Column{
					{
						Name:  "created_at",
						Exp:   ">=",
						Value: "2021-01-01T00:00:00+07:00",
					},
					{
						Name:  "created_at",
						Exp:   "<",
						Value: "2021-01-02T00:00:00+07:00",
					},
				},
			},
			want:    "created_at >= ? AND created_at < ?",
			want1:   []interface{}{time.Date(2021, 1, 1, 0, 0, 0, 0, time.FixedZone("", 7*3600)), time.Date(2021, 1, 2, 0, 0, 0, 0, time.FixedZone("", 7*3600))},
			wantErr: false,
		},

		// ---------------------------- error ----------------------------------------------
		{
			name: "exp type err",
			args: args{
				columns: []Column{
					{
						Name:  "gender",
						Value: "male",
						Exp:   "xxxxxx",
					},
				},
			},
			want:    "",
			want1:   nil,
			wantErr: true,
		},
		{
			name: "logic type err",
			args: args{
				columns: []Column{
					{
						Name:  "gender",
						Value: "male",
						Logic: "xxxxxx",
					},
				},
			},
			want:    "",
			want1:   nil,
			wantErr: true,
		},
		{
			name: "empty",
			args: args{
				columns: nil,
			},
			want:    "",
			want1:   nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := &Params{
				Columns: tt.args.columns,
			}
			got, got1, err := params.ConvertToGormConditions()
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertToGormConditions() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ConvertToGormConditions() got = [%v], want = [%v]", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("ConvertToGormConditions() got1 = [%v], want = [%v]", got1, tt.want1)
			}

			got = strings.Replace(got, "?", "%v", -1)
			t.Logf(got, got1...)
		})
	}
}

func TestConditions_ConvertToGormConditions_Error(t *testing.T) {
	p := &Params{
		Limit: 10,
		Columns: []Column{
			{
				Name:  "age",
				Value: 10,
			},
			{
				Name:  "email",
				Value: "foo@bar.com",
			},
		}}

	whitelists := map[string]bool{"name": true, "age": true}
	_, _, err := p.ConvertToGormConditions(WithWhitelistNames(whitelists))
	t.Log(err)
	assert.Error(t, err)

	fn := func(columns []Column) error {
		for _, col := range columns {
			if col.Value == "foo@bar.com" {
				return errors.New("'foo@bar.com' is not allowed")
			}
		}
		return nil
	}
	_, _, err = p.ConvertToGormConditions(WithValidateFn(fn))
	t.Log(err)
	assert.Error(t, err)
}

func TestConditions_ConvertToGorm(t *testing.T) {
	c := Conditions{
		Columns: []Column{
			{
				Name:  "name",
				Value: "ZhangSan",
			},
			{
				Name:  "gender",
				Value: "male",
			},
		}}
	str, values, err := c.ConvertToGorm()
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, "name = ? AND gender = ?", str)
	assert.Equal(t, len(values), 2)
}

func TestConditions_ConvertToGorm_Error(t *testing.T) {
	c := Conditions{Columns: []Column{
		{
			Name:  "age",
			Value: 10,
		},
		{
			Name:  "email",
			Value: "foo@bar.com",
		},
	}}

	whitelists := map[string]bool{"name": true, "age": true}
	_, _, err := c.ConvertToGorm(WithWhitelistNames(whitelists))
	t.Log(err)
	assert.Error(t, err)

	fn := func(columns []Column) error {
		for _, col := range columns {
			if col.Value == "foo@bar.com" {
				return errors.New("'foo@bar.com' is not allowed")
			}
		}
		return nil
	}
	_, _, err = c.ConvertToGorm(WithValidateFn(fn))
	t.Log(err)
	assert.Error(t, err)
}

func Test_convertValue(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name string
		args args
		want interface{}
	}{
		{
			name: "string 1",
			args: args{
				v: "foo",
			},
			want: "foo",
		},
		{
			name: "string 2",
			args: args{
				v: "'123'",
			},
			want: "'123'",
		},
		{
			name: "string 3",
			args: args{
				v: "\"123\"",
			},
			want: "123",
		},
		{
			name: "int 1",
			args: args{
				v: "123",
			},
			want: 123,
		},
		{
			name: "int 2",
			args: args{
				v: 123,
			},
			want: 123,
		},
		{
			name: "float",
			args: args{
				v: 123.456,
			},
			want: 123.456,
		},
		{
			name: "float string",
			args: args{
				v: "123.456",
			},
			want: 123.456,
		},
		{
			name: "bool",
			args: args{
				v: true,
			},
			want: true,
		},
		{
			name: "bool string",
			args: args{
				v: "true",
			},
			want: true,
		},
		{
			name: "datetime 1",
			args: args{
				v: "2023-05-15T14:30:00Z",
			},
			want: time.Date(2023, 5, 15, 14, 30, 0, 0, time.UTC),
		},
		{
			name: "datetime 2",
			args: args{
				v: "2023-05-15T14:30:00+07:00",
			},
			want: time.Date(2023, 5, 15, 14, 30, 0, 0, time.FixedZone("UTC+7", 25200)),
		},
		{
			name: "datetime 3",
			args: args{
				v: "2023-05-15T14:30:00.123Z",
			},
			want: time.Date(2023, 5, 15, 14, 30, 0, 123000000, time.UTC),
		},
		{
			name: "datetime 4",
			args: args{
				v: "2023-05-15T14:30:00+0700",
			},
			want: time.Date(2023, 5, 15, 14, 30, 0, 0, time.FixedZone("UTC+7", 25200)),
		},
		{
			name: "datetime 5",
			args: args{
				v: "2023-05-15T14:30:00.123+07:00",
			},
			want: time.Date(2023, 5, 15, 14, 30, 0, 123000000, time.FixedZone("UTC+7", 25200)),
		},
		{
			name: "datetime 6",
			args: args{
				v: "2023-05-15",
			},
			want: time.Date(2023, 5, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "datetime 7",
			args: args{
				v: "2023-05-15 14:30:00",
			},
			want: time.Date(2023, 5, 15, 14, 30, 0, 0, time.UTC),
		},
		{
			name: "datetime 8",
			args: args{
				v: "2023-05-15 14:30:00.123",
			},
			want: time.Date(2023, 5, 15, 14, 30, 0, 123000000, time.UTC),
		},
		{
			name: "datetime 9",
			args: args{
				v: "2023-05-15 14:30:00 +07:00",
			},
			want: time.Date(2023, 5, 15, 14, 30, 0, 0, time.FixedZone("UTC+7", 25200)),
		},
		{
			name: "datetime 10",
			args: args{
				v: "2023-05-15 14:30:00.123 +07:00",
			},
			want: time.Date(2023, 5, 15, 14, 30, 0, 123000000, time.FixedZone("UTC+7", 25200)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertValue(tt.args.v)
			if dt, ok := got.(time.Time); ok {
				assert.Equal(t, tt.want.(time.Time).In(time.UTC), dt.In(time.UTC))
			} else {
				assert.Equalf(t, tt.want, got, "convertValue(%v)", tt.args.v)
			}
		})
	}
}
