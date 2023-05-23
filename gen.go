package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"unicode"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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

const annotationTemplate = `{{ if .Auth }}// @Security ApiKeyAuth{{ end }}
// @Param {{ .HandlerName }} {{ .ParamType }} {{ .Req }} true "请求参数"
// @Success 200	{object} util.Response{data={{ .Resp }}}
// @Router {{ .Group }}{{ .Path }} [{{ .Method|ToLower }}]`

type AnnotationData struct {
	Auth        bool
	HandlerName string
	ParamType   string
	Req         string
	Resp        string
	Group       string
	Path        string
	Method      string
}

func addSwagAnnotation(info TypeInfo) string {
	tmpl, err := template.New("annotation").Funcs(template.FuncMap{
		"ToLower": strings.ToLower,
	}).Parse(annotationTemplate)
	if err != nil {
		log.Fatal(err)
	}

	var sb strings.Builder
	err = tmpl.Execute(&sb, AnnotationData{
		Auth:        info.Auth,
		HandlerName: info.HandlerName,
		ParamType:   getParamType(info.Method),
		Req:         info.Req,
		Resp:        info.Resp,
		Group:       info.Group,
		Path:        info.Path,
		Method:      info.Method,
	})
	if err != nil {
		log.Fatal(err)
	}

	return strings.TrimLeftFunc(sb.String(), unicode.IsSpace)
}

