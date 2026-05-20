// cmd/bot/main.go — QQ 群聊 Bot 启动入口
package main

import (
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"qq-group-bot/application"
	"qq-group-bot/infrastructure/config"
	infrallm "qq-group-bot/infrastructure/llm"
	"qq-group-bot/infrastructure/persistence"
	"qq-group-bot/interfaces/onebot"
	"qq-group-bot/interfaces/scheduler"

	_ "modernc.org/sqlite"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("QQ 群聊 Bot 启动中...")

	// ---------- 1. 加载 .env + 配置 ----------
	config.LoadDotEnv(".env")
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	log.Printf("配置加载成功, 监听端口: %s", cfg.WS_PORT)

	// ---------- 2. 初始化数据库 ----------
	if err := os.MkdirAll("data", 0755); err != nil {
		log.Fatalf("创建 data 目录失败: %v", err)
	}
	db, err := sql.Open("sqlite", cfg.DB_PATH)
	if err != nil {
		log.Fatalf("打开数据库失败: %v", err)
	}
	defer db.Close()

	if err := initDB(db); err != nil {
		log.Fatalf("初始化数据库表失败: %v", err)
	}
	log.Println("数据库初始化完成")

	// ---------- 3. 创建仓储 ----------
	msgRepo := persistence.NewSQLiteMessageRepo(db)
	summaryRepo := persistence.NewSQLiteSummaryRepo(db)

	// ---------- 4. 创建 LLM Provider ----------
	llmProvider := infrallm.NewOpenAIProvider(cfg.LLM_API_URL, cfg.LLM_API_KEY, cfg.LLM_MODEL)

	// ---------- 5. 创建 Application Service ----------
	chatService := application.NewChatService(msgRepo)
	askService := application.NewAskService(msgRepo, llmProvider)
	summaryService := application.NewSummaryService(msgRepo, summaryRepo, llmProvider)
	compactService := application.NewCompactService(msgRepo, llmProvider)

	// ---------- 6. 创建 OneBot 接口 ----------
	eventHandler := onebot.NewEventHandler(chatService, askService)
	wsAddr := ":" + cfg.WS_PORT
	wsServer := onebot.NewWSServer(wsAddr, cfg.WS_PATH, cfg.ACCESS_TOKEN, eventHandler)

	// ---------- 7. 创建定时任务 ----------
	groups := loadGroupIDs(cfg)
	dailyJob := scheduler.NewDailyJob(summaryService, compactService, cfg.SUMMARY_HOUR, groups)

	// ---------- 8. 启动服务 ----------
	// WebSocket 服务器
	go func() {
		log.Printf("OneBot WebSocket 启动于 %s%s", wsAddr, cfg.WS_PATH)
		if err := wsServer.Start(); err != nil {
			log.Fatalf("WebSocket 服务器错误: %v", err)
		}
	}()

	// 定时总结任务
	go func() {
		log.Printf("每日总结任务已配置，执行时间: %02d:00", cfg.SUMMARY_HOUR)
		for {
			now := time.Now()
			// 计算下次执行时间
			next := time.Date(now.Year(), now.Month(), now.Day(), cfg.SUMMARY_HOUR, 0, 0, 0, now.Location())
			if now.After(next) {
				next = next.Add(24 * time.Hour)
			}
			wait := next.Sub(now)
			log.Printf("下次总结时间: %s (等待 %s)", next.Format("2006-01-02 15:04"), wait.Round(time.Second))

			time.Sleep(wait)

			log.Println("执行每日总结...")
			errs := dailyJob.Run()
			for _, e := range errs {
				log.Printf("总结错误: %v", e)
			}
			log.Println("每日总结完成")

			// 避免立即重入
			time.Sleep(time.Minute)
		}
	}()

	// ---------- 9. 等待关闭信号 ----------
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh
	log.Printf("收到信号 %v, 关闭中...", sig)
}

// initDB 创建数据库表
func initDB(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS messages (
		id TEXT PRIMARY KEY,
		group_id TEXT NOT NULL,
		user_id TEXT NOT NULL,
		nickname TEXT DEFAULT '',
		content TEXT DEFAULT '',
		timestamp INTEGER NOT NULL,
		msg_type TEXT DEFAULT 'text',
		compacted INTEGER DEFAULT 0
	);
	CREATE INDEX IF NOT EXISTS idx_messages_group_date ON messages(group_id, timestamp);
	CREATE INDEX IF NOT EXISTS idx_messages_content ON messages(content);

	CREATE TABLE IF NOT EXISTS summaries (
		id TEXT PRIMARY KEY,
		group_id TEXT NOT NULL,
		date TEXT NOT NULL,
		content TEXT DEFAULT '',
		msg_count INTEGER DEFAULT 0,
		created_at INTEGER NOT NULL,
		UNIQUE(group_id, date)
	);
	`
	_, err := db.Exec(schema)
	return err
}

// loadGroupIDs 从环境变量加载群列表
// 支持 GROUP_IDS 环境变量，逗号分隔，如: "123456,789012"
func loadGroupIDs(cfg *config.Config) []string {
	ids := os.Getenv("GROUP_IDS")
	if ids == "" {
		log.Println("警告: GROUP_IDS 未设置，定时任务不会处理任何群")
		return nil
	}
	var result []string
	for _, id := range splitAndTrim(ids, ",") {
		if id != "" {
			result = append(result, id)
		}
	}
	log.Printf("已配置 %d 个群: %v", len(result), result)
	return result
}

func splitAndTrim(s, sep string) []string {
	var result []string
	for {
		idx := 0
		for idx = 0; idx < len(s); idx++ {
			if s[idx] == sep[0] {
				break
			}
		}
		if idx == 0 && s != "" {
			// 跳过空段
			if idx < len(s) {
				s = s[idx+len(sep):]
			} else {
				break
			}
			continue
		}
		part := s[:idx]
		// 手动 trim 空格
		for len(part) > 0 && part[0] == ' ' {
			part = part[1:]
		}
		for len(part) > 0 && part[len(part)-1] == ' ' {
			part = part[:len(part)-1]
		}
		if part != "" {
			result = append(result, part)
		}
		if idx < len(s) {
			s = s[idx+len(sep):]
		} else {
			break
		}
	}
	return result
}
