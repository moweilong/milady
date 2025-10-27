package parser

import (
	"encoding/json"
	"fmt"
	"strings"
	"text/template"

	"github.com/jinzhu/inflection"
)

// CrudInfo crud info for cache, dao, handler, service, protobuf, error
type CrudInfo struct {
	TableNameCamel          string `json:"tableNameCamel"`          // pascal case, example: FooBar. 帕斯卡命名法
	TableNameCamelFCL       string `json:"tableNameCamelFCL"`       // camel case and first character lower, example: fooBar. 驼峰命名法，首字母小写
	TableNamePluralCamel    string `json:"tableNamePluralCamel"`    // plural, pascal case, example: FooBars. 帕斯卡命名法，复数形式
	TableNamePluralCamelFCL string `json:"tableNamePluralCamelFCL"` // plural, camel case, example: fooBars. 驼峰命名法，复数形式，首字母小写

	ColumnName               string `json:"columnName"`               // column name, example: first_name. 蛇形命名法
	ColumnNameCamel          string `json:"columnNameCamel"`          // column name, pascal case, example: FirstName, 帕斯卡命名法
	ColumnNameCamelFCL       string `json:"columnNameCamelFCL"`       // column name, camel case and first character lower, example: firstName, 驼峰命名法，首字母小写
	ColumnNamePluralCamel    string `json:"columnNamePluralCamel"`    // column name, plural, pascal case, example: FirstNames, 帕斯卡命名法，复数形式
	ColumnNamePluralCamelFCL string `json:"columnNamePluralCamelFCL"` // column name, plural, camel case and first character lower, example: firstNames, 驼峰命名法，复数形式，首字母小写

	GoType       string `json:"goType"`       // go type, example: string, uint64
	GoTypeFCU    string `json:"goTypeFCU"`    // go type, first character upper, example: String, Uint64
	ProtoType    string `json:"protoType"`    // proto type, example: string, uint64
	IsStringType bool   `json:"isStringType"` // go type is string or not

	PrimaryKeyColumnName string `json:"PrimaryKeyColumnName"` // primary key, example: id
	IsCommonType         bool   `json:"isCommonType"`         // custom primary key name and type
	IsStandardPrimaryKey bool   `json:"isStandardPrimaryKey"` // standard primary key id
}

// isDesiredGoType define the desired（期望） go type, check if the go type is desired type
func isDesiredGoType(t string) bool {
	switch t {
	case "string", "uint64", "int64", "uint", "int", "uint32", "int32": //nolint
		return true
	}
	return false
}

// setCrudInfo set crud info from tmplField
func setCrudInfo(field tmplField) *CrudInfo {
	primaryKeyName := ""
	if field.IsPrimaryKey {
		primaryKeyName = field.ColName
	}
	// 复数形式
	pluralName := inflection.Plural(field.Name)

	// tmplField to CrudInfo
	info := &CrudInfo{
		ColumnName:               field.ColName,
		ColumnNameCamel:          field.Name,
		ColumnNameCamelFCL:       customFirstLetterToLower(field.Name),
		ColumnNamePluralCamel:    customEndOfLetterToLower(field.Name, pluralName), // TODO 和结构体注释不一致，不是 Pascal 命名法
		ColumnNamePluralCamelFCL: customFirstLetterToLower(customEndOfLetterToLower(field.Name, pluralName)),
		GoType:                   field.GoType,
		GoTypeFCU:                firstLetterToUpper(field.GoType),
		ProtoType:                simpleGoTypeToProtoType(field.GoType),
		IsStringType:             field.GoType == "string",
		PrimaryKeyColumnName:     primaryKeyName,
		IsStandardPrimaryKey:     field.ColName == "id",
	}

	if info.ColumnNameCamel == info.ColumnNamePluralCamel {
		info.ColumnNamePluralCamel += "s" // if singular and plural are the same, force the suffix 's' to distinguish them
	}
	if info.ColumnNameCamelFCL == info.ColumnNamePluralCamelFCL {
		info.ColumnNamePluralCamelFCL += "s" // if singular and plural are the same, force the suffix 's' to distinguish them
	}

	return info
}

