package bot_service

import (
	"context"
	"testing"

	"github.com/NoahAmethyst/go-cqhttp/coolq"
	"github.com/NoahAmethyst/go-cqhttp/protocol/pb/qqbot_pb"
)

type fakeScheduledPushNotifier struct {
	result  coolq.NotifyResult
	content string
}

func (n *fakeScheduledPushNotifier) NotifySubscribers(content string) coolq.NotifyResult {
	n.content = content
	return n.result
}

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

func TestNotifyReturnsNotInitializedErrorWithoutBot(t *testing.T) {
	previousBot := QQBot
	QQBot = nil
	t.Cleanup(func() {
		QQBot = previousBot
	})

	_, err := (BotService{}).Notify(context.Background(), &qqbot_pb.NotifyReq{Content: "notice"})
	if err == nil || err.Error() != "QQ bot not initialize yet" {
		t.Fatalf("Notify() error = %v, want QQ bot not initialize yet", err)
	}
}
