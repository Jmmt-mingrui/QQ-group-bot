package config

import (
	"os"
	"testing"
)

// 清空所有相关的环境变量，确保测试隔离
func clearEnv() {
	os.Unsetenv("BOT_QQ")
	os.Unsetenv("BOT_WS_PORT")
	os.Unsetenv("ONEBOT_WS_PATH")
	os.Unsetenv("ONEBOT_ACCESS_TOKEN")
	os.Unsetenv("LLM_API_URL")
	os.Unsetenv("LLM_API_KEY")
	os.Unsetenv("LLM_MODEL")
	os.Unsetenv("SUMMARY_HOUR")
	os.Unsetenv("DB_PATH")
}

// 设置测试用的环境变量
func setTestEnv() {
	os.Setenv("BOT_QQ", "123456")
	os.Setenv("BOT_WS_PORT", "9000")
	os.Setenv("ONEBOT_WS_PATH", "/onebot")
	os.Setenv("ONEBOT_ACCESS_TOKEN", "mytoken")
	os.Setenv("LLM_API_URL", "https://api.openai.com/v1")
	os.Setenv("LLM_API_KEY", "sk-test123")
	os.Setenv("LLM_MODEL", "gpt-4")
	os.Setenv("SUMMARY_HOUR", "8")
	os.Setenv("DB_PATH", "/tmp/test.db")
}

// TestLoadConfig_AllFields 测试从环境变量加载所有字段
func TestLoadConfig_AllFields(t *testing.T) {
	clearEnv()
	setTestEnv()
	defer clearEnv()

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("期望无错误，但得到: %v", err)
	}

	if cfg.BOT_QQ != "123456" {
		t.Errorf("BOT_QQ 期望 '123456'，但得到 '%s'", cfg.BOT_QQ)
	}
	if cfg.WS_PORT != "9000" {
		t.Errorf("WS_PORT 期望 '9000'，但得到 '%s'", cfg.WS_PORT)
	}
	if cfg.WS_PATH != "/onebot" {
		t.Errorf("WS_PATH 期望 '/onebot'，但得到 '%s'", cfg.WS_PATH)
	}
	if cfg.ACCESS_TOKEN != "mytoken" {
		t.Errorf("ACCESS_TOKEN 期望 'mytoken'，但得到 '%s'", cfg.ACCESS_TOKEN)
	}
	if cfg.LLM_API_URL != "https://api.openai.com/v1" {
		t.Errorf("LLM_API_URL 期望 'https://api.openai.com/v1'，但得到 '%s'", cfg.LLM_API_URL)
	}
	if cfg.LLM_API_KEY != "sk-test123" {
		t.Errorf("LLM_API_KEY 期望 'sk-test123'，但得到 '%s'", cfg.LLM_API_KEY)
	}
	if cfg.LLM_MODEL != "gpt-4" {
		t.Errorf("LLM_MODEL 期望 'gpt-4'，但得到 '%s'", cfg.LLM_MODEL)
	}
	if cfg.SUMMARY_HOUR != 8 {
		t.Errorf("SUMMARY_HOUR 期望 8，但得到 %d", cfg.SUMMARY_HOUR)
	}
	if cfg.DB_PATH != "/tmp/test.db" {
		t.Errorf("DB_PATH 期望 '/tmp/test.db'，但得到 '%s'", cfg.DB_PATH)
	}
}

// TestLoadConfig_Defaults 测试可选字段使用默认值
func TestLoadConfig_Defaults(t *testing.T) {
	clearEnv()
	// 仅设置必填字段
	os.Setenv("BOT_QQ", "123456")
	os.Setenv("LLM_API_URL", "https://api.openai.com/v1")
	os.Setenv("LLM_API_KEY", "sk-test123")
	os.Setenv("LLM_MODEL", "gpt-4")
	defer clearEnv()

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("期望无错误，但得到: %v", err)
	}

	if cfg.WS_PORT != "8080" {
		t.Errorf("WS_PORT 默认值期望 '8080'，但得到 '%s'", cfg.WS_PORT)
	}
	if cfg.WS_PATH != "/ws" {
		t.Errorf("WS_PATH 默认值期望 '/ws'，但得到 '%s'", cfg.WS_PATH)
	}
	if cfg.SUMMARY_HOUR != 23 {
		t.Errorf("SUMMARY_HOUR 默认值期望 23，但得到 %d", cfg.SUMMARY_HOUR)
	}
	if cfg.DB_PATH != "./data/bot.db" {
		t.Errorf("DB_PATH 默认值期望 './data/bot.db'，但得到 '%s'", cfg.DB_PATH)
	}
}

// TestLoadConfig_MissingRequired 测试必填字段缺失时报错
func TestLoadConfig_MissingRequired(t *testing.T) {
	tests := []struct {
		name      string
		unsetEnv  string // 取消设置的环境变量名
	}{
		{"缺失 BOT_QQ", "BOT_QQ"},
		{"缺失 LLM_API_URL", "LLM_API_URL"},
		{"缺失 LLM_API_KEY", "LLM_API_KEY"},
		{"缺失 LLM_MODEL", "LLM_MODEL"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnv()
			// 设置所有必填
			os.Setenv("BOT_QQ", "123456")
			os.Setenv("LLM_API_URL", "https://api.openai.com/v1")
			os.Setenv("LLM_API_KEY", "sk-test123")
			os.Setenv("LLM_MODEL", "gpt-4")
			// 取消指定的必填字段
			os.Unsetenv(tt.unsetEnv)
			defer clearEnv()

			_, err := LoadConfig()
			if err == nil {
				t.Errorf("期望错误，但得到 nil")
			}
		})
	}
}

// TestConfigStruct 验证配置结构体能正确暴露所有字段
func TestConfigStruct(t *testing.T) {
	// 编译期检查：确认结构体字段存在
	cfg := Config{
		BOT_QQ:       "test",
		WS_PORT:      "8080",
		WS_PATH:      "/ws",
		ACCESS_TOKEN: "token",
		LLM_API_URL:  "https://example.com",
		LLM_API_KEY:  "key",
		LLM_MODEL:    "model",
		SUMMARY_HOUR: 23,
		DB_PATH:      "./data/bot.db",
	}

	if cfg.BOT_QQ != "test" {
		t.Error("BOT_QQ 字段暴露异常")
	}
	if cfg.WS_PORT != "8080" {
		t.Error("WS_PORT 字段暴露异常")
	}
	if cfg.WS_PATH != "/ws" {
		t.Error("WS_PATH 字段暴露异常")
	}
	if cfg.ACCESS_TOKEN != "token" {
		t.Error("ACCESS_TOKEN 字段暴露异常")
	}
	if cfg.LLM_API_URL != "https://example.com" {
		t.Error("LLM_API_URL 字段暴露异常")
	}
	if cfg.LLM_API_KEY != "key" {
		t.Error("LLM_API_KEY 字段暴露异常")
	}
	if cfg.LLM_MODEL != "model" {
		t.Error("LLM_MODEL 字段暴露异常")
	}
	if cfg.SUMMARY_HOUR != 23 {
		t.Error("SUMMARY_HOUR 字段暴露异常")
	}
	if cfg.DB_PATH != "./data/bot.db" {
		t.Error("DB_PATH 字段暴露异常")
	}
}