// newCrudInfo create crud info from tmplData
func newCrudInfo(data tmplData) *CrudInfo {
	if len(data.Fields) == 0 {
		return nil
	}

	// find primary key
	var info *CrudInfo
	for _, field := range data.Fields {
		if field.IsPrimaryKey {
			info = setCrudInfo(field)
			break
		}
	}

	// if not found primary key, find the first xxx_id column as primary key
	if info == nil {
		for _, field := range data.Fields {
			if strings.HasSuffix(field.ColName, "_id") && isDesiredGoType(field.GoType) { // xxx_id
				info = setCrudInfo(field)
				break
			}
		}
	}

	// if not found xxx_id field, use the first field of integer or string type
	if info == nil {
		for _, field := range data.Fields {
			if isDesiredGoType(field.GoType) {
				info = setCrudInfo(field)
				break
			}
		}
	}

	// use the first column as primary key
	if info == nil {
		info = setCrudInfo(data.Fields[0])
	}

	info.TableNameCamel = data.TableName
	info.TableNameCamelFCL = data.TName

	pluralName := inflection.Plural(data.TableName)
	info.TableNamePluralCamel = customEndOfLetterToLower(data.TableName, pluralName) // TODO 和结构体注释不一致，不是 Pascal 命名法
	info.TableNamePluralCamelFCL = customFirstLetterToLower(customEndOfLetterToLower(data.TableName, pluralName))

	return info
}

// getCode return crud info json string
func (info *CrudInfo) getCode() string {
	if info == nil {
		return ""
	}
	pkData, _ := json.Marshal(info)
	return string(pkData)
}

// CheckCommonType check if the primary key is custom primary key, not standard primary key id
func (info *CrudInfo) CheckCommonType() bool {
	if info == nil {
		return false
	}
	return info.IsCommonType
}

// isIDPrimaryKey check if the primary key is standard primary key id
func (info *CrudInfo) isIDPrimaryKey() bool {
	if info == nil {
		return false
	}
	if info.ColumnName == "id" && (info.GoType == "uint64" ||
		info.GoType == "int64" ||
		info.GoType == "uint" ||
		info.GoType == "int" ||
		info.GoType == "uint32" ||
		info.GoType == "int32") {
		return true
	}
	return false
}

// GetGRPCProtoValidation return grpc proto validation tag
func (info *CrudInfo) GetGRPCProtoValidation() string {
	if info == nil {
		return ""
	}
	if info.ProtoType == "string" {
		return `[(validate.rules).string.min_len = 1]`
	}
	return fmt.Sprintf(`[(validate.rules).%s.gt = 0]`, info.ProtoType)
}

// GetWebProtoValidation return web proto validation tag
func (info *CrudInfo) GetWebProtoValidation() string {
	if info == nil {
		return ""
	}
	if info.ProtoType == "string" {
		return fmt.Sprintf(`[(validate.rules).string.min_len = 1, (tagger.tags) = "uri:\"%s\""]`, info.ColumnNameCamelFCL)
	}
	return fmt.Sprintf(`[(validate.rules).%s.gt = 0, (tagger.tags) = "uri:\"%s\""]`, info.ProtoType, info.ColumnNameCamelFCL)
}

func getCommonHandlerStructCodes(data tmplData, jsonNamedType int) (string, error) {
	// 处理字段的 JSON 名称和 Go 类型
	newFields := []tmplField{}
	for _, field := range data.Fields {
		if jsonNamedType == 0 { // snake case
			field.JSONName = customToSnake(field.ColName)
		} else {
			field.JSONName = customToCamel(field.ColName) // camel case (default)
		}
		field.GoType = getHandlerGoType(&field) // 处理 Go 类型，确保在处理器层使用合适的数据类型
		newFields = append(newFields, field)
	}
	data.Fields = newFields

	// TODO handlerCreateStructCommonTmpl 的用途待更新
	postStructCode, err := tmplExecuteWithFilter(data, handlerCreateStructCommonTmpl)
	if err != nil {
		return "", fmt.Errorf("handlerCreateStructTmpl error: %v", err)
	}

	// TODO handlerUpdateStructCommonTmpl 的用途待更新
	putStructCode, err := tmplExecuteWithFilter(data, handlerUpdateStructCommonTmpl, columnID)
	if err != nil {
		return "", fmt.Errorf("handlerUpdateStructTmpl error: %v", err)
	}

	// TODO handlerDetailStructCommonTmpl 的用途待更新
	getStructCode, err := tmplExecuteWithFilter(data, handlerDetailStructCommonTmpl, columnID, columnCreatedAt, columnUpdatedAt)
	if err != nil {
		return "", fmt.Errorf("handlerDetailStructTmpl error: %v", err)
	}

	return postStructCode + putStructCode + getStructCode, nil
}

