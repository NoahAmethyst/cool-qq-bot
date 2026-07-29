# Codex Project Context

## Purpose and scope

This repository is a customized QQ bot built on the go-cqhttp codebase and MiraiGo client. This document is a fast navigation baseline for Codex: it points to the modules, interfaces, runtime flow, and deployment artifacts that matter most. Source code remains authoritative. Repository configuration describes intended deployment, but any statement about the current Kubernetes context, workload health, image, or logs must be verified with read-only checks against the live cluster.

## User-facing capabilities

- Login supports a configured password, a cached session token, or a QR code, followed by token-based and normal-login reconnect paths.
- Group and private messages share a `#`-command experience. The bot also exposes OneBot-style HTTP calls and event posts, forward WebSocket and reverse WebSocket transports, and a QQBot gRPC interface with `SendMsg`, `Self`, `Friends`, and `Groups`.
- On-demand news covers Weibo, 36Kr, Wall Street CN, Caixin, and Zhihu. Scheduled reporters currently cover Weibo, 36Kr, Wall Street CN, and Caixin at different cron frequencies, plus a daily summary; Zhihu is on-demand in the current job registration. Users can enable or disable scheduled delivery for a group or private chat.
- Finance and calculation features include cryptocurrency prices, gold prices and fluctuation reports, exchange rates, mathematical expressions, and the Kelly strategy.
- Translation rotates across the translation adapters that initialized successfully.
- Assistant modes include ChatGPT, GPT-4, Bing Copilot, Baidu Qianfan/Ernie, and DeepSeek. Model selection is stored per user, while short-lived conversation context is kept in memory per user and separately for group and private chats.
- State is stored in local JSON, synchronized to Tencent COS where implemented, and optionally in LevelDB for message history and reply/recall context.
- AI image generation has a complete shared implementation and group/private adapters, but the image command is commented out of both current command lists, so it is not enabled by default.

## Repository map

| Path | Responsibility |
| --- | --- |
| `main.go`, `shutdown.go` | Process entry point, timezone and logging setup, outbound Spider gRPC client initialization, daily-record upload scheduling, and shutdown notification wiring. |
| `cmd/gocq/` | CLI flags, working-directory/config initialization, device and login flows, reconnect, list synchronization, bot construction, transport startup, and process lifetime. |
| `coolq/` | Core `CQBot` implementation: QQ events, message conversion and sending, commands, assistants, reporters, jobs, state, CQ codes, and generated OneBot-call targets. |
| `server/` | HTTP, reverse HTTP reporting, forward WebSocket, reverse WebSocket, middleware, daemon, and serverless transport adapters. |
| `modules/api/` | `Getter`/middleware call routing and generated dispatch from OneBot action names to `CQBot` calls. |
| `modules/servers/` | Registry and runner that connects configured transports to a `CQBot`. |
| `modules/config/`, `internal/base/` | YAML generation/parsing and environment expansion; CLI flags and normalized runtime configuration. |
| `cluster/`, `bot_service/` | QQBot gRPC server, middleware, outbound client connections and Spider wrapper, plus the QQBot gRPC method implementation. |
| `protocol/` | QQBot and Spider protobuf definitions and generated Go bindings. |
| `db/` | Persistence interfaces, multi-backend coordination, LevelDB adapter, and an available but not default-registered MongoDB implementation. |
| `util/` | Cron, news sources and daily records, finance, AI clients, translation rotation, math, file/COS, HTTP, encryption, and content helpers. |
| `internal/`, `global/` | Internal cache, B-tree, update/diagnosis, flags/version and shared filesystem, network, logging, signal, buffer, parameter, and codec helpers. |
| `.github/workflows/`, `Dockerfile`, `qq_bot.yml` | Container build/push automation, pod-refresh automation, runtime image definition, and Kubernetes Deployment/Service declaration. |

## Startup and runtime flow

