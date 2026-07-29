# Codex Project Context

## Purpose and scope

This repository is a customized QQ bot built on go-cqhttp and MiraiGo. This document is a navigation baseline for Codex: it identifies the active modules, runtime flows, configuration, and deployment artifacts. Source code remains authoritative. Tracked automation and the local deployment manifest describe deployment intent; live Kubernetes state must be verified with read-only checks.

## Required Jarvis development rule

- `/Users/amethyst/Jarvis/jarvis/api/grpc` is the authoritative Jarvis Proto definition directory.
- Before changing `protocol/jarvis.proto` or regenerating its Go bindings, compare against that directory.
- The authoritative directory is a developer-workstation rule, not a container runtime dependency.

## User-facing capabilities

- Login supports configured password, cached session token, or QR code, followed by token-based and normal-login reconnect paths.
- Group and private messages share a `#` command experience. The bot also exposes OneBot-style HTTP calls and event posts, forward/reverse WebSocket transports, and a QQBot gRPC interface with `SendMsg`, `Self`, `Friends`, and `Groups`.
- On-demand news covers Weibo, 36Kr, Wall Street CN, Caixin, and Zhihu. Scheduled reporters cover Weibo, 36Kr, Wall Street CN, and Caixin; gold fluctuation delivery runs hourly at minute 30. Zhihu remains on-demand.
- Private chats expose `#天气 城市` to set a daily 08:00 weather push, `#天气` to inspect it, and `#天气 关闭` to cancel it. Group chats do not list or handle the weather command.
- Finance and calculation features include cryptocurrency prices, gold prices, exchange rates, mathematical expressions, and the Kelly strategy.
- The hourly gold reporter fetches one price per run, compares it with the previous scheduled observation, records every valid observation, and fans a 10%-or-greater rise/fall alert out to both group and private report subscribers. Overlapping runs are skipped, invalid or non-finite prices do not change the baseline, and manual `#黄金` queries do not change the scheduled comparison baseline.
- Translation rotates across the translation adapters that initialized successfully.
- Jarvis gRPC is the only text-assistant/model path. OpenAI is retained only for DALL·E 3 image generation.
- State is stored in local JSON, synchronized to Tencent COS where implemented, and optionally in LevelDB for message history and reply/recall context.
- `CMDImage` is omitted from bare-`#` group/private help lists, but both handler maps still register it, so direct `#图片` remains callable.

## Repository map

| Path | Responsibility |
| --- | --- |
| `main.go`, `shutdown.go` | Process entry point, timezone/logging, outbound Spider gRPC initialization, daily-record upload scheduling, and shutdown notification wiring. |
| `cmd/gocq/` | CLI/config initialization, blocking Jarvis startup and deferred close, login/reconnect, bot construction, transports, and process lifetime. |
| `coolq/` | Core bot, events, commands, single Jarvis assistant adapter, reporters/jobs, Jarvis summary formatting, state, CQ codes, and generated OneBot call targets. |
| `cluster/jarvis_svc/` | Reusable Jarvis gRPC connection, per-operation deadlines, chat/ingest/summary calls, stable Wall Street knowledge identity, and close lifecycle. |
| `protocol/jarvis.proto`, `protocol/pb/jarvis_pb/` | Repository copy of the Jarvis API and generated Go bindings. |
| `server/`, `modules/servers/`, `modules/api/` | HTTP/WebSocket transports, server registry, middleware, and generated OneBot action routing. |
| `modules/config/`, `internal/base/` | YAML generation/parsing, environment expansion, flags, and normalized runtime configuration including Jarvis. |
| `cluster/`, `bot_service/` | QQBot gRPC server, outbound Spider client wrapper, and QQBot gRPC implementation. |
| `protocol/spider.proto`, `protocol/qqbot.proto`, `protocol/pb/` | Spider/QQBot protobuf definitions and generated Go bindings. |
| `db/` | Persistence interfaces, multi-backend coordination, LevelDB adapter, and an available but not default-registered MongoDB implementation. |
| `util/top_list/` | News sources, deduplication, daily records, and Wall Street incremental collection data. |
| `util/ai_util/` | OpenAI DALL·E 3 image client only; legacy text-provider implementations have been removed. |
| `util/` | Cron, finance, translation, math, file/COS, HTTP, encryption, and content helpers. |
| `.github/workflows/`, `Dockerfile` | Image build/push, Pod refresh automation, and runtime image. |
| `qq_bot.yml` | Local, untracked, potentially sensitive Kubernetes manifest in the primary workspace; absent from Git history. |

## Startup and runtime flow

