package middleware

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/text/language"
)

const langContextKey = "lang"

// I18nMiddleware 从 Accept-Language/X-Lang/query 参数提取语言
// 优先级：X-Lang > Accept-Language > query lang > 默认 zh-CN
// 支持的语言：zh-CN, en-US
func I18nMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := c.GetHeader("X-Lang")
		if lang == "" {
			accept := c.GetHeader("Accept-Language")
			if accept != "" {
				if idx := strings.IndexByte(accept, ','); idx != -1 {
					lang = accept[:idx]
				} else {
					lang = accept
				}
			}
		}

		if lang == "" {
			lang = c.Query("lang")
		}
		if lang == "" {
			lang = "zh-CN"
		}

		lang = strings.TrimSpace(lang)
		if tag, err := language.Parse(lang); err == nil {
			lang = tag.String()
		}

		supported := false
		for _, item := range []string{"zh-CN", "en-US"} {
			if lang == item {
				supported = true
				break
			}
		}
		if !supported {
			lang = "zh-CN"
		}

		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), langContextKey, lang))
		c.Header("X-Lang", lang)
		c.Next()
	}
}

// LangFromContext 从 context 读取语言标签
func LangFromContext(ctx context.Context) string {
	if lang, ok := ctx.Value(langContextKey).(string); ok {
		return lang
	}
	return "zh-CN"
}
