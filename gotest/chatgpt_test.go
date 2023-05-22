package gotest

import (
	"testing"
)

func Test_Chatgpt(t *testing.T) {
	replyMsg, err := ai_util.AskChatGpt("hello")
	if err != nil {
		panic(err)
	}
	t.Logf("%+v", replyMsg)
}

func Test_OpenAIGenerateImg(t *testing.T) {
	imgResp, err := ai_util.GenerateImage(" A dream of a distant galaxy, by Caspar David Friedrich, matte painting trending on artstation HQ.")
	if err != nil {
		panic(err)
	}

	t.Logf("%+v", imgResp)
}

func Test_OpenAIEditImg(t *testing.T) {
	imgResp, err := ai_util.GenerateImage("Color photo of a beautiful Chinese girl sitting on a traditional wooden bench in a serene garden. She wears a flowing red dress with intricate gold embroidery and a delicate flower in her hair. Her porcelain skin glows in the soft sunlight, and her almond-shaped eyes gaze off into the distance with a look of contemplation. The garden around her is filled with vibrant greenery, blooming flowers, and a peaceful koi pond.\n\nNikon D850, Fujifilm Pro 400H film, 85mm lens, natural light.\n\nDirectors: Ang Lee, Wong Kar-wai\nCinematographers: Christopher Doyle, Mark Lee Ping-bin\nPhotographers: Zhang Jingna, Chen Man\nFashion designers: Guo Pei, Vera Wang, Jason Wu\n—c 10 —ar 2:3")
	if err != nil {
		panic(err)
	}

	t.Logf("%+v", imgResp)
}

func Test_AIAssistant(t *testing.T) {
	assistant, err := ai_util.AskAIAssistant("1+1=?")
	if err != nil {
		panic(err)
	}
	t.Logf("%s", assistant.Text)

	assistant, err = ai_util.AskAIAssistant("这个结果的二次方等于多少", assistant.ID)
	if err != nil {
		panic(err)
	}
	t.Logf("%s", assistant.Text)
}
