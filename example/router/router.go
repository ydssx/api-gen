package router

import (
	"github.com/gin-gonic/gin"
	"github.com/ydssx/api-gen/example/handler"
)

type routerGroup func(*gin.RouterGroup)

// this is a comments
func UserRouter(rg *gin.RouterGroup) {
	rg.POST("/login", handler.LoginHandler)
	rg.POST("/login", handler.LoginHandler)
	rg.GET("/register", handler.RegisterHandler)
	rg.POST("/login", handler.LoginHandler)
	rg.GET("/register", handler.RegisterHandler)
	rg.POST("/login", handler.LoginHandler)
	rg.GET("/register", handler.RegisterHandler)
	rg.POST("/login", handler.LoginHandler)
	rg.GET("/register", handler.RegisterHandler)
}

// this is api router
func ApiRouter() {

}

// 用户管理
func init() {
	gin.SetMode(gin.DebugMode)
}
