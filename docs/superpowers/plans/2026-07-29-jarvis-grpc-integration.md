# Jarvis gRPC Integration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace every text-LLM path in go-cqhttp with Jarvis gRPC, ingest Wall Street CN news into Jarvis, generate cross-day chaptered summaries, and publish the verified implementation through the existing GitHub Actions and Kubernetes flow.

**Architecture:** Generate a Go client from the authoritative Jarvis Proto and wrap it in a focused `cluster/jarvis_svc` adapter with separate connection, chat, ingest, and summary deadlines. Keep OpenAI only for image generation, let Jarvis own conversation history, and use a dedicated Wall Street knowledge identity for ingestion and daily summaries. Preserve existing user-owned working-tree changes and selectively stage only the Jarvis hunk in `config.yml`.

**Tech Stack:** Go 1.19, protobuf, gRPC, YAML, QQ/MiraiGo messaging, GitHub Actions, Alibaba Cloud Container Registry, Kubernetes.

**Verification constraint:** Per the user's explicit instruction, do not add or run unit tests. Use formatting, Proto generation checks, local compilation, GitHub Actions monitoring, and remote startup/integration checks.

---

## File Structure

- Create `protocol/jarvis.proto`: build-time copy of the authoritative Jarvis Proto.
- Create `protocol/pb/jarvis_pb/jarvis.pb.go`: generated Jarvis messages.
- Create `protocol/pb/jarvis_pb/jarvis_grpc.pb.go`: generated Jarvis client/server bindings.
- Create `cluster/jarvis_svc/jarvis_svc.go`: Jarvis connection lifecycle and typed `Chat`, `Summary`, and `Ingest` calls.
- Create `coolq/jarvis_summary.go`: trusted URL filtering, news prompt construction, chapter parsing, and chapter delivery helpers.
- Modify `modules/config/config.go`: Jarvis YAML structure and defaults.
- Modify `modules/config/default_config.yml`: generated-config Jarvis defaults.
- Modify `config.yml`: local Jarvis block while preserving existing account/log changes.
- Modify `internal/base/flag.go`: expose normalized Jarvis runtime configuration.
- Modify `cmd/gocq/main.go`: initialize and close the Jarvis gRPC client.
- Modify `coolq/chat_assistant.go`: single Jarvis assistant, stable user IDs, and read-only mode command.
- Modify `coolq/cmd.go`: Jarvis-only command help.
- Modify `coolq/state.go`: remove model selection and local assistant context lifecycle.
- Modify `coolq/reporter.go`: ingest newly collected Wall Street news and invoke chaptered Jarvis summary.
- Modify `util/top_list/wallstreetcn_news.go`: expose newly collected news and collection timestamps.
- Modify `util/top_list/daily_record.go`: safely copy Wall Street records for summary construction.
- Modify `util/ai_util/ai_cli.go`: retain only the OpenAI image client.
- Modify `coolq/env.go`: retain runtime OpenAI image-key refresh and remove Chimera-specific refresh.
- Modify `cluster/spider_svc/spider_svc.go`: remove the obsolete Bing Copilot wrapper while preserving news and finance RPC wrappers.
- Delete `util/ai_util/models.go`, `chatgpt.go`, `deepseek.go`, `ernie.go`, and `bing_chat.go`.
- Modify `gotest/ai_test.go`: remove obsolete provider integration checks; retain image-generation checks without running them.
- Modify `constant/preference.go`: remove unused text-provider constants.
- Modify `go.mod` and `go.sum`: remove unused Bing/Ernie/Qianfan dependencies with `go mod tidy`.
- Modify `README.md`: document the implemented Jarvis architecture, commands, config, and remaining image/translation environment variables.
- Modify `docs/CODEX_PROJECT_CONTEXT.md`: synchronize features, runtime flow, state, dependencies, scheduled summary, and deployment verification.

### Task 1: Add the authoritative Jarvis protocol and configuration

**Files:**

- Create: `protocol/jarvis.proto`
- Create: `protocol/pb/jarvis_pb/jarvis.pb.go`
- Create: `protocol/pb/jarvis_pb/jarvis_grpc.pb.go`
- Modify: `modules/config/config.go`
- Modify: `modules/config/default_config.yml`
- Modify: `config.yml`
- Modify: `internal/base/flag.go`

- [ ] **Step 1: Add and compare the authoritative Proto**

