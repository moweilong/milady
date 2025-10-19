package goast

import (
	"fmt"
	"go/format"
	"os"
	"strings"
)

type CodeAstOption func(*CodeAst)

func defaultClientOptions() *CodeAst {
	return &CodeAst{
		ignoreFuncNameMap: make(map[string]struct{}),
	}
}

func (a *CodeAst) apply(opts ...CodeAstOption) {
	for _, opt := range opts {
		opt(a)
	}
}

// WithCoverSameFunc sets cover same function in the merged code
func WithCoverSameFunc() CodeAstOption {
	return func(a *CodeAst) {
		a.isCoverSameFunc = true
	}
}

// WithIgnoreMergeFunc sets ignore to merge the same function name in the two code
func WithIgnoreMergeFunc(funcName ...string) CodeAstOption {
	return func(a *CodeAst) {
		for _, name := range funcName {
			a.ignoreFuncNameMap[name] = struct{}{}
		}
	}
}

// CodeAst is the struct for code
type CodeAst struct {
	FilePath string
	Code     string

	AstInfos    []*AstInfo
	packageInfo *AstInfo
	importInfos []*AstInfo
	constInfos  []*AstInfo
	varInfos    []*AstInfo
	typeInfos   []*AstInfo
	funcInfos   []*AstInfo

	nonExistedConstCode    []string
	nonExistedVarCode      []string
	nonExistedTypeInfoMap  map[string]*TypeInfo // key is type name
	mergedStructMethodsMap map[string]struct{}  // key is struct name
	ignoreFuncNameMap      map[string]struct{}  // key is function name

	changeCodeFlag  bool
	isCoverSameFunc bool
}

// NewCodeAst creates a new CodeAst object from file path
func NewCodeAst(filePath string, opts ...CodeAstOption) (*CodeAst, error) {
	o := defaultClientOptions()
	o.apply(opts...)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	astInfos, err := ParseGoCode(filePath, data)
	if err != nil {
		return nil, err
	}

	codeAst := &CodeAst{
		FilePath:               filePath,
		Code:                   string(data),
		AstInfos:               astInfos,
		nonExistedTypeInfoMap:  make(map[string]*TypeInfo),
		mergedStructMethodsMap: make(map[string]struct{}),
		isCoverSameFunc:        o.isCoverSameFunc,
		ignoreFuncNameMap:      o.ignoreFuncNameMap,
	}
	codeAst.setSlices()

	return codeAst, nil
}

// NewCodeAstFromData creates a new CodeAst object from data
func NewCodeAstFromData(data []byte, opts ...CodeAstOption) (*CodeAst, error) {
	o := defaultClientOptions()
	o.apply(opts...)

	astInfos, err := ParseGoCode("", data)
	if err != nil {
		return nil, err
	}

	codeAst := &CodeAst{
		Code:                   string(data),
		AstInfos:               astInfos,
		nonExistedTypeInfoMap:  make(map[string]*TypeInfo),
		mergedStructMethodsMap: make(map[string]struct{}),
		isCoverSameFunc:        o.isCoverSameFunc,
		ignoreFuncNameMap:      o.ignoreFuncNameMap,
	}
	codeAst.setSlices()

	return codeAst, nil
}

func (a *CodeAst) setSlices() {
	for _, astInfo := range a.AstInfos {
		switch astInfo.Type {
		case PackageType:
			a.packageInfo = astInfo
		case ImportType:
			a.importInfos = append(a.importInfos, astInfo)
		case ConstType:
			a.constInfos = append(a.constInfos, astInfo)
		case VarType:
			a.varInfos = append(a.varInfos, astInfo)
		case TypeType:
			a.typeInfos = append(a.typeInfos, astInfo)
		case FuncType:
			a.funcInfos = append(a.funcInfos, astInfo)
		}
	}
}

