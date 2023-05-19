package util

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

const (
	ERROR      = -1
	SUCCESS    = 0
	ErrorMsg   = "操作失败"
	SuccessMsg = "操作成功"
)

func result(c *gin.Context, code int, data interface{}, msg string) {
	c.JSON(http.StatusOK, Response{
		Code: code,
		Msg:  msg,
		Data: data,
	})
}

func OK(c *gin.Context) {
	result(c, 0, nil, SuccessMsg)
}

func OKWithData(c *gin.Context, data interface{}) {
	result(c, SUCCESS, data, SuccessMsg)
}

func FailWithMsg(c *gin.Context, msg string) {
	result(c, ERROR, nil, msg)
}