1. `main.go` fixes the local timezone at UTC+8, installs the log formatter, initializes every registered outbound gRPC client (currently Spider), installs the shutdown notifier, schedules daily-record upload at 00:05, and calls `gocq.Main`.
2. `gocq.Main` parses CLI flags, handles help/daemon/working-directory modes, loads `config.yml`, configures rotating logs, creates media/cache directories, initializes caches, selects configured database adapters from the registered drivers, and opens them.
3. It prepares device information and chooses cached-token, password, or QR login. Reconnect first retries the saved token and can fall back to normal password login where applicable.
4. After login it saves `session.token`, reloads friends and groups, applies the configured presence, constructs `coolq.NewQQBot`, and starts configured transports through `modules/servers.Run`.
5. It assigns the constructed bot to `bot_service.QQBot`. A background goroutine then checks for updates, runs network diagnosis, and starts the QQBot gRPC server on `:9090`; the main path waits on the shared signal handler.
6. `NewQQBot` initializes persisted and in-memory state, subscribes the active QQ events, registers reporter jobs, announces startup in every loaded group, starts heartbeat dispatch, and starts a shutdown listener.
7. Active advanced group and private message subscriptions each try the calculator, assistant, and `#`-command flows for the same incoming message. A message can therefore match more than one flow. Most standard OneBot message/notice/request event subscriptions in `NewQQBot` are currently commented out; only the explicitly active subscriptions should be assumed.
8. Shutdown handling is asynchronous: signal listeners notify the bot, which saves subscription/global/sent-news state and backs up news records. Model choices are persisted when changed. Be cautious when changing channel initialization, process exit, or persistence ordering, because delivery and cleanup run in goroutines.

## Command and feature flow

- `coolq/cmd.go` maintains separate group and private command lists and handler maps. A bare `#` prints the relevant list; a recognized leading `#command` dispatches to that chat type's handler.
- News commands fetch Weibo, 36Kr, Wall Street CN, Caixin, or Zhihu. Finance commands fetch coin, gold, or exchange-rate data, and the Kelly handler performs allocation math.
- `#翻译` parses the remaining text, chooses Chinese-to-English or other-language-to-Chinese, and uses the rotating translation adapter set.
- `#模式` reads or changes the sender's assistant model.
- Assistant requests trigger on `?`, `？`, or a group mention of the bot. The group/private adapters select context by sender, keep group and private sessions separate, and expire supported in-memory context after a short idle window.
- `coolq.State` owns scheduled-push recipients, sent-news hashes used for deduplication, global values, per-user model choices, and in-memory assistant sessions. The persisted portions use JSON files with COS upload/download fallback.

## Interfaces and seams

The following terms use the codebase-design vocabulary: a **module** presents an **interface** at a **seam**, and an **adapter** supplies an implementation at that seam.

- `modules/servers` is the seam between configured transport modules and `coolq.CQBot`. Registration keeps transport selection out of the bot; `server` supplies HTTP/reverse-HTTP, forward-WebSocket, and reverse-WebSocket adapters.
- `modules/api.Caller` is the seam from a transport's `Getter` plus middleware chain to generated `CQBot` calls. Rate limiting and long polling are adapters at that middleware seam.
- `db.Database` is the persistence seam. Enabled adapters all receive writes, while reads intentionally use the first configured adapter. LevelDB is registered by the blank import in `main.go` but is still enabled or disabled by configuration. MongoDB code exists, while its `main.go` blank import remains commented out.
- `Assistant`, `Translator`, `Calculator`, and `ImgGenerator` are shared feature interfaces. Their group and private adapters isolate chat-specific extraction and reply behavior while reusing one implementation, providing leverage and locality.
- `bot_service.BotService` implements the QQBot gRPC methods `SendMsg`, `Self`, `Friends`, and `Groups` over the active `CQBot`.
- `cluster/spider_svc` is a wrapper module over the generated Spider client. It is the call seam used by news, finance, and Bing Copilot features.

## Configuration, state, and persistence

`config.yml` contains account and reconnect behavior, heartbeat, message formatting and handling, output/logging, default transport middleware, database settings, and the configured HTTP/WebSocket transport list. `modules/config` expands both `${NAME}` and `${NAME:default}` before YAML decoding; the default is used only when the named environment variable is empty.

The Kubernetes manifest declares these host-path mounts:

