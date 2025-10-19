// Package goast is a library for parsing Go code and extracting information from it.
package goast

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

const (
	// ast types

	PackageType = "package"
	ImportType  = "import"
	ConstType   = "const"
	VarType     = "var"
	FuncType    = "func"
	TypeType    = "type"

	// for TypeType

	StructType    = "struct"
	InterfaceType = "interface"
	ArrayType     = "array"
	MapType       = "map"
	ChanType      = "chan"
)

// AstInfo Go code block information
type AstInfo struct {
	// Type is the type of the code block, such as "func", "type", "const", "var", "import", "package".
	Type string

	// Names is the name of the code block, such as "func Name", "type Names", "const Names", "var Names", "import Paths".
	// If Type is "func", a standalone function without a receiver has a single name.
	// If the function is a method belonging to a struct, it has two names: the first
	// represents the function name, and the second represents the struct name.
	Names []string

	Comment string

	Body string
}

func (a *AstInfo) IsPackageType() bool {
	return a.Type == PackageType
}
func (a *AstInfo) IsImportType() bool {
	return a.Type == ImportType
}
func (a *AstInfo) IsConstType() bool {
	return a.Type == ConstType
}
func (a *AstInfo) IsVarType() bool {
	return a.Type == VarType
}
func (a *AstInfo) IsTypeType() bool {
	return a.Type == TypeType
}
func (a *AstInfo) IsFuncType() bool {
	return a.Type == FuncType
}

func (a *AstInfo) GetName() string {
	return strings.Join(a.Names, ",")
}

// ParseFile parses a go file and returns a list of AstInfo
func ParseFile(goFilePath string) ([]*AstInfo, error) {
	filename := filepath.Base(goFilePath)
	data, err := os.ReadFile(goFilePath)
	if err != nil {
		return nil, err
	}
	return ParseGoCode(filename, data)
}

// ParseGoCode parses a go code and returns a list of AstInfo
func ParseGoCode(filename string, data []byte) ([]*AstInfo, error) {
	src := string(data)
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, src, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var astInfos []*AstInfo

	pkgNames, pkgComment, pkgBody := getPackageCode(fset, file, src)
	astInfos = append(astInfos, &AstInfo{Type: PackageType, Names: pkgNames, Comment: pkgComment, Body: pkgBody})

	// traverse AST code blocks
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			receiverName, comment, body := getFuncDeclCode(fset, node, src)
			names := []string{node.Name.Name}
			if receiverName != "" {
				names = append(names, receiverName)
			}
			astInfos = append(astInfos, &AstInfo{Type: FuncType, Names: names, Comment: comment, Body: body})

		case *ast.GenDecl:
			names, comment, body := getGenDeclCode(fset, node, src)
			astInfos = append(astInfos, &AstInfo{Type: node.Tok.String(), Names: names, Comment: comment, Body: body})

			//case *ast.BadDecl:
			//	code := getBadDeclCode(fset, node, src)
			//	println(code)
		}
		return true
	})

	return astInfos, nil
}

func getPackageCode(fset *token.FileSet, f *ast.File, src string) (names []string, comment string, body string) {
	if f.Doc != nil {
		var comments []string
		for _, cmt := range f.Doc.List {
			comments = append(comments, cmt.Text)
		}
		comment = strings.Join(comments, "\n")
	}

	packagePos := fset.Position(f.Package)
	body = src[packagePos.Offset : packagePos.Offset+len("package "+f.Name.Name)]

	return []string{f.Name.Name}, comment, body
}

func getFuncDeclCode(fset *token.FileSet, fn *ast.FuncDecl, src string) (receiverName string, comment string, body string) {
	if fn == nil {
		return "", "", ""
	}

	if fn.Recv != nil && len(fn.Recv.List) > 0 {
		recvType := fn.Recv.List[0].Type
		switch t := recvType.(type) {
		case *ast.StarExpr:
			if ident, ok := t.X.(*ast.Ident); ok {
				receiverName = ident.Name
			}
		case *ast.Ident:
			receiverName = t.Name
		}
	}

	commentText := ""
	if fn.Doc != nil {
		var parts []string
		for _, c := range fn.Doc.List {
			parts = append(parts, strings.TrimSpace(c.Text))
		}
		commentText = strings.Join(parts, "\n")
	}

	start := fn.Type.Func     // the starting position of the func keyword
	end := fn.Body.Rbrace + 1 // end position of function body
	return receiverName, commentText, getCodeFromPos(fset, start, end, src)
}

