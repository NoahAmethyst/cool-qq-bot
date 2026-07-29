# Codex Project Context Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add an automatically loaded repository rule file and a detailed, security-safe project cognition document for future Codex conversations.

**Architecture:** Keep the automatically loaded `AGENTS.md` short and prescriptive, with one mandatory link to the detailed `docs/CODEX_PROJECT_CONTEXT.md`. Keep confirmed code and deployment facts in the context document, while treating live cluster state and all credentials as external, runtime-only information.

**Tech Stack:** Markdown, Go 1.19 project structure, GitHub Actions, Docker Buildx, Alibaba Cloud Container Registry, Kubernetes.

---

## File Structure

- Create `AGENTS.md`: repository-wide Codex loading, development, verification, security, and remote-operation rules.
- Create `docs/CODEX_PROJECT_CONTEXT.md`: detailed feature map, module map, runtime flow, interfaces, persistence, deployment flow, testing posture, and known verification boundaries.
- Do not modify `.dockerignore`, `.gitignore`, `config.yml`, `crun.sh`, `qq_bot.yml`, or `qrcode.png`; these already contain user-owned working-tree changes.

### Task 1: Add the repository-wide Codex rules

**Files:**

- Create: `AGENTS.md`
- Reference: `docs/superpowers/specs/2026-07-29-codex-project-context-design.md`

- [ ] **Step 1: Confirm the target is not replacing an existing repository file**

Run:

```bash
test ! -e AGENTS.md
```

Expected: exit code `0` and no output. If the file exists, stop and merge its instructions instead of replacing it.

- [ ] **Step 2: Create `AGENTS.md` with the approved rules**

Create exactly this rule structure:

```markdown
# Repository Instructions

## Required project context

- Before working on this repository, read `docs/CODEX_PROJECT_CONTEXT.md`.
- Treat that document as a navigation baseline, not a substitute for reading the source files involved in the current task.
- Update the context document whenever a change materially alters features, module responsibilities, runtime flow, CI/CD, or deployment.

## Development and verification

- This is a Go 1.19 project. Follow the existing package layout and Go style; avoid unrelated refactors.
- Preserve user-owned working-tree changes and never revert unrelated files.
- Add or update tests for changed code.
- Run focused tests first, then the widest practical local test set.
- Tests under `gotest/` may require network access, cloud services, or API credentials; report unavailable prerequisites separately from deterministic unit-test failures.
- Inspect `.github/workflows/` when a change affects build or deployment behavior.

## Deployment environment

- Remote host: `1.15.231.125`
- SSH user: `root`
- Remote runtime-data directory: `/root/qqbot`
- Orchestrator: Kubernetes
- Deployment manifest in this repository: `qq_bot.yml`
- Image registry: Alibaba Cloud Container Registry, Hangzhou region
- Image build workflow: `.github/workflows/docker.yml`
- Post-build update workflow: `.github/workflows/update_pod.yml`
- The repository file is named `qq_bot.yml`; do not silently substitute `qq-bot.yaml`.

## Secrets and sensitive data

- Never write server passwords, QQ credentials, API keys, cookies, cloud credentials, registry credentials, or GitHub Secrets into tracked files.
- Never echo sensitive values in responses, patches, logs, commit messages, or persistent shell history.
- Obtain remote credentials only through an approved secure channel or the currently authorized session.
- Record environment-variable and Secret names only, never their values.
- Treat `config.yml`, `qq_bot.yml`, device/session files, QR codes, and runtime logs as potentially sensitive; redact them before displaying content.

## Remote-operation rules

- Start with read-only checks of Kubernetes context, namespace, workloads, Pods, events, logs, and image state.
- Do not edit remote files, apply manifests, change Secrets, restart workloads, or delete Pods without explicit authorization for that action.
- Prefer the existing GitHub Actions → Alibaba Cloud image → Kubernetes update path for normal releases.
- Do not replace CI with a manual server build unless the user explicitly requests emergency handling.
- Before a deployment change, confirm the Kubernetes context, namespace, target Deployment, current image, and rollback point.
- After an authorized change, verify rollout status, Pod readiness, container logs, and exposed ports.
```

- [ ] **Step 3: Validate the loading link and sensitive-data boundary**

Run:

