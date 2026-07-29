package coolq

import "testing"

func TestToPlainQQTextRemovesMarkdownAndPreservesEmoji(t *testing.T) {
	input := "# 📊 **市场重点**\n" +
		"- 美元走强\n" +
		"> `风险`：关注波动\n" +
		"[原文](https://wallstreetcn.com/articles/123)\n" +
		"| 资产 | 方向 |\n" +
		"| --- | --- |\n" +
		"| 黄金 | **上涨** |"
	want := "📊 市场重点\n" +
		"美元走强\n" +
		"风险：关注波动\n" +
		"原文\n" +
		"资产   方向\n" +
		"黄金   上涨"

	if got := toPlainQQText(input); got != want {
		t.Fatalf("toPlainQQText() = %q, want %q", got, want)
	}
}
