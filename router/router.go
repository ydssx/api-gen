package router

import "github.com/gin-gonic/gin"

type routerGroup func(*gin.RouterGroup)

func UserRouter(rg *gin.RouterGroup) {
	rg.GET("/")
}