```bash
test -f docs/CODEX_PROJECT_CONTEXT.md || true
rg -n 'docs/CODEX_PROJECT_CONTEXT\.md|1\.15\.231\.125|/root/qqbot|qq_bot\.yml' AGENTS.md
rg -ni '(password|passwd|token|secret|api[_-]?key)[[:space:]]*[:=][[:space:]]*[^<${`[:space:]][^[:space:]]+' AGENTS.md
```

Expected:

- The first command may be non-zero until Task 2 creates the linked document.
- The second command prints the required context link and non-sensitive deployment metadata.
- The third command prints nothing and exits `1`, showing that no literal credential assignment was introduced.

- [ ] **Step 4: Commit only the rule file**

Run:

```bash
git add AGENTS.md
git diff --cached --name-only
git diff --cached --check
git commit -m "docs: add repository instructions"
```

Expected:

- The staged file list contains only `AGENTS.md`.
- `git diff --cached --check` has no output.
- The commit succeeds without including user-owned files.

### Task 2: Add the detailed project cognition document

**Files:**

- Create: `docs/CODEX_PROJECT_CONTEXT.md`
- Reference: `main.go`
- Reference: `cmd/gocq/main.go`
- Reference: `coolq/bot.go`
- Reference: `coolq/event.go`
- Reference: `coolq/cmd.go`
- Reference: `coolq/chat_assistant.go`
- Reference: `coolq/state.go`
- Reference: `server/http.go`
- Reference: `server/websocket.go`
- Reference: `cluster/server.go`
- Reference: `cluster/rpc/rpc.go`
- Reference: `bot_service/bot_service.go`
- Reference: `db/multidb.go`
- Reference: `.github/workflows/docker.yml`
- Reference: `.github/workflows/update_pod.yml`
- Reference: `Dockerfile`
- Reference: `qq_bot.yml`

- [ ] **Step 1: Create the detailed context with source-backed sections**

Create `docs/CODEX_PROJECT_CONTEXT.md` with all of these sections and facts:

```markdown
# Codex Project Context

## Purpose and scope

This repository is a customized QQ bot built on go-cqhttp and MiraiGo. This document is a fast navigation baseline for Codex. Source code remains authoritative, and live Kubernetes state must be checked before operational conclusions.

## User-facing capabilities

- QQ account login by password, token recovery, or QR code, with reconnect support.
- Group and private-message command handling.
- OneBot-style HTTP and forward/reverse WebSocket communication.
- QQBot gRPC methods for sending messages and reading bot, friend, and group information.
- On-demand and scheduled Weibo, 36Kr, Wall Street CN, Caixin, and Zhihu reports.
- Cryptocurrency price, gold price, and exchange-rate queries.
- Mathematical-expression and Kelly-strategy calculations.
- Translation through initialized Youdao, Tencent, Alibaba Cloud, and Volcano Engine adapters in rotation.
- ChatGPT, GPT-4, Bing Copilot, Baidu Qianfan, and DeepSeek conversations with per-user model selection and short-lived context.
- Local state files, Tencent COS backup, and optional LevelDB message storage.
- AI image-generation implementation exists, but its command is currently commented out of both group and private command lists and is not enabled by default.

## Repository map

| Path | Responsibility |
| --- | --- |
| `main.go`, `shutdown.go` | Process entry, timezone/log initialization, Spider client initialization, daily backup scheduling, signal handling. |
| `cmd/gocq/` | CLI/config startup, login, token recovery, reconnect, friend/group loading, bot and server startup. |
| `coolq/` | Core bot module: sends messages, converts CQCode, handles events and commands, runs AI/translation/calculation/report features, and owns runtime state. |
| `server/` | HTTP, forward WebSocket, reverse WebSocket, serverless, middleware, event delivery, and OneBot action transport adapters. |
| `modules/api/` | Generated OneBot action dispatcher and middleware-aware caller. |
| `modules/servers/` | Registry that starts transport adapters configured in `config.yml`. |
| `modules/config/`, `internal/base/` | YAML parsing, defaults, environment expansion, command flags, and runtime configuration globals. |
| `cluster/`, `bot_service/` | QQBot gRPC server, recovery/logging interceptors, and Spider gRPC client wrappers. |
| `protocol/` | QQBot and Spider protobuf definitions plus generated Go code. |
| `db/` | Message-storage interface and adapters; LevelDB is imported by default, MongoDB is present but disabled in `main.go`. |
| `util/` | AI clients, news sources, finance, translation, cron, HTTP, file, hash, and cloud-storage helpers. |
| `internal/`, `global/` | Cache, BTree, self-diagnosis, self-update, filesystem, logging, networking, signals, and shared values. |
| `.github/workflows/`, `Dockerfile`, `qq_bot.yml` | Container build, image publication, Pod refresh, and Kubernetes resources. |

## Startup and runtime flow

