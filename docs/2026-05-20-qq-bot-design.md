# QQ 群聊 Bot 架构设计

> 日期：2026-05-20
> 架构：DDD 四层 + OneBot v11

---

## 一、项目概述

一个 QQ 群聊 Bot，使用独立 QQ 号加入群聊，通过 NapCatQQ（OneBot v11 协议）收发消息。

**核心功能：**
- 每日聊天自动总结
- 聊天记录自动 Compact 压缩
- @Bot 问答（RAG 检索群聊历史）

---

## 二、技术栈

| 层 | 选择 | 说明 |
|----|------|------|
| QQ 协议 | NapCatQQ + OneBot v11 | WebSocket 双向通信 |
| 后端 | Go 1.22+ | DDD 四层架构 |
| 数据库 | SQLite | 零配置，单文件 |
| LLM | OpenAI 兼容 API | 总结 + 问答 |
| 配置 | .env 文件 | 敏感信息隔离 |

---

## 三、DDD 架构

### 3.1 目录结构

```
qq-group-bot/
├── cmd/bot/main.go                 # 启动入口
├── .env.example                    # 配置模板
├── .gitignore
│
├── interfaces/                     # 接口层
│   ├── onebot/
│   │   ├── ws_server.go           # WebSocket 服务端
│   │   ├── event_handler.go       # OneBot 事件解析与分发
│   │   └── api_handler.go         # OneBot API（发消息等）
│   └── scheduler/
│       └── daily_job.go           # 定时任务
│
├── application/                    # 应用层
│   ├── chat_service.go            # 消息处理主流程
│   ├── summary_service.go         # 每日总结编排
│   ├── ask_service.go             # @问答编排
│   └── compact_service.go         # 聊天压缩编排
│
├── domain/                         # 领域层
│   ├── message/
│   │   ├── message.go             # 消息实体
│   │   ├── repository.go          # 消息仓储接口
│   │   └── compact.go             # Compact 领域逻辑
│   ├── summary/
│   │   ├── summary.go             # 总结实体
│   │   └── repository.go          # 总结仓储接口
│   └── group/
│       └── group.go               # 群组实体
│
├── infrastructure/                 # 基础设施层
│   ├── persistence/
│   │   ├── sqlite_message_repo.go # 消息仓储实现
│   │   └── sqlite_summary_repo.go # 总结仓储实现
│   ├── llm/
│   │   └── openai_provider.go     # OpenAI LLM 实现
│   └── config/
│       └── config.go              # 配置管理
│
├── skills/                         # Bot 技能提示词
│   └── summary.md                 # 总结 System Prompt
│
└── data/                           # 运行时数据（gitignore）
    └── bot.db                      # SQLite 数据库
```

### 3.2 依赖方向

```
interfaces ──→ application ──→ domain ←── infrastructure
```

- domain 无外部依赖
- infrastructure 实现 domain 定义的接口
- application 编排领域对象
- interfaces 接入外部通信

---

## 四、核心流程

### 4.1 消息采集

```
QQ群消息 → NapCatQQ → WebSocket → ws_server.Receive()
  → event_handler.Parse() → chat_service.Handle(event)
    → message.NewMessage() → messageRepo.Save(msg)
```

### 4.2 每日总结

```
scheduler 定时触发 → summary_service.Run(groupID)
  → messageRepo.QueryByDate(groupID, today)
  → llm.ChatCompletion(messages) → summaryRepo.Save(summary)
  → onebot.SendGroupMessage(summary)
```

### 4.3 @问答

```
收到 @消息 → ask_service.Answer(groupID, userID, question)
  → messageRepo.Search(groupID, question)  // 关键词检索
  → llm.ChatCompletion(context + question)  // RAG
  → onebot.SendGroupMessage(answer)
```

### 4.4 Compact 压缩

```
scheduler 定时或触发 → compact_service.Run(groupID)
  → messageRepo.QueryOldMessages(groupID, before)
  → llm.ChatCompletion("压缩以下为结构化摘要...")
  → messageRepo.Replace(before, compacted)
```

---

## 五、数据模型

### 5.1 消息实体

```go
type Message struct {
    ID        string    // 消息 ID
    GroupID   string    // 群号
    UserID    string    // 发送者 QQ
    Nickname  string    // 群昵称
    Content   string    // 消息文本
    Timestamp time.Time // 发送时间
    MsgType   string    // text / image / at
    Compacted bool      // 是否已被压缩
}
```

### 5.2 总结实体

```go
type Summary struct {
    ID        string
    GroupID   string
    Date      string    // "2026-05-20"
    Content   string    // 总结文本
    MsgCount  int       // 原始消息数
    CreatedAt time.Time
}
```

### 5.3 群组实体

```go
type Group struct {
    ID   string  // 群号
    Name string  // 群名
}
```

---

## 六、配置项设计 (.env)

| 变量 | 说明 | 必填 |
|------|------|------|
| BOT_QQ | Bot 登录 QQ 号 | 是 |
| BOT_WS_PORT | WebSocket 监听端口 | 是 |
| ONEBOT_WS_PATH | WebSocket 路径 | 否 |
| ONEBOT_ACCESS_TOKEN | OneBot 鉴权 token | 否 |
| LLM_API_URL | LLM API 地址 | 是 |
| LLM_API_KEY | LLM API Key | 是 |
| LLM_MODEL | LLM 模型名称 | 是 |
| SUMMARY_HOUR | 每日总结时间 | 否(默认 23) |
| DB_PATH | SQLite 数据库路径 | 否 |

---

## 七、开发顺序

1. **domain/** — 实体 + 仓储接口（零依赖，先定义好契约）
2. **infrastructure/config/** — 配置加载
3. **infrastructure/persistence/** — SQLite 实现
4. **infrastructure/llm/** — LLM Provider
5. **application/** — 业务编排
6. **interfaces/** — WebSocket + 定时任务
7. **cmd/bot/** — 启动入口 + 依赖注入