Use `apply_patch` to create `protocol/jarvis.proto` with the exact authoritative content:

```protobuf
syntax = "proto3";

package jarvis;

service JarvisService {
  rpc Chat(ChatRequest) returns (ChatResponse);
  rpc Ingest(IngestRequest) returns (IngestResponse);
  rpc GetMemory(MemoryRequest) returns (MemoryResponse);
  rpc DeleteMemory(MemoryRequest) returns (DeleteResponse);
}

message ChatRequest {
  string message = 1;
  string user_id = 2;
  string llm = 3;
  string reflect_llm = 4;
}

message ChatResponse {
  string answer = 1;
  bool low_confidence = 2;
}

message IngestRequest {
  string content = 1;
  string source_url = 2;
  string user_id = 3;
}

message IngestResponse {
  bool success = 1;
}

message MemoryRequest {
  string user_id = 1;
}

message MemoryResponse {
  repeated ConversationEntry entries = 1;
}

message ConversationEntry {
  string role = 1;
  string content = 2;
  string created_at = 3;
}

message DeleteResponse {
  bool success = 1;
}
```

Then run:

```bash
cmp /Users/amethyst/Jarvis/jarvis/api/grpc/jarvis.proto protocol/jarvis.proto
```

Expected: `cmp` exits `0` with no output.

- [ ] **Step 2: Generate native Go plugins in a temporary directory**

Run:

```bash
mkdir -p /private/tmp/jarvis-protoc-bin
GOBIN=/private/tmp/jarvis-protoc-bin go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.31.0
GOBIN=/private/tmp/jarvis-protoc-bin go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0
```

Expected: both commands exit `0`; no repository files change.

- [ ] **Step 3: Generate Go bindings through Jarvis's working Python protoc**

Run:

```bash
mkdir -p protocol/pb/jarvis_pb
PATH="/private/tmp/jarvis-protoc-bin:$PATH" \
  /Users/amethyst/Jarvis/.venv/bin/python-grpc-tools-protoc \
  -I protocol \
  --go_out=protocol/pb/jarvis_pb \
  --go_opt=paths=source_relative \
  --go_opt=Mjarvis.proto=github.com/NoahAmethyst/go-cqhttp/protocol/pb/jarvis_pb \
  --go-grpc_out=protocol/pb/jarvis_pb \
  --go-grpc_opt=paths=source_relative \
  --go-grpc_opt=Mjarvis.proto=github.com/NoahAmethyst/go-cqhttp/protocol/pb/jarvis_pb \
  protocol/jarvis.proto
```

Expected: generated files declare package `jarvis_pb`, contain `JarvisServiceClient`, `ChatRequest`, and `IngestRequest`, and carry generated-code headers.

- [ ] **Step 4: Add the typed Jarvis YAML configuration**

Add to `modules/config/config.go`:

```go
type Jarvis struct {
	GRPCTarget             string `yaml:"grpc-target"`
	ConnectTimeoutSeconds  int    `yaml:"connect-timeout-seconds"`
	ChatTimeoutSeconds     int    `yaml:"chat-timeout-seconds"`
	IngestTimeoutSeconds   int    `yaml:"ingest-timeout-seconds"`
	SummaryTimeoutSeconds  int    `yaml:"summary-timeout-seconds"`
	WallStreetUserID       string `yaml:"wallstreet-user-id"`
}

func DefaultJarvis() Jarvis {
	return Jarvis{
		GRPCTarget:            "jarvis:9090",
		ConnectTimeoutSeconds: 10,
		ChatTimeoutSeconds:    300,
		IngestTimeoutSeconds:  60,
		SummaryTimeoutSeconds: 600,
		WallStreetUserID:      "go-cqhttp:wallstreet",
	}
}
```

Add `Jarvis Jarvis` to `Config`. In `Parse`, initialize `config := &Config{Jarvis: DefaultJarvis()}` before YAML decoding so an older mounted remote `config.yml` remains deployable, then validate that the target, user ID, and four timeouts are non-empty/positive.

- [ ] **Step 5: Add the YAML blocks without replacing user values**

Add the following top-level block to both YAML files:

```yaml
jarvis:
  grpc-target: ${JARVIS_GRPC_TARGET:jarvis:9090}
  connect-timeout-seconds: 10
  chat-timeout-seconds: 300
  ingest-timeout-seconds: 60
  summary-timeout-seconds: 600
  wallstreet-user-id: go-cqhttp:wallstreet
```