1. `main.main` fixes the process timezone to UTC+8, configures colored logrus output, initializes every registered outbound gRPC client, registers shutdown handling, schedules daily-record upload, and calls `gocq.Main`.
2. `gocq.Main` parses flags and `config.yml`, configures rotating logs, creates media/cache directories, initializes configured database adapters, and restores or performs QQ login.
3. After login it saves the session token, loads friend and group lists, sets online status, creates `coolq.CQBot`, and starts configured HTTP/WebSocket adapters through `servers.Run`.
4. The bot instance is assigned to `bot_service.QQBot`; a background goroutine runs self-update/network diagnosis and starts the QQBot gRPC server on port `9090`.
5. `coolq.NewQQBot` initializes persisted state, subscribes advanced group/private message handlers and friend events, registers report cron jobs, emits startup messages to loaded groups, starts heartbeat delivery, and starts the state-save listener.
6. `groupMessageAdvancedEvent` and `privateMessageAdvancedEvent` independently attempt calculation, AI handling, and `#command` dispatch for each incoming message.
7. Standard OneBot message/event subscriptions remain mostly commented out in `NewQQBot`; do not assume every event method in `coolq/event.go` is currently active.
8. Shutdown notification saves bot state and top-list backups. Process termination behavior must be verified carefully because signal handling and the state listener are asynchronous.

## Command and feature flow

- A text element beginning with `#` is matched by `coolq/cmd.go` and dispatched through separate group/private handler maps.
- `#微博`, `#36`, `#华尔街`, `#财新`, and `#知乎` call reporter methods backed by `util/top_list` and, for several sources, the Spider gRPC service.
- `#比特币`, `#黄金`, and `#汇率` call finance helpers.
- `#翻译` selects the next initialized translation adapter.
- `#模式` reads or changes the sender's AI model.
- AI questions are accepted when the text contains a question mark or the bot is mentioned in a group. OpenAI-compatible and Ernie contexts are retained per sender and expire after an inactivity window.
- Push subscriptions and deduplication state are held by `coolq.State` and persisted as JSON.

## Interfaces and seams

- `modules/servers` is the seam between configured communication transports and `coolq.CQBot`; HTTP and WebSocket are adapters registered during package initialization.
- `modules/api.Caller` is the seam between transport parameter getters/middleware and generated calls to `CQBot` methods.
- `db.Database` is the message-storage seam. Writes fan out to every configured adapter; reads use the first configured adapter.
- `coolq.Assistant`, `Translator`, `Calculator`, and `ImgGenerator` let group/private adapters share feature implementations.
- QQBot gRPC exposes `SendMsg`, `Self`, `Friends`, and `Groups`.
- `cluster/spider_svc` wraps the external Spider gRPC interface for news, Copilot, and finance data.

## Configuration, state, and persistence

- `config.yml` controls account, reconnect, heartbeat, message format, logging, transport adapters, middleware, and database adapters.
- `${NAME}` and `${NAME:default}` expressions in YAML are expanded from environment variables.
- Kubernetes mounts `/root/qqbot/config.yml`, `address.txt`, `device.json`, `session.token`, log/data directories, and a temporary file directory into `/opt` paths.
- Default LevelDB data lives at `data/leveldb-v3`.
- JSON files persist report recipients, sent-news hashes, global state, and selected AI models; the implementation also attempts Tencent COS download/upload.
- AI conversation context is process memory, not durable storage.
- Sensitive runtime inputs include QQ credentials, AI/translation/cloud keys, cookies, registry credentials, and Kubernetes/GitHub Secrets. Never reproduce their values.

## Build and deployment

1. `.github/workflows/docker.yml` runs for pushes to `master` except README-only changes, and for pull requests targeting `master`.
2. It uses QEMU and Docker Buildx to build a `linux/amd64` image from `Dockerfile`.
3. It authenticates with GitHub Secrets and pushes `registry.cn-hangzhou.aliyuncs.com/lexmargin/qq-bot:latest`.
4. `.github/workflows/update_pod.yml` runs when the `Docker Image CI` workflow completes and calls `NoahAmethyst/kuber-opt-action@v1.0.0` with `delete-pod` for `app=qq-bot`.
5. The update workflow does not explicitly test `workflow_run.conclusion == 'success'`; completion and success are not equivalent.
6. `qq_bot.yml` defines one-replica Deployment `qq-bot`, `imagePullPolicy: Always`, Alibaba Cloud pull secret `aliyun-docker-register`, hostPath persistence, and a NodePort Service.

## Ports and runtime addresses

| Purpose | Container/Service port | NodePort |
| --- | ---: | ---: |
| Bot transport declared by the manifest | `8888` | `30088` |
| QQBot gRPC | `9090` | `30091` |

The actual OneBot HTTP/WebSocket listener is controlled by the mounted `config.yml`; the manifest alone does not prove which protocol listens on `8888`. The outbound Spider address comes from `SPIDER_SERVICE_ADDR`.

## Tests and verification boundaries

