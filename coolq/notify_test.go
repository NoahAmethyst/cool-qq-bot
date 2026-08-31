package coolq

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/Mrs4s/MiraiGo/message"
	log "github.com/sirupsen/logrus"
)

type recordingNotifySender struct {
	groupIDs        []int64
	privateIDs      []int64
	groupContents   []string
	privateContents []string
	groupStatus     map[int64]int32
	privateStatus   map[int64]int32
}

func (s *recordingNotifySender) SendGroupMessage(id int64, msg *message.SendingMessage) int32 {
	s.groupIDs = append(s.groupIDs, id)
	s.groupContents = append(s.groupContents, notifyMessageText(msg))
	if status, ok := s.groupStatus[id]; ok {
		return status
	}
	return 0
}

func (s *recordingNotifySender) SendPrivateMessage(id int64, _ int64, msg *message.SendingMessage) int32 {
	s.privateIDs = append(s.privateIDs, id)
	s.privateContents = append(s.privateContents, notifyMessageText(msg))
	if status, ok := s.privateStatus[id]; ok {
		return status
	}
	return 0
}

func notifyMessageText(msg *message.SendingMessage) string {
	if msg == nil || len(msg.Elements) != 1 {
		return ""
	}
	text, _ := msg.Elements[0].(*message.TextElement)
	if text == nil {
		return ""
	}
	return text.Content
}

func captureStandardLog(t *testing.T) *bytes.Buffer {
	t.Helper()

	logger := log.StandardLogger()
	previousOutput := logger.Out
	previousFormatter := logger.Formatter
	previousLevel := logger.GetLevel()
	buffer := new(bytes.Buffer)
	logger.SetOutput(buffer)
	logger.SetFormatter(&log.TextFormatter{DisableTimestamp: true})
	logger.SetLevel(log.DebugLevel)
	t.Cleanup(func() {
		logger.SetOutput(previousOutput)
		logger.SetFormatter(previousFormatter)
		logger.SetLevel(previousLevel)
	})
	return buffer
}

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
	for _, content := range append(sender.groupContents, sender.privateContents...) {
		if content != "scheduled notice" {
			t.Fatalf("content = %q, want %q", content, "scheduled notice")
		}
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