Do not alter the existing local account or log-level values in `config.yml`.

- [ ] **Step 6: Expose normalized runtime config**

Add `Jarvis config.Jarvis` to `internal/base`, assign `Jarvis = conf.Jarvis` in `Init`, and keep all time conversion inside the Jarvis adapter.

### Task 2: Implement the Jarvis gRPC client lifecycle

**Files:**

- Create: `cluster/jarvis_svc/jarvis_svc.go`
- Modify: `cmd/gocq/main.go`

- [ ] **Step 1: Implement a single reusable connection**

Create a package-level client guarded by a mutex. `Init` must:

```go
ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.ConnectTimeoutSeconds)*time.Second)
defer cancel()

conn, err := grpc.DialContext(
	ctx,
	cfg.GRPCTarget,
	grpc.WithTransportCredentials(insecure.NewCredentials()),
	grpc.WithBlock(),
)
```

On success, store the connection, generated client, ordinary chat timeout, ingest timeout, summary timeout, and Wall Street user ID. Reject empty or non-positive values before dialing.

- [ ] **Step 2: Implement typed operations**

Expose:

```go
func Chat(message, userID string) (*jarvis_pb.ChatResponse, error)
func Summary(message string) (*jarvis_pb.ChatResponse, error)
func Ingest(content, sourceURL string) (*jarvis_pb.IngestResponse, error)
func WallStreetUserID() string
func Close() error
```

`Chat` uses `chat-timeout-seconds`; `Summary` uses `summary-timeout-seconds`; `Ingest` uses `ingest-timeout-seconds` and the configured Wall Street identity. None retries. Return errors unchanged so callers can classify them with `status.Code`.

- [ ] **Step 3: Initialize before QQ login and close on exit**

Immediately after `base.Init()` in `gocq.Main`, call:

```go
if err := jarvis_svc.Init(base.Jarvis); err != nil {
	log.Fatalf("Jarvis gRPC 初始化失败: %v", err)
}
defer func() {
	if err := jarvis_svc.Close(); err != nil {
		log.Warnf("关闭 Jarvis gRPC 连接失败: %v", err)
	}
}()
```

The adapter must log only the target and successful connection state, never user content or service error details.

### Task 3: Replace assistant modes and local context with Jarvis

**Files:**

- Modify: `coolq/chat_assistant.go`
- Modify: `coolq/cmd.go`
- Modify: `coolq/state.go`

- [ ] **Step 1: Reduce the Assistant interface**

Remove `Model`, `ChangeModel`, and `Session`. Add:

```go
JarvisUserID() string
```

Implement IDs as:

```go
func (p *PrivateAssistant) JarvisUserID() string {
	return fmt.Sprintf("qq:private:%d", p.Sender())
}

func (p *GroupAssistant) JarvisUserID() string {
	return fmt.Sprintf("qq:group:%d:%d", p.Chat(), p.Sender())
}
```

- [ ] **Step 2: Make `#模式` read-only**

Replace numeric parsing and persistence:

```go
func ShowAssistantMode(assistant Assistant) {
	value := strings.TrimSpace(strings.ReplaceAll(assistant.GetText().Content, "#模式", ""))
	if value == "" {
		assistant.Reply("当前助手模式：Jarvis")
		return
	}
	assistant.Reply("模型已由 Jarvis 服务统一管理，不支持在机器人端切换")
}
```

Update group/private handlers to call `ShowAssistantMode`. Change command help to “查看当前助手模式；模型由 Jarvis 服务统一管理”.

- [ ] **Step 3: Replace provider dispatch with one Jarvis call**

Keep the existing question/mention trigger. After 10 seconds, send “Jarvis 正在响应，请稍后...”. Call:

```go
resp, err := jarvis_svc.Chat(textEle.Content, assistant.JarvisUserID())
```

On error, log `status.Code(err)` plus the non-sensitive user ID and reply “调用 Jarvis 失败，请稍后重试”. On an empty answer, reply “Jarvis 返回空回答，请稍后重试”. Otherwise remove `*` as the current UX does and reply with the answer. Do not log the answer or prompt.

