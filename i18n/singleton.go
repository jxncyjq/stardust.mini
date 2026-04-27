package i18n

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"

	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

var (
	globalBundle *goi18n.Bundle
	msgKeyMap    map[int]string // errCode -> msgKey 映射表
	once         sync.Once
)

// Init 初始化全局 i18n Bundle（仅调用一次）。
// bundleDir: 包含 messages.*.json 和 error_msg_keys.json 的目录。
// defaultLang: 默认语言（通常为 "zh-CN"）。
func Init(bundleDir string, defaultLang string) error {
	var err error
	once.Do(func() {
		defaultTag := language.MustParse(defaultLang)
		globalBundle = goi18n.NewBundle(defaultTag)
		globalBundle.RegisterUnmarshalFunc("json", json.Unmarshal)

		files, readErr := os.ReadDir(bundleDir)
		if readErr == nil {
			for _, file := range files {
				if file.IsDir() {
					continue
				}
				name := file.Name()
				if !strings.HasPrefix(name, "messages.") || filepath.Ext(name) != ".json" {
					continue
				}
				filePath := filepath.Join(bundleDir, name)
				if _, loadErr := globalBundle.LoadMessageFile(filePath); loadErr != nil {
					continue
				}
			}
		}

		msgKeyPath := filepath.Join(bundleDir, "error_msg_keys.json")
		if data, readErr := os.ReadFile(msgKeyPath); readErr == nil {
			_ = json.Unmarshal(data, &msgKeyMap)
		}
		if msgKeyMap == nil {
			msgKeyMap = make(map[int]string)
		}
	})
	return err
}

// GetLocalizer 按请求语言创建 Localizer。
func GetLocalizer(ctx context.Context, acceptLangs ...string) *goi18n.Localizer {
	if globalBundle == nil {
		panic("i18n not initialized. Call i18n.Init() in main()")
	}
	return goi18n.NewLocalizer(globalBundle, acceptLangs...)
}

// MessageByCode 按错误码返回多语言文案。
func MessageByCode(ctx context.Context, errCode int, fallback string) string {
	if globalBundle == nil {
		return fallback
	}

	msgKey := msgKeyMap[errCode]
	if msgKey == "" {
		return fallback
	}

	lang := LangFromContext(ctx)
	localizer := goi18n.NewLocalizer(globalBundle, lang)
	msg, err := localizer.Localize(&goi18n.LocalizeConfig{
		MessageID: msgKey,
		DefaultMessage: &goi18n.Message{
			ID:    msgKey,
			Other: fallback,
		},
	})
	if err != nil {
		return fallback
	}
	return msg
}

// LangFromContext 从 context 读取语言标签。
func LangFromContext(ctx context.Context) string {
	if lang, ok := ctx.Value("lang").(string); ok {
		return lang
	}
	return "zh-CN"
}

// ResetForTest 重置全局单例（仅用于测试）。
func ResetForTest() {
	once = sync.Once{}
	globalBundle = nil
	msgKeyMap = nil
}
