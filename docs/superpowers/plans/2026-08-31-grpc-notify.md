# gRPC Notify 广播推送 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为 QQBot gRPC 服务提供仅接收文本内容的 `Notify` 接口，并向全部定时推送订阅者广播。

**Architecture:** `coolq` 新增一个可测试的订阅者广播器：它从 `reportState` 取得群聊和私聊快照、逐一投递，并记录每次失败但不中断后续投递。`bot_service` 只将该结果翻译成 gRPC 语义：无订阅者为成功提示，至少一次成功为成功，全部失败为错误。Proto 生成绑定继续由 `protocol/gen.sh` 生成，绝不手改。

**Tech Stack:** Go 1.19、Protocol Buffers、gRPC、MiraiGo、logrus。

---

## 文件结构

- `protocol/qqbot.proto`：声明 `Notify` 及其仅含文本内容的请求。
- `protocol/pb/qqbot_pb/qqbot.pb.go`：由 `protocol/gen.sh` 生成；提供 `NotifyReq`、客户端和服务端绑定。
- `coolq/notify.go`：定时推送订阅者快照、逐一投递、结果汇总和不含消息正文的失败日志。
- `coolq/notify_test.go`：验证群/私聊广播及失败日志。
- `bot_service/bot_service.go`：实现 gRPC `Notify` 并将投递结果转换为约定的响应或错误。
- `bot_service/bot_service_test.go`：验证 gRPC 层的无订阅者、部分成功、全失败及未初始化语义。
- `docs/CODEX_PROJECT_CONTEXT.md`：记录新的 gRPC 广播能力。

### Task 1: 增加 Proto 契约并生成 Go 绑定

**Files:**
- Modify: `protocol/qqbot.proto:7-32`
- Modify generated: `protocol/pb/qqbot_pb/qqbot.pb.go`

- [ ] **Step 1: 在服务声明中加入 Notify，并定义其请求消息**

在 `Groups` RPC 后加入以下定义。保留 `SendMsgReq` 的字段号和行为，不复用它。

```proto
  rpc Notify(NotifyReq) returns(Resp){}
}

message NotifyReq{
  string content = 1;
}

message Empty{
```

- [ ] **Step 2: 重新生成绑定**

Run: `cd protocol && bash gen.sh`

Expected: exit 0；`protocol/pb/qqbot_pb/qqbot.pb.go` 包含 `type NotifyReq`、`GetContent()`、`QQBotServiceClient.Notify`、`QQBotServiceServer.Notify` 和 `/proto.QQBotService/Notify` handler。

- [ ] **Step 3: 确认生成的 API 完整**

Run: `rg -n 'type NotifyReq|func \(c \*qQBotServiceClient\) Notify|Notify\(context.Context, \*NotifyReq\)|/proto.QQBotService/Notify' protocol/pb/qqbot_pb/qqbot.pb.go`

Expected: 四类符号均至少出现一次。

- [ ] **Step 4: 提交 Proto 契约**

Run: `git add protocol/qqbot.proto protocol/pb/qqbot_pb/qqbot.pb.go`

Run: `git commit -m "feat: add QQBot Notify grpc contract"`

### Task 2: 以失败测试定义订阅者广播与失败日志

**Files:**
- Create: `coolq/notify.go`
- Create: `coolq/notify_test.go`

- [ ] **Step 1: 编写失败测试**

在 `coolq/notify_test.go` 中创建 `recordingNotifySender`，实现群和私聊发送方法；它为每个目标保存文本，并按 `groupStatus` / `privateStatus` 返回配置的 `int32` 状态。添加以下测试：

```go
func TestNotifySubscribersFansOutToGroupAndPrivateRecipients(t *testing.T) {
	sender := &recordingNotifySender{}
	result := notifySubscribers(sender, []int64{101, 102}, []int64{201}, "scheduled notice")

	if result.Attempted != 3 || result.Succeeded != 3 {
		t.Fatalf("result = %#v, want 3 attempted and 3 succeeded", result)
	}
	if got, want := sender.groupIDs, []int64{101, 102}; !reflect.DeepEqual(got, want) {
		t.Fatalf("group IDs = %v, want %v", got, want)
	}
	if got, want := sender.privateIDs, []int64{201}; !reflect.DeepEqual(got, want) {
		t.Fatalf("private IDs = %v, want %v", got, want)
	}
}

func TestNotifySubscribersLogsEveryFailedRecipientWithoutContent(t *testing.T) {
	sender := &recordingNotifySender{
		groupStatus:   map[int64]int32{102: -1},
		privateStatus: map[int64]int32{201: -1},
	}
	logs := captureStandardLog(t)
	result := notifySubscribers(sender, []int64{101, 102}, []int64{201}, "secret scheduled notice")

	if result.Attempted != 3 || result.Succeeded != 1 {
		t.Fatalf("result = %#v, want 3 attempted and 1 succeeded", result)
	}
	for _, want := range []string{"recipient_type=group", "recipient_id=102", "recipient_type=private", "recipient_id=201", "status=-1"} {
		if !strings.Contains(logs.String(), want) {
			t.Fatalf("log = %q, want %q", logs.String(), want)
		}
	}
	if strings.Contains(logs.String(), "secret scheduled notice") {
		t.Fatalf("log unexpectedly contains notification content: %q", logs.String())
	}
}
```

