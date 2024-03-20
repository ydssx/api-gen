package gen

import (
	"fmt"
	"strings"

	"github.com/dave/dst"
)

type RouteNode struct {
	Caller   string
	Path     string
	Children []*RouteNode
}

func BuildRouteTree(routerFile, routerFunc string) (*RouteNode, error) {
	_, targetFunc, err := searchFunc(routerFile, routerFunc)
	if err != nil {
		return nil, err
	}

	root := &RouteNode{Caller: findRootRG(targetFunc), Path: "", Children: []*RouteNode{}}

	parseFunction(targetFunc.Body.List, root)

	return root, nil
}

func parseFunction(body []dst.Stmt, parent *RouteNode) {
	for i, stmt := range body {
		switch s := stmt.(type) {
		case *dst.AssignStmt:
			if len(s.Rhs) > 0 {
				if call, ok := s.Rhs[0].(*dst.CallExpr); ok && isSelectorExpr(call.Fun, "Group") && isTargetRG(call.Fun, parent.Caller) {
					if len(s.Lhs) > 0 {
						if lhs, ok := s.Lhs[0].(*dst.Ident); ok {
							funcNode := &RouteNode{Caller: lhs.Name, Path: getGroupName(call), Children: []*RouteNode{}}
							parseFunction(body[i+1:], funcNode)
							parent.Children = append(parent.Children, funcNode)
						}
					}
				}
			}
		case *dst.BlockStmt:
			parseFunction(s.List, parent)
		}
	}
}

func isTargetRG(expr dst.Expr, rgName string) bool {
	return expr.(*dst.SelectorExpr).X.(*dst.Ident).Name == rgName
}

func getGroupPath(routerFile, routerFunc, group string) string {

	tree, err := BuildRouteTree(routerFile, routerFunc)
	if err != nil {
		fmt.Println("Failed to build route tree:", err)
		return ""
	}

	// printRouteTree(tree, 0)
	path := DFSPath(tree, group)
	if len(path) < 1 {
		return ""
	}
	return strings.Join(path[0], "/")
}

func printRouteTree(node *RouteNode, depth int) {
	indent := strings.Repeat("  ", depth)
	fmt.Printf("%s%s\n", indent, node.Caller)

	for _, child := range node.Children {
		printRouteTree(child, depth+1)
	}
}

// DFS搜索路径
func DFSPath(root *RouteNode, target string) [][]string {
	paths := [][]string{}
	if root == nil {
		return paths
	}

	// 辅助函数，递归搜索路径
	var dfs func(node *RouteNode, path []string)
	dfs = func(node *RouteNode, path []string) {
		if node == nil {
			return
		}

		// 将当前节点值添加到路径中
		path = append(path, node.Path)

		// 检查当前节点是否为目标节点
		if node.Path == target {
			// 将找到的路径添加到结果中
			paths = append(paths, append([]string{}, path...))
		}

		// 递归搜索子节点
		for _, child := range node.Children {
			dfs(child, path)
		}

		// 回溯，移除当前节点值，继续搜索其他路径
		path = path[:len(path)-1]
	}

	// 从根节点开始搜索路径
	dfs(root, []string{})

	return paths
}

// findRootRG finds the root router group variable name declared in the given function.
// It looks for common patterns of initializing a router group and assigning it to a variable.
// Returns the variable name if found, empty string if not.
func findRootRG(funcDecl *dst.FuncDecl) string {
	paramList := funcDecl.Type.Params.List
	if len(paramList) > 0 {
		return paramList[0].Names[0].Name
	}
	for _, stmt := range funcDecl.Body.List {
		switch s := stmt.(type) {
		case *dst.AssignStmt:
			if len(s.Rhs) > 0 {
				if call, ok := s.Rhs[0].(*dst.CallExpr); ok && (isSelectorExpr(call.Fun, "New") || isSelectorExpr(call.Fun, "Default")) && isTargetRG(call.Fun, "gin") {
					if len(s.Lhs) > 0 {
						if lhs, ok := s.Lhs[0].(*dst.Ident); ok {
							return lhs.Name
						}
					}
				}
			}
		}
	}
	return ""
}
