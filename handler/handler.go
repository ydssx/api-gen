package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/ydssx/api-gen/logic"
	"github.com/ydssx/api-gen/types"
	"github.com/ydssx/api-gen/util"
)

// @Success 200	{object} model.Admin
// @Router /accounts/{id} [get]
func loginHandler(c *gin.Context) {
	var req types.LoginReq
	if err := c.ShouldBind(&req); err != nil {
		util.FailWithMsg(c, util.WrapErrMsg(err))
		return
	}

	resp, err := logic.LoginLogic(req)
	if err != nil {
		util.FailWithMsg(c, err.Error())
		return
	}

	util.OKWithData(c, resp)
}