`captureStandardLog` 必须保存并在 `t.Cleanup` 恢复标准 logrus logger 的输出、格式化器和日志级别；测试不得调用 `t.Parallel()`。

- [ ] **Step 2: 确认测试因缺少广播器而失败**

Run: `go test ./coolq -run 'TestNotifySubscribers' -count=1`

Expected: FAIL，原因是 `recordingNotifySender` 的测试依赖的 `notifySubscribers` 与 `NotifyResult` 尚未定义。

- [ ] **Step 3: 实现最小广播器与 CQBot 入口**

在 `coolq/notify.go` 定义以下类型和函数。`-1` 是现有 `SendGroupMessage` 和 `SendPrivateMessage` 的失败返回值；所有其他返回值都算作一次成功投递。

```go
package coolq

import (
	"fmt"

	"github.com/Mrs4s/MiraiGo/message"
	log "github.com/sirupsen/logrus"
)

type notifySender interface {
	SendGroupMessage(int64, *message.SendingMessage) int32
	SendPrivateMessage(int64, int64, *message.SendingMessage) int32
}

type NotifyResult struct {
	Attempted int
	Succeeded int
}

func notifySubscribers(sender notifySender, groupIDs, privateIDs []int64, content string) NotifyResult {
	result := NotifyResult{Attempted: len(groupIDs) + len(privateIDs)}
	send := func(recipientType string, recipientID int64, status int32) {
		if status != -1 {
			result.Succeeded++
			return
		}
		log.WithFields(log.Fields{
			"recipient_type": recipientType,
			"recipient_id":   recipientID,
			"status":         status,
		}).WithError(fmt.Errorf("message send returned status %d", status)).Error("Notify delivery failed")
	}
	for _, id := range groupIDs {
		send("group", id, sender.SendGroupMessage(id, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(content)}}))
	}
	for _, id := range privateIDs {
		send("private", id, sender.SendPrivateMessage(id, 0, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(content)}}))
	}
	return result
}

func (bot *CQBot) NotifySubscribers(content string) NotifyResult {
	if bot == nil || bot.state == nil || bot.state.reportState == nil {
		return NotifyResult{}
	}
	return notifySubscribers(bot, bot.state.reportState.getReportList(true), bot.state.reportState.getReportList(false), content)
}
```

- [ ] **Step 4: 确认广播测试通过**

Run: `go test ./coolq -run 'TestNotifySubscribers' -count=1`

Expected: PASS。

- [ ] **Step 5: 提交广播器**

Run: `git add coolq/notify.go coolq/notify_test.go`

Run: `git commit -m "feat: broadcast scheduled push notifications"`

### Task 3: 以失败测试定义 gRPC 返回语义

**Files:**
- Modify: `bot_service/bot_service.go:14-37`
- Create: `bot_service/bot_service_test.go`

- [ ] **Step 1: 编写失败测试**

在 `bot_service/bot_service_test.go` 中创建 `fakeScheduledPushNotifier`，实现 `NotifySubscribers(content string) coolq.NotifyResult` 并记录其 `content`。使用 `BotService{notifier: fake}` 添加下列测试：

```go
func TestNotifyReturnsNoRecipientsMessage(t *testing.T) {
	resp, err := (BotService{notifier: &fakeScheduledPushNotifier{result: coolq.NotifyResult{}}}).Notify(
		context.Background(), &qqbot_pb.NotifyReq{Content: "notice"},
	)
	if err != nil {
		t.Fatalf("Notify() error = %v", err)
	}
	if got, want := resp.Message, "没有可推送的订阅者"; got != want {
		t.Fatalf("message = %q, want %q", got, want)
	}
}

func TestNotifySucceedsWhenAtLeastOneRecipientReceivesMessage(t *testing.T) {
	fake := &fakeScheduledPushNotifier{result: coolq.NotifyResult{Attempted: 2, Succeeded: 1}}
	_, err := (BotService{notifier: fake}).Notify(context.Background(), &qqbot_pb.NotifyReq{Content: "notice"})
	if err != nil {
		t.Fatalf("Notify() error = %v", err)
	}
	if fake.content != "notice" {
		t.Fatalf("content = %q, want %q", fake.content, "notice")
	}
}

func TestNotifyReturnsErrorWhenEveryRecipientFails(t *testing.T) {
	_, err := (BotService{notifier: &fakeScheduledPushNotifier{result: coolq.NotifyResult{Attempted: 2}}}).Notify(
		context.Background(), &qqbot_pb.NotifyReq{Content: "notice"},
	)
	if err == nil {
		t.Fatal("Notify() error = nil, want error")
	}
}
```

