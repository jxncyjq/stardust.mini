package httpServer

import (
	"context"
	"fmt"
	"os"

	"github.com/jxncyjq/stardust.mini/core/jwt"
	"github.com/jxncyjq/stardust.mini/core/redis"
	"github.com/jxncyjq/stardust.mini/core/utils"
	"github.com/labstack/echo/v4"
)

// Access 用于判断是否具有访问权限
func Access() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			//在这里处理拦截请求的逻辑
			jwtStr := c.Request().Header.Get("jwt")

			if jwtStr != "" {
				appName := os.Getenv("APP_NAME")
				appVersion := os.Getenv("APP_VERSION")
				secret := fmt.Sprintf("%s-%s", appName, appVersion)
				jwtObj, ok := jwt.JWTDecrypt(jwtStr, secret)
				if !ok || jwtObj == nil || jwtObj["token"] == nil || jwtObj["id"] == nil {
					return c.JSON(401, map[string]interface{}{
						"errCode": 2,
						"errMsg":  "jwt解析错误",
					})
				}

				id, ok := jwtObj["id"].(string)
				if !ok {
					return c.JSON(401, map[string]interface{}{
						"errCode": 2,
						"errMsg":  "用户信息获取失败",
					})
				}
				redisCmd := redis.GetRedisDb()
				key := fmt.Sprintf("%s:user:token:%s", appName, id)
				oldToken, err1 := redisCmd.Get(context.Background(), key).Result()
				if err1 != nil {
					return c.JSON(401, map[string]interface{}{
						"errCode": 2,
						"errMsg":  "获取token失败",
					})
				}
				if jwtObj["token"] != oldToken {
					return c.JSON(401, map[string]interface{}{
						"errCode": 2,
						"errMsg":  "账户已经在其他终端登录",
					})
				}
				c.QueryParams().Add("id", id)
			} else {
				ip := echo.ExtractIPFromXFFHeader()(c.Request())
				// 判断是否内网IP
				ok := utils.IsInnerIp(ip)
				if !ok {
					return c.JSON(401, map[string]interface{}{
						"errCode": 2,
						"errMsg":  "服务器繁忙,请稍后再试!",
					})
				}
			}
			return next(c)
		}
	}
}

func SetHeaders(headers map[string]string, next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		for key, value := range headers {
			c.Response().Header().Set(key, value)
		}
		return next(c)
	}
}
