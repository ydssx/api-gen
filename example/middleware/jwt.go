package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

func JwtMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tokenString := ctx.Request.Header.Get("Authorization")
		if tokenString == "" {
			// Token 不存在或无效，返回错误响应
			ctx.String(401, "%s", "Unauthorized")
			ctx.Abort()
			return
		}

		// 解析和验证 JWT Token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// 根据需要配置密钥获取方法
			return []byte("123456"), nil
		})
		if err != nil || !token.Valid {
			// Token 无效，返回错误响应
			ctx.String(402, "%s", "Unauthorized")
			ctx.Abort()
			return
		}
		ctx.Next()

	}
}