1. `main.go` fixes local time at UTC+8, configures logging, initializes the registered outbound Spider gRPC client, installs shutdown notification, schedules daily-record upload at 00:05, and calls `gocq.Main`.
2. `gocq.Main` parses flags and loads `config.yml` through `base.Init`.
3. Immediately after configuration, it performs a blocking `jarvis_svc.Init(base.Jarvis)`. Failure is fatal; a successful connection is reused for calls and closed by `defer` when `gocq.Main` returns.
4. The program configures rotating logs/directories/caches and databases, prepares device information, then uses cached-token, password, or QR login.
5. After login it reloads friends/groups, applies presence, constructs `coolq.NewQQBot`, starts configured transports, assigns `bot_service.QQBot`, and starts update diagnosis plus the QQBot gRPC server on `:9090`.
6. `NewQQBot` initializes persisted state, subscribes active QQ events, registers reporter jobs including the private weather job at 08:00, announces startup to loaded groups, starts heartbeat dispatch, and listens for shutdown.
7. Active group/private message subscriptions each try calculator, Jarvis assistant, and `#` command flows for the same incoming message. Most standard OneBot event subscriptions remain commented out; assume only explicitly active subscriptions.
8. Shutdown saves subscription/global/sent-news state and backs up news records. There is no local assistant model or conversation-session state to persist.

## Jarvis assistant and command flow

- Assistant requests trigger on `?`, `？`, or a group mention of the bot.
- Private conversations use `qq:private:<user>`.
- Group conversations use `qq:group:<group>:<user>`, isolating the same sender across groups.
- Jarvis manages model selection, history, and conversation lifecycle. go-cqhttp has no local context-expiry window and does not claim a 30-minute window.
- `#模式` with no argument replies `当前助手模式：Jarvis`.
- `#模式 <argument>` replies `模型已由 Jarvis 服务统一管理，不支持在机器人端切换`.
- `coolq.State` now owns scheduled-push recipients, sent-news hashes, and global values. Legacy per-user assistant-model choices and in-memory group/private dialogue sessions are removed.

## Wall Street knowledge and summary flow

- This integration currently covers Wall Street CN only; other news sources retain their existing reporter behavior.
- The scheduled Wall Street reporter runs at minutes 05/15/25/35/45/55 from 09:00 through 23:59. Each fetch atomically marks a URL/title/summary fingerprint, records an RFC3339 collection time, stores new or changed items in the daily record, and asynchronously ingests trusted URLs deduplicated within that fetch batch.
- Ingest content contains the Asia/Shanghai news date, title, and summary; `source_url` contains an allowlisted Wall Street CN HTTPS URL. The stable date plus unchanged title, summary, and URL produce deterministic same-day knowledge input. Jarvis is not asked to scrape that URL during ingest.
- At 23:56, `SummaryReporter` fetches group and private subscribers and calls `ReportSummary` once. One Jarvis result is then fanned out to all recipients.
- The summary uses the same stable Wall Street Jarvis identity as ingestion. Its prompt includes the current collected material, asks Jarvis to use historical knowledge to build a dated cross-day causal timeline, and permits selective `web_scrape` calls only for relevant allowlisted URLs at summary time.
- The response must contain exactly four ordered chapter markers. The bot sends these fixed chapters separately:
  1. `今日全球经济形势概览`
  2. `重点事件、原文补充与影响分析`
  3. `跨日时间线及因果传导`
  4. `未来展望、风险与机会`
- Only an individual chapter longer than the message limit is split further. Low-confidence responses receive a warning. Build, empty-response, RPC, or marker-parse failures send a summary-failure message.

## Private weather flow

- Weather settings are private-chat only. `#天气 城市` saves or replaces the sender's city, `#天气` reports the current city and 08:00 schedule, and `#天气 关闭` removes the subscription.
- The scheduled job snapshots the private weather state, groups recipients by city, and calls Jarvis once for each distinct city. The same city result is fanned out to all subscribed private users.
- Each request includes the current Asia/Shanghai date and asks Jarvis to query that city's daily forecast, then provide travel, sun-protection, clothing, and safety advice without inventing unavailable weather values.
- The prompt allows normal QQ text, emoji, and symbolic expressions but forbids Markdown. The bot also removes common Markdown headings, list markers, emphasis, code fences, tables, quotes, and links before sending without filtering emoji.
- Jarvis requests use a stable hashed identity per city. RPC, nil-response, empty-response, or empty-sanitized-response failures produce a short private failure message for that city's recipients without stopping other cities.

## Interfaces and seams

- `modules/servers` separates configured transports from `coolq.CQBot`; `server` supplies HTTP/reverse-HTTP and forward/reverse-WebSocket adapters.
- `modules/api.Caller` connects a transport `Getter` plus middleware to generated `CQBot` calls.
- `db.Database` is the persistence seam. Enabled adapters all receive writes, while reads use `backends[0]`; map traversal means the first backend is not guaranteed to match configuration order.
- `Assistant`, `Translator`, `Calculator`, and `ImgGenerator` are shared feature interfaces. Group/private assistant adapters provide chat-specific extraction and stable Jarvis identities.
- `cluster/jarvis_svc` is the outbound Jarvis seam. It owns one reusable insecure in-cluster gRPC connection and operation-specific context deadlines; it does not retry calls.
- The weather reporter uses the normal Jarvis chat deadline and a city-isolated identity; it does not add another gRPC client or connection.
- `cluster/spider_svc` wraps the generated Spider client for active news and finance calls. Legacy Bing text-assistant use has been removed.
- `bot_service.BotService` implements the inbound QQBot gRPC methods.

## Configuration, state, and persistence