func getCodeFromPos(fset *token.FileSet, start, end token.Pos, src string) string {
	file := fset.File(start)
	if file == nil {
		return ""
	}
	startOffset := file.Offset(start)
	endOffset := file.Offset(end)
	if startOffset < 0 || endOffset > len(src) || startOffset >= endOffset {
		return ""
	}
	return src[startOffset:endOffset]
}

func getGenDeclCode(fset *token.FileSet, gen *ast.GenDecl, src string) (names []string, comment string, body string) {
	if gen == nil {
		return nil, "", ""
	}

	commentText := ""
	if gen.Doc != nil {
		var parts []string
		for _, c := range gen.Doc.List {
			parts = append(parts, strings.TrimSpace(c.Text))
		}
		commentText = strings.Join(parts, "\n")
	}

	start := gen.TokPos // keyword starting position
	var end token.Pos
	if gen.Rparen.IsValid() {
		end = gen.Rparen + 1 // end position of parentheses
	} else if len(gen.Specs) > 0 {
		lastSpec := gen.Specs[len(gen.Specs)-1]
		end = lastSpec.End() // end position of the last Spec
	} else {
		end = start + token.Pos(len(gen.Tok.String())) // in the case of keywords only
	}

	return getGenName(gen), commentText, getCodeFromPos(fset, start, end, src)
}

func getGenName(gen *ast.GenDecl) []string {
	var names []string
	switch gen.Tok {
	case token.IMPORT:
		for _, spec := range gen.Specs {
			imp := spec.(*ast.ImportSpec)
			names = append(names, imp.Path.Value)
		}

	case token.CONST:
		for _, spec := range gen.Specs {
			val := spec.(*ast.ValueSpec)
			names = append(names, val.Names[0].Name)
		}

	case token.TYPE:
		for _, spec := range gen.Specs {
			typ := spec.(*ast.TypeSpec)
			names = append(names, typ.Name.Name)
		}

	case token.VAR:
		for _, spec := range gen.Specs {
			val := spec.(*ast.ValueSpec)
			names = append(names, val.Names[0].Name)
		}
	}
	return names
}

// -----------------------------------------------------------------------------------

func adaptPackage(src string) string {
	if len(src) > 50 {
		if strings.Contains(src[:50], "package ") {
			return src
		}
	}
	if strings.Contains(src, "\npackage ") {
		return src
	}
	return "package parse\n\n" + src
}

// nolint
func parseBody(body string) (*token.FileSet, *ast.File, string, error) {
	src := adaptPackage(body)
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		f, err = parser.ParseFile(fset, "", "package parse\n\n"+src, parser.ParseComments)
		if err != nil {
			return nil, nil, "", err
		}
	}
	return fset, f, src, nil
}

type ImportInfo struct {
	Path    string
	Alias   string
	Comment string
	Body    string
}

// ParseImportGroup parse import group from source code
func ParseImportGroup(body string) ([]*ImportInfo, error) {
	fset, f, src, err := parseBody(body)
	if err != nil {
		return nil, err
	}

	var srcLines = strings.Split(src, "\n")
	var imports []*ImportInfo

	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.IMPORT {
			continue
		}

		for _, spec := range genDecl.Specs {
			importSpec := spec.(*ast.ImportSpec)

			// get path
			path := strings.Trim(importSpec.Path.Value, `"`)

			// get alias
			var alias string
			if importSpec.Name != nil {
				alias = importSpec.Name.Name
			}

			// get comment doc
			var comment string
			if importSpec.Doc != nil {
				comment = getSrcContent(srcLines, fset.Position(importSpec.Doc.List[0].Pos()).Line,
					fset.Position(importSpec.Doc.List[len(importSpec.Doc.List)-1].End()).Line)
			}

			// get source code of import path
			code := getSrcContent(srcLines, fset.Position(importSpec.Pos()).Line, fset.Position(importSpec.End()).Line)

			imports = append(imports, &ImportInfo{
				Path:    path,
				Alias:   alias,
				Comment: comment,
				Body:    code,
			})
		}
	}

	return imports, nil
}

type ConstInfo struct {
	Name    string
	Value   string
	Comment string
	Body    string
}

