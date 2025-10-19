package query

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestPage(t *testing.T) {
	page := DefaultPage(-1)
	t.Log(page.Page(), page.Limit(), page.Sort(), page.Skip())
	page = NewPage(0, 20, "")
	t.Log(page.Page(), page.Limit(), page.Sort(), page.Skip())

	SetMaxSize(1)
	page = NewPage(0, 20, "_id")
	t.Log(page.Page(), page.Limit(), page.Sort(), page.Skip())
}

func TestParams_ConvertToPage(t *testing.T) {
	p := &Params{
		Page:  0,
		Limit: 20,
		Sort:  "age,-name",
	}
	order, limit, offset := p.ConvertToPage()
	t.Logf("order=%v, limit=%d, skip=%d", order, limit, offset)
}

func TestParams_ConvertToMongoFilter(t *testing.T) {
	type args struct {
		columns []Column
	}
	tests := []struct {
		name    string
		args    args
		want    bson.M
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
			want:    bson.M{"name": "ZhangSan"},
			wantErr: false,
		},
		{
			name: "1 column neq",
			args: args{
				columns: []Column{
					{
						Name:  "name",
						Exp:   "!=",
						Value: "ZhangSan",
					},
				},
			},
			want:    bson.M{"name": bson.M{"$ne": "ZhangSan"}},
			wantErr: false,
		},
		{
			name: "1 column gt",
			args: args{
				columns: []Column{
					{
						Name:  "age",
						Exp:   ">",
						Value: 20,
					},
				},
			},
			want:    bson.M{"age": bson.M{"$gt": 20}},
			wantErr: false,
		},
		{
			name: "1 column gte",
			args: args{
				columns: []Column{
					{
						Name:  "age",
						Exp:   ">=",
						Value: 20,
					},
				},
			},
			want:    bson.M{"age": bson.M{"$gte": 20}},
			wantErr: false,
		},
		{
			name: "1 column lt",
			args: args{
				columns: []Column{
					{
						Name:  "age",
						Exp:   "<",
						Value: 20,
					},
				},
			},
			want:    bson.M{"age": bson.M{"$lt": 20}},
			wantErr: false,
		},
		{
			name: "1 column lte",
			args: args{
				columns: []Column{
					{
						Name:  "age",
						Exp:   "<=",
						Value: 20,
					},
				},
			},
			want:    bson.M{"age": bson.M{"$lte": 20}},
			wantErr: false,
		},
		{
			name: "1 column like",
			args: args{
				columns: []Column{
					{
						Name:  "name",
						Exp:   Like,
						Value: "Li",
					},
				},
			},
			want:    bson.M{"name": bson.M{"$options": "i", "$regex": "Li"}},
			wantErr: false,
		},
		{
			name: "1 column IN (string)",
			args: args{
				columns: []Column{
					{
						Name:  "name",
						Exp:   In,
						Value: "ab,cd,ef",
					},
				},
			},
			want:    bson.M{"name": bson.M{"$in": []interface{}{"ab", "cd", "ef"}}},
			wantErr: false,
		},
		{
			name: "1 column IN (int)",
			args: args{
				columns: []Column{
					{
						Name:  "level",
						Exp:   In,
						Value: "3, 4, 5",
					},
				},
			},
			want:    bson.M{"level": bson.M{"$in": []interface{}{3, 4, 5}}},
			wantErr: false,
		},
		{
			name: "1 column IN (string)",
			args: args{
				columns: []Column{
					{
						Name:  "level",
						Exp:   In,
						Value: "'3', '4', \"5\"",
					},
				},
			},
			want:    bson.M{"level": bson.M{"$in": []interface{}{"3", "4", "5"}}},
			wantErr: false,
		},
		{
			name: "1 column IN ([]interface{})",
			args: args{
				columns: []Column{
					{
						Name:  "level",
						Exp:   In,
						Value: []interface{}{3, 4, 5},
					},
				},
			},
			want:    bson.M{"level": bson.M{"$in": []interface{}{3, 4, 5}}},
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
			want:    bson.M{"$and": []bson.M{{"name": "ZhangSan"}, {"gender": "male"}}},
			wantErr: false,
		},
		{
			name: "2 columns neq and",
			args: args{
				columns: []Column{
					{
						Name:  "name",
						Exp:   "!=",
						Value: "ZhangSan",
					},
					{
						Name:  "name",
						Exp:   "!=",
						Value: "LiSi",
					},
				},
			},
			want:    bson.M{"$and": []bson.M{{"name": bson.M{"$ne": "ZhangSan"}}, {"name": bson.M{"$ne": "LiSi"}}}},
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
						Exp:   ">",
						Value: 20,
					},
				},
			},
			want:    bson.M{"$and": []bson.M{{"gender": "male"}, {"age": bson.M{"$gt": 20}}}},
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
						Exp:   ">=",
						Value: 20,
					},
				},
			},
			want:    bson.M{"$and": []bson.M{{"gender": "male"}, {"age": bson.M{"$gte": 20}}}},
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
						Exp:   "<",
						Value: 20,
					},
				},
			},
			want:    bson.M{"$and": []bson.M{{"gender": "female"}, {"age": bson.M{"$lt": 20}}}},
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
						Exp:   "<=",
						Value: 20,
					},
				},
			},
			want:    bson.M{"$and": []bson.M{{"gender": "female"}, {"age": bson.M{"$lte": 20}}}},
			wantErr: false,
		},
		{
			name: "2 columns range and",
			args: args{
				columns: []Column{
					{
						Name:  "age",
						Exp:   ">=",
						Value: 10,
					},
					{
						Name:  "age",
						Exp:   "<=",
						Value: 20,
					},
				},
			},
			want:    bson.M{"$and": []bson.M{{"age": bson.M{"$gte": 10}}, {"age": bson.M{"$lte": 20}}}},
			wantErr: false,
		},
		{
			name: "2 columns eq or",
			args: args{
				columns: []Column{
					{
						Name:  "name",
						Value: "LiSi",
						Logic: "||",
					},
					{
						Name:  "gender",
						Value: "female",
					},
				},
			},
			want:    bson.M{"$or": []bson.M{{"name": "LiSi"}, {"gender": "female"}}},
			wantErr: false,
		},
		{
			name: "2 columns neq or",
			args: args{
				columns: []Column{
					{
						Name:  "name",
						Value: "LiSi",
						Logic: "||",
					},
					{
						Name:  "gender",
						Exp:   "!=",
						Value: "male",
					},
				},
			},
			want:    bson.M{"$or": []bson.M{{"name": "LiSi"}, {"gender": bson.M{"$ne": "male"}}}},
			wantErr: false,
		},
		{
			name: "2 columns eq and in",
			args: args{
				columns: []Column{
					{
						Name:  "gender",
						Value: "male",
					},
					{
						Name:  "name",
						Exp:   In,
						Value: "LiSi,ZhangSan,WangWu",
					},
				},
			},
			want:    bson.M{"$and": []bson.M{{"gender": "male"}, {"name": bson.M{"$in": []interface{}{"LiSi", "ZhangSan", "WangWu"}}}}},
			wantErr: false,
		},

		// --------------------------- query 3 columns  ------------------------------
		{
			name: "3 columns and",
			args: args{
				columns: []Column{
					{
						Name:  "gender",
						Value: "male",
					},
					{
						Name:  "name",
						Value: "ZhangSan",
					},
					{
						Name:  "age",
						Exp:   "<",
						Value: 12,
					},
				},
			},
			want:    bson.M{"$and": []bson.M{{"gender": "male"}, {"name": "ZhangSan"}, {"age": bson.M{"$lt": 12}}}},
			wantErr: false,
		},
		{
			name: "3 columns or",
			args: args{
				columns: []Column{
					{
						Name:  "gender",
						Value: "male",
						Logic: "or",
					},
					{
						Name:  "name",
						Value: "ZhangSan",
						Logic: "or",
					},
					{
						Name:  "age",
						Exp:   "<",
						Value: 12,
					},
				},
			},
			want:    bson.M{"$or": []bson.M{{"gender": "male"}, {"name": "ZhangSan"}, {"age": bson.M{"$lt": 12}}}},
			wantErr: false,
		},
		{
			name: "3 columns mix (or and)",
			args: args{
				columns: []Column{
					{
						Name:  "gender",
						Value: "male",
						Logic: "and",
					},
					{
						Name:  "name",
						Value: "ZhangSan",
						Logic: "or",
					},
					{
						Name:  "age",
						Exp:   "<",
						Value: 12,
					},
				},
			},
			want:    bson.M{"$or": []bson.M{{"$and": []bson.M{{"gender": "male"}, {"name": "ZhangSan"}}}, {"age": bson.M{"$lt": 12}}}},
			wantErr: false,
		},

		// --------------------------- query 4 columns  ------------------------------
		{
			name: "4 columns mix (or and)",
			args: args{
				columns: []Column{
					{
						Name:  "gender",
						Value: "male",
						Logic: "or",
					},
					{
						Name:  "name",
						Value: "ZhangSan",
					},
					{
						Name:  "age",
						Exp:   "<",
						Value: 12,
						Logic: "or",
					},
					{
						Name:  "city",
						Value: "canton",
					},
				},
			},
			want:    bson.M{"$or": []bson.M{{"gender": "male"}, {"$and": []bson.M{{"name": "ZhangSan"}, {"age": bson.M{"$lt": 12}}}}, {"city": "canton"}}},
			wantErr: false,
		},

		// --------------------------- parentheses group  ------------------------------
		{
			name: "parentheses group 1",
			args: args{
				columns: []Column{
					{Name: "salary", Exp: ">=", Value: 10000, Logic: "or:("},
					{Name: "level", Exp: "in", Value: "3,4,5", Logic: "and:)"},
					{Name: "dept", Value: "mkt", Logic: "and"},
				},
			},
			want: bson.M{
				"$and": []bson.M{
					{
						"$or": []bson.M{{"salary": bson.M{"$gte": 10000}}, {"level": bson.M{"$in": []interface{}{3, 4, 5}}}},
					},
					{
						"dept": "mkt",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "parentheses group 2",
			args: args{
				columns: []Column{
					{Name: "dept", Value: "rd", Logic: "and:("},
					{Name: "salary", Exp: ">=", Value: 10000, Logic: "or:)"},
					{Name: "dept", Value: "mkt", Logic: "and:("},
					{Name: "level", Exp: "in", Value: "3,4,5", Logic: "and:)"},
				},
			},
			want: bson.M{
				"$or": []bson.M{
					{
						"dept": "rd", "salary": bson.M{"$gte": 10000},
					},
					{
						"dept": "mkt", "level": bson.M{"$in": []interface{}{3, 4, 5}},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "parentheses group 3",
			args: args{
				columns: []Column{
					{Name: "salary", Exp: ">=", Value: 10000, Logic: "or:("},
					{Name: "level", Exp: "in", Value: "3,4,5", Logic: "and:)"},
					{Name: "dept", Value: "mkt", Logic: "or:("},
					{Name: "dept", Value: "rd", Logic: "and:)"},
					{Name: "age", Exp: "<=", Value: 30, Logic: "or:("},
					{Name: "age", Exp: ">", Value: 40, Logic: "and:)"},
				},
			},
			want: bson.M{
				"$and": []bson.M{
					{
						"$or": []bson.M{{"salary": bson.M{"$gte": 10000}}, {"level": bson.M{"$in": []interface{}{3, 4, 5}}}},
					},
					{
						"$or": []bson.M{{"dept": "mkt"}, {"dept": "rd"}}},
					{
						"$or": []bson.M{{"age": bson.M{"$lte": 30}}, {"age": bson.M{"$gt": 40}}},
					},
				},
			},
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
			want:    bson.M{"$and": []bson.M{{"created_at": bson.M{"$gte": time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)}}, {"created_at": bson.M{"$lt": time.Date(2021, 1, 2, 0, 0, 0, 0, time.UTC)}}}},
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
			want:    bson.M{"$and": []bson.M{{"created_at": bson.M{"$gte": time.Date(2021, 1, 1, 0, 0, 0, 0, time.FixedZone("", 6*3600))}}, {"created_at": bson.M{"$lt": time.Date(2021, 1, 2, 0, 0, 0, 0, time.FixedZone("", 6*3600))}}}},
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
			want:    bson.M{"$and": []bson.M{{"created_at": bson.M{"$gte": time.Date(2021, 1, 1, 0, 0, 0, 0, time.FixedZone("", 7*3600))}}, {"created_at": bson.M{"$lt": time.Date(2021, 1, 2, 0, 0, 0, 0, time.FixedZone("", 7*3600))}}}},
			wantErr: false,
		},

		// --------------------------- object id ------------------------------
		{
			name: "convert to object id 1",
			args: args{
				columns: []Column{
					{
						Name:  "id",
						Value: "65ce48483f11aff697e30d6d",
					},
					{
						Name:  "order_id:oid",
						Value: "65ce48483f11aff697e30d6d",
					},
				},
			},
			want:    bson.M{"$and": []bson.M{{"_id": primitive.ObjectID{0x65, 0xce, 0x48, 0x48, 0x3f, 0x11, 0xaf, 0xf6, 0x97, 0xe3, 0xd, 0x6d}}, {"order_id": primitive.ObjectID{0x65, 0xce, 0x48, 0x48, 0x3f, 0x11, 0xaf, 0xf6, 0x97, 0xe3, 0xd, 0x6d}}}},
			wantErr: false,
		},

		{
			name: "convert to object id 2",
			args: args{
				columns: []Column{
					{
						Name:  "userId",
						Value: "65ce48483f11aff697e30d6d",
					},
					{
						Name:  "orderID",
						Value: "65ce48483f11aff697e30d6d",
					},
				},
			},
			want:    bson.M{"$and": []bson.M{{"userId": primitive.ObjectID{0x65, 0xce, 0x48, 0x48, 0x3f, 0x11, 0xaf, 0xf6, 0x97, 0xe3, 0xd, 0x6d}}, {"orderID": primitive.ObjectID{0x65, 0xce, 0x48, 0x48, 0x3f, 0x11, 0xaf, 0xf6, 0x97, 0xe3, 0xd, 0x6d}}}},
			wantErr: false,
		},

		{
			name: "convert to object id 3",
			args: args{
				columns: []Column{
					{
						Name:  "_id",
						Value: "65ce48483f11aff697e30d6d",
					},
					{
						Name:  "my_order",
						Value: "65ce48483f11aff697e30d6d",
					},
				},
			},
			want:    bson.M{"$and": []bson.M{{"_id": primitive.ObjectID{0x65, 0xce, 0x48, 0x48, 0x3f, 0x11, 0xaf, 0xf6, 0x97, 0xe3, 0xd, 0x6d}}, {"my_order": primitive.ObjectID{0x65, 0xce, 0x48, 0x48, 0x3f, 0x11, 0xaf, 0xf6, 0x97, 0xe3, 0xd, 0x6d}}}},
			wantErr: false,
		},

		// ---------------------------- error ----------------------------------------------
		{
			name: "exp type err",
			args: args{
				columns: []Column{
					{
						Name:  "gender",
						Exp:   "xxxxxx",
						Value: "male",
					},
				},
			},
			want:    nil,
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
			want:    nil,
			wantErr: true,
		},
		{
			name: "name empty",
			args: args{
				columns: []Column{
					{
						Name:  "",
						Value: "male",
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "value empty",
			args: args{
				columns: []Column{
					{
						Name:  "name",
						Value: nil,
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "empty",
			args: args{
				columns: nil,
			},
			want:    primitive.M{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := &Params{
				Columns: tt.args.columns,
			}
			got, err := params.ConvertToMongoFilter()
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertToMongoFilter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConvertToMongoFilter() got = %#v, want = %#v", got, tt.want)
			}
		})
	}
}

func TestParams_ConvertToMongoFilter_Error(t *testing.T) {
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
	_, err := p.ConvertToMongoFilter(WithWhitelistNames(whitelists))
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
	_, err = p.ConvertToMongoFilter(WithValidateFn(fn))
	t.Log(err)
	assert.Error(t, err)
}

func TestConditions_ConvertToMongo(t *testing.T) {
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
	got, err := c.ConvertToMongo()
	if err != nil {
		t.Error(err)
	}
	want := bson.M{"$and": []bson.M{{"name": "ZhangSan"}, {"gender": "male"}}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ConvertToMongo() got = %+v, want %+v", got, want)
	}
}

func TestConditions_checkValid(t *testing.T) {
	// empty error
	c := Conditions{}
	err := c.CheckValid()
	assert.Error(t, err)

	// value is empty error
	c = Conditions{
		Columns: []Column{
			{
				Name:  "foo",
				Value: nil,
			},
		},
	}
	err = c.CheckValid()
	assert.Error(t, err)

	// exp error
	c = Conditions{
		Columns: []Column{
			{
				Name:  "foo",
				Value: "bar",
				Exp:   "unknown-exp",
			},
		},
	}
	err = c.CheckValid()
	assert.Error(t, err)

	// logic error
	c = Conditions{
		Columns: []Column{
			{
				Name:  "foo",
				Value: "bar",
				Logic: "unknown-logic",
			},
		},
	}
	err = c.CheckValid()
	assert.Error(t, err)

	// success
	c = Conditions{
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
	err = c.CheckValid()
	assert.NoError(t, err)
}

func Test_groupingIndex(t *testing.T) {
	type args struct {
		l         int
		orIndexes []int
	}
	tests := []struct {
		name string
		args args
		want [][]int
	}{
		{
			name: "4 index 1",
			args: args{
				l:         4,
				orIndexes: []int{0, 2},
			},
			want: [][]int{{0}, {1, 2}, {3}},
		},
		{
			name: "4 index 2",
			args: args{
				l:         4,
				orIndexes: []int{1},
			},
			want: [][]int{{0, 1}, {2, 3}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := groupingIndex(tt.args.l, tt.args.orIndexes)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("groupingIndex got = %#v, want = %#v", got, tt.want)
			}
			t.Log(got)
		})
	}
}

func Test_getSort(t *testing.T) {
	names := []string{
		"", "id", "-id", "gender", "gender,id", "-gender,-id",
	}
	for _, name := range names {
		d := getSort(name)
		t.Log(d)
	}
}

func TestConditions_ConvertToMongo_Error(t *testing.T) {
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
	_, err := c.ConvertToMongo(WithWhitelistNames(whitelists))
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
	_, err = c.ConvertToMongo(WithValidateFn(fn))
	t.Log(err)
	assert.Error(t, err)
}

func groupingIndex(l int, orIndexes []int) [][]int {
	groupIndexes := [][]int{}
	lastIndex := 0
	for _, index := range orIndexes {
		group := []int{}
		for i := lastIndex; i <= index; i++ {
			group = append(group, i)
		}
		groupIndexes = append(groupIndexes, group)
		if lastIndex == index {
			lastIndex++
		} else {
			lastIndex = index
		}
	}
	group := []int{}
	for i := lastIndex + 1; i < l; i++ {
		group = append(group, i)
	}
	groupIndexes = append(groupIndexes, group)
	return groupIndexes
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
