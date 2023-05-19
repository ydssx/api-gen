package main

import (
	"fmt"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"log"
)

var logicInfo LogicInfo
var apiInfo ApiInfo
var structPkgName string
var apiStruct ApiStruct
var handlerMap = make(map[string][]string, 0)

func main1() {

	logicFile := "logic/logic.go"
	files := []string{"types/example.go", logicFile}
	for _, file := range files {
		parse(file)
	}
	genLogic(logicFile, apiStruct)
	parse(logicFile)
	genHandler("handler/handler.go", apiStruct, apiInfo, logicInfo)
}

func parse(filename string) {
	fset := token.NewFileSet()

	astFile, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	// 创建*ast.Package对象
	pkg := &ast.Package{
		Name:  astFile.Name.Name,
		Files: make(map[string]*ast.File),
	}
	pkg.Files[pkg.Name] = astFile

	// 使用go/doc包的New函数构建文档对象
	docPkg := doc.New(pkg, "example", 0)
	// 遍历结构体类型声明
	for _, typ := range docPkg.Types {
		if typ.Name == "LoginResp" {
			// 获取结构体类型的注释
			comment := typ.Doc

			// 打印结构体类型的注释
			fmt.Println("Struct Comment:", comment)
		}
	}

	for _, decl := range astFile.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			// 获取函数的注释
			if d.Doc != nil {
				funcName := d.Name.Name
				commentText := d.Doc.Text()

				fmt.Printf("Function: %s\nComment: %s\n\n", funcName, commentText)
			}
			logicInfo = parseLogic(pkg.Name, d)

		case *ast.GenDecl:

			// 遍历结构体声明
			for _, spec := range d.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok {
					if _, ok := typeSpec.Type.(*ast.StructType); ok {
						// 获取结构体的注释
						if d.Doc != nil {
							structName := typeSpec.Name.Name
							commentText := d.Doc.Text()

							fmt.Printf("Struct: %s\nComment: %s\n\n", structName, commentText)
							apiInfo = ParseComments(commentText)
							handlerMap[apiInfo.Path] = append(handlerMap[apiInfo.Path], typeSpec.Name.Name)
						}
						structPkgName = pkg.Name
						// for _, field := range structType.Fields.List {
						// 	fmt.Println(field.Names, field.Comment.Text())
						// }

					}
				}
			}
		}
	}
	apiStruct = parseStructs(structPkgName, handlerMap["/login"])
}

// func main1() {
// 	g := gin.Default()
// 	g.POST("/login", handler.LoginHandler)
// 	g.Run()
// }


func parseFile(filename string) (*ast.Package, *doc.Package, error) {
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

	return pkg, docPkg, nil
}

func processStructs(pkg *ast.Package, docPkg *doc.Package) {
    for _, typ := range docPkg.Types {
        if typ.Name == "LoginResp" {
            comment := typ.Doc
            fmt.Println("Struct Comment:", comment)
        }
    }
}

func processFuncs(pkg *ast.Package, astFile *ast.File) {
    for _, decl := range astFile.Decls {
        switch d := decl.(type) {
        case *ast.FuncDecl:
            if d.Doc != nil {
                funcName := d.Name.Name
                commentText := d.Doc.Text()
                fmt.Printf("Function: %s\nComment: %s\n\n", funcName, commentText)
            }
            logicInfo = parseLogic(pkg.Name, d)

        case *ast.GenDecl:
            for _, spec := range d.Specs {
                if typeSpec, ok := spec.(*ast.TypeSpec); ok {
                    if _, ok := typeSpec.Type.(*ast.StructType); ok {
                        if d.Doc != nil {
                            structName := typeSpec.Name.Name
                            commentText := d.Doc.Text()
                            fmt.Printf("Struct: %s\nComment: %s\n\n", structName, commentText)
                            apiInfo = ParseComments(commentText)
                            handlerMap[apiInfo.Path] = append(handlerMap[apiInfo.Path], typeSpec.Name.Name)
                        }
                        structPkgName = pkg.Name
                    }
                }
            }
        }
    }
	apiStruct = parseStructs(structPkgName, handlerMap["/login"])
}

func parseAndProcess(filename string) {
    pkg, docPkg, err := parseFile(filename)
    if err != nil {
        log.Fatal(err)
    }

    processStructs(pkg, docPkg)
    processFuncs(pkg, pkg.Files[pkg.Name])
}


func main() {
	logicFile := "logic/logic.go"
	files := []string{"types/example.go", logicFile}
	for _, file := range files {
		pkg, docPkg, err := parseFile(file)
		if err != nil {
			log.Fatal(err)
		}

		processStructs(pkg, docPkg)
		processFuncs(pkg, pkg.Files[pkg.Name])
	}

	genLogic(logicFile, apiStruct)

	pkg, docPkg, err := parseFile(logicFile)
	if err != nil {
		log.Fatal(err)
	}
	processStructs(pkg, docPkg)
	processFuncs(pkg, pkg.Files[pkg.Name])

	genHandler("handler/handler.go", apiStruct, apiInfo, logicInfo)

}