// ParseConstGroup parse const group from source code
func ParseConstGroup(body string) ([]*ConstInfo, error) {
	fset, f, src, err := parseBody(body)
	if err != nil {
		return nil, err
	}

	var srcLines = strings.Split(src, "\n")
	var consts []*ConstInfo

	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.CONST {
			continue
		}

		singleComment := ""
		if genDecl.Doc != nil {
			singleComment = getSrcContent(srcLines, fset.Position(genDecl.Doc.List[0].Pos()).Line,
				fset.Position(genDecl.Doc.List[len(genDecl.Doc.List)-1].End()).Line)
		}

		for _, spec := range genDecl.Specs {
			valueSpec := spec.(*ast.ValueSpec)
			for i, name := range valueSpec.Names {
				constName := name.Name

				// get line content
				var comment string
				if valueSpec.Doc != nil {
					comment = getSrcContent(srcLines, fset.Position(valueSpec.Doc.List[0].Pos()).Line,
						fset.Position(valueSpec.Doc.List[len(valueSpec.Doc.List)-1].End()).Line)
				}
				if len(genDecl.Specs) == 1 && singleComment != "" && comment == "" {
					comment = singleComment
				}

				// get code content
				code := getSrcContent(srcLines, fset.Position(valueSpec.Pos()).Line, fset.Position(valueSpec.End()).Line)

				// get value (if exists)
				var constValue string
				if i < len(valueSpec.Values) {
					if basicLit, ok := valueSpec.Values[i].(*ast.BasicLit); ok {
						constValue = basicLit.Value
					}
				}

				consts = append(consts, &ConstInfo{
					Name:    constName,
					Value:   constValue,
					Comment: comment,
					Body:    code,
				})
			}
		}
	}

	return consts, nil
}

type VarInfo struct {
	Name    string
	Value   string
	Comment string
	Body    string
}

// ParseVarGroup parse var group from source code
func ParseVarGroup(body string) ([]*VarInfo, error) {
	fset, f, src, err := parseBody(body)
	if err != nil {
		return nil, err
	}

	var srcLines = strings.Split(src, "\n")
	var vars []*VarInfo

	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.VAR {
			continue
		}

		singleComment := ""
		if genDecl.Doc != nil {
			singleComment = getSrcContent(srcLines, fset.Position(genDecl.Doc.List[0].Pos()).Line,
				fset.Position(genDecl.Doc.List[len(genDecl.Doc.List)-1].End()).Line)
		}

		for _, spec := range genDecl.Specs {
			valueSpec := spec.(*ast.ValueSpec)
			for i, name := range valueSpec.Names {
				varName := name.Name

				// get comment
				var comment string
				if valueSpec.Doc != nil {
					comment = getSrcContent(srcLines, fset.Position(valueSpec.Doc.List[0].Pos()).Line,
						fset.Position(valueSpec.Doc.List[len(valueSpec.Doc.List)-1].End()).Line)
				}
				if len(genDecl.Specs) == 1 && singleComment != "" && comment == "" {
					comment = singleComment
				}

				// get code content
				code := getSrcContent(srcLines, fset.Position(valueSpec.Pos()).Line, fset.Position(valueSpec.End()).Line)

				// get var value (if exists)
				var varValue string
				if i < len(valueSpec.Values) {
					if basicLit, ok := valueSpec.Values[i].(*ast.BasicLit); ok {
						varValue = basicLit.Value
					}
				}

				vars = append(vars, &VarInfo{
					Name:    varName,
					Value:   varValue,
					Comment: comment,
					Body:    code,
				})
			}
		}
	}
	return vars, nil
}

type TypeInfo struct {
	Type    string
	Name    string
	Comment string
	Body    string
	IsIdent bool
}