另加一个 `QQBot = nil` 且未注入 `notifier` 的测试，断言与现有接口相同地返回 `QQ bot not initialize yet`。

- [ ] **Step 2: 确认测试因服务方法不存在而失败**

Run: `go test ./bot_service -run TestNotify -count=1`

Expected: FAIL，原因是 `BotService` 尚无 `notifier` 依赖和 `Notify` 方法。

- [ ] **Step 3: 实现最小的服务层适配器**

在 `bot_service/bot_service.go` 添加依赖接口和字段；不改变 `QQBot` 全局变量及其现有调用方。

```go
type scheduledPushNotifier interface {
	NotifySubscribers(string) coolq.NotifyResult
}

type BotService struct {
	notifier scheduledPushNotifier
}

func (b BotService) scheduledPushNotifier() (scheduledPushNotifier, error) {
	if b.notifier != nil {
		return b.notifier, nil
	}
	if QQBot == nil {
		return nil, errors.New("QQ bot not initialize yet")
	}
	return QQBot, nil
}

func (b BotService) Notify(_ context.Context, req *qqbot_pb.NotifyReq) (*qqbot_pb.Resp, error) {
	notifier, err := b.scheduledPushNotifier()
	if err != nil {
		return nil, err
	}
	result := notifier.NotifySubscribers(req.GetContent())
	if result.Attempted == 0 {
		return &qqbot_pb.Resp{Message: "没有可推送的订阅者"}, nil
	}
	if result.Succeeded == 0 {
		return nil, errors.New("failed to send notification to all scheduled push subscribers")
	}
	return &qqbot_pb.Resp{}, nil
}
```

- [ ] **Step 4: 确认 gRPC 语义测试通过**

Run: `go test ./bot_service -run TestNotify -count=1`

Expected: PASS。

- [ ] **Step 5: 提交服务实现**

Run: `git add bot_service/bot_service.go bot_service/bot_service_test.go`

Run: `git commit -m "feat: implement QQBot Notify grpc service"`

### Task 4: 格式化、回归验证并更新项目上下文

**Files:**
- Modify: `docs/CODEX_PROJECT_CONTEXT.md:27, 96-100`

- [ ] **Step 1: 记录运行时能力**

将“User-facing capabilities”中的 QQBot gRPC 接口列表扩展为 `SendMsg`、`Notify`、`Self`、`Friends`、`Groups`，并说明 `Notify` 对所有已开启定时推送的群聊与私聊订阅者广播；无订阅者返回成功提示，部分失败记录目标和原因但不影响至少一次成功的结果。

- [ ] **Step 2: 格式化修改的 Go 文件**

Run: `gofmt -w coolq/notify.go coolq/notify_test.go bot_service/bot_service.go bot_service/bot_service_test.go`

Expected: exit 0。

- [ ] **Step 3: 运行聚焦测试和编译验证**

Run: `go test ./coolq ./bot_service -count=1`

Expected: exit 0。

Run: `go build ./...`

Expected: exit 0。

- [ ] **Step 4: 检查契约和工作区范围**

Run: `git diff --check && git status --short`

Expected: 无空白错误；仅暂存本计划列出的实现文件，保留并不暂存既有 `config.yml`、`qq_bot.yml`、`qrcode.png` 与无关文档改动。

- [ ] **Step 5: 提交最终文档**

Run: `git add docs/CODEX_PROJECT_CONTEXT.md`

Run: `git commit -m "docs: describe grpc notify broadcasts"`

## 部署后验证

本地测试不能证明真实 QQ 收件人可达。按仓库部署规则发布后，使用一个开启群聊订阅和一个开启私聊订阅的测试环境调用 `Notify`，确认两端均收到相同文本；再临时使一个测试收件人不可投递，确认另一个仍收到消息、服务端日志含失败目标与状态、RPC 仍成功。测试完成后恢复测试订阅状态。
