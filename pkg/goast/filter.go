package goast

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// FuncInfo represents function information
type FuncInfo struct {
	Name    string
	Comment string
}

// ExtractComment extracts function comments in Go code
func (f FuncInfo) ExtractComment() string {
	if f.Comment == "" {
		return ""
	}
	comment := f.Comment

	// regular matching `//` or `/* */` comments
	lineComment := regexp.MustCompile(`(?m)^//\s?`)
	blockComment := regexp.MustCompile(`(?m)/\*|\*/`)

	// remove the `//` or `/* */` tags
	comment = lineComment.ReplaceAllString(comment, "")
	comment = blockComment.ReplaceAllString(comment, "")

	// remove the space at the beginning of the line and split the line
	lines := strings.Split(comment, "\n")
	for i := range lines {
		lines[i] = strings.TrimSpace(lines[i])
	}

	// output the comment string
	commentStr := strings.Join(lines, "\n")
	commentStr = strings.TrimSpace(commentStr)
	commentStr = strings.TrimPrefix(commentStr, f.Name)
	return strings.TrimSpace(commentStr)
}

// containsPanicCall determine if there is a panic("implement me"), or customized flag, e.g. panic("ai to do")
func containsPanicCall(fn *ast.FuncDecl, customFlag ...string) bool {
	found := false
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		ce, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		ident, ok := ce.Fun.(*ast.Ident)
		if !ok || ident.Name != "panic" {
			return true
		}
		if len(ce.Args) > 0 {
			basicLit, ok := ce.Args[0].(*ast.BasicLit)
			if ok && basicLit.Kind == token.STRING {
				s, err := strconv.Unquote(basicLit.Value)
				if err == nil {
					if strings.HasPrefix(s, "implement me") {
						found = true
						return false // stop traversing if you find it.
					}
					for _, flag := range customFlag {
						if strings.Contains(s, flag) {
							found = true
							return false // stop traversing if you find it.
						}
					}
				}
			}
		}
		return true
	})
	return found
}

// interval indicates an area to delete
type interval struct {
	start token.Pos
	end   token.Pos
}

// FilterFuncCodeByFile filters out the code of functions that contain panic("implement me") or customized flag, e.g. panic("ai to do")
func FilterFuncCodeByFile(goFilePath string, customFlag ...string) ([]byte, []FuncInfo, error) {
	filename := filepath.Base(goFilePath)
	data, err := os.ReadFile(goFilePath)
	if err != nil {
		return nil, nil, err
	}

	return FilterFuncCode(filename, data, customFlag...)
}

// FilterFuncCode filters out the code of functions that contain panic("implement me") or customized flag, e.g. panic("ai to do")
func FilterFuncCode(filename string, data []byte, customFlag ...string) ([]byte, []FuncInfo, error) {
	fset := token.NewFileSet()
	// parse source code for comments
	f, err := parser.ParseFile(fset, filename, string(data), parser.ParseComments)
	if err != nil {
		return nil, nil, err
	}

	// used to record the code interval corresponding to the function to be deleted (including its Doc comment)
	var removeIntervals []interval
	// used to collect function names and comment information that contain panic ("implementation me")
	var panicFuncInfos []FuncInfo

	var isMatch bool

	// traverse declarations in the file, keeping only qualified function declarations
	var decls []ast.Decl
	for _, decl := range f.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok {
			decls = append(decls, decl)
			continue
		}

		// preserve if function name starts with New
		if strings.HasPrefix(funcDecl.Name.Name, "New") {
			decls = append(decls, decl)
			continue
		}
		// if the function body is nil, it remains
		if funcDecl.Body == nil {
			decls = append(decls, decl)
			continue
		}

		// if matches call panic("implement me") and has function comment, the function is retained
		if containsPanicCall(funcDecl, customFlag...) {
			commentText := ""
			if funcDecl.Doc != nil {
				var parts []string
				for _, c := range funcDecl.Doc.List {
					parts = append(parts, strings.TrimSpace(c.Text))
				}
				commentText = strings.Join(parts, "\n")
			}
			panicFuncInfos = append(panicFuncInfos, FuncInfo{
				Name:    funcDecl.Name.Name,
				Comment: commentText,
			})
			if commentText != "" {
				decls = append(decls, decl)
				isMatch = true
			}
			continue
		}

		// delete other cases: record deletion interval
		start := funcDecl.Pos()
		if funcDecl.Doc != nil {
			start = funcDecl.Doc.Pos()
		}
		removeIntervals = append(removeIntervals, interval{start: start, end: funcDecl.End()})
	}
	if !isMatch {
		return nil, nil, fmt.Errorf("no function satisfies both conditions: 1. the function body contains the" +
			" panic(\"implement me\") marker, and 2. the function includes a comment describing its functionality")
	}

	f.Decls = decls

	// filter comment groups to remove comments that fall within the deletion interval
	var newComments []*ast.CommentGroup
COMMENT:
	for _, cg := range f.Comments {
		for _, rem := range removeIntervals {
			if cg.Pos() >= rem.start && cg.End() <= rem.end {
				continue COMMENT
			}
		}
		newComments = append(newComments, cg)
	}
	f.Comments = newComments

	var buf bytes.Buffer
	if err = printer.Fprint(&buf, fset, f); err != nil {
		return nil, nil, err
	}
	return buf.Bytes(), panicFuncInfos, nil
}
