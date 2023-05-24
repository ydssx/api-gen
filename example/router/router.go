package router

import (
	"github.com/gin-gonic/gin"
	"github.com/ydssx/api-gen/example/handler"
)

type routerGroup func(*gin.RouterGroup)

// this is a comments
func UserRouter(rg *gin.RouterGroup) {
	// asadsafdf
	rg.POST("/login", handler.LoginHandler)
	user := rg.Group("user")
	{
		// this is dfdf
		user.GET("/register", handler.RegisterHandler)
	}
	api := rg.Group("api")
	{
		// commnet api
		v2 := api.Group("apiv2")
		{
			// comment apiv2
			v2.GET("api2222")
		}
		api.POST("/login", handler.LoginHandler)
	}
	// comentde dsaas
	rg.DELETE("delete")
	rg.GET("/register", handler.RegisterHandler)
}

func ApiRouter() {
	// this is api router

}

// 用户管理
func init() {
	gin.SetMode(gin.DebugMode)
}
