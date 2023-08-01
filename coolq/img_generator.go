package coolq

import (
	"fmt"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Mrs4s/go-cqhttp/util/ai_util"
	"github.com/Mrs4s/go-cqhttp/util/file_util"
	"github.com/pkg/errors"
	"github.com/sashabaranov/go-openai"
	log "github.com/sirupsen/logrus"
	"io"
	"time"
)

type ImgGenerator interface {
	SendMessage(content string)
	SendImg(stream io.ReadSeeker, filePath, url string) error
	GetText() []string
	Check() bool
	Target() int64
}

type PrivateImgGenerator struct {
	bot *CQBot
	m   *message.PrivateMessage
}

func (p *PrivateImgGenerator) Check() bool {
	return p.bot != nil && p.m != nil
}

func (p *PrivateImgGenerator) Target() int64 {
	return p.m.Chat()
}

func (p *PrivateImgGenerator) SendMessage(content string) {

	p.bot.SendPrivateMessage(p.Target(), 0, &message.SendingMessage{Elements: []message.IMessageElement{
		message.NewText(
			content)}})
}

func (p *PrivateImgGenerator) SendImg(_ io.ReadSeeker, filepath, url string) error {

	result := p.bot.SendPrivateMessage(p.Target(), 0, &message.SendingMessage{
		Elements: []message.IMessageElement{
			&LocalImageElement{
				File: filepath,
				URL:  url,
			},
			message.NewText(fmt.Sprintf("图片链接：%s", url)),
		},
	})
	if result < 0 {
		return errors.New("发送图片失败")
	}
	return nil

}

func (p *PrivateImgGenerator) GetText() []string {
	return p.m.Texts()
}

type GroupImgGenerator struct {
	bot *CQBot
	m   *message.GroupMessage
}

func (p *GroupImgGenerator) Check() bool {
	return p.bot != nil && p.m != nil
}

func (p *GroupImgGenerator) SendMessage(content string) {

	p.bot.SendGroupMessage(p.Target(), &message.SendingMessage{Elements: []message.IMessageElement{
		message.NewText(
			content)}})
}

func (p *GroupImgGenerator) SendImg(_ io.ReadSeeker, filepath, url string) error {

	result := p.bot.SendGroupMessage(p.Target(), &message.SendingMessage{
		Elements: []message.IMessageElement{
			&LocalImageElement{
				File: filepath,
				URL:  url,
			},
		},
	})
	if result < 0 {
		return errors.New("发送图片失败")
	}

	return nil
}

func (p *GroupImgGenerator) GetText() []string {
	return p.m.Texts()
}

func (p *GroupImgGenerator) Target() int64 {
	return p.m.Chat()
}

func GenerateImage(generator ImgGenerator) {
	if generator == nil || !generator.Check() {
		log.Warnf("invalid image generator")
		return
	}

	texts := generator.GetText()
	if len(texts) == 0 {
		return
	}
	text, done := parseSourceText(texts[0])
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

	r, err := ai_util.GenerateImage(text, openai.CreateImageSize512x512)

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
		log.Errorf("上传图片失败：%s", err.Error())
		generator.SendMessage(fmt.Sprintf("%s，图片连接：%s", err.Error(), r.Data[0].URL))
	}
}
