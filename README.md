<p align="center">
  <img src="https://raw.githubusercontent.com/DGZSbot/ai-icon/refs/heads/main/WorkBuddy.png" alt="WorkBuddy2API" width="110">
</p>

<h1 align="center">WorkBuddy2API Panel</h1>

<p align="center">
  <b>把腾讯 CodeBuddy / WorkBuddy 账号包装成多协议兼容 API 的自托管网关 · 带 Web 管理面板</b><br>
  三协议入站 · OAuth 浏览器登录 · 多账号池轮转 · 熔断与冷却 · 会话粘性 · 联网检索 · 定时任务 · 成长任务一键完成 · 媒体生成
</p>

<p align="center">
  <img alt="Go" src="https://img.shields.io/badge/Go-1.26.10-00ADD8?logo=go&logoColor=white&style=flat-square">
  <img alt="API" src="https://img.shields.io/badge/API-OpenAI%20%7C%20Codex%20%7C%20Anthropic-412991?style=flat-square">
  <img alt="Deploy" src="https://img.shields.io/badge/Deploy-Single_Binary%20%7C%20Docker-2496ED?style=flat-square">
  <img alt="License" src="https://img.shields.io/badge/License-MIT-green?style=flat-square">
</p>

---

## 这是什么

一个**自托管的多协议反向代理网关**：把 CodeBuddy / WorkBuddy 账号通过 OAuth 设备授权拿到凭证，在网关侧做 token 自动刷新、账号池调度和流量治理，对客户端暴露三种协议：

| 协议 | 端点 | 认证头 |
|---|---|---|
| OpenAI Chat Completions | `POST /v1/chat/completions` | `Authorization: Bearer <api_key>` |
| Codex Responses | `POST /v1/responses` | `Authorization: Bearer <api_key>` |
| Anthropic Messages | `POST /v1/messages`（也接受 `POST /messages`） | `x-api-key: <api_key>` |

面向**个人多账号**场景：账号共享、单号故障自动换号、冷却/熔断防雪崩、会话粘性保多轮上下文。模型名原样透传给上游，不做改写。

