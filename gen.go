package main

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"os"
	"strings"
)

var logicTmp = `
func %sLogic(req %s) (resp %s, err error) {
	// TODO: add your logic here and delete this line

	return
}
`

func genLogicFunc(filename string, api TypeInfo) FuncInfo {
	content := fmt.Sprintf(logicTmp, api.HandlerName, api.Req, api.Resp)
	return writeDecl(filename, content)
}

var handlerTmp = `
// @Success 200	{object} util.Response{data=%s}
// @Router %s [%s]
func %sHandler(c *gin.Context) {
	var req %s
	if err := c.ShouldBind(&req); err != nil {
		util.FailWithMsg(c, util.WrapValidateErrMsg(err))
		return
	}

	%s := %s.%s(req)
	if err != nil {
		util.FailWithMsg(c, err.Error())
		return
	}

	util.OKWithData(c, %s)
}
`

func genHandlerFunc(filename string, def TypeInfo, logic FuncInfo) FuncInfo {

	// 要追加的内容
	content := fmt.Sprintf(handlerTmp, def.Resp, def.Path, strings.ToLower(def.Method), def.HandlerName, def.Req, strings.Join(logic.Results, ", "), logic.Pkg, logic.FuncName, logic.Results[0])

	return writeDecl(filename, content)
}

func writeDecl(filename, decl string) (info FuncInfo) {
	// 解析文件
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	// 将新函数的源代码解析为语法树
	funcAST, err := parser.ParseFile(fset, "", "package main\n\n"+decl, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	newFunc := funcAST.Decls[0].(*ast.FuncDecl)
	info = parseFunc(file.Name.Name, newFunc)

	// 将新函数添加到原始文件的语法树中
	file.Decls = append(file.Decls, newFunc)

	output, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer output.Close()

	// 格式化生成的代码
	err = format.Node(output, fset, file)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("New function added to", filename)
	return
}

func addRouter(routerFile, routerFunc string, apiInfo TypeInfo, handlerFunc FuncInfo) (err error) {

	// 解析Go文件
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, routerFile, nil, parser.ParseComments)
	if err != nil {
		fmt.Println("Failed to parse file:", err)
		return
	}

	// 查找目标函数
	var targetFunc *ast.FuncDecl
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == routerFunc {
			targetFunc = fn
			break
		}
	}
	if targetFunc == nil {
		return fmt.Errorf("Failed to find target func :%s ,%v", routerFunc, err)
	}

	// 创建新的CallExpr节点
	newCallExpr := &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   ast.NewIdent(targetFunc.Type.Params.List[0].Names[0].Name),
				Sel: ast.NewIdent(apiInfo.Method),
			},
			Args: []ast.Expr{
				ast.NewIdent(`"` + apiInfo.Path + `"`),
				ast.NewIdent(handlerFunc.Pkg + "." + handlerFunc.FuncName),
			},
		},
	}
	funcBody := targetFunc.Body
	// 在目标函数体的语句列表中找到适当的位置插入新的调用表达式
	insertIndex := findInsertIndex(funcBody.List, targetFunc.Pos(), targetFunc.End())
	funcBody.List = append(funcBody.List[:insertIndex], append([]ast.Stmt{newCallExpr}, funcBody.List[insertIndex:]...)...) // 创建新的路由语句

	outputFile, err := os.OpenFile(routerFile, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("Failed to create file:", err)
		return
	}
	defer outputFile.Close()

	err = formatAndWriteFile(outputFile, fset, file)
	if err != nil {
		fmt.Println("Failed to write file:", err)
		return
	}

	fmt.Println("New statement added to", routerFile)
	return nil
}

func formatAndWriteFile(outputFile *os.File, fset *token.FileSet, file *ast.File) error {
	// 创建缓冲区来保存格式化后的代码
	var buf strings.Builder

	// 使用格式化配置进行代码格式化
	err := printer.Fprint(&buf, fset, file)
	if err != nil {
		return err
	}

	// 格式化后的代码写入文件
	_, err = outputFile.WriteString(buf.String())
	if err != nil {
		return err
	}

	return nil
}

func findInsertIndex(stmts []ast.Stmt, startPos, endPos token.Pos) int {
	for i, stmt := range stmts {
		if stmt.Pos() >= endPos {
			return i
		}
	}
	return len(stmts)
}
