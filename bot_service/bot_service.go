package bot_service

import (
	"context"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/NoahAmethyst/go-cqhttp/constant"
	"github.com/NoahAmethyst/go-cqhttp/coolq"
	"github.com/NoahAmethyst/go-cqhttp/protocol/pb/qqbot_pb"
	"github.com/pkg/errors"
	"os"
	"strconv"
)

type scheduledPushNotifier interface {
	NotifySubscribers(string) coolq.NotifyResult
}

type BotService struct {
	notifier scheduledPushNotifier
}

var QQBot *coolq.CQBot

func (b BotService) scheduledPushNotifier() (scheduledPushNotifier, error) {
	if b.notifier != nil {
		return b.notifier, nil
	}
	if QQBot == nil {
		return nil, errors.New("QQ bot not initialize yet")
	}
	return QQBot, nil
}

func (b BotService) SendMsg(_ context.Context, req *qqbot_pb.SendMsgReq) (resp *qqbot_pb.Resp, err error) {
	resp = new(qqbot_pb.Resp)
	if QQBot == nil {
		err = errors.New("QQ bot not initialize yet")
		return
	}
	if req.Group {
		QQBot.SendGroupMessage(req.Chat, &message.SendingMessage{Elements: []message.IMessageElement{
			message.NewText(req.Content)}})
	} else {
		QQBot.SendPrivateMessage(req.Chat, 0, &message.SendingMessage{Elements: []message.IMessageElement{
			message.NewText(req.Content)}})
	}
	return
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

func (b BotService) Self(_ context.Context, _ *qqbot_pb.Empty) (resp *qqbot_pb.Resp, err error) {
	resp = new(qqbot_pb.Resp)
	if QQBot == nil {
		err = errors.New("QQ bot not initialize yet")
		return
	}

	owner := int64(-1)
	if _owner, err := strconv.ParseInt(os.Getenv(constant.OWNER), 10, 64); err == nil {
		owner = _owner
	}
	resp.Self = &qqbot_pb.User{
		Nickname: QQBot.Client.Nickname,
		Code:     QQBot.Client.Uin,
		Owner:    owner,
	}
	return
}

func (b BotService) Friends(_ context.Context, _ *qqbot_pb.Empty) (resp *qqbot_pb.Resp, err error) {
	resp = new(qqbot_pb.Resp)
	if QQBot == nil {
		err = errors.New("QQ bot not initialize yet")
		return
	}

	for _, _friend := range QQBot.Client.FriendList {
		resp.Friends = append(resp.Friends, &qqbot_pb.User{
			Nickname: _friend.Nickname,
			Code:     _friend.Uin,
			Remark:   _friend.Remark,
		})
	}
	return
}

func (b BotService) Groups(_ context.Context, _ *qqbot_pb.Empty) (resp *qqbot_pb.Resp, err error) {
	resp = new(qqbot_pb.Resp)
	if QQBot == nil {
		err = errors.New("QQ bot not initialize yet")
		return
	}

	for _, _group := range QQBot.Client.GroupList {
		group := &qqbot_pb.Group{
			Code:            _group.Code,
			Name:            _group.Name,
			Owner:           _group.OwnerUin,
			GroupCreateTime: _group.GroupCreateTime,
			GroupLevel:      _group.GroupLevel,
			MemberCount:     uint64(_group.MemberCount),
			MaxMemberCount:  uint64(_group.MaxMemberCount),
			Members:         nil,
		}
		for _, _member := range _group.Members {
			group.Members = append(group.Members, &qqbot_pb.User{
				Nickname:      _member.Nickname,
				Code:          _member.Uin,
				CardName:      _member.CardName,
				JoinTime:      _member.JoinTime,
				LastSpeakTime: _member.LastSpeakTime,
				SpecialTitle:  _member.SpecialTitle,
			})
		}
		resp.Groups = append(resp.Groups, group)
	}
	return
}