> ⚠️ 本项目是**非官方**网关，仅限**本人授权账号**、本机或私有环境使用。见[合规与风险](#合规与风险)。

## 血缘与维护状态

```
Sliverkiss/workbuddy2api（源头，已删库）
        └─ linguo2625469/workbuddy2api-panel（2026-10-21 停更）
                └─ FFFFFire0/workbuddy2api-panel（本仓，自主维护中）
                        └─ 已合入 misakano7545 活跃线全部 P0/P1 功能
```

- **直接上游** [linguo2625469/workbuddy2api-panel](https://github.com/linguo2625469/workbuddy2api-panel) 已停更，不再跟进接口变化。本仓是它的 fork，由本仓维护者**自主维护**。
- 2026-11-23 合入了 [misakano7545/workbuddy2api-panel](https://github.com/misakano7545/workbuddy2api-panel) 活跃线（相对共同祖先 866 个提交）的全部功能与修复：三协议入站、11140 风控连坐保护、OOM 护栏、多站点登录、成长任务 25 项、补领清扫、模型降级路由、TTS/ASR、托管出图、余额趋势等（见[与上游的差异](#与上游的差异)）。
- 同时保留了上游 fork 独有的资产：`mergeAdjacentToolCalls`（DeepSeek 并行工具调用合并）等。
- Go 工具链 **1.26.10**。

> 风险自评：上游随时可能进一步收紧非官方接入导致通道不可用；因使用非官方客户端带来的账号风控/封禁，由使用者自行承担。

## 核心能力

| 能力 | 说明 |
|---|---|
| 🔑 **OAuth 一键登录** | 面板「添加账号」或 `login.sh` 完成设备授权，凭证落盘并热加载进池（免重启）；支持 CN / 官方国际站 / 旧桌面端三个站点 |
| 🔄 **多账号池** | 快过期积分优先 + 成本分层 + 可选选号策略（加权随机 / 轮询 / 顺序耗尽，热生效）+ 防惊群 |
| 🛡️ **熔断与冷却** | 429 软冷却指数退避、404 短冷却、402 硬冷却到次日 04:00、连续失败熔断、在途租约限流；11140 风控命中就地停手并全池暂停派发，避免连坐 |
| 🧲 **会话粘性** | 同一会话尽量绑定同一账号，TTL 滚动续期，失败自动解绑；客户端不发会话 ID 时按内容派生 |
| ⏰ **定时任务** | 签到、活跃上报、猫猫旅行、token 保活、成长任务队列，各自独立开关与时段 |
| ⚡ **流式 + 非流式** | 出站强制流式，SSE 帧按白名单重建；非流式本地聚合成单响应 |
| 🧠 **推理模型兼容** | DeepSeek 思维链注入、reasoning 多轮回填、effort 档位对齐模型实际支持范围 |
| 🗑️ **指纹脱敏** | 出站请求体黑名单字段清洗 + 网关自有系统提示词，两层消灭指纹误报 |
| 📊 **可观测** | 每请求表格日志（TTFB / token 速率）、按天落盘、上游劣化滑窗告警、`/healthz` 探活、`/metrics` Prometheus 指标 |
| 💾 **状态持久化** | 池状态原子落盘，可选 Upstash Redis 异步镜像，重启择新恢复 |
| 🌐 **联网检索** | `/v1/search` / `/v1/fetch` 透传官方检索；`/v1/messages`、`/v1/responses` 支持服务端联网工具 |
| 🎬 **媒体生成 + MCP** | 生图 / 改图 / 出视频 / 3D（默认关）、`POST /mcp`（Streamable HTTP），记账与换号复用同一路径 |
| 🖥️ **Web 管理面板** | 内嵌单页面板（明暗主题），账号运维 / 模型档位 / 在线改配置（热生效）/ 运行日志 / 任务中心 / 余额趋势 |

## 快速开始

### 方式一：Docker Compose（推荐）

```bash
git clone https://github.com/FFFFFire0/workbuddy2api-panel.git
cd workbuddy2api-panel
cp config.example.json config.json
docker compose up -d --build
```

`docker-compose.yml` 默认**从源码构建**，保证跑的是本仓代码；端口 `7863`。

- 容器以 uid 10001 运行，若挂载目录权限报错：`PUID=$(id -u) PGID=$(id -g) docker compose up -d --build`。
- 容器内已显式放开监听（`WB2A_LISTEN=0.0.0.0:7863`），宿主机部署的程序默认只绑 `127.0.0.1`。
- 想改用预编译镜像：注释 `build` 段、取消 `image` 行注释，镜像由本仓 CI 推送到 `ghcr.io/<你的仓库>`。

### 方式二：本地二进制

需要 Go 1.26.10（本机若无，见文末工具链说明）：

```bash
go build -o wb2api ./cmd/server
cp config.example.json config.json
./wb2api -config config.json
```

首次启动若无 `config.json` 会自动生成（含随机 `api_key`）。打开面板：<http://127.0.0.1:7863/panel/>

### 添加账号

- **面板**：打开 `/panel/` → 「添加账号」→ 浏览器完成设备授权，凭证落盘并自动入池；
- **命令行**：`./login.sh`（CN）、`./login.sh --site=intl`（国际站）、`./login.sh --site=workbuddy`（旧桌面端）。Docker 下用 `docker exec -it workbuddy2api ./login.sh`。

### 验证

```bash
curl http://127.0.0.1:7863/healthz
curl http://127.0.0.1:7863/v1/models \
  -H "Authorization: Bearer <你的 api_key>"
```

## API 端点速查

所有端点除 `/healthz` 外均需鉴权（`/v1/messages` 用 `x-api-key`，其余用 Bearer）。

| 方法 | 路径 | 用途 |
|---|---|---|
| POST | `/v1/chat/completions` | OpenAI Chat Completions |
| POST | `/v1/responses` | Codex Responses |
| POST | `/v1/messages` | Anthropic Messages |
| POST | `/v1/messages/count_tokens` | Anthropic token 计数 |
| GET | `/v1/models` | 模型目录（含能力标记） |
| GET | `/v1/capacity` | 池容量 |
| GET | `/status` | 运行状态 |
| GET | `/healthz` | 探活（无鉴权） |
| GET | `/metrics` | Prometheus 指标 |
| POST | `/v1/images/generations` / `/v1/images/edits` | 文生图 / 改图 |
| POST | `/v1/videos` / `/v1/videos/generations` | 提交视频任务 |
| GET | `/v1/videos/{task_id}` `/content` | 查任务 / 取成品 |
| POST | `/v1/3d/generations` / `/v1/3d/tasks` | 3D 生成（默认关） |
| POST | `/v1/images/strip-watermark` | 去水印（默认关） |
| POST | `/v1/audio/speech` / `/v1/audio/transcriptions` | TTS / ASR（默认关） |
| POST | `/v1/embeddings` | 嵌入（默认关） |
| POST | `/v1/search` / `/v1/fetch` | 联网检索 / 抓取 |
| POST | `/mcp` | MCP（JSON-RPC 2.0，Streamable HTTP） |
| GET | `/panel/` | Web 管理面板 |

默认功能开关在 `config.json` 的 `features` 段：`tts_enabled` / `asr_enabled` / `embeddings_enabled` / `three_d_enabled` / `watermark_strip_enabled` 均为 `false`，按需打开（面板保存后热生效）。

## 配置要点

配置即 `config.json`（模板 `config.example.json` 内每个字段都有 `_comment` 说明）。常用项：

| 配置 | 默认 | 说明 |
|---|---|---|
| `listen` | `127.0.0.1:7863` | 只绑本机；对外暴露请放在反代后并加认证 |
| `api_key` | — | 客户端鉴权密钥，首启随机生成 |
| `server.max_payload_bytes` | `67108864` | 入站请求体上限（64 MiB），OOM 护栏 |
| `cooldown.soft_rate` | `600s` | 429 软冷却起点 |
| `upstream.restriction_hold_seconds` | `1800` | 11140 风控连坐后全池暂停秒数，`0` = 只停手不暂停 |
| `pool.model_fallback` | 空 | 6004 模型级限流时的等价模型降级表 |
| `pool.reserve_credits` | `10` | 保留积分，低于此值停止接单（防余额用到 0） |
| `schedule.*` | 见模板 | 签到 / 旅行 / 活跃 / 保活 / 成长任务排程 |
| `features.*` | 见模板 | 各类功能开关，多数热生效 |

面板「配置」页可在线修改并热生效；装配期字段（如 `listen`）需重启。写入采用深合并 + 原子替换。

## 与上游的差异

本仓在 linguo 停更点之上，主要增量来自 misakano 活跃线的 P0/P1 合入：

- **P0**：三协议入站、11140 风控连坐保护、11102 归请求级、入站请求体 OOM 护栏、多站点登录、Go 1.26.10。
- **P1**：成长任务 25 项、补领清扫、6004 模型降级路由、TTS/ASR、托管出图、余额趋势、models.dev 能力声明、请求来源记录、jsonscan 优化等。
- **保留的 fork 独有资产**：`mergeAdjacentToolCalls`（DeepSeek 11148 并行工具调用合并，在 `internal/upstream/tool_pairing.go`，`payload.go` 已接线）。
- **未引入**：AutoBot 自动开 PR 流水线、住宅代理返利推广。
- module 名保持 `github.com/linguo2625469/workbuddy2api-panel`，以延续本 fork 的身份。

验证状态：`go build ./...` 通过，`go test ./...` 全绿（29 包）。

## 合规与风险

- 本项目**非官方**，使用 CodeBuddy / WorkBuddy 账号作为上游，仅限本人授权账号、本机 / 私有环境测试。
- 请勿用于搭建对外中转站、批量注册小号或转售额度；这类滥用违背上游初衷，也是账号风控和项目停摆的主因。
- 上游接口与风控策略随时可能变化，通道不可用时需跟进修复；非官方客户端导致的账号风险由使用者自担。

## 本地开发 / 工具链

本机若无 Go 1.26.10，可从 Go 官方 module proxy 下载（境内可靠源）：

```bash
# 下载 golang.org/toolchain 的 1.26.10 包（约 73MB），解压后将 GOROOT 指向其中的
# golang.org/toolchain@v0.0.1-go1.26.10.windows-amd64 目录
$env:GOROOT = "<解压目录>\golang.org\toolchain@v0.0.1-go1.26.10.windows-amd64"
$env:GOPROXY = "https://goproxy.cn,direct"
$env:GOSUMDB = "sum.golang.google.cn"
go version          # go1.26.10 windows/amd64
go build ./...
go test ./...
```

## License

[MIT](LICENSE)

> 本 README 描述的是 `FFFFFire0/workbuddy2api-panel` 自主维护分支。上游设计中的账号池调度、错误分类与提示词体系原样保留，感谢所有上游维护者。
