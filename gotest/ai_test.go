package gotest

import (
	"fmt"
	"github.com/NoahAmethyst/go-cqhttp/util/ai_util"
	go_ernie "github.com/anhao/go-ernie"
	"github.com/sashabaranov/go-openai"
	"strings"
	"testing"
)

func Test_Models(t *testing.T) {
	if models, err := ai_util.OpenAiModels(); err != nil {
		panic(err)
	} else {
		for _, _model := range models.Models {
			t.Logf("%+v", _model.ID)
		}
	}

}

func Test_ChatGpt(t *testing.T) {
	ctx := make([]openai.ChatCompletionMessage, 0, 4)
	ctx = append(ctx, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: "魔术的起源是什么",
	})
	replyMsg, err := ai_util.AskChatGpt(ctx)
	if err != nil {
		panic(err)
	}
	t.Logf("%+v", replyMsg)
}

func Test_ChatGptWithContext(t *testing.T) {
	ctx := make([]openai.ChatCompletionMessage, 0, 4)
	ctx = append(ctx, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: "南京今天天气怎么样",
	})
	replyMsg, err := ai_util.AskChatGpt(ctx)
	if err != nil {
		panic(err)
	}
	t.Logf("%+v", replyMsg)

	ctx = append(ctx, openai.ChatCompletionMessage{
		Role:    replyMsg.Choices[0].Message.Role,
		Content: replyMsg.Choices[0].Message.Content,
	})

	ctx = append(ctx, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: "还有其他的方法吗",
	})

	replyMsg, err = ai_util.AskChatGpt(ctx)
	if err != nil {
		panic(err)
	}
	t.Logf("%+v", replyMsg)
}

func Test_ChatGpt4WithContext(t *testing.T) {
	ctx := make([]openai.ChatCompletionMessage, 0, 4)
	ctx = append(ctx, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: "How is the weather today in Nanjing?",
	})
	replyMsg, err := ai_util.AskChatGpt4(ctx)
	if err != nil {
		panic(err)
	}
	t.Logf("%+v", replyMsg)

	ctx = append(ctx, openai.ChatCompletionMessage{
		Role:    replyMsg.Choices[0].Message.Role,
		Content: replyMsg.Choices[0].Message.Content,
	})

	ctx = append(ctx, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: "And what day is it today?",
	})

	replyMsg, err = ai_util.AskChatGpt4(ctx)
	if err != nil {
		panic(err)
	}
	t.Logf("%+v", replyMsg)
}

func Test_OpenAIGenerateImg(t *testing.T) {

	prompt := `Create an high quality image of a vast, awe-inspiring scene that blends elements of science fiction, cyberpunk, and the infinite expanse of the universe. The image should feature an enormous planet dominating the background, with intricate details on the surface visible from afar, such as glowing cities, sprawling forests, and deep canyons. The sky around the planet should be filled with swirling clouds, electric storms, and perhaps even space debris or satellites adding to the intrigue.`
	imgResp, err := ai_util.GenerateImage(prompt, openai.CreateImageSize1024x1024)
	if err != nil {
		panic(err)
	}

	t.Logf("%+v", imgResp)
}

func Test_OpenAIEditImg(t *testing.T) {
	imgResp, err := ai_util.GenerateImage("Color photo of a beautiful Chinese girl sitting on a traditional wooden bench in a serene garden. She wears a flowing red dress with intricate gold embroidery and a delicate flower in her hair. Her porcelain skin glows in the soft sunlight, and her almond-shaped eyes gaze off into the distance with a look of contemplation. The garden around her is filled with vibrant greenery, blooming flowers, and a peaceful koi pond.\n\nNikon D850, Fujifilm Pro 400H film, 85mm lens, natural light.\n\nDirectors: Ang Lee, Wong Kar-wai\nCinematographers: Christopher Doyle, Mark Lee Ping-bin\nPhotographers: Zhang Jingna, Chen Man\nFashion designers: Guo Pei, Vera Wang, Jason Wu\n—c 10 —ar 2:3",
		openai.CreateImageSize1024x1024)
	if err != nil {
		panic(err)
	}

	t.Logf("%+v", imgResp)
}

func Test_BingChat(t *testing.T) {

	bingChatCli, err := ai_util.NewBingChat()
	if err != nil {
		panic(err)
	}

	answer, err := ai_util.AskBingChat(bingChatCli, "南京今天天气怎么样")
	if err != nil {
		panic(err)
	}

	var strBuilder strings.Builder
	strBuilder.WriteString(answer.Answer)

	if len(answer.Reference) > 0 {
		strBuilder.WriteString("\n\n参考资料:")
	}
	for title, link := range answer.Reference {
		strBuilder.WriteString(fmt.Sprintf("\n%s %s", title, link))
	}

	if len(answer.Suggestions) > 0 {
		strBuilder.WriteString("\n\n您也可以这样提问")
	}
	for i, suggest := range answer.Suggestions {
		strBuilder.WriteString(fmt.Sprintf("\n%d: %s", i+1, suggest))
	}

	t.Logf("%s", strBuilder.String())

}

func Test_Ernie(t *testing.T) {
	completion, err := ai_util.AskErnie([]go_ernie.ChatCompletionMessage{
		{Role: go_ernie.MessageRoleUser,
			Content: "怎么制作一个可乐鸡翅"},
	})

	if err != nil {
		panic(err)
	}

	t.Logf("%+v", completion)
}