func getCommonServiceStructCode(data tmplData) (string, error) {
	builder := strings.Builder{}
	err := serviceStructCommonTmpl.Execute(&builder, data)
	if err != nil {
		return "", err
	}
	code := builder.String()

	serviceCreateStructCode, err := tmplExecuteWithFilter(data, serviceCreateStructCommonTmpl)
	if err != nil {
		return "", fmt.Errorf("handle serviceCreateStructTmpl error: %v", err)
	}
	serviceCreateStructCode = strings.ReplaceAll(serviceCreateStructCode, "ID:", "Id:")

	serviceUpdateStructCode, err := tmplExecuteWithFilter(data, serviceUpdateStructCommonTmpl, columnID)
	if err != nil {
		return "", fmt.Errorf("handle serviceUpdateStructTmpl error: %v", err)
	}
	serviceUpdateStructCode = strings.ReplaceAll(serviceUpdateStructCode, "ID:", "Id:")

	code = strings.ReplaceAll(code, "// serviceCreateStructCode", serviceCreateStructCode)
	code = strings.ReplaceAll(code, "// serviceUpdateStructCode", serviceUpdateStructCode)

	return code, nil
}

func getCommonProtoFileCode(data tmplData, jsonNamedType int, isWebProto bool, isExtendedAPI bool) (string, error) {
	data.Fields = goTypeToProto(data.Fields, jsonNamedType, true)

	var err error
	builder := strings.Builder{}
	if isWebProto {
		if isExtendedAPI {
			err = protoFileForWebCommonTmpl.Execute(&builder, data)
		} else {
			err = protoFileForSimpleWebCommonTmpl.Execute(&builder, data)
		}
		if err != nil {
			return "", err
		}
	} else {
		if isExtendedAPI {
			err = protoFileCommonTmpl.Execute(&builder, data)
		} else {
			err = protoFileSimpleCommonTmpl.Execute(&builder, data)
		}
		if err != nil {
			return "", err
		}
	}
	code := builder.String()

	protoMessageCreateCode, err := tmplExecuteWithFilter2(data, protoMessageCreateCommonTmpl)
	if err != nil {
		return "", fmt.Errorf("handle protoMessageCreateCommonTmpl error: %v", err)
	}

	protoMessageUpdateCode, err := tmplExecuteWithFilter2(data, protoMessageUpdateCommonTmpl, columnID)
	if err != nil {
		return "", fmt.Errorf("handle protoMessageUpdateCommonTmpl error: %v", err)
	}
	if !isWebProto {
		srcStr := fmt.Sprintf(`, (tagger.tags) = "uri:\"%s\""`, getProtoFieldName(data.Fields))
		protoMessageUpdateCode = strings.ReplaceAll(protoMessageUpdateCode, srcStr, "")
	}

	protoMessageDetailCode, err := tmplExecuteWithFilter2(data, protoMessageDetailCommonTmpl, columnID, columnCreatedAt, columnUpdatedAt)
	if err != nil {
		return "", fmt.Errorf("handle protoMessageDetailCommonTmpl error: %v", err)
	}

	code = strings.ReplaceAll(code, "// protoMessageCreateCode", protoMessageCreateCode)
	code = strings.ReplaceAll(code, "// protoMessageUpdateCode", protoMessageUpdateCode)
	code = strings.ReplaceAll(code, "// protoMessageDetailCode", protoMessageDetailCode)
	code = strings.ReplaceAll(code, "*time.Time", "int64")
	code = strings.ReplaceAll(code, "time.Time", "int64")
	code = strings.ReplaceAll(code, "left_curly_bracket", "{")
	code = strings.ReplaceAll(code, "right_curly_bracket", "}")

	code = adaptedDbType2(data, isWebProto, code)

	return code, nil
}

func tmplExecuteWithFilter2(data tmplData, tmpl *template.Template, reservedColumns ...string) (string, error) {
	var newFields = []tmplField{}
	for _, field := range data.Fields {
		if isIgnoreFields(field.ColName, reservedColumns...) {
			continue
		}
		newFields = append(newFields, field)
	}
	data.Fields = newFields

	builder := strings.Builder{}
	err := tmpl.Execute(&builder, data)
	if err != nil {
		return "", fmt.Errorf("tmpl.Execute error: %v", err)
	}
	return builder.String(), nil
}

