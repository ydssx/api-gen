package types

// @handler login
// @router /login [post]
type (
	LoginReq struct {
		Name     string `json:"username" binding:"required"` //用户名
		Password string `json:"password"`
	}
	
	LoginResp struct {
		User string `json:"user"`
	}
)
