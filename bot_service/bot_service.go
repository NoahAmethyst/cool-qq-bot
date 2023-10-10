package bot_service

import (
	"context"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Mrs4s/go-cqhttp/constant"
	"github.com/Mrs4s/go-cqhttp/coolq"
	"github.com/Mrs4s/go-cqhttp/protocol/pb/qqbot_pb"
	"github.com/pkg/errors"
	"os"
	"strconv"
)

type BotService struct{}

var QQBot *coolq.CQBot

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
