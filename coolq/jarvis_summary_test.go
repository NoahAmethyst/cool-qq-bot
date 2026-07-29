package coolq

import (
	"strings"
	"testing"

	"github.com/NoahAmethyst/go-cqhttp/util/top_list"
)

func TestBuildWallStreetSummaryPromptRequiresPlainQQTextAndAllowsEmoji(t *testing.T) {
	records := map[string][]top_list.WallStreetNews{
		"2026-07-30 12:00": {
			{
				Title:       "测试新闻",
				Content:     "测试摘要",
				Url:         "https://wallstreetcn.com/articles/123",
				CollectedAt: "2026-07-30T12:00:00+08:00",
			},
		},
	}

	prompt, err := buildWallStreetSummaryPrompt(records)
	if err != nil {
		t.Fatalf("buildWallStreetSummaryPrompt() error = %v", err)
	}

	for _, required := range []string{
		"适合直接发送到QQ的普通文本",
		"不要使用Markdown",
		"可以使用emoji和普通符号",
	} {
		if !strings.Contains(prompt, required) {
			t.Errorf("prompt does not contain %q", required)
		}
	}
}

func TestParsePlainJarvisSummarySanitizesEveryChapter(t *testing.T) {
	answer := "<<<CHAPTER_1>>>\n# 📈 **概览**\n" +
		"<<<CHAPTER_2>>>\n- 重点事件\n" +
		"<<<CHAPTER_3>>>\n> 因果链条\n" +
		"<<<CHAPTER_4>>>\n`风险`与机会"

	chapters, err := parsePlainJarvisSummary(answer)
	if err != nil {
		t.Fatalf("parsePlainJarvisSummary() error = %v", err)
	}

	want := [4]string{"📈 概览", "重点事件", "因果链条", "风险与机会"}
	if chapters != want {
		t.Fatalf("parsePlainJarvisSummary() = %#v, want %#v", chapters, want)
	}
}

func TestParsePlainJarvisSummaryRejectsChapterEmptyAfterCleanup(t *testing.T) {
	answer := "<<<CHAPTER_1>>>\n---\n" +
		"<<<CHAPTER_2>>>\n重点事件\n" +
		"<<<CHAPTER_3>>>\n因果链条\n" +
		"<<<CHAPTER_4>>>\n风险与机会"

	if _, err := parsePlainJarvisSummary(answer); err == nil {
		t.Fatal("parsePlainJarvisSummary() error = nil, want sanitized-empty chapter error")
	}
}