// ParseTypeGroup parse type group from source code
func ParseTypeGroup(body string) ([]*TypeInfo, error) {
	fset, f, src, err := parseBody(body)
	if err != nil {
		return nil, err
	}

	var srcLines = strings.Split(src, "\n")
	var types []*TypeInfo

	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		singleComment := ""
		if genDecl.Doc != nil {
			singleComment = getSrcContent(srcLines, fset.Position(genDecl.Doc.List[0].Pos()).Line,
				fset.Position(genDecl.Doc.List[len(genDecl.Doc.List)-1].End()).Line)
		}

		for _, spec := range genDecl.Specs {
			typeSpec := spec.(*ast.TypeSpec)
			typeName := typeSpec.Name.Name

			// get comment
			var comment string
			if typeSpec.Doc != nil {
				comment = getSrcContent(srcLines, fset.Position(typeSpec.Doc.List[0].Pos()).Line,
					fset.Position(typeSpec.Doc.List[len(typeSpec.Doc.List)-1].End()).Line)
			}
			if len(genDecl.Specs) == 1 && singleComment != "" && comment == "" {
				comment = singleComment
			}

			// get code content
			code := getSrcContent(srcLines, fset.Position(typeSpec.Pos()).Line, fset.Position(typeSpec.End()).Line)

			// get type definition
			var isIdent bool
			var typeDef string
			switch t := typeSpec.Type.(type) {
			case *ast.StructType:
				typeDef = StructType
			case *ast.InterfaceType:
				typeDef = InterfaceType
			case *ast.FuncType:
				typeDef = FuncType
			case *ast.MapType:
				typeDef = MapType
			case *ast.ArrayType:
				typeDef = ArrayType
			case *ast.ChanType:
				typeDef = ChanType
			case *ast.Ident:
				typeDef = t.Name
				isIdent = true
			default:
				typeDef = fmt.Sprintf("%T", t)
			}

			types = append(types, &TypeInfo{
				Type:    typeDef,
				Name:    typeName,
				Comment: comment,
				Body:    code,
				IsIdent: isIdent,
			})
		}
	}

	return types, nil
}

type InterfaceInfo struct {
	Name        string
	Comment     string
	MethodInfos []*MethodInfo
}

// ParseInterface parse interface group from source code
func ParseInterface(body string) ([]*InterfaceInfo, error) {
	fset, f, src, err := parseBody(body)
	if err != nil {
		return nil, err
	}

	var srcLines = strings.Split(src, "\n")
	var interfaceInfos []*InterfaceInfo

	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		singleComment := ""
		if genDecl.Doc != nil {
			singleComment = getSrcContent(srcLines, fset.Position(genDecl.Doc.List[0].Pos()).Line,
				fset.Position(genDecl.Doc.List[len(genDecl.Doc.List)-1].End()).Line)
		}

		var methodInfos []*MethodInfo
		for _, spec := range genDecl.Specs {
			typeSpec := spec.(*ast.TypeSpec)
			interfaceType, ok := typeSpec.Type.(*ast.InterfaceType)
			if !ok {
				continue
			}

			interfaceName := typeSpec.Name.Name

			// get interface comment
			var interfaceComment string
			if typeSpec.Doc != nil {
				interfaceComment = getSrcContent(srcLines, fset.Position(typeSpec.Doc.List[0].Pos()).Line,
					fset.Position(typeSpec.Doc.List[len(typeSpec.Doc.List)-1].End()).Line)
			}
			if len(genDecl.Specs) == 1 && singleComment != "" && interfaceComment == "" {
				interfaceComment = singleComment
			}

			var isIdent bool
			for _, method := range interfaceType.Methods.List {
				// get method name
				var methodName string
				switch t := method.Type.(type) {
				case *ast.FuncType:
					if len(method.Names) > 0 {
						methodName = method.Names[0].Name
					}
				case *ast.Ident: // embedded interface
					methodName = t.Name
					isIdent = true
				default:
					continue
				}

				// get method comment
				var methodComment string
				if method.Doc != nil {
					methodComment = getSrcContent(srcLines, fset.Position(method.Doc.List[0].Pos()).Line,
						fset.Position(method.Doc.List[len(method.Doc.List)-1].End()).Line)
				}

				// get method line content
				code := getSrcContent(srcLines, fset.Position(method.Pos()).Line, fset.Position(method.End()).Line)

				methodInfos = append(methodInfos, &MethodInfo{
					Name:         methodName,
					Comment:      methodComment,
					Body:         code,
					ReceiverName: interfaceName,
					IsIdent:      isIdent,
				})
			}
			interfaceInfos = append(interfaceInfos, &InterfaceInfo{
				Name:        interfaceName,
				Comment:     interfaceComment,
				MethodInfos: methodInfos,
			})
		}
	}
	return interfaceInfos, nil
}

// MethodInfo method function info
type MethodInfo struct {
	Name         string
	Comment      string
	Body         string
	ReceiverName string
	IsIdent      bool
}