// simpleGoTypeToProtoType convert go type to proto type
func simpleGoTypeToProtoType(goType string) string {
	var protoType string
	switch goType {
	case "int", "int32":
		protoType = "int32"
	case "uint", "uint32":
		protoType = "uint32"
	case "int64":
		protoType = "int64"
	case "uint64":
		protoType = "uint64"
	case "string":
		protoType = "string"
	case "time.Time", "*time.Time":
		protoType = "string"
	case "float32":
		protoType = "float"
	case "float64":
		protoType = "double"
	case goTypeInts, "[]int64":
		protoType = "repeated int64"
	case "[]int32":
		protoType = "repeated int32"
	case "[]byte":
		protoType = "string"
	case goTypeStrings:
		protoType = "repeated string"
	case jsonTypeName:
		protoType = "string"
	default:
		protoType = "string"
	}
	return protoType
}

func adaptedDbType2(data tmplData, isWebProto bool, code string) string {
	if isWebProto {
		code = replaceProtoMessageFieldCode(code, webDefaultProtoMessageFieldCodes)
	} else {
		code = replaceProtoMessageFieldCode(code, grpcDefaultProtoMessageFieldCodes)
	}

	if data.ProtoSubStructs != "" {
		code += "\n" + data.ProtoSubStructs
	}

	return code
}

func firstLetterToUpper(str string) string {
	if len(str) == 0 {
		return str
	}

	if (str[0] >= 'A' && str[0] <= 'Z') || (str[0] >= 'a' && str[0] <= 'z') {
		return strings.ToUpper(str[:1]) + str[1:]
	}

	return str
}

// customFirstLetterToLower convert first letter to lower case, special case: ID -> id, IP -> ip
func customFirstLetterToLower(str string) string {
	str = firstLetterToLower(str)

	if len(str) == 2 {
		switch str {
		case "iD":
			str = "id"
		case "iP":
			str = "ip"
		}
	} else if len(str) == 3 {
		switch str {
		case "iDs":
			str = "ids"
		case "iPs":
			str = "ips"
		}
	}

	return str
}

// customEndOfLetterToLower 将单词复数形式的结尾字母从大写转为小写
// 保持单词主体部分的大小写不变，只处理复数后缀
//
// 参数:
//
//	srcStr: 原始单词 (如 "iD")
//	str: 原始单词的复数形式 (如 "iDS")
//
// 返回值:
//
//	处理后的复数形式，复数后缀保持小写 (如 "iDs")
//
// 示例:
//
//	srcStr: "iD", str: "iDS" → return: "iDs"
//	srcStr: "bUS", str: "bUSES" → return: "bUSes"
func customEndOfLetterToLower(srcStr string, str string) string {
	l := len(str) - len(srcStr)
	switch l {
	case 1:
		if str[len(str)-1] == 'S' {
			return str[:len(str)-1] + "s"
		}
	case 2:
		if str[len(str)-2:] == "ES" {
			return str[:len(str)-2] + "es"
		}
	}

	return str
}

// getHandlerGoType 根据字段信息获取适合处理器使用的 Go 类型
// 根据数据库驱动和字段类型进行特殊处理，确保在处理器层使用合适的数据类型
//
// 参数:
//
//	field: 模板字段信息，包含数据库驱动类型和原始字段类型等信息
//
// 返回值:
//
//	处理后的 Go 类型字符串，如 "string", "*bool", "*time.Time" 等
//
// 处理规则:
//  1. 对于 MySQL, PostgreSQL, TiDB 等数据库:
//     - JSON 类型转换为 string
//     - 布尔类型转换为 *bool (指针类型允许空值)
//     - DECIMAL 类型转换为 string
//  2. 对于 time.Time 类型，统一转换为 *time.Time
func getHandlerGoType(field *tmplField) string {
	var goType = field.GoType
	if field.DBDriver == DBDriverMysql || field.DBDriver == DBDriverPostgresql || field.DBDriver == DBDriverTidb {
		if field.rewriterField != nil {
			switch field.rewriterField.goType {
			case jsonTypeName:
				goType = "string"
			case boolTypeName, boolTypeTinyName:
				goType = "*bool"
			case decimalTypeName:
				goType = "string"
			}
		}
	}
	if field.GoType == "time.Time" {
		goType = "*time.Time"
	}
	return goType
}
