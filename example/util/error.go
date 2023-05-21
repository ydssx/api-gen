package util

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

func WrapValidateErrMsg(err error) (msg string) {
	switch v := err.(type) {
	case *json.UnmarshalTypeError:
		msg = fmt.Sprintf("请求参数`%s`类型错误，应为%s类型", v.Field, v.Type.Name())
	case validator.ValidationErrors:
		for _, e := range v {
			msg += fmt.Sprintf("缺少必要参数：`%s`", strings.ToLower(e.Field()))
		}
	default:
		msg = err.Error()
	}
	return
}