- Deterministic local tests cover configuration expansion, HTTP parameter handling, CQCode conversion/escaping, version comparison, and the internal BTree.
- Several tests under `gotest/` call AI, news, translation, COS, file-download, or other external services and may require credentials and network access.
- Current GitHub Actions contains image build and Pod-update workflows, but no separate lint/test workflow.
- Before reporting live deployment health, perform read-only checks for Kubernetes context, namespace, Deployment, ReplicaSet, Pods, events, image IDs, and recent logs.

## Maintenance rules

- Update this document when module responsibilities, active event subscriptions, commands, persistence, ports, workflows, image names, or manifests change.
- Keep confirmed repository facts separate from live-environment observations.
- Do not place passwords, tokens, cookies, QQ identifiers, QR images, or API-key values in this document.
```

- [ ] **Step 2: Check every important repository reference**

Run:

```bash
for path in main.go cmd/gocq/main.go coolq/bot.go coolq/cmd.go server/http.go server/websocket.go cluster/server.go bot_service/bot_service.go db/multidb.go .github/workflows/docker.yml .github/workflows/update_pod.yml Dockerfile qq_bot.yml; do test -e "$path" || exit 1; done
```

Expected: exit code `0` and no output.

- [ ] **Step 3: Validate document completeness and secret safety**

Run:

```bash
rg -n '^## (User-facing capabilities|Repository map|Startup and runtime flow|Command and feature flow|Interfaces and seams|Configuration, state, and persistence|Build and deployment|Ports and runtime addresses|Tests and verification boundaries|Maintenance rules)$' docs/CODEX_PROJECT_CONTEXT.md
rg -ni '(password|passwd|token|secret|api[_-]?key)[[:space:]]*[:=][[:space:]]*[^<${`[:space:]][^[:space:]]+' AGENTS.md docs/CODEX_PROJECT_CONTEXT.md
rg -n 'qq-bot\.yaml' AGENTS.md docs/CODEX_PROJECT_CONTEXT.md
git diff --check -- AGENTS.md docs/CODEX_PROJECT_CONTEXT.md
```

Expected:

- The first command prints every required section.
- The password scan prints nothing and exits `1`.
- The incorrect filename scan prints nothing and exits `1`.
- The diff check prints nothing and exits `0`.

- [ ] **Step 4: Commit only the project cognition document**

Run:

```bash
git add docs/CODEX_PROJECT_CONTEXT.md
git diff --cached --name-only
git diff --cached --check
git commit -m "docs: add codex project context"
```

Expected:

- The staged file list contains only `docs/CODEX_PROJECT_CONTEXT.md`.
- The commit succeeds without including user-owned working-tree changes.

### Task 3: Verify the complete documentation contract

**Files:**

- Verify: `AGENTS.md`
- Verify: `docs/CODEX_PROJECT_CONTEXT.md`
- Verify: `docs/superpowers/specs/2026-07-29-codex-project-context-design.md`

- [ ] **Step 1: Verify the loading contract and repository facts**

Run:

```bash
test -f AGENTS.md
test -f docs/CODEX_PROJECT_CONTEXT.md
rg -n 'Before working on this repository, read `docs/CODEX_PROJECT_CONTEXT\.md`' AGENTS.md
rg -n 'registry\.cn-hangzhou\.aliyuncs\.com/lexmargin/qq-bot:latest|30088|30091|SPIDER_SERVICE_ADDR' docs/CODEX_PROJECT_CONTEXT.md
```

Expected: every command exits `0`; searches print the required loading rule and deployment facts.

- [ ] **Step 2: Run deterministic Go tests relevant to documented behavior**

Run:

```bash
go test ./modules/config ./server ./coolq ./global ./internal/btree
```

Expected: exit code `0`. If dependency download is blocked, report that separately and retry only with approved network access. Do not run credential-dependent `gotest` integration tests as if they were deterministic unit tests.

- [ ] **Step 3: Confirm no unrelated user files entered documentation commits**

Run:

```bash
git show --name-only --format= HEAD~1
git show --name-only --format= HEAD
git status --short
```

Expected:

- The two implementation commits contain only `AGENTS.md` and `docs/CODEX_PROJECT_CONTEXT.md`.
- Existing user-owned changes remain visible and uncommitted.

- [ ] **Step 4: Final sensitive-value and Markdown verification**

Run:

```bash
rg -ni '(password|passwd|token|secret|api[_-]?key)[[:space:]]*[:=][[:space:]]*[^<${`[:space:]][^[:space:]]+' AGENTS.md docs/CODEX_PROJECT_CONTEXT.md docs/superpowers/specs/2026-07-29-codex-project-context-design.md
git diff --check HEAD~2..HEAD
```

Expected:

- The sensitive-value scan prints nothing and exits `1`.
- The documentation diff check prints nothing and exits `0`.
