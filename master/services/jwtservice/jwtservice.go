package jwtservice

import (
	"errors"
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
	claims := userClaims{
		UserID: UserID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // 存在时间
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
