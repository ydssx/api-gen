package main

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"log"
	"os"
	"strings"

	"github.com/pkg/errors"
)

var logicTmp = `
// this is logic
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

	if isFunctionExists(file, newFunc.Name.Name) {
		// 如果函数名重复，可以选择跳过添加或者进行替换
		log.Println("Function", newFunc.Name.Name, "already exists. Skipping...")
		return
	}

	fileAppend(filename, decl)
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

func fileAppend(filename, content string) error {
	// 打开文件，如果文件不存在则创建，以追加模式打开
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return errors.Wrap(err, "Failed to open file")
	}
	defer file.Close()

	// 将内容写入文件
	if _, err := io.WriteString(file, content); err != nil {
		return errors.Wrap(err, "Failed to write to file")
	}

	fmt.Println("New function added to", filename)
	return nil
}

// 检查函数名是否存在
func isFunctionExists(file *ast.File, functionName string) bool {
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == functionName {
			return true
		}
	}
	return false
}

func addRouter1(routerFile, routerFunc string, apiInfo TypeInfo, handlerFunc FuncInfo) (err error) {
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

	// 创建新的文件节点，并按原始顺序将函数添加到该节点中
	newFile := &ast.File{
		Name:  file.Name,
		Decls: make([]ast.Decl, len(file.Decls)),
	}
	copy(newFile.Decls, file.Decls)

	// 在目标函数体的语句列表中找到适当的位置插入新的调用表达式
	insertIndex := findInsertIndex(targetFunc.Body.List, targetFunc.Body.Lbrace+1, targetFunc.Body.Rbrace-1)
	targetFunc.Body.List = append(targetFunc.Body.List[:insertIndex], append([]ast.Stmt{newCallExpr}, targetFunc.Body.List[insertIndex:]...)...)

	// 将目标函数替换为修改后的函数
	for i, decl := range newFile.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == targetFunc.Name.Name {
			newFile.Decls[i] = targetFunc
			break
		}
	}

	outputFile, err := os.OpenFile(routerFile, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("Failed to create file:", err)
		return
	}
	defer outputFile.Close()

	// 重新写入文件，保留原始文件的格式和注释
	err = format.Node(outputFile, fset, newFile)
	if err != nil {
		fmt.Println("Failed to write file:", err)
		return
	}

	fmt.Println("New statement added to", routerFile)
	return nil
}