func (a *CodeAst) mergeImportCode(genAst *CodeAst) error {
	if len(genAst.importInfos) == 0 {
		return nil
	}

	// 1. append import code to package

	if len(a.importInfos) == 0 {
		srcStr := a.packageInfo.Body
		dstStr := ""
		for _, info := range genAst.AstInfos {
			if info.IsImportType() {
				dstStr += info.Body + "\n"
			}
		}
		if strings.Count(a.Code, srcStr) > 1 {
			return errDuplication("mergeImportCode", srcStr)
		}

		a.Code = strings.Replace(a.Code, srcStr, srcStr+"\n\n"+dstStr, 1)
		a.changeCodeFlag = true
		return nil
	}

	// 2. append import code to import

	srcImportInfos, err := a.parseImportCode()
	if err != nil {
		return err
	}
	genImportInfos, err := genAst.parseImportCode()
	if err != nil {
		return err
	}

	srcLen := len(srcImportInfos)
	srcImportInfoMap := make(map[string]struct{}, srcLen)
	for _, srcIi := range srcImportInfos {
		srcImportInfoMap[srcIi.Path] = struct{}{}
	}

	var nonExistedImportInfos []*ImportInfo
	//var nonExistedImportPaths []string
	for _, genIfi := range genImportInfos {
		if _, ok := srcImportInfoMap[genIfi.Path]; !ok {
			nonExistedImportInfos = append(nonExistedImportInfos, genIfi)
			//nonExistedImportPaths = append(nonExistedImportPaths, genIfi.Path)
		}
	}

	if len(nonExistedImportInfos) > 0 {
		var srcStr = a.packageInfo.Body
		var dstStr = "import (\n"
		for _, info := range srcImportInfos {
			if info.Comment != "" {
				dstStr += info.Comment + "\n"
			}
			dstStr += "\t" + trimBody(info.Body, ImportType) + "\n"
		}
		for _, info := range nonExistedImportInfos {
			if info.Comment != "" {
				dstStr += info.Comment + "\n"
			}
			dstStr += "\t" + trimBody(info.Body, ImportType) + "\n"
		}
		dstStr += ")"

		a.Code = strings.Replace(a.Code, srcStr, srcStr+"\n\n"+dstStr, 1)
		a.changeCodeFlag = true
		for _, info := range a.importInfos {
			a.Code = strings.Replace(a.Code, info.Body, "", 1)
		}
	}

	return nil
}

func (a *CodeAst) compareConstCode(genAst *CodeAst) error {
	if len(genAst.constInfos) == 0 {
		return nil
	}

	if len(a.constInfos) == 0 {
		dstStr := ""
		for _, info := range genAst.AstInfos {
			if info.IsConstType() {
				if info.Comment != "" {
					dstStr += info.Comment + "\n"
				}
				dstStr += info.Body + "\n"
			}
		}
		a.nonExistedConstCode = append(a.nonExistedConstCode, dstStr)
		return nil
	}

	srcConstInfos, err := a.parseConstCode()
	if err != nil {
		return err
	}
	genConstInfos, err := genAst.parseConstCode()
	if err != nil {
		return err
	}

	srcLen := len(srcConstInfos)
	srcConstInfoMap := make(map[string]struct{}, srcLen)
	for _, srcCi := range srcConstInfos {
		srcConstInfoMap[srcCi.Name] = struct{}{}
	}

	var nonExistedConstInfos []*ConstInfo
	//var nonExistedConstNames []string
	for _, genCi := range genConstInfos {
		if _, ok := srcConstInfoMap[genCi.Name]; !ok {
			nonExistedConstInfos = append(nonExistedConstInfos, genCi)
			//nonExistedConstNames = append(nonExistedConstNames, genCi.Name)
		}
	}

	if len(nonExistedConstInfos) > 0 {
		var dstStr = "const (\n"
		for _, info := range nonExistedConstInfos {
			if info.Comment != "" {
				if !strings.HasPrefix(info.Comment, "\t") {
					info.Comment = "\t" + info.Comment
				}
				dstStr += info.Comment + "\n"
			}
			dstStr += "\t" + trimBody(info.Body, ConstType) + "\n"
		}
		dstStr += ")\n"
		a.nonExistedConstCode = append(a.nonExistedConstCode, dstStr)
	}

	return nil
}

