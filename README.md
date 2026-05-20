# qq-group-bot

基于 OneBot v11 协议的 QQ 群聊机器人，支持自动总结、聊天记录搜索、@问答等功能。

## 功能

- **每日总结** — 定时生成群聊每日摘要，按主题归类，标注热点和关键决策
- **@ 问答** — @Bot 提问，自动检索相关聊天记录后回答
- **消息存储** — 群聊消息持久化到 SQLite，支持全文搜索
- **聊天 compact** — 将旧消息压缩归档，保留语义信息的同时减少 token 消耗

## 架构

```
cmd/bot/          # 启动入口
application/      # 应用服务层（AskService, ChatService, SummaryService, CompactService）
domain/           # 领域层（Message, Group, Summary, LLM Provider 接口）
infrastructure/   # 基础设施层（LLM Provider 实现, SQLite 仓储, 配置）
interfaces/       # 接口层（OneBot WebSocket 服务端, 定时任务调度）
```

## 快速开始

### 1. 前置条件

- Go 1.21+
- [NapCatQQ](https://github.com/NapNeko/NapCatQQ) 或其他 OneBot v11 客户端
- 任选其一即可的 LLM API：
  - [ModelScope](https://modelscope.cn) — 国内免费，注册即用
  - [OpenAI](https://platform.openai.com) — GPT 系列
  - [Ollama](https://ollama.com) — 本地运行
  - 任何 OpenAI 兼容接口（vLLM, OneAPI, DeepSeek 等）

### 2. 配置

```bash
cp .env.example .env
# 编辑 .env，填写必要信息
```

`.env` 说明：

| 变量 | 必填 | 说明 |
|------|------|------|
| `BOT_QQ` | 是 | Bot 登录的 QQ 号 |
| `GROUP_IDS` | 是 | 监控的群号，逗号分隔 |
| `LLM_API_URL` | 是 | OpenAI 兼容 API 地址 |
| `LLM_API_KEY` | 是 | API Key |
| `LLM_MODEL` | 是 | 模型名称 |
| `BOT_WS_PORT` | 否 | WebSocket 端口，默认 `8080` |
| `ONEBOT_WS_PATH` | 否 | WebSocket 路径，默认 `/ws` |
| `ONEBOT_ACCESS_TOKEN` | 否 | 与 NapCatQQ 一致的 access_token |
| `SUMMARY_HOUR` | 否 | 每日总结时间（小时），默认 `23` |
| `DB_PATH` | 否 | SQLite 数据库路径，默认 `./data/bot.db` |

### 3. 各平台 LLM 配置示例

**ModelScope（魔塔社区）**

```bash
LLM_API_URL=https://api-inference.modelscope.cn/v1
LLM_API_KEY=ms-xxxxxxxxxxxxxxxx
LLM_MODEL=Qwen/Qwen3-Next-80B-A3B-Instruct
```

**OpenAI**

```bash
LLM_API_URL=https://api.openai.com/v1
LLM_API_KEY=sk-xxxxxxxxxxxxxxxx
LLM_MODEL=gpt-4o
```

**Ollama 本地**

```bash
LLM_API_URL=http://localhost:11434/v1
LLM_API_KEY=ollama
LLM_MODEL=qwen2.5:7b
```

**DeepSeek**

```bash
LLM_API_URL=https://api.deepseek.com/v1
LLM_API_KEY=sk-xxxxxxxxxxxxxxxx
LLM_MODEL=deepseek-chat
```

### 4. 构建与运行

```bash
# 安装依赖
go mod tidy

# 构建
go build -o bot ./cmd/bot/

# 运行
./bot
```

### 5. 配置 NapCatQQ

1. 打开 NapCatQQ WebUI（默认 `http://127.0.0.1:6099/webui/`）
2. 添加 WebSocket 客户端连接：
   - 协议：`WebSocket`
   - 连接类型：`正向连接`（即 NapCatQQ 作为客户端连到 Bot）
   - 地址：`ws://localhost:8080/ws`（端口和路径与 `.env` 中一致）
   - Access Token：与 `.env` 中 `ONEBOT_ACCESS_TOKEN` 一致
3. 连接成功后在群里 @Bot 提问

### 6. 运行测试

```bash
go test ./...
```

## 项目结构

```
qq-group-bot/
├── cmd/bot/main.go              # 启动入口
├── application/                  # 应用服务
│   ├── ask_service.go           # @问答（RAG 检索 + LLM）
│   ├── chat_service.go          # 消息保存
│   ├── summary_service.go       # 每日总结生成
│   └── compact_service.go       # 消息压缩
├── domain/                       # 领域层
│   ├── llm/provider.go          # LLM Provider 接口
│   ├── message/                 # 消息领域
│   │   ├── message.go
│   │   ├── repository.go        # 消息仓储接口
│   │   └── compact.go           # 压缩逻辑
│   ├── group/                   # 群聊领域
│   └── summary/                 # 总结领域
├── infrastructure/               # 基础设施
│   ├── config/config.go         # 配置加载与校验
│   ├── llm/openai_provider.go   # OpenAI 兼容 LLM Provider
│   └── persistence/             # SQLite 仓储实现
├── interfaces/                   # 接口层
│   ├── onebot/                  # OneBot v11 实现
│   │   ├── ws_server.go         # WebSocket 服务端
│   │   ├── event_handler.go     # 事件解析与分发
│   │   └── api_handler.go       # OneBot API 调用
│   └── scheduler/daily_job.go   # 定时任务
├── skills/summary.md            # 总结的 System Prompt 模板
├── .env.example                 # 环境变量示例
└── docs/                        # 设计文档
```

## 工作原理

### 消息流程

```
QQ群消息 → NapCatQQ → WebSocket → ws_server → event_handler
                                                    ├── @消息 → AskService (RAG 检索 → LLM → 回复)
                                                    └── 普通消息 → ChatService (存入 SQLite)
```

### 每日总结

`scheduler` 按 `SUMMARY_HOUR` 定时执行：
1. 读取当天所有群消息
2. 压缩老旧消息（compact）
3. 调用 LLM 生成摘要
4. 摘要持久化到 SQLite

### RAG 问答

用户 @Bot 提问时：
1. 用问题全文检索 SQLite 中的聊天记录
2. 构建上下文 Prompt（问题 + 检索到的消息）
3. 调用 LLM 生成回答
4. 通过 OneBot API 回复到群

## License

MIT
