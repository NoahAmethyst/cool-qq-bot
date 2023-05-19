package coolq

import (
	"fmt"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Mrs4s/go-cqhttp/util/file_util"
	"github.com/Mrs4s/go-cqhttp/util/openai_util"
	log "github.com/sirupsen/logrus"
	"time"
)

func (bot *CQBot) generateImgInPrivate(m *message.PrivateMessage) {
	var textEle *message.TextElement
	for _, _ele := range m.Elements {
		switch _ele.Type() {
		case message.Text:
			textEle = _ele.(*message.TextElement)
		default:

		}
	}

	if textEle == nil {
		return
	}

	text, done := parseSourceText(textEle)
	if done {
		bot.SendPrivateMessage(m.Sender.Uin, 0, &message.SendingMessage{Elements: []message.IMessageElement{
			message.NewText(
				"缺少生成图片的描述内容")}})
		return
	}

	recvChan := make(chan struct{}, 1)
	defer close(recvChan)
	go func(uid int64) {
		select {
		case <-recvChan:
			return
		case <-time.After(time.Second * 10):
			bot.SendPrivateMessage(m.Sender.Uin, 0, &message.SendingMessage{Elements: []message.IMessageElement{
				message.NewText("OPENAI正在生成图片，请稍等...")}})
		}
	}(m.Sender.Uin)

	r, err := openai_util.GenerateImage(text)

	recvChan <- struct{}{}

	if err != nil {
		bot.SendPrivateMessage(m.Sender.Uin, 0, &message.SendingMessage{Elements: []message.IMessageElement{
			message.NewText(
				fmt.Sprintf("DELL.2生成图片失败：%s", err.Error()))}})
	} else {
		if len(r.Data) == 0 {
			bot.SendPrivateMessage(m.Sender.Uin, 0, &message.SendingMessage{Elements: []message.IMessageElement{
				message.NewText("DELL.2生成图片为空")}})
		} else {
			bot.SendPrivateMessage(m.Sender.Uin, 0, &message.SendingMessage{Elements: []message.IMessageElement{
				message.NewText("正在上传图片，请稍后...")}})
			if f, path, err := file_util.DownloadImgFromUrl(r.Data[0].URL); err == nil {
				if msg, err := bot.uploadLocalImage(message.Source{
					SourceType: message.SourcePrivate,
					PrimaryID:  m.Sender.Uin,
				}, &LocalImageElement{
					Stream: f,
					File:   path,
					URL:    r.Data[0].URL,
				}); err != nil {
					log.Error("上传图片失败：%s", err.Error())
					bot.SendPrivateMessage(m.Sender.Uin, 0, &message.SendingMessage{Elements: []message.IMessageElement{
						message.NewText(fmt.Sprintf("上传图片失败(%s)，图片地址：%s", err.Error(), r.Data[0].URL))}})
				} else {
					bot.SendPrivateMessage(m.Sender.Uin, 0, &message.SendingMessage{
						Elements: []message.IMessageElement{
							msg,
						},
					})
				}
			} else {
				log.Error("读取图片转换为reader失败：%s", err.Error())
				bot.SendPrivateMessage(m.Sender.Uin, 0, &message.SendingMessage{Elements: []message.IMessageElement{
					message.NewText(fmt.Sprintf("读取图片转换为reader失败(%s)，图片地址：%s", err.Error(), r.Data[0].URL))}})
			}
		}

	}
}

func (bot *CQBot) generateImgInGroup(m *message.GroupMessage) {
	var textEle *message.TextElement
	for _, _ele := range m.Elements {
		switch _ele.Type() {
		case message.Text:
			textEle = _ele.(*message.TextElement)
		default:

		}
	}

	if textEle == nil {
		return
	}

	text, done := parseSourceText(textEle)
	if done {
		bot.SendGroupMessage(m.GroupCode, &message.SendingMessage{Elements: []message.IMessageElement{
			message.NewText(
				"缺少生成图片的描述内容")}})
		return
	}

	recvChan := make(chan struct{}, 1)
	defer close(recvChan)
	go func(uid int64) {
		select {
		case <-recvChan:
			return
		case <-time.After(time.Second * 10):
			bot.SendGroupMessage(m.GroupCode, &message.SendingMessage{Elements: []message.IMessageElement{
				message.NewText("OPENAI正在生成图片，请稍等...")}})
		}
	}(m.Sender.Uin)

	r, err := openai_util.GenerateImage(text)

	recvChan <- struct{}{}

	if err != nil {
		bot.SendGroupMessage(m.GroupCode, &message.SendingMessage{Elements: []message.IMessageElement{
			message.NewText(
				fmt.Sprintf("DELL.2生成图片失败：%s", err.Error()))}})
	} else {

		if len(r.Data) == 0 {
			bot.SendGroupMessage(m.GroupCode, &message.SendingMessage{Elements: []message.IMessageElement{
				message.NewText("DELL.2生成图片为空")}})
		} else {
			bot.SendGroupMessage(m.GroupCode, &message.SendingMessage{Elements: []message.IMessageElement{
				message.NewText("正在上传图片，请稍后...")}})
			if f, path, err := file_util.DownloadImgFromUrl(r.Data[0].URL); err == nil {
				if msg, err := bot.uploadLocalImage(message.Source{
					SourceType: message.SourceGroup,
					PrimaryID:  m.GroupCode,
				}, &LocalImageElement{
					Stream: f,
					File:   path,
					URL:    r.Data[0].URL,
				}); err != nil {
					log.Error("上传图片失败：%s", err.Error())
					bot.SendGroupMessage(m.GroupCode, &message.SendingMessage{Elements: []message.IMessageElement{
						message.NewText(fmt.Sprintf("上传图片失败(%s)，图片地址：%s", err.Error(), r.Data[0].URL))}})
				} else {
					bot.SendGroupMessage(m.GroupCode, &message.SendingMessage{
						Elements: []message.IMessageElement{
							msg,
						},
					})
				}
			} else {
				log.Error("读取图片转换为reader失败：%s", err.Error())
				bot.SendGroupMessage(m.GroupCode, &message.SendingMessage{Elements: []message.IMessageElement{
					message.NewText(fmt.Sprintf("读取图片转换为reader失败(%s)，图片地址：%s", err.Error(), r.Data[0].URL))}})
			}
		}

	}

}