func (a *CodeAst) compareVarCode(genAst *CodeAst) error {
	if len(genAst.varInfos) == 0 {
		return nil
	}

	if len(a.varInfos) == 0 {
		dstStr := ""
		for _, info := range genAst.AstInfos {
			if info.IsVarType() {
				if info.Comment != "" {
					dstStr += info.Comment + "\n"
				}
				dstStr += info.Body + "\n"
			}
		}
		a.nonExistedVarCode = append(a.nonExistedVarCode, dstStr)
		return nil
	}

	srcVarInfos, err := a.parseVarCode()
	if err != nil {
		return err
	}
	genVarInfos, err := genAst.parseVarCode()
	if err != nil {
		return err
	}

	srcLen := len(srcVarInfos)
	srcVarInfoMap := make(map[string]struct{}, srcLen)
	for _, srcVi := range srcVarInfos {
		srcVarInfoMap[srcVi.Name] = struct{}{}
	}

	var nonExistedVarInfos []*VarInfo
	//var nonExistedVarNames []string
	for _, genVi := range genVarInfos {
		if _, ok := srcVarInfoMap[genVi.Name]; !ok {
			nonExistedVarInfos = append(nonExistedVarInfos, genVi)
			//nonExistedVarNames = append(nonExistedVarNames, genVi.Name)
		} else {
			if genVi.Name == "_" && !strings.Contains(a.Code, strings.TrimSpace(genVi.Body)) {
				nonExistedVarInfos = append(nonExistedVarInfos, genVi)
				//nonExistedVarNames = append(nonExistedVarNames, genVi.Name)
			}
		}
	}

	if len(nonExistedVarInfos) > 0 {
		var dstStr = "var (\n"
		for _, info := range nonExistedVarInfos {
			if info.Comment != "" {
				if !strings.HasPrefix(info.Comment, "\t") {
					info.Comment = "\t" + info.Comment
				}
				dstStr += info.Comment + "\n"
			}
			dstStr += "\t" + trimBody(info.Body, VarType) + "\n"
		}
		dstStr += ")\n"
		a.nonExistedVarCode = append(a.nonExistedVarCode, dstStr)
	}

	return nil
}

