package gen

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

	"github.com/dave/dst"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type ApiInfo struct {
	Path        string
	Method      string
	HandlerName string
	Auth        bool
	Group       string
	Summary     string
}

func ParseComments(comment string) (info ApiInfo) {
	info.Auth = true
	list := strings.Fields(comment)
	for i := 0; i < len(list); i++ {
		switch strings.ToLower(list[i]) {
		case "@handler":
			info.HandlerName = cases.Title(language.English).String(list[i+1])
		case "@router":
			info.Path = list[i+1]
			info.Method = strings.Trim(strings.ToUpper(list[i+2]), "[]")
		case "@auth":
			if list[i+1] == "false" {
				info.Auth = false
			}
		case "@group":
			info.Group = list[i+1]
		case "@summary":
			info.Summary = list[i+1]
		}
	}
	return
}

type TypeInfo struct {
	Req     string
	Resp    string
	PkgName string
	ApiInfo
}

func parseStructs(pkgName string, structNames []string) (r TypeInfo) {
	for _, v := range structNames {
		if suffix := strings.TrimSuffix(v, "Req"); suffix != v {
			r.Req = pkgName + "." + v
		} else if suffix := strings.TrimSuffix(v, "Resp"); suffix != v {
			r.Resp = pkgName + "." + v
		}
	}
	r.PkgName = pkgName
	return
}

type FuncInfo struct {
	Pkg      string
	FuncName string
	Results  []string
}

func parseFunc(pkg string, dec *dst.FuncDecl) (l FuncInfo) {
	l.Pkg = pkg
	l.FuncName = dec.Name.Name
	results := dec.Type.Results
	result := []string{}
	if results != nil {
		for _, v := range results.List {
			result = append(result, v.Names[0].Name)
		}
	}
	l.Results = result
	return
}

func parseCodeTmp(code string) (*ast.FuncDecl, error) {
	code = "package main\n\n" + code
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", code, parser.ParseComments)
	if err != nil {
		panic(err)
	}
	if len(file.Decls) == 0 {
		panic("no statements")
	}
	fn, ok := file.Decls[0].(*ast.FuncDecl)
	if !ok {
		panic("not a function")
	}
	// if len(fn.Body.List) == 0 {
	// 	panic("no statements")
	// }
	return fn, nil
}

func parseTypes(filename, path string) (info TypeInfo) {
	var apiInfo ApiInfo
	var types []string

	fset := token.NewFileSet()

	astFile, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	for _, decl := range astFile.Decls {
		if v, ok := decl.(*ast.GenDecl); ok {
			for _, spec := range v.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok {
					if _, ok := typeSpec.Type.(*ast.StructType); ok {
						if v.Doc != nil {
							// structName := typeSpec.Name.Name
							commentText := v.Doc.Text()
							info := ParseComments(commentText)
							if info.Path == path {
								apiInfo = info
								types = append(types, typeSpec.Name.Name)
							}
						}
					}
				}
			}

		}
	}
	info = parseStructs(astFile.Name.Name, types)
	info.ApiInfo = apiInfo
	return
}