Correct `PrivateAssistant.Reply` so a long answer sends `_msg` for the current fragment and uses `len(_splittedMsg) > 1` to decide whether pagination labels are needed; do not resend the full answer for every fragment.

- [ ] **Step 4: Remove local assistant state**

Delete:

- `assistantModel`;
- `AiAssistantSession`;
- model JSON/COS load and save paths;
- group/private dialogue session fields and initialization;
- Bing conversation objects;
- OpenAI and Ernie message context arrays;
- activity channels and timeout goroutines.

Keep report recipients, sent-news cache, global state, and owner handling unchanged.

### Task 4: Ingest newly collected Wall Street news

**Files:**

- Modify: `util/top_list/wallstreetcn_news.go`
- Modify: `util/top_list/daily_record.go`
- Modify: `coolq/reporter.go`
- Create: `coolq/jarvis_summary.go`

- [ ] **Step 1: Preserve collection timestamps and expose new items**

Extend:

```go
type WallStreetNews struct {
	Title       string
	Content     string
	Url         string
	CollectedAt string
}
```

Create `LoadWallStreetNewsWithUpdates() (all []WallStreetNews, updates []WallStreetNews, err error)`. Use one `time.Now()` value and RFC3339 timestamp for every item in that fetch. Keep `LoadWallStreetNews()` as a compatibility wrapper for integration-style callers.

- [ ] **Step 2: Correct Wall Street record locking**

Change `wallStreetNewsDailyRecord.GetData` to hold `RLock` through the copy:

```go
d.RLock()
defer d.RUnlock()
```

Do not refactor unrelated record types.

- [ ] **Step 3: Add trusted URL filtering**

In `coolq/jarvis_summary.go`, accept only URLs where:

```go
u.Scheme == "https"
u.User == nil
host == "wallstreetcn.com" || strings.HasSuffix(host, ".wallstreetcn.com")
```

Normalize the hostname to lowercase. Invalid/untrusted URLs may still be shown by the original news push, but must not be sent to Jarvis for scraping.

- [ ] **Step 4: Ingest only updates**

After the reporter fetches `updates`, deduplicate by URL and call:

```go
content := fmt.Sprintf(
	"新闻时间：%s\n标题：%s\n摘要：%s",
	news.CollectedAt,
	news.Title,
	news.Content,
)
_, err := jarvis_svc.Ingest(content, news.Url)
```

Continue on failure. Log only the trusted URL and `status.Code(err)`. Preserve original push behavior.

### Task 5: Generate and deliver a cross-day chaptered summary

**Files:**

- Modify: `coolq/reporter.go`
- Modify: `coolq/jarvis_summary.go`

- [ ] **Step 1: Build a deduplicated summary prompt**

From `WallStreetNewsDailyRecord.GetData`, deduplicate by URL and sort by collection time. Include timestamp, title, summary, and trusted URL. The instruction must require historical knowledge retrieval, selective URL scraping, explicit dates, fact/analysis separation, and these exact markers:

```text
<<<CHAPTER_1>>>
<<<CHAPTER_2>>>
<<<CHAPTER_3>>>
<<<CHAPTER_4>>>
```

Map them to:

1. 今日全球经济形势概览
2. 重点事件、原文补充与影响分析
3. 跨日时间线及因果传导
4. 未来展望、风险与机会

- [ ] **Step 2: Call the summary deadline**

Replace `AskDeepSeek` with:

```go
resp, err := jarvis_svc.Summary(prompt)
```

Do not retry. On failure, send “华尔街每日资讯总结失败，请稍后查看” without exposing gRPC details.

- [ ] **Step 3: Parse all four chapters**

Scan the answer between exact markers. Return a structure error unless every marker appears once, in order, with non-empty content. Ignore leading text before the first marker rather than sending it as an unstructured page.

- [ ] **Step 4: Send chapters, not the whole document**

For each chapter, prepend the fixed chapter title. If a chapter exceeds 1000 runes, use `content_util.SplitLongMessage` and send:

```text
【第三章 2/3】
<chapter fragment>
```

Otherwise send:

```text
第三章：跨日时间线及因果传导
<chapter content>
```

If `resp.LowConfidence` is true, send “提示：Jarvis 将本次总结标记为低置信度，请结合原始新闻核验。” before chapter one.

### Task 6: Remove obsolete text-provider clients and dependencies

**Files:**