| Host path | Container path | Purpose |
| --- | --- | --- |
| `/root/qqbot/file` | `/tmp` | Generated/downloaded files and JSON state when `FILE_ROOT` selects that location. |
| `/root/qqbot/config.yml` | `/opt/config.yml` | Runtime configuration. |
| `/root/qqbot/address.txt` | `/opt/address.txt` | Login/address state. |
| `/root/qqbot/device.json` | `/opt/device.json` | MiraiGo device identity. |
| `/root/qqbot/session.token` | `/opt/session.token` | Login session cache. |
| `/root/qqbot/log` | `/opt/log` | Declared log mount. |
| `/root/qqbot/data` | `/opt/data` | Internal data, including the LevelDB store at `data/leveldb-v3` when that adapter is enabled. |

Local JSON state includes group/private report recipients, sent-news deduplication, global values, assistant-model choices, and per-source daily records. Implemented save/load paths synchronize selected JSON and records with COS. Assistant conversation context is memory-only and expires; it is not durable state.

Sensitive inputs belong in runtime configuration or secret stores, never in this file. Categories include QQ passwords, session tokens and QR codes; OneBot access tokens; AI and translation provider keys/secrets; Bing cookies; COS/cloud credentials; container-registry and cluster credentials; and QQ/owner/account identifiers.

## Build and deployment

- `.github/workflows/docker.yml` runs `Docker Image CI` for pushes to `master` except README-only changes, and for pull requests targeting `master`. It sets up QEMU and Buildx, logs in with repository Secrets, builds `linux/amd64`, and pushes `registry.cn-hangzhou.aliyuncs.com/lexmargin/qq-bot:latest`.
- `.github/workflows/update_pod.yml` runs after the `Docker Image CI` workflow reaches the `completed` event and calls `NoahAmethyst/kuber-opt-action@v1.0.0` with `delete-pod` for app `qq-bot`. It does **not** check `workflow_run.conclusion == 'success'`, so a completed failed build can still trigger pod deletion.
- `Dockerfile` builds the Go binary in an Alpine-based multi-stage image, exposes 8888, and runs `/opt/go-cqhttp`.
- `qq_bot.yml` declares a one-replica `qq-bot` Deployment using `imagePullPolicy: Always`, the `aliyun-docker-register` pull secret, the host-path mounts above, and a NodePort Service. It also declares the outbound `SPIDER_SERVICE_ADDR`; the RPC code constructs the same name from `SPIDER` plus `_SERVICE_ADDR`.

## Ports and runtime addresses

| Container/Service port | NodePort | Intended role |
| --- | --- | --- |
| `8888` | `30088` | Manifest-declared bot/OneBot-facing port. |
| `9090` | `30091` | QQBot gRPC server started by `cluster.StartServer`. |

The manifest and Dockerfile expose 8888, but they do not prove that an OneBot HTTP or WebSocket adapter listens there. Actual OneBot listeners and reverse targets are determined by the mounted `config.yml` and must be checked at runtime. Spider is outbound gRPC; its address comes from `SPIDER_SERVICE_ADDR`.

## Tests and verification boundaries

Deterministic local tests cover configuration expansion (`modules/config`), HTTP parameter precedence and aliases (`server`), CQCode conversion/escaping (`coolq`), version comparison (`global`), and the file-backed B-tree (`internal/btree`). Prefer those focused packages when validating local changes.

The `gotest` package contains integration-style checks for AI providers, translation, COS, downloads, feeds, finance, and environment-dependent behavior; it is not a hermetic unit-test suite. At this baseline, `go test ./...` also has unrelated known failures from an absent coin helper import and vet format issues in `cluster`/`coolq`. The GitHub Actions workflows build and deploy the image but do not define an independent lint or test job.

Before reporting live health, perform read-only checks for the current kube context and namespace, Deployment, ReplicaSet, Pods, events, running image IDs, and relevant logs. A clean repository manifest or local test run is not evidence that the live workload is healthy.

## Maintenance rules

- Update this document when code ownership, active event subscriptions, commands, scheduled jobs, state formats, persistence adapters, ports, workflows, image names, or Kubernetes manifests change.
- Keep repository facts separate from dated live observations. Re-verify the cluster before making any operational claim.
- Never add passwords, tokens, cookies, QQ identifiers, QR codes, API key values, cloud credential values, or copied secret payloads.
- Preserve exact artifact names, especially `qq_bot.yml`, and keep listener claims tied to `config.yml` rather than inferring them from exposed ports alone.
