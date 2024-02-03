package jwtservice

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// jwtkey
var jwtKey = []byte("my-secret-key")

type userClaims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}

func GenerateToken(UserID uint) (string, error) {
	var expireTime = 24 * time.Hour
	// 测试环境下token有效期为15s
	if os.Getenv("ENV") == "test" {
		timestr := os.Getenv("TOKEN_EXPIRE_TIME")
		expireTime, _ = time.ParseDuration(timestr)
	}
	claims := userClaims{
		UserID: UserID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expireTime)), // 存在时间
		},
	}
	// 生成token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// 生成签名字符串
	return token.SignedString(jwtKey)
}

func ParseToken(tokenString string) (*userClaims, error) {
	var mc = new(userClaims)
	// 解析token
	token, err := jwt.ParseWithClaims(tokenString, mc, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtKey), nil
	})
	if err != nil {

		return nil, err
	}
	// 对token对象中claim进行断言
	if token.Valid {
		return mc, nil
	}
	return nil, errors.New("invalid token")
}