- Modify: `util/ai_util/ai_cli.go`
- Modify: `coolq/env.go`
- Modify: `constant/preference.go`
- Modify: `cluster/spider_svc/spider_svc.go`
- Delete: `util/ai_util/models.go`
- Delete: `util/ai_util/chatgpt.go`
- Delete: `util/ai_util/deepseek.go`
- Delete: `util/ai_util/ernie.go`
- Delete: `util/ai_util/bing_chat.go`
- Modify: `gotest/ai_test.go`
- Modify: `go.mod`
- Modify: `go.sum`

- [ ] **Step 1: Keep only the OpenAI image client**

Simplify `ai_cli.go` to initialize `openaiCli` from `OPENAI_API_KEY` and the existing proxy settings. Keep `SetOpenaiKey` so the owner-only `#ENV` command can refresh image credentials. Remove model listing, Chimera, Ernie, and DeepSeek clients and attributes.

- [ ] **Step 2: Remove obsolete constants and handlers**

Remove direct text-provider constants that no production path uses. Remove Chimera special handling from `envSetHandler`; arbitrary environment variables may still use the default branch. Keep `OPENAI_API_KEY`, `OPENAI_PROXY`, and `NOT_MIRROR` for image generation.

Remove `cluster/spider_svc.AskBingCopilot`; preserve all news and finance RPC wrappers in that file.

- [ ] **Step 3: Remove obsolete integration-style test functions**

From `gotest/ai_test.go`, remove provider model listing, ChatGPT, GPT-4, DeepSeek, Bing, and Ernie functions and imports. Retain only the two existing OpenAI image functions. Do not add or run tests.

- [ ] **Step 4: Tidy dependencies**

Run:

```bash
go mod tidy
```

Expected: Bing Chat, go-ernie, Qianfan, and unused transitive provider dependencies disappear; `go-openai` remains for images; gRPC/protobuf remain for Jarvis and existing services.

### Task 7: Synchronize README and project cognition

**Files:**

- Modify: `README.md`
- Modify: `docs/CODEX_PROJECT_CONTEXT.md`

- [ ] **Step 1: Rewrite README as implemented behavior**

Document:

- Jarvis as the only text assistant;
- `#模式` read-only behavior;
- private/group memory isolation;
- Jarvis-owned conversation history;
- Wall Street real-time knowledge ingestion;
- 23:56 cross-day, URL-enriched, four-chapter summary;
- preserved OpenAI image generation;
- the `jarvis` YAML block with all four deadlines;
- `JARVIS_GRPC_TARGET` override;
- cluster default `jarvis:9090`;
- remaining environment-variable categories without secret values;
- build and Kubernetes deployment through GitHub Actions.

Remove deprecated claims about ChatGPT/Bing/Ernie/DeepSeek switching, ten-minute local context, Chimera support, and obsolete provider credentials.

- [ ] **Step 2: Update project context**

Replace the five-mode assistant and local lifecycle descriptions. Add the Jarvis Proto package, `cluster/jarvis_svc`, stable user IDs, knowledge ingestion, summary chapters, gRPC target, config defaults, and the requirement to monitor both GitHub Actions workflows after a `master` push.

### Task 8: Format, compile, review, and create the implementation commit

**Files:**

- All implementation files above.

- [ ] **Step 1: Format Go code**

Run:

```bash
gofmt -w cluster/jarvis_svc/jarvis_svc.go cluster/spider_svc/spider_svc.go coolq/chat_assistant.go coolq/jarvis_summary.go coolq/reporter.go coolq/state.go coolq/cmd.go coolq/env.go util/top_list/wallstreetcn_news.go util/top_list/daily_record.go util/ai_util/ai_cli.go modules/config/config.go internal/base/flag.go cmd/gocq/main.go gotest/ai_test.go constant/preference.go
```

Expected: exit `0`.

- [ ] **Step 2: Verify Proto source and generated bindings**

Run:

```bash
cmp /Users/amethyst/Jarvis/jarvis/api/grpc/jarvis.proto protocol/jarvis.proto
rg -n 'type JarvisServiceClient|type ChatRequest|type IngestRequest' protocol/pb/jarvis_pb
```

Expected: Proto comparison exits `0`; all generated types are found.

- [ ] **Step 3: Compile without running tests**

Run:

```bash
go build -o /private/tmp/go-cqhttp-jarvis .
```

