package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/ydssx/api-gen/example/logic"
	"github.com/ydssx/api-gen/example/types"
	"github.com/ydssx/api-gen/example/util"
)


// @Security ApiKeyAuth
// @Param Login body types.LoginReq true "请求参数"
// @Success 200	{object} util.Response{data=types.LoginResp}
// @Router user//login [post]
func LoginHandler(c *gin.Context) {
	var req types.LoginReq
	if err := c.ShouldBind(&req); err != nil {
		util.FailWithMsg(c, util.WrapValidateErrMsg(err))
		return
	}

	resp, err := logic.LoginLogic(req)
	if err != nil {
		util.FailWithMsg(c, err.Error())
		return
	}

	util.OKWithData(c, resp)
}

// @Param Register query types.RegisterReq true "请求参数"
// @Success 200	{object} util.Response{data=types.RegisterResp}
// @Router user//register [get]
func RegisterHandler(c *gin.Context) {
	var req types.RegisterReq
	if err := c.ShouldBind(&req); err != nil {
		util.FailWithMsg(c, util.WrapValidateErrMsg(err))
		return
	}

	resp, err := logic.RegisterLogic(req)
	if err != nil {
		util.FailWithMsg(c, err.Error())
		return
	}

	util.OKWithData(c, resp)
}
