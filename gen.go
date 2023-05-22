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
	"net/http"
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
%s
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

func addSwagAnnotation(info TypeInfo) (s string) {
	s = `// @Param %s %s %s true "请求参数"
// @Success 200	{object} util.Response{data=%s}
// @Router %s [%s]`
	if info.Auth {
		s = `// @Security ApiKeyAuth
// @Param %s %s %s true "请求参数"
// @Success 200	{object} util.Response{data=%s}
// @Router %s [%s]`
	}

	paramType := "body"
	if info.Method == http.MethodGet {
		paramType = "query"
	}

	s = fmt.Sprintf(s, info.HandlerName, paramType, info.Req, info.Resp, info.Group+"/"+info.Path, strings.ToLower(info.Method))
	return
}

func genHandlerFunc(filename string, def TypeInfo, logic FuncInfo) FuncInfo {

	// 要追加的内容
	content := fmt.Sprintf(handlerTmp, addSwagAnnotation(def), def.HandlerName, def.Req, strings.Join(logic.Results, ", "), logic.Pkg, logic.FuncName, logic.Results[0])

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

func findInsertPos(stmts []ast.Stmt, newCallExpr ast.Stmt, group string, startPos, endPos token.Pos) []ast.Stmt {
	var found bool
	for i, stmt := range stmts {
		if v, ok := stmt.(*ast.ExprStmt); ok {
			if call, ok := v.X.(*ast.CallExpr); ok {
				if indent, ok := call.Args[0].(*ast.BasicLit); ok && indent.Value == fmt.Sprintf(`"%s"`, group) {
					found = true
					startPos = call.Rparen
					findInsertPos(stmts[i+1:], newCallExpr, group, startPos, endPos)
				}
			}
		} else if v, ok := stmt.(*ast.BlockStmt); ok && found {
			v.List = append(v.List, newCallExpr)
			return stmts
		}
	}
	stmts = append(stmts, newCallExpr)
	return stmts
}

func fileAppend(filename, content string) error {
	// 打开文件，如果文件不存在则创建，以追加模式打开
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return errors.Wrap(err, "Failed to open file")
	}
	defer file.Close()

	if hasTrailingEmptyLine, _ := hasEmptyLineAtEnd(filename); !hasTrailingEmptyLine {
		content = string('\n') + content
	}

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
				// ast.NewIdent(handlerFunc.Pkg + "." + handlerFunc.FuncName),
				&ast.SelectorExpr{X: ast.NewIdent(handlerFunc.Pkg), Sel: ast.NewIdent(handlerFunc.FuncName)},
			},
		},
	}

	if isRouterAdded(targetFunc.Body.List, targetFunc.Type.Params.List[0].Names[0].Name, apiInfo.Method, `"`+apiInfo.Path+`"`, handlerFunc.Pkg+"."+handlerFunc.FuncName) {
		log.Println("router", apiInfo.Path, "already exists. Skipping...")
		return
	}

	// 创建新的文件节点，并按原始顺序将函数添加到该节点中
	newFile := &ast.File{
		Name:  file.Name,
		Decls: make([]ast.Decl, len(file.Decls)),
	}
	copy(newFile.Decls, file.Decls)

	// 在目标函数体的语句列表中找到适当的位置插入新的调用表达式
	if apiInfo.Group != "" {
		targetFunc.Body.List = findInsertPos(targetFunc.Body.List, newCallExpr, apiInfo.Group, targetFunc.Body.Lbrace+1, targetFunc.Body.Rbrace-1)
	} else {
		insertIndex := findInsertIndex(targetFunc.Body.List, targetFunc.Body.Lbrace+1, targetFunc.Body.Rbrace-1)
		targetFunc.Body.List = append(targetFunc.Body.List[:insertIndex], append([]ast.Stmt{newCallExpr}, targetFunc.Body.List[insertIndex:]...)...)
	}
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

// 判断文件末尾是否有空行
func hasEmptyLineAtEnd(filename string) (bool, error) {
	// 读取文件内容
	content, err := os.ReadFile(filename)
	if err != nil {
		return false, err
	}

	// 检查最后一个字符是否是换行符
	if len(content) > 0 && content[len(content)-1] == '\n' {
		return true, nil
	}

	return false, nil
}

func isRouterAdded(stmts []ast.Stmt, fName, method, path, handlerName string) bool {
	for _, stmt := range stmts {
		if stmt, ok := stmt.(*ast.ExprStmt); ok {
			if call, ok := stmt.X.(*ast.CallExpr); ok {
				funcName := call.Fun.(*ast.SelectorExpr).X.(*ast.Ident).Name
				meth := call.Fun.(*ast.SelectorExpr).Sel.Name
				if len(call.Args) != 2 {
					continue
				}
				arg1 := call.Args[0].(*ast.BasicLit).Value
				pkg := call.Args[1].(*ast.SelectorExpr).X.(*ast.Ident).Name
				handler := call.Args[1].(*ast.SelectorExpr).Sel.Name
				arg2 := pkg + "." + handler
				if fName == funcName && meth == method && path == arg1 && handlerName == arg2 {
					return true
				}
			}
		}
		if stmt, ok := stmt.(*ast.BlockStmt); ok {
			if isRouterAdded(stmt.List, fName, method, path, handlerName) {
				return true
			}
		}
	}
	return false
}