func getParamType(method string) string {
	if method == http.MethodGet {
		return "query"
	}
	return "body"
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

type RouterExprInfo struct {
	RG         string
	Method     string
	PathArg    string
	HandlerArg struct {
		HandlerPkg  string
		HandlerFunc string
	}
}

// TODO:生成router时保留函数内部注释
// func addRouter(routerFile, routerFunc string, apiInfo TypeInfo, handlerFunc FuncInfo) (err error) {
// 	// 解析Go文件
// 	fset := token.NewFileSet()
// 	file, err := parser.ParseFile(fset, routerFile, nil, parser.ParseComments)
// 	if err != nil {
// 		fmt.Println("Failed to parse file:", err)
// 		return
// 	}

// 	// 查找目标函数
// 	var targetFunc *ast.FuncDecl
// 	for _, decl := range file.Decls {
// 		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == routerFunc {
// 			targetFunc = fn
// 			break
// 		}
// 	}
// 	if targetFunc == nil {
// 		return fmt.Errorf("Failed to find target func :%s ,%v", routerFunc, err)
// 	}

// 	x := targetFunc.Type.Params.List[0].Names[0].Name
// 	info := RouterExprInfo{
// 		RG:      x,
// 		Method:  apiInfo.Method,
// 		PathArg: `"` + apiInfo.Path + `"`,
// 		HandlerArg: struct {
// 			HandlerPkg  string
// 			HandlerFunc string
// 		}{handlerFunc.Pkg, handlerFunc.FuncName},
// 	}
// 	if apiInfo.Group != "" {
// 		if g := findRouterGroup(targetFunc.Body.List, apiInfo.Group); g != "" {
// 			info.RG = g
// 		} else {
// 			logrus.Warningf("Failed to find target group :%s", apiInfo.Group)
// 		}
// 	}
// 	// 创建新的CallExpr节点
// 	newCallExpr := &ast.ExprStmt{
// 		X: &ast.CallExpr{
// 			Fun: &ast.SelectorExpr{
// 				X:   ast.NewIdent(info.RG),
// 				Sel: ast.NewIdent(info.Method),
// 			},
// 			Args: []ast.Expr{
// 				ast.NewIdent(info.PathArg),
// 				// ast.NewIdent(handlerFunc.Pkg + "." + handlerFunc.FuncName),
// 				&ast.SelectorExpr{X: ast.NewIdent(info.HandlerArg.HandlerPkg), Sel: ast.NewIdent(info.HandlerArg.HandlerFunc)},
// 			},
// 		},
// 	}

// 	if isRouterAdded(targetFunc.Body.List, info) {
// 		log.Println("router", apiInfo.Path, "already exists. Skipping...")
// 		return
// 	}

// 	// 创建新的文件节点，并按原始顺序将函数添加到该节点中
// 	newFile := &ast.File{
// 		Name:  file.Name,
// 		Decls: make([]ast.Decl, len(file.Decls)),
// 	}
// 	copy(newFile.Decls, file.Decls)

// 	// 在目标函数体的语句列表中找到适当的位置插入新的调用表达式
// 	if apiInfo.Group != "" && x != info.RG {
// 		findAndInsert(targetFunc.Body.List, newCallExpr, apiInfo.Group)
// 	} else {
// 		insertIndex := findInsertIndex(targetFunc.Body.List, targetFunc.Body.Lbrace+1, targetFunc.Body.Rbrace-1)
// 		targetFunc.Body.List = append(targetFunc.Body.List[:insertIndex], append([]ast.Stmt{newCallExpr}, targetFunc.Body.List[insertIndex:]...)...)
// 	}
// 	// 将目标函数替换为修改后的函数
// 	for i, decl := range newFile.Decls {
// 		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == targetFunc.Name.Name {
// 			newFile.Decls[i] = targetFunc
// 			break
// 		}
// 	}

// 	outputFile, err := os.OpenFile(routerFile, os.O_WRONLY|os.O_CREATE, 0644)
// 	if err != nil {
// 		fmt.Println("Failed to create file:", err)
// 		return
// 	}
// 	defer outputFile.Close()

// 	// 重新写入文件，保留原始文件的格式和注释
// 	err = format.Node(outputFile, fset, newFile)
// 	if err != nil {
// 		fmt.Println("Failed to write file:", err)
// 		return
// 	}

// 	fmt.Println("New statement added to", routerFile)
// 	return nil
// }

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

func isRouterAdded(stmts []dst.Stmt, info RouterExprInfo) bool {
	for _, stmt := range stmts {
		switch stmt := stmt.(type) {
		case *dst.ExprStmt:
			if callExpr, ok := stmt.X.(*dst.CallExpr); ok {
				if selectorExpr, ok := callExpr.Fun.(*dst.SelectorExpr); ok {
					funcName := ""
					if indent, ok := selectorExpr.X.(*dst.Ident); ok {
						funcName = indent.Name
					}
					method := selectorExpr.Sel.Name
					if funcName != info.RG || method != info.Method || len(callExpr.Args) != 2 {
						continue
					}
				}
				path := ""
				handlerName := ""
				handlerPkg := ""
				if pathLit, ok := callExpr.Args[0].(*dst.BasicLit); ok {
					path = pathLit.Value
				}
				if selExpr, ok := callExpr.Args[1].(*dst.SelectorExpr); ok {
					handlerPkg = selExpr.X.(*dst.Ident).Name
					handlerName = selExpr.Sel.Name
				}
				if path == info.PathArg && handlerName == info.HandlerArg.HandlerFunc && handlerPkg == info.HandlerArg.HandlerPkg {
					return true
				}
			}

		case *dst.BlockStmt:
			if isRouterAdded(stmt.List, info) {
				return true
			}
		}
	}
	return false
}

func findRouterGroup(stmts []dst.Stmt, group string) string {
	for _, stmt := range stmts {
		switch stmt := stmt.(type) {
		case *dst.ExprStmt:
			if call, ok := stmt.X.(*dst.CallExpr); ok {
				if isRouterGroupCall(call, group) {
					return getRouterGroupName(call)
				}
			}
		case *dst.BlockStmt:
			groupName := findRouterGroup(stmt.List, group)
			if groupName != "" {
				return groupName
			}
		case *dst.AssignStmt:
			if len(stmt.Rhs) > 0 {
				if call, ok := stmt.Rhs[0].(*dst.CallExpr); ok && isRouterGroupCall(call, group) {
					if len(stmt.Lhs) > 0 {
						if lhs, ok := stmt.Lhs[0].(*dst.Ident); ok {
							return lhs.Name
						}
					}
				}
			}
		}
	}
	return ""
}

func isRouterGroupCall(call *dst.CallExpr, group string) bool {
	if len(call.Args) < 1 {
		return false
	}
	arg, ok := call.Args[0].(*dst.BasicLit)
	if !ok || arg.Kind != token.STRING {
		return false
	}
	return arg.Value == fmt.Sprintf(`"%s"`, group) && isSelectorExpr(call.Fun, "Group")
}

func isSelectorExpr(expr dst.Expr, name string) bool {
	if selector, ok := expr.(*dst.SelectorExpr); ok {
		return selector.Sel.Name == name
	}
	return false
}

func getRouterGroupName(call *dst.CallExpr) string {
	selector := call.Fun.(*dst.SelectorExpr)
	ident := selector.X.(*dst.Ident)
	return ident.Name
}

func findAndInsert(stmts []dst.Stmt, newCallExpr dst.Stmt, group string) []dst.Stmt {
	var found bool
	for i, stmt := range stmts {
		switch stmt := stmt.(type) {
		case *dst.ExprStmt:
			if call, ok := stmt.X.(*dst.CallExpr); ok {
				if isRouterGroupCall(call, group) {
					stmts = append(stmts[:i+1], insertBlock(stmts[i+1:], newCallExpr)...)
					return stmts
				}
			}
		case *dst.BlockStmt:
			findAndInsert(stmt.List, newCallExpr, group)
			if found {
				stmt.List = append(stmt.List, newCallExpr)
				return stmts
			}
		case *dst.AssignStmt:
			if len(stmt.Rhs) > 0 {
				if call, ok := stmt.Rhs[0].(*dst.CallExpr); ok && isRouterGroupCall(call, group) {
					stmts = append(stmts[:i+1], insertBlock(stmts[i+1:], newCallExpr)...)
					return stmts
				}
			}

		}
	}
	if !found {
		stmts = append(stmts, newCallExpr)
	}
	return stmts
}

func insertBlock(stmts []dst.Stmt, newCallExpr dst.Stmt) []dst.Stmt {
	for _, stmt := range stmts {
		switch stmt := stmt.(type) {
		case *dst.BlockStmt:
			stmt.List = append(stmt.List, newCallExpr)
			return stmts
		}
	}
	stmts = append(stmts, newCallExpr)
	return stmts
}

func addRouter(routerFile, routerFunc string, apiInfo TypeInfo, handlerFunc FuncInfo) (err error) {
	// 解析Go文件
	fset := token.NewFileSet()
	file, err := decorator.ParseFile(fset, routerFile, nil, parser.ParseComments)
	if err != nil {
		fmt.Println("Failed to parse file:", err)
		return
	}

	// 查找目标函数
	var targetFunc *dst.FuncDecl
	for _, decl := range file.Decls {
		if fn, ok := decl.(*dst.FuncDecl); ok && fn.Name.Name == routerFunc {
			targetFunc = fn
			break
		}
	}
	if targetFunc == nil {
		return fmt.Errorf("Failed to find target func :%s ,%v", routerFunc, err)
	}

	x := targetFunc.Type.Params.List[0].Names[0].Name
	info := RouterExprInfo{
		RG:      x,
		Method:  apiInfo.Method,
		PathArg: `"` + apiInfo.Path + `"`,
		HandlerArg: struct {
			HandlerPkg  string
			HandlerFunc string
		}{handlerFunc.Pkg, handlerFunc.FuncName},
	}
	if apiInfo.Group != "" {
		if g := findRouterGroup(targetFunc.Body.List, apiInfo.Group); g != "" {
			info.RG = g
		} else {
			logrus.Warningf("Failed to find target group :%s", apiInfo.Group)
		}
	}
	// 创建新的CallExpr节点
	newCallExpr := &dst.ExprStmt{
		X: &dst.CallExpr{
			Fun: &dst.SelectorExpr{
				X:   dst.NewIdent(info.RG),
				Sel: dst.NewIdent(info.Method),
			},
			Args: []dst.Expr{
				dst.NewIdent(info.PathArg),
				// ast.NewIdent(handlerFunc.Pkg + "." + handlerFunc.FuncName),
				&dst.SelectorExpr{X: dst.NewIdent(info.HandlerArg.HandlerPkg), Sel: dst.NewIdent(info.HandlerArg.HandlerFunc)},
			},
		},
	}

	if isRouterAdded(targetFunc.Body.List, info) {
		log.Println("router", apiInfo.Path, "already exists. Skipping...")
		return
	}

	// 创建新的文件节点，并按原始顺序将函数添加到该节点中
	// newFile := &dst.File{
	// 	Name:  file.Name,
	// 	Decls: make([]dst.Decl, len(file.Decls)),
	// }
	// copy(newFile.Decls, file.Decls)

	// 在目标函数体的语句列表中找到适当的位置插入新的调用表达式
	if apiInfo.Group != "" && x != info.RG {
		findAndInsert(targetFunc.Body.List, newCallExpr, apiInfo.Group)
	} else {
		// insertIndex := findInsertIndex(targetFunc.Body.List, targetFunc.Body.Lbrace+1, targetFunc.Body.Rbrace-1)
		insertIndex := len(targetFunc.Body.List)
		targetFunc.Body.List = append(targetFunc.Body.List[:insertIndex], append([]dst.Stmt{newCallExpr}, targetFunc.Body.List[insertIndex:]...)...)
	}
	// 将目标函数替换为修改后的函数
	for i, decl := range file.Decls {
		if fn, ok := decl.(*dst.FuncDecl); ok && fn.Name.Name == targetFunc.Name.Name {
			file.Decls[i] = targetFunc
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
	err = decorator.Fprint(outputFile, file)
	if err != nil {
		fmt.Println("Failed to write file:", err)
		return
	}

	fmt.Println("New statement added to", routerFile)
	return nil
}