`modules/config` expands `${NAME}` and `${NAME:default}` before YAML decoding. The Jarvis configuration has code defaults, so a config that omits the entire section still uses the following values:

```yaml
jarvis:
  grpc-target: ${JARVIS_GRPC_TARGET:jarvis:9090}
  connect-timeout-seconds: 10
  chat-timeout-seconds: 300
  ingest-timeout-seconds: 120
  summary-timeout-seconds: 600
  wallstreet-user-id: go-cqhttp:wallstreet
```

`JARVIS_GRPC_TARGET` optionally overrides the in-cluster default `jarvis:9090`. The runtime uses Jarvis over gRPC only. The target is an outbound Jarvis address and is distinct from the bot's inbound QQBot gRPC listener on `:9090`.

Local JSON state includes group/private report recipients, private weather cities, sent-news deduplication, global values, and per-source daily records. Private weather settings use `weatherPrivateState.json` with a `QQ user ID -> city` mapping and synchronize through the existing COS cache path. Assistant model choices and assistant conversations are no longer local or durable go-cqhttp state.

Sensitive runtime inputs include QQ/session/device data, OneBot tokens, OpenAI image credentials, translation and COS credentials, external or credential-bearing service addresses, registry/cluster credentials, owner/account identifiers, and QR codes. Never record those values in documentation, logs, patches, commits, or temporary credential files. Stable in-cluster Service DNS names such as `jarvis:9090` may be documented when they contain no credentials.

The primary workspace's local `qq_bot.yml` declares these host-path mounts. It is untracked and must be re-read before any deployment:

| Host path | Container path | Purpose |
| --- | --- | --- |
| `/root/qqbot/file` | `/tmp` | Generated/downloaded files and JSON state when selected by `FILE_ROOT`. |
| `/root/qqbot/config.yml` | `/opt/config.yml` | Runtime configuration. |
| `/root/qqbot/address.txt` | `/opt/address.txt` | Login/address state. |
| `/root/qqbot/device.json` | `/opt/device.json` | MiraiGo device identity. |
| `/root/qqbot/session.token` | `/opt/session.token` | Login session cache. |
| `/root/qqbot/log` | `/opt/log` | Declared log mount. |
| `/root/qqbot/data` | `/opt/data` | Internal data, including LevelDB at `data/leveldb-v3` when enabled. |

## Dependencies

- `go.mod` targets Go 1.19; the Docker builder uses Go 1.19.7.
- `google.golang.org/grpc` and `google.golang.org/protobuf` support Jarvis, Spider, and QQBot RPCs.
- `github.com/sashabaranov/go-openai` remains only for DALL·E 3 image generation.
- ChatGPT/GPT-4, Bing Copilot, Ernie/Qianfan, DeepSeek, and Chimera text-provider code and provider-specific dependencies have been removed.
- Translation, Tencent COS, MiraiGo, LevelDB, transport, and finance dependencies remain.

## Build, deployment, and verification

- Local build sequence: `go mod download`, then `go build`. The Jarvis change is locally checked through formatting, authoritative-Proto comparison, and Go compilation.
- Real Jarvis RPC and QQ message behavior must be tested after deployment and process startup; do not describe the current local change as live-tested.
- `.github/workflows/docker.yml` runs `Docker Image CI` for non-README-only pushes to `master` and PRs targeting `master`. It builds/pushes `registry.cn-hangzhou.aliyuncs.com/lexmargin/qq-bot:latest` for `linux/amd64`.
- `.github/workflows/update_pod.yml` runs `Update App` after `Docker Image CI` reaches `completed` and invokes `delete-pod` for app `qq-bot`. It does not check `workflow_run.conclusion == 'success'`.
- Therefore, after every `master` push, monitor both workflows through completion, investigate failures, then read-only verify the Kubernetes Deployment rollout, running image, Pod readiness, restart count, and relevant logs. Workflow completion alone is not deployment evidence.
- Before any deployment operation, confirm Kubernetes context/namespace, target Deployment, current image, and rollback point. Do not make remote changes without explicit authorization.
- The actual manifest is `qq_bot.yml`; never silently replace or rename it to `qq-bot.yaml`.

## Ports and runtime addresses

| Container/Service port | NodePort | Intended role |
| --- | --- | --- |
| `8888` | `30088` | Local-manifest bot/OneBot-facing port. |
| `9090` | `30091` | Local-manifest mapping for the inbound QQBot gRPC server. |

The Dockerfile exposes 8888. The NodePorts come from the local untracked manifest and require re-verification. Neither artifact proves that an OneBot transport listens on 8888; listeners come from the mounted `config.yml`. Spider is outbound gRPC through `SPIDER_SERVICE_ADDR`; Jarvis is outbound gRPC through the `jarvis` configuration.

## Verification boundaries and maintenance

- Tests under `gotest/` include network/credential integrations and are not hermetic.
- For the current Jarvis integration, the user explicitly selected deployment-time functional testing instead of local unit tests. Do not claim tests were run when they were not.
- Update this document when modules, runtime flows, commands, state formats, jobs, configuration, dependencies, workflows, image names, ports, or local deployment manifests materially change.
- Keep repository facts separate from dated live observations and never include secrets or QQ identifiers.
