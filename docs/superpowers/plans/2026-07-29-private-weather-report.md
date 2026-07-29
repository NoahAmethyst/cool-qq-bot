# Private Weather Report Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a private-chat-only `#天气` subscription command and send Jarvis-generated, QQ-safe non-Markdown weather advice, including optional emoji, every day at 08:00.

**Architecture:** Persist a concurrency-safe `QQ user ID -> city` map independently from existing news subscriptions. At 08:00, group subscribers by city, call Jarvis once per city with a date-specific weather prompt, sanitize common Markdown syntax while preserving emoji, and fan the QQ text out to private recipients.

**Tech Stack:** Go 1.19, MiraiGo messages, robfig/cron, Jarvis gRPC, existing JSON/Tencent COS cache utilities.

**Verification exception:** The user explicitly requested no new tests and no local verification commands. Runtime validation will occur after user-managed deployment.

---

### Task 1: Persist private weather subscriptions

**Files:**
- Modify: `coolq/state.go`

- [ ] **Step 1: Add weather state to the root state**

Add `weatherState *weatherState` to `State`, initialize it in `initState`, and call `weatherState.saveCache()` from `State.SaveCache`.

- [ ] **Step 2: Add concurrency-safe weather operations**

Define:

```go
type weatherState struct {
    sync.RWMutex
    Cities map[int64]string
}

func (s *weatherState) set(userID int64, city string)
func (s *weatherState) del(userID int64) bool
func (s *weatherState) get(userID int64) (string, bool)
func (s *weatherState) snapshot() map[int64]string
func (s *weatherState) saveCache()
func initWeatherState() *weatherState
```

`set` and `del` persist immediately after releasing the write lock. `snapshot` returns a copy so the scheduled task never holds the state lock during Jarvis or QQ calls.

- [ ] **Step 3: Reuse the repository cache convention**

Store `weatherPrivateState.json` beneath `file_util.GetFileRoot()`. If local loading fails, download `cache/weatherPrivateState.json` from Tencent COS and retry. Upload the file after every successful save without logging cities or user IDs.

### Task 2: Implement private command interaction

**Files:**
- Modify: `coolq/cmd.go`
- Create: `coolq/weather.go`

- [ ] **Step 1: Register the command only for private chat**

Add `CMDWeather = "天气"`, its help text, and add it only to `privateCmdList` and `privateCmdHandlers`:

```go
CMDWeather: func(bot *CQBot, privateMessage *message.PrivateMessage) {
    bot.handlePrivateWeather(privateMessage)
},
```

Do not add it to `groupCmdList` or `groupCmdHandlers`.

- [ ] **Step 2: Parse and execute weather settings**

In `coolq/weather.go`, implement `handlePrivateWeather` with these exact outcomes:

- `#天气`: reply with the configured city and `每天 08:00 推送`, or the setup usage when none exists.
- `#天气 杭州`: trim the argument, reject control characters or more than 40 runes, save it, and confirm the daily schedule.
- `#天气 关闭`: delete the setting and report whether weather push was closed or was already closed.

All replies use `SendPrivateMessage` with one `message.TextElement`.

### Task 3: Query Jarvis and enforce QQ-safe non-Markdown text

**Files:**
- Create: `coolq/weather.go`

- [ ] **Step 1: Build a date-specific prompt**

Implement:

```go
func buildWeatherPrompt(city string, now time.Time) string
func weatherJarvisUserID(city string) string
```

The prompt must request the selected city's current-day weather, temperature range, rain, wind and UV when available, plus travel, sun-protection, clothing and safety advice. It may use emoji and symbolic expressions, but must forbid invented values and explicitly prohibit Markdown headings, bullets, emphasis, code fences, tables and link syntax.

Use a stable, city-isolated Jarvis identity prefixed with `go-cqhttp:weather:` so one city's history does not pollute another city's request.

- [ ] **Step 2: Sanitize Jarvis output**

Implement:

```go
func plainQQText(answer string) string
```

Normalize line endings, convert Markdown links to labels, remove code fences, common heading/quote/list prefixes, emphasis/backtick markers and table pipes, preserve emoji, collapse excessive blank lines, and trim the result. An empty sanitized answer is a failure.

- [ ] **Step 3: Fan out once per city**

Implement:

```go
func (bot *CQBot) ReportWeather(subscriptions map[int64]string)
```

Group recipients by city, sort cities and user IDs for deterministic processing, call `jarvis_svc.Chat(prompt, weatherJarvisUserID(city))` once per city, then send the sanitized answer to each recipient. On RPC or empty-answer failure, log only the gRPC status code and send `今日天气查询失败，请稍后重试` to affected recipients. If Jarvis marks the response low-confidence, prepend a plain-text caution.

### Task 4: Schedule the 08:00 job

**Files:**
- Modify: `coolq/job.go`
- Modify: `coolq/bot.go`

- [ ] **Step 1: Keep weather out of group startup announcements**

Change `ReportJob.RunJob` so it writes to `JobModels` only when `Model` is non-empty. This preserves existing news announcements while allowing non-news scheduled jobs.

- [ ] **Step 2: Add the weather reporter**

Add:

```go
func (bot *CQBot) WeatherReporter(corn string) *ReportJob {
    return &ReportJob{
        Report: func() {
            bot.ReportWeather(bot.state.weatherState.snapshot())
        },
        Corn: corn,
    }
}
```

- [ ] **Step 3: Register the schedule**

In `NewQQBot`, register:

```go
bot.RegisterJob(bot.WeatherReporter("0 0 8 * * *"))
```

The process already fixes its runtime timezone to UTC+8 in `main.go`, so this fires at 08:00 Asia/Shanghai.

### Task 5: Synchronize repository documentation

**Files:**
- Modify: `docs/CODEX_PROJECT_CONTEXT.md`

- [ ] **Step 1: Update capability and runtime descriptions**

Document the private-only `#天气` command, daily 08:00 task, per-city Jarvis request/fan-out, QQ plain-text output constraint, and `weatherPrivateState.json` persistence. Explicitly state that group chat does not expose the weather command.

- [ ] **Step 2: Review the final diff without running verification**

Inspect only the scoped diff to ensure no existing sensitive or unrelated files are included. Do not run tests, builds, formatting commands, deployment, or remote operations because the user reserved runtime verification for the post-deployment environment.
