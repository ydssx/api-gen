package main

import (
	"fmt"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"log"
)

var logicInfo FuncInfo
var apiInfo ApiInfo
var structPkgName string
var apiStruct TypeInfo
var PathTypes = make(map[string][]string, 0)

// func main1() {
// 	g := gin.Default()
// 	g.POST("/login", handler.LoginHandler)
// 	g.Run()
// }

func parseFile(filename string) (*ast.File, *doc.Package, error) {
	fset := token.NewFileSet()

	astFile, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, nil, err
	}

	pkg := &ast.Package{
		Name:  astFile.Name.Name,
		Files: make(map[string]*ast.File),
	}
	pkg.Files[pkg.Name] = astFile

	docPkg := doc.New(pkg, "example", 0)

	return astFile, docPkg, nil
}

func processStructs(docPkg *doc.Package) {
	for _, typ := range docPkg.Types {
		if typ.Name == "LoginResp" {
			comment := typ.Doc
			fmt.Println("Struct Comment:", comment)
		}
	}
}

func main() {
	typeFile := "types/example.go"
	logicFile := "logic/logic.go"

	typeInfo := parseTypes(typeFile, "/login")

	logicFunc := genLogicFunc(logicFile, typeInfo)

	handlerFunc := genHandlerFunc("handler/handler.go", typeInfo, logicFunc)

	err := addRouter("router/router.go", "UserRouter", typeInfo, handlerFunc)
	if err != nil {
		log.Fatal(err)
	}
}
