package types

// @group user
// @handler login
// @router /login [post]
type (
	LoginReq struct {
		Name     string `json:"name" binding:"required"` //用户名
		Password string `json:"password"`
	}

	LoginResp struct {
		User string `json:"user"`
	}
)

// @group user
// @auth false
// @handler register
// @router /register [get]
type (
	RegisterReq struct {
		Name     string `form:"name" binding:"required"` //用户名
		Password string `form:"password"`
	}

	RegisterResp struct {
		User string `json:"user"`
	}
)
