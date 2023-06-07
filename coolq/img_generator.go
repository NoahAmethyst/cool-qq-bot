package coolq

import (
	"fmt"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Mrs4s/go-cqhttp/util/ai_util"
	"github.com/Mrs4s/go-cqhttp/util/file_util"
	log "github.com/sirupsen/logrus"
	"io"
	"reflect"
	"time"
)

type imgGenerator interface {
	SendMessage(content string)
	SendImg(stream io.ReadSeeker, filePath, url string) error
	GetText() *message.TextElement
	Check() bool
	Target() int64
}

type privateImgGenerator struct {
	bot *CQBot
	m   *message.PrivateMessage
}

func (p *privateImgGenerator) Check() bool {
	return p.bot != nil && p.m != nil
}

func (p *privateImgGenerator) Target() int64 {
	return p.m.Sender.Uin
}

func (p *privateImgGenerator) SendMessage(content string) {

	p.bot.SendPrivateMessage(p.Target(), 0, &message.SendingMessage{Elements: []message.IMessageElement{
		message.NewText(
			content)}})
}

func (p *privateImgGenerator) SendImg(stream io.ReadSeeker, filePath, url string) error {

	img, err := p.bot.uploadLocalImage(message.Source{
		SourceType: message.SourcePrivate,
		PrimaryID:  p.Target(),
	}, &LocalImageElement{
		Stream: stream,
		File:   filePath,
		URL:    url,
	})
	if err != nil {
		log.Error("上传图片失败：%s", err.Error())
		err = fmt.Errorf("上传图片失败(%s)，图片地址：%s", err.Error(), url)
	} else {
		p.bot.SendPrivateMessage(p.Target(), 0, &message.SendingMessage{
			Elements: []message.IMessageElement{
				img,
			},
		})
	}

	return err
}

func (p *privateImgGenerator) GetText() *message.TextElement {
	var textEle *message.TextElement
	for _, _ele := range p.m.Elements {
		switch _ele.Type() {
		case message.Text:
			textEle = _ele.(*message.TextElement)
		default:
		}
	}
	return textEle
}

type groupImgGenerator struct {
	bot *CQBot
	m   *message.GroupMessage
}

func (p *groupImgGenerator) Check() bool {
	return p.bot != nil && p.m != nil
}

func (p *groupImgGenerator) SendMessage(content string) {

	p.bot.SendGroupMessage(p.Target(), &message.SendingMessage{Elements: []message.IMessageElement{
		message.NewText(
			content)}})
}

func (p *groupImgGenerator) SendImg(stream io.ReadSeeker, filePath, url string) error {

	img, err := p.bot.uploadLocalImage(message.Source{
		SourceType: message.SourceGroup,
		PrimaryID:  p.Target(),
	}, &LocalImageElement{
		Stream: stream,
		File:   filePath,
		URL:    url,
	})
	if err != nil {
		log.Error("上传图片失败：%s", err.Error())
		err = fmt.Errorf("上传图片失败(%s)，图片地址：%s", err.Error(), url)
	} else {
		p.bot.SendGroupMessage(p.Target(), &message.SendingMessage{
			Elements: []message.IMessageElement{
				img,
			},
		})
	}
	return err
}

func (p *groupImgGenerator) GetText() *message.TextElement {
	var textEle *message.TextElement
	for _, _ele := range p.m.Elements {
		switch _ele.Type() {
		case message.Text:
			textEle = _ele.(*message.TextElement)
		default:
		}
	}
	return textEle
}

func (p *groupImgGenerator) Target() int64 {
	return p.m.GroupCode
}

func generateImg(generator imgGenerator) {
	if generator == nil {
		log.Warnf("nil image generator")
		return
	}
	if !generator.Check() {
		log.Warnf("generator[%s] not set bot or message", reflect.TypeOf(generator).Name())
		return
	}

	textEle := generator.GetText()
	if textEle == nil {
		return
	}
	text, done := parseSourceText(textEle)
	if done {
		generator.SendMessage("缺少生成图片的描述内容")
		return
	}

	recvChan := make(chan struct{}, 1)
	defer close(recvChan)
	go func(uid int64) {
		select {
		case <-recvChan:
			return
		case <-time.After(time.Second * 10):
			generator.SendMessage("OPENAI正在生成图片，请稍等...")
		}
	}(generator.Target())

	r, err := ai_util.GenerateImage(text)

	recvChan <- struct{}{}

	if err != nil {
		generator.SendMessage(fmt.Sprintf("DELL.2生成图片失败：%s", err.Error()))
		return
	}

	if len(r.Data) == 0 {
		generator.SendMessage("DELL.2生成图片为空")
		return
	}
	generator.SendMessage("正在上传图片，请稍后...")
	f, path, err := file_util.DownloadImgFromUrl(r.Data[0].URL)
	if err != nil {
		log.Error("读取图片转换为reader失败：%s", err.Error())
		generator.SendMessage(fmt.Sprintf("读取图片转换为reader失败(%s)，图片地址：%s", err.Error(), r.Data[0].URL))
		return
	}

	if err := generator.SendImg(f, path, r.Data[0].URL); err != nil {
		log.Error("上传图片失败：%s", err.Error())
		generator.SendMessage(err.Error())
	}
}
