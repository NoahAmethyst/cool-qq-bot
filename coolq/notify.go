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

// NotifyResult describes the delivery outcome for one Notify request.
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
		send("group", id, sender.SendGroupMessage(id, &message.SendingMessage{
			Elements: []message.IMessageElement{message.NewText(content)},
		}))
	}
	for _, id := range privateIDs {
		send("private", id, sender.SendPrivateMessage(id, 0, &message.SendingMessage{
			Elements: []message.IMessageElement{message.NewText(content)},
		}))
	}
	return result
}

// NotifySubscribers sends content to all group and private scheduled-push subscribers.
func (bot *CQBot) NotifySubscribers(content string) NotifyResult {
	if bot == nil || bot.state == nil || bot.state.reportState == nil {
		return NotifyResult{}
	}
	return notifySubscribers(
		bot,
		bot.state.reportState.getReportList(true),
		bot.state.reportState.getReportList(false),
		content,
	)
}
