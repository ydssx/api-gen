package main

import (
	"fmt"
	"go/ast"
	"io"
	"os"
	"strings"

	"github.com/pkg/errors"
)

type ApiInfo struct {
	Path        string
	Method      string
	HandlerName string
}

func ParseComments(comment string) (info ApiInfo) {
	list := strings.Fields(comment)
	for i := 0; i < len(list); i++ {
		switch list[i] {
		case "@handler":
			info.HandlerName = list[i+1]
		case "@router":
			info.Path = list[i+1]
			info.Method = strings.ToUpper(list[i+2])
		}
	}
	return
}

type ApiStruct struct {
	Req     string
	Resp    string
	PkgName string
}

func parseStructs(pkgName string, structNames []string) (r ApiStruct) {
	for _, v := range structNames {
		if suffix := strings.TrimSuffix(v, "Req"); suffix != v {
			r.Req = pkgName + "." + v
		} else if suffix := strings.TrimSuffix(v, "Resp"); suffix != v {
			r.Resp = pkgName + "." + v
		}
	}
	return
}

type LogicInfo struct {
	Pkg      string
	FuncName string
	Receiver string
	Results  []string
}

func parseLogic(pkg string, dec *ast.FuncDecl) (l LogicInfo) {
	l.Pkg = pkg
	l.FuncName = dec.Name.Name
	resultList := dec.Type.Results.List
	result := []string{}
	for _, v := range resultList {
		result = append(result, v.Names[0].Name)
	}
	l.Results = result
	return
}

var logicTmp = `
func LoginLogic(req %s) (resp %s, err error) {
	// TODO: add your logic here and delete this line

	return
}
`

func genLogic(filename string, api ApiStruct) {
	content := fmt.Sprintf(logicTmp, api.Req, api.Resp)
	fileAppend(filename, content)
}

var handlerTmp = `
// @Success 200	{object} model.Admin
// @Router /accounts/{id} [get]
func %sHandler(c *gin.Context) {
	var req %s
	if err := c.ShouldBind(&req); err != nil {
		util.FailWithMsg(c, util.WrapErrMsg(err))
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

func genHandler(filename string, def ApiStruct, api ApiInfo, logic LogicInfo) {

	// 要追加的内容
	content := fmt.Sprintf(handlerTmp, api.HandlerName, def.Req, strings.Join(logic.Results, ", "), logic.Pkg, logic.FuncName, logic.Results[0])

	fileAppend(filename, content)

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

	fmt.Println("Content appended to file.")
	return nil
}
