package router

import (
	"github.com/gin-gonic/gin"
	"github.com/ydssx/api-gen/handler"
)

type routerGroup func(*gin.RouterGroup)

// this is a comments
func UserRouter(rg *gin.RouterGroup) {
	rg.GET("/register", handler.RegisterHandler)
	rg.POST("/user/login", handler.LoginHandler)
}

// this is api router
func ApiRouter() {

}

// 用户管理
func init() {
	gin.SetMode(gin.DebugMode)
}
