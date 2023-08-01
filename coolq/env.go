package coolq

import (
	"fmt"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Mrs4s/go-cqhttp/constant"
	"github.com/Mrs4s/go-cqhttp/util/ai_util"
	"github.com/Mrs4s/go-cqhttp/util/file_util"
	"os"
	"strings"
)

func (bot *CQBot) SetENV(privateMsg *message.PrivateMessage) {
	texts := privateMsg.Texts()
	if len(texts) == 0 {
		bot.SendPrivateMessage(privateMsg.Chat(), 0, &message.SendingMessage{
			Elements: []message.IMessageElement{
				message.NewText("请使用#ENV {NAME}={VALUE}的形式设置环境变量")},
		})
		return
	}

	source, ok := parseSourceText(texts[0])
	if !ok {
		bot.SendPrivateMessage(privateMsg.Chat(), 0, &message.SendingMessage{
			Elements: []message.IMessageElement{
				message.NewText("请使用#ENV {NAME}={VALUE}的形式设置环境变量")},
		})
		return
	}

	v := strings.Split(source, "=")
	if len(v) != 2 {
		bot.SendPrivateMessage(privateMsg.Chat(), 0, &message.SendingMessage{
			Elements: []message.IMessageElement{
				message.NewText("请使用#ENV {NAME}={VALUE}的形式设置环境变量")},
		})
		return
	}

	bot.SendPrivateMessage(privateMsg.Chat(), 0, &message.SendingMessage{
		Elements: []message.IMessageElement{
			message.NewText(envSetHandler(v[0], v[1])),
		}})
}

func envSetHandler(env, value string) string {
	var res string
	switch env {
	case constant.OPENAI_API_KEY:
		ai_util.SetOpenaiKey(value)
		res = "Openai API Key 设置完成"
	case constant.FILE_ROOT:
		file_util.SetFileRoot(value)
		res = "文件本地存储根目录 设置完成"
	default:
		_ = os.Setenv(env, value)
		res = fmt.Sprintf("环境变量 %s 设置完成", env)
	}
	return res
}
