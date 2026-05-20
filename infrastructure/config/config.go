// Package config 提供应用配置的加载和验证功能。
// 配置从环境变量读取，支持默认值，对必填字段进行校验。
package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// LoadDotEnv 读取 .env 文件并将键值对设置到环境变量中（不覆盖已有值）。
func LoadDotEnv(filepath string) error {
	f, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// 支持 KEY=VALUE 格式
		idx := strings.IndexByte(line, '=')
		if idx == -1 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		val = strings.Trim(val, "\"'")
		if key == "" {
			continue
		}
		// 不覆盖已有的环境变量
		if os.Getenv(key) == "" {
			os.Setenv(key, val)
		}
	}
	return scanner.Err()
}

// Config 应用配置结构体
type Config struct {
	BOT_QQ       string // Bot 登录的 QQ 号（必填）
	WS_PORT      string // WebSocket 监听端口，默认 "8080"
	WS_PATH      string // WebSocket 路径，默认 "/ws"
	ACCESS_TOKEN string // OneBot access_token
	LLM_API_URL  string // OpenAI 兼容 API 地址（必填）
	LLM_API_KEY  string // API Key（必填）
	LLM_MODEL    string // 模型名称（必填）
	SUMMARY_HOUR int    // 每日总结时间（小时），默认 23
	DB_PATH      string // SQLite 数据库路径，默认 "./data/bot.db"
}

// LoadConfig 从环境变量加载配置，设置默认值，并验证必填字段。
func LoadConfig() (*Config, error) {
	cfg := &Config{
		BOT_QQ:       os.Getenv("BOT_QQ"),
		WS_PORT:      getEnvWithDefault("BOT_WS_PORT", "8080"),
		WS_PATH:      getEnvWithDefault("ONEBOT_WS_PATH", "/ws"),
		ACCESS_TOKEN: os.Getenv("ONEBOT_ACCESS_TOKEN"),
		LLM_API_URL:  os.Getenv("LLM_API_URL"),
		LLM_API_KEY:  os.Getenv("LLM_API_KEY"),
		LLM_MODEL:    os.Getenv("LLM_MODEL"),
		DB_PATH:      getEnvWithDefault("DB_PATH", "./data/bot.db"),
	}

	// 解析 SUMMARY_HOUR，使用默认值 23
	summaryHourStr := os.Getenv("SUMMARY_HOUR")
	if summaryHourStr == "" {
		cfg.SUMMARY_HOUR = 23
	} else {
		hour, err := strconv.Atoi(summaryHourStr)
		if err != nil {
			return nil, fmt.Errorf("SUMMARY_HOUR 不是有效的整数: %s", summaryHourStr)
		}
		cfg.SUMMARY_HOUR = hour
	}

	// 验证必填字段
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// getEnvWithDefault 读取环境变量，若为空则返回默认值。
func getEnvWithDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// validate 检查必填字段是否已设置，返回第一个缺失的字段错误。
func (c *Config) validate() error {
	required := []struct {
		value string
		name  string
	}{
		{c.BOT_QQ, "BOT_QQ"},
		{c.LLM_API_URL, "LLM_API_URL"},
		{c.LLM_API_KEY, "LLM_API_KEY"},
		{c.LLM_MODEL, "LLM_MODEL"},
	}

	for _, r := range required {
		if r.value == "" {
			return fmt.Errorf("必填字段 %s 未设置", r.name)
		}
	}
	return nil
}