// ParseStructMethods parse struct methods from ast infos
func ParseStructMethods(astInfos []*AstInfo) map[string][]*MethodInfo {
	var m = make(map[string][]*MethodInfo) // map[structName][]*MethodInfo

	for _, info := range astInfos {
		if !info.IsFuncType() {
			continue
		}
		if len(info.Names) == 2 {
			funcName := info.Names[0]
			structName := info.Names[1]
			methodAst := &MethodInfo{
				Name:         funcName,
				Comment:      info.Comment,
				Body:         info.Body,
				ReceiverName: structName,
			}
			if methodInfos, ok := m[structName]; !ok {
				m[structName] = []*MethodInfo{methodAst}
			} else {
				m[structName] = append(methodInfos, methodAst)
			}
		}
	}

	return m
}

type StructInfo struct {
	Name    string
	Comment string
	Fields  []*StructFieldInfo
}

type StructFieldInfo struct {
	Name    string
	Type    string
	Comment string
	Body    string
}

// ParseStruct parse struct info from source code
func ParseStruct(body string) (map[string]*StructInfo, error) { //nolint
	fset, f, src, err := parseBody(body)
	if err != nil {
		return nil, err
	}

	var srcLines = strings.Split(src, "\n")
	var structInfos = make(map[string]*StructInfo)

	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		singleComment := ""
		if genDecl.Doc != nil {
			singleComment = getSrcContent(srcLines, fset.Position(genDecl.Doc.List[0].Pos()).Line,
				fset.Position(genDecl.Doc.List[len(genDecl.Doc.List)-1].End()).Line)
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}
			structName := typeSpec.Name.Name

			// get struct comment
			var structComment string
			if typeSpec.Doc != nil {
				structComment = getSrcContent(srcLines, fset.Position(typeSpec.Doc.List[0].Pos()).Line,
					fset.Position(typeSpec.Doc.List[len(typeSpec.Doc.List)-1].End()).Line)
			}
			if len(genDecl.Specs) == 1 && singleComment != "" && structComment == "" {
				structComment = singleComment
			}

			var fields []*StructFieldInfo
			for _, field := range structType.Fields.List {
				var fieldNames []string
				if len(field.Names) > 0 {
					for _, name := range field.Names {
						fieldNames = append(fieldNames, name.Name)
					}
				} else {
					// 处理嵌入字段
					switch x := field.Type.(type) {
					case *ast.Ident:
						fieldNames = append(fieldNames, x.Name)
					case *ast.StarExpr:
						if ident, ok := x.X.(*ast.Ident); ok {
							fieldNames = append(fieldNames, "*"+ident.Name)
						}
					}
				}

				// get field name
				var fieldName string
				if len(fieldNames) > 0 {
					fieldName = fieldNames[0]
				}

				// get comment
				var comment string
				if field.Doc != nil {
					comment = getSrcContent(srcLines, fset.Position(field.Doc.List[0].Pos()).Line,
						fset.Position(field.Doc.List[len(field.Doc.List)-1].End()).Line)
				}

				// get source code of field
				code := getSrcContent(srcLines, fset.Position(field.Pos()).Line, fset.Position(field.End()).Line)

				fields = append(fields, &StructFieldInfo{
					Name:    fieldName,
					Type:    getTypeString(field.Type),
					Comment: comment,
					Body:    code,
				})
			}
			structInfos[structName] = &StructInfo{
				Name:    structName,
				Comment: structComment,
				Fields:  fields,
			}
		}
	}

	return structInfos, nil
}

func getTypeString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.ArrayType:
		return "[]" + getTypeString(t.Elt)
	case *ast.StructType:
		return "struct"
	case *ast.SelectorExpr:
		return getTypeString(t.X) + "." + t.Sel.Name
	case *ast.StarExpr:
		return "*" + getTypeString(t.X)
	case *ast.MapType:
		return "map[" + getTypeString(t.Key) + "]" + getTypeString(t.Value)
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.ChanType:
		return "chan " + getTypeString(t.Value)
	default:
		return "unknown"
	}
}

func getSrcContent(srcLines []string, start, end int) string {
	var srcContent string
	l := len(srcLines)
	if start < 1 || start > l || end < 1 || end > l {
		return ""
	}
	if start == end {
		srcContent = srcLines[start-1]
	} else {
		srcContent = strings.Join(srcLines[start-1:end], "\n")
	}
	return srcContent
}