func (a *CodeAst) mergeExistedTypeCode(genAst *CodeAst) error {
	srcTypeInfosMap, err := a.parseTypeCode()
	if err != nil {
		return err
	}
	genTypeInfosMap, err := genAst.parseTypeCode()
	if err != nil {
		return err
	}

	var srcTypeNameMap = make(map[string]struct{})
	for _, srcTypeInfos := range srcTypeInfosMap {
		for _, info := range srcTypeInfos {
			srcTypeNameMap[info.Name] = struct{}{}
		}
	}
	var nonExistedTypeInfoMap = make(map[string]*TypeInfo)

	for typeName, genTypeInfos := range genTypeInfosMap {
		// get non-existed type infos
		for _, info := range genTypeInfos {
			if _, ok := srcTypeNameMap[info.Name]; !ok {
				nonExistedTypeInfoMap[info.Name] = info
			}
		}

		// merge existed interface method code and struct fields code
		srcTypeInfos, ok := srcTypeInfosMap[typeName]
		if !ok {
			continue
		}
		srcTypeInfoMap := make(map[string]*TypeInfo, len(srcTypeInfos))
		for _, srcTi := range srcTypeInfos {
			srcTypeInfoMap[srcTi.Name] = srcTi
		}
		for _, genTypeInfo := range genTypeInfos {
			switch genTypeInfo.Type {
			case InterfaceType:
				if srcTypeInfo, ok := srcTypeInfoMap[genTypeInfo.Name]; ok {
					err = a.mergeInterfaceMethodCode(srcTypeInfo, genTypeInfo)
					if err != nil {
						return err
					}
				}
			case StructType:
				if srcTypeInfo, ok := srcTypeInfoMap[genTypeInfo.Name]; ok {
					err = a.mergeStructFieldsCode(srcTypeInfo, genTypeInfo)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	a.nonExistedTypeInfoMap = nonExistedTypeInfoMap

	return nil
}

func (a *CodeAst) mergeInterfaceMethodCode(srcTypeInfo *TypeInfo, genTypeInfo *TypeInfo) error {
	srcInterfaceInfos, err := ParseInterface(srcTypeInfo.Body)
	if err != nil {
		return err
	}
	genInterfaceInfos, err := ParseInterface(genTypeInfo.Body)
	if err != nil {
		return err
	}

	if len(srcInterfaceInfos) == 1 && len(genInterfaceInfos) == 1 {
		srcLastMethodStr := ""
		mLen := len(srcInterfaceInfos[0].MethodInfos)
		srcInterfaceMethodInfoMap := make(map[string]struct{}, mLen)
		for i, srcIm := range srcInterfaceInfos[0].MethodInfos {
			srcInterfaceMethodInfoMap[srcIm.Name] = struct{}{}
			if i == mLen-1 {
				srcLastMethodStr = srcIm.Body
			}
		}
		var newMethods []string
		for _, genMethodInfo := range genInterfaceInfos[0].MethodInfos {
			if _, ok := srcInterfaceMethodInfoMap[genMethodInfo.Name]; !ok {
				newMethodStr := ""
				if genMethodInfo.Comment != "" {
					newMethodStr += genMethodInfo.Comment + "\n"
				}
				newMethodStr += genMethodInfo.Body
				newMethods = append(newMethods, newMethodStr)
			}
		}
		if len(newMethods) > 0 {
			srcStr := srcTypeInfo.Body
			if strings.Count(a.Code, srcStr) > 1 {
				return errDuplication("mergeInterfaceMethodCode", srcStr)
			}
			dstStr := ""
			if len(srcInterfaceInfos[0].MethodInfos) == 0 {
				dstStr = "type " + srcInterfaceInfos[0].Name + " interface {\n" + strings.Join(newMethods, "\n") + "\n}"
			} else {
				dstStr = strings.Replace(srcStr, srcLastMethodStr, srcLastMethodStr+"\n"+strings.Join(newMethods, "\n"), 1)
			}
			a.Code = strings.Replace(a.Code, srcStr, dstStr, 1)
			a.changeCodeFlag = true
		}
	}
	return nil
}

func (a *CodeAst) mergeStructFieldsCode(srcTypeInfo *TypeInfo, genTypeInfo *TypeInfo) error {
	srcStructInfos, err := ParseStruct(srcTypeInfo.Body)
	if err != nil {
		return err
	}
	genStructInfos, err := ParseStruct(genTypeInfo.Body)
	if err != nil {
		return err
	}

	for name, genStructInfo := range genStructInfos {
		srcStructInfo, ok := srcStructInfos[name]
		if !ok {
			continue
		}

		fLen := len(srcStructInfo.Fields)
		srcLastFieldStr := ""
		srcFieldMap := make(map[string]struct{}, fLen)
		for i, field := range srcStructInfo.Fields {
			srcFieldMap[field.Name] = struct{}{}
			if i == fLen-1 {
				srcLastFieldStr = field.Body
			}
		}

		var newFields []string
		for _, field := range genStructInfo.Fields {
			newFieldStr := ""
			if _, ok := srcFieldMap[field.Name]; !ok {
				if field.Comment != "" {
					newFieldStr += field.Comment + "\n"
				}
				newFieldStr += field.Body
				newFields = append(newFields, newFieldStr)
			}
		}
		if len(newFields) > 0 {
			srcStr := srcTypeInfo.Body
			if strings.Count(a.Code, srcStr) > 1 {
				return errDuplication("mergeStructFieldsCode", srcStr)
			}
			dstStr := ""
			if len(srcStructInfo.Fields) == 0 {
				dstStr = "type " + srcStructInfo.Name + " struct {\n" + strings.Join(newFields, "\n") + "\n}"
			} else {
				dstStr = strings.Replace(srcStr, srcLastFieldStr, srcLastFieldStr+"\n"+strings.Join(newFields, "\n"), 1)
			}
			a.Code = strings.Replace(a.Code, srcStr, dstStr, 1)
			a.changeCodeFlag = true
		}
	}

	return nil
}

func (a *CodeAst) mergeStructMethodsCode(genAst *CodeAst) error {
	srcImportInfoMap := ParseStructMethods(a.AstInfos)
	genImportInfoMap := ParseStructMethods(genAst.AstInfos)

	for structName, genMethods := range genImportInfoMap {
		var nonExistedImports []string
		var lastMethodFuncCode string
		if srcMethods, ok := srcImportInfoMap[structName]; ok {
			var srcMethodMap = make(map[string]struct{}, len(srcMethods))
			for i, srcMethod := range srcMethods {
				srcMethodMap[srcMethod.Name] = struct{}{}
				if i == len(srcMethods)-1 {
					lastMethodFuncCode = srcMethod.Body
				}
			}
			for _, genMethod := range genMethods {
				if _, isExisted := srcMethodMap[genMethod.Name]; !isExisted {
					nonExistedImports = append(nonExistedImports, genMethod.Comment+"\n"+genMethod.Body)
				}
			}
		}
		if len(nonExistedImports) > 0 {
			srcStr := lastMethodFuncCode
			if strings.Count(a.Code, srcStr) > 1 {
				return errDuplication("mergeStructMethodsCode", srcStr)
			}
			dstStr := lastMethodFuncCode + "\n\n" + strings.Join(nonExistedImports, "\n\n")
			a.Code = strings.Replace(a.Code, srcStr, dstStr, 1)
			a.changeCodeFlag = true
			a.mergedStructMethodsMap[structName] = struct{}{}
		}
	}

	return nil
}

func (a *CodeAst) coverFuncCode(genAst *CodeAst) {
	var srcFuncNameMap = make(map[string]*AstInfo)
	for _, srcFuncInfo := range a.funcInfos {
		srcFuncNameMap[srcFuncInfo.GetName()] = srcFuncInfo
	}
	for _, genFuncInfo := range genAst.funcInfos {
		genFuncName := genFuncInfo.GetName()
		if genFuncName == "init" || genFuncName == "_" {
			continue
		}

		var ignoreFuncName string
		if len(genFuncInfo.Names) == 2 {
			ignoreFuncName = genFuncInfo.Names[0]
		} else {
			ignoreFuncName = genFuncName
		}
		if _, ok := a.ignoreFuncNameMap[ignoreFuncName]; ok {
			continue
		}

		if srcFuncInfo, ok := srcFuncNameMap[genFuncName]; ok {
			srcStr := srcFuncInfo.Body
			dstStr := genFuncInfo.Body
			comment := ""
			if srcFuncInfo.Comment == "" && genFuncInfo.Comment != "" {
				comment = genFuncInfo.Comment
			}
			if comment != "" {
				dstStr = comment + "\n" + dstStr
			}
			a.Code = strings.Replace(a.Code, srcStr, dstStr, 1)
			a.changeCodeFlag = true
		}
	}
}

// appends non-existed code to the end of the source code.
func (a *CodeAst) appendNonExistedCode(genAsts []*AstInfo) error { // nolint
	srcNameAstMap := make(map[string]struct{}, len(a.AstInfos))
	for _, info := range a.AstInfos {
		srcNameAstMap[info.GetName()] = struct{}{}
	}

	if len(a.nonExistedConstCode) > 0 {
		a.Code += "\n" + strings.Join(a.nonExistedConstCode, "\n")
		a.changeCodeFlag = true
	}
	if len(a.nonExistedVarCode) > 0 {
		a.Code += "\n" + strings.Join(a.nonExistedVarCode, "\n")
		a.changeCodeFlag = true
	}

	var appendCodes []string
	for _, genAst := range genAsts {
		if genAst.IsPackageType() || genAst.IsImportType() || genAst.IsConstType() || genAst.IsVarType() {
			continue
		}
		if genAst.IsFuncType() && len(genAst.Names) == 2 {
			if _, ok := a.mergedStructMethodsMap[genAst.Names[1]]; ok {
				continue
			}
		}

		isNeedAppend := false
		name := genAst.GetName()
		if _, ok := srcNameAstMap[name]; !ok {
			if genAst.IsTypeType() && len(genAst.Names) > 1 {
				var nonExistedTypes []string
				for _, name := range genAst.Names {
					var dstStr string
					if info, ok := a.nonExistedTypeInfoMap[name]; ok {
						if info.Comment != "" {
							dstStr += info.Comment + "\n"
						}
						dstStr += info.Body
						nonExistedTypes = append(nonExistedTypes, dstStr)
					}
				}
				if len(nonExistedTypes) > 0 {
					appendCodes = append(appendCodes, "type (\n"+strings.Join(nonExistedTypes, "\n")+"\n)")
					continue
				}
			}

			isNeedAppend = true
		} else {
			if name == "_" && !strings.Contains(a.Code, genAst.Body) {
				isNeedAppend = true
			}
			if genAst.IsFuncType() && name == "init" && !strings.Contains(a.Code, genAst.Body) {
				isNeedAppend = true
			}
		}
		if isNeedAppend {
			comment := ""
			if genAst.Comment != "" {
				comment = genAst.Comment
			}
			appendCodes = append(appendCodes, comment+"\n"+genAst.Body)
		}
	}

	if len(appendCodes) > 0 {
		a.Code += strings.Join(appendCodes, "\n\n") + "\n"
		a.changeCodeFlag = true
	}

	return nil
}

func (a *CodeAst) parseImportCode() ([]*ImportInfo, error) {
	body := ""
	for _, info := range a.importInfos {
		if info.Comment != "" {
			body += info.Comment + "\n"
		}
		body += info.Body + "\n"
	}
	return ParseImportGroup(body)
}

func (a *CodeAst) parseConstCode() ([]*ConstInfo, error) {
	body := ""
	for _, info := range a.constInfos {
		if info.Comment != "" {
			body += info.Comment + "\n"
		}
		body += info.Body + "\n"
	}
	return ParseConstGroup(body)
}

func (a *CodeAst) parseVarCode() ([]*VarInfo, error) {
	body := ""
	for _, info := range a.varInfos {
		if info.Comment != "" {
			body += info.Comment + "\n"
		}
		body += info.Body + "\n"
	}
	return ParseVarGroup(body)
}

func (a *CodeAst) parseTypeCode() (map[string][]*TypeInfo, error) {
	body := ""
	for _, info := range a.typeInfos {
		if info.Comment != "" {
			body += info.Comment + "\n"
		}
		body += info.Body + "\n"
	}

	typeInfos, err := ParseTypeGroup(body)
	if err != nil {
		return nil, err
	}

	typeMap := make(map[string][]*TypeInfo, len(typeInfos))
	for _, info := range typeInfos {
		if info.Name == "" {
			continue
		}
		if _, ok := typeMap[info.Name]; !ok {
			typeMap[info.Name] = []*TypeInfo{}
		}
		typeMap[info.Name] = append(typeMap[info.Name], info)
	}
	return typeMap, nil
}

func errDuplication(marker string, srcStr string) error {
	return fmt.Errorf("%s: multiple duplicate string `%s` exists, please modify the source code to ensure uniqueness", marker, srcStr)
}

func trimBody(body string, codeType string) string {
	body = strings.TrimSpace(body)
	return strings.TrimPrefix(body, codeType+" ")
}

// MergeGoFile merges two Go code files into one.
func MergeGoFile(srcFile string, genFile string, opts ...CodeAstOption) (*CodeAst, error) {
	srcAst, err := NewCodeAst(srcFile, opts...)
	if err != nil {
		return nil, err
	}

	genAst, err := NewCodeAst(genFile)
	if err != nil {
		return nil, err
	}

	return mergeCode(srcAst, genAst)
}

// MergeGoCode merges two Go code strings into one.
func MergeGoCode(srcCode []byte, genCode []byte, opts ...CodeAstOption) (*CodeAst, error) {
	srcAst, err := NewCodeAstFromData(srcCode, opts...)
	if err != nil {
		return nil, err
	}

	genAst, err := NewCodeAstFromData(genCode, opts...)
	if err != nil {
		return nil, err
	}

	return mergeCode(srcAst, genAst)
}

func mergeCode(srcAst *CodeAst, genAst *CodeAst) (*CodeAst, error) {
	// merge import code
	err := srcAst.mergeImportCode(genAst)
	if err != nil {
		return nil, err
	}

	// compare const code
	err = srcAst.compareConstCode(genAst)
	if err != nil {
		return nil, err
	}

	// compare var code
	err = srcAst.compareVarCode(genAst)
	if err != nil {
		return nil, err
	}

	// merge interface method and struct fields code
	err = srcAst.mergeExistedTypeCode(genAst)
	if err != nil {
		return nil, err
	}

	// merge struct method function code
	err = srcAst.mergeStructMethodsCode(genAst)
	if err != nil {
		return nil, err
	}

	if srcAst.isCoverSameFunc {
		// cover same function code
		srcAst.coverFuncCode(genAst)
	}

	// append non-existed code
	err = srcAst.appendNonExistedCode(genAst.AstInfos)
	if err != nil {
		return nil, err
	}

	if srcAst.changeCodeFlag {
		data, err := format.Source([]byte(srcAst.Code))
		if err == nil {
			srcAst.Code = string(data)
		}
	}

	return srcAst, nil
}
