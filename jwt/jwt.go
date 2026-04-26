package jwt

import (
	"errors"
	"sync"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/jxncyjq/stardust.mini/logs"
	"go.uber.org/zap"
)

var (
	jwtLogger     *zap.Logger
	jwtLoggerOnce sync.Once
)

func logger() *zap.Logger {
	jwtLoggerOnce.Do(func() {
		defer func() {
			if recover() != nil {
				jwtLogger = zap.NewNop()
			}
		}()
		jwtLogger = logs.GetLogger("jwt")
	})
	return jwtLogger
}

func JWTDecrypt(tokenString, secret string) (jwt.MapClaims, bool) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		logger().Warn("jwt decrypt failed", zap.Error(err))
		return nil, false
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, true
	} else {
		logger().Warn("jwt claims invalid")
		return nil, false
	}
}

type MyCustomClaims struct {
	ID    string `json:"id"`
	Token string `json:"token"`
	jwt.RegisteredClaims
}

func JWTEncrypt(id string, myToken string, secret string) string {
	return JWTEncryptWithExpiry(id, myToken, secret, 24*time.Hour)
}

func JWTEncryptWithExpiry(id string, myToken string, secret string, expiry time.Duration) string {
	mySigningKey := []byte(secret)
	claims := MyCustomClaims{
		id,
		myToken,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			Issuer:    "admin",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtStr, err := token.SignedString(mySigningKey)
	if err != nil {
		logger().Error("jwt sign failed", zap.Error(err))
		return ""
	}
	return jwtStr
}