Expected: exit `0` and a binary at `/private/tmp/go-cqhttp-jarvis`.

- [ ] **Step 4: Perform static repository checks**

Run:

```bash
git diff --check
rg -n 'AskChatGpt|AskChatGpt4|AskDeepSeek|AskErnie|AskBingChat|AskBingCopilot|ChatGPT|BingCopilot|assistantModel|AiAssistantSession' --glob '*.go' coolq util/ai_util cluster/spider_svc
rg -n 'OPENAI_API_KEY\\s*[:=]\\s*[^[:space:]]|DEEPSEEK_API_KEY\\s*[:=]\\s*[^[:space:]]' README.md docs AGENTS.md
```

Expected: `git diff --check` exits `0`; the obsolete production-symbol search has no hits; documentation contains no credential values.

- [ ] **Step 5: Review the exact implementation diff**

Use:

```bash
git diff --stat
git diff --name-status
git status --short
```

Confirm unrelated `.dockerignore`, `.gitignore`, `crun.sh`, `qq_bot.yml`, and `qrcode.png` changes remain unstaged and unchanged.

- [ ] **Step 6: Stage only the Jarvis hunk from the dirty `config.yml`**

Stage all clean implementation paths normally. For `config.yml`, interactively select only the new top-level `jarvis` block:

```bash
git diff -- config.yml
git add -p config.yml
```

Answer `n` for the existing account and log-level hunks and `y` only for the Jarvis configuration hunk.

Then verify:

```bash
git diff --cached --name-status
git diff --cached --check
git diff --cached -- config.yml
```

Expected: the cached `config.yml` diff contains the Jarvis block but not the user's account or log-level changes.

- [ ] **Step 7: Create one implementation-and-documentation commit**

Run:

```bash
git commit -m "FEATURE -- Integrate Jarvis gRPC assistant"
```

Expected: one commit containing implementation, generated code, config defaults, README, and project context.

### Task 9: Push, monitor workflows, and verify the remote rollout

**Files:**

- No new local files.

- [ ] **Step 1: Capture the live rollback point before pushing**

Over existing key-based SSH, run fixed read-only commands for:

```bash
kubectl config current-context
kubectl get deployment qq-bot -n default -o wide
kubectl get pods -n default -l app=qq-bot -o wide
kubectl get deployment qq-bot -n default -o jsonpath='{.spec.template.spec.containers[0].image}'
```

Record the current Pod name, image tag, image ID, readiness, and restart count without reading Secrets or full environment values.

- [ ] **Step 2: Push `master`**

Run:

```bash
git push github master
```

Expected: push succeeds and includes the implementation commit.

- [ ] **Step 3: Monitor Docker Image CI**

Identify the workflow run for the pushed commit and watch it through completion. Confirm the Build and push step succeeds. If it fails, inspect build logs, fix locally, create a new commit, and repeat; do not claim deployment success.

- [ ] **Step 4: Monitor Update App**

Because `update_pod.yml` triggers on every completed Docker workflow regardless of conclusion, independently verify the update workflow and never infer image success from Pod deletion. Confirm its action succeeds only after the image workflow is green.

- [ ] **Step 5: Verify Kubernetes rollout and image**

Run fixed read-only checks:

```bash
kubectl rollout status deployment/qq-bot -n default --timeout=180s
kubectl get pods -n default -l app=qq-bot -o wide
kubectl describe pod -n default -l app=qq-bot
kubectl logs -n default -l app=qq-bot --since=15m
```

Confirm a new Pod is Ready, restart count is stable, the running image ID reflects the new build, and logs show successful Jarvis gRPC initialization without configuration or Proto errors.

- [ ] **Step 6: Verify in-cluster Jarvis reachability**

From the qq-bot Pod, verify DNS and TCP connectivity to:

```text
jarvis:9090
```

Do not use the external NodePort for the application path.

- [ ] **Step 7: Perform the live functional checks**

Use the deployed bot to confirm:

- a private question receives a Jarvis answer;
- the same user in different groups uses isolated identities;
- `#模式` reports Jarvis;
- `#模式 1` refuses local switching;
- the next Wall Street reporter cycle logs successful or classified `Ingest` results without content leakage;
- the next 23:56 run sends four chapters and includes a cross-day timeline.

If the time-dependent reporter checks cannot occur in the current verification window, report them as pending operational observations rather than passing them by assumption.
