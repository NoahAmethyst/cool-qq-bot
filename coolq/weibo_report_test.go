package coolq

import (
	"testing"
	"time"

	"github.com/NoahAmethyst/go-cqhttp/util/top_list"
)

func TestFormatWeiboTagsUsesEmojiForKnownLabels(t *testing.T) {
	got := formatWeiboTags([]string{"当前爆词", "好友正在看", "上午霸榜", "凌晨霸榜", "04:06登顶"})
	const want = "🔥 👀 🌅 🌙 👑"
	if got != want {
		t.Fatalf("formatWeiboTags() = %q, want %q", got, want)
	}
}

func TestFormatWeiboTagsKeepsUnknownLabels(t *testing.T) {
	const want = "新标签"
	if got := formatWeiboTags([]string{want}); got != want {
		t.Fatalf("formatWeiboTags() = %q, want %q", got, want)
	}
}

func TestFormatWeiboHotReportUsesCompactRows(t *testing.T) {
	now := time.Date(2026, 8, 28, 10, 30, 0, 0, time.Local)
	got := formatWeiboHotReport([]top_list.WeiboHot{{
		Rank:  1,
		Title: "标题",
		Tags:  []string{"当前爆词"},
	}}, now)
	const want = "╭─ 微博实时热搜\n│ 2026-08-28 10:30\n╰────────────\n01. 标题  🔥\n────────────\n发送 #微博 1,2,3 获取详情"
	if got != want {
		t.Fatalf("formatWeiboHotReport() = %q, want %q", got, want)
	}
}

func TestFormatWeiboHotReportOmitsSpacingForEmptyTags(t *testing.T) {
	now := time.Date(2026, 8, 28, 10, 30, 0, 0, time.Local)
	got := formatWeiboHotReport([]top_list.WeiboHot{{Rank: 2, Title: "无标签标题"}}, now)
	const want = "╭─ 微博实时热搜\n│ 2026-08-28 10:30\n╰────────────\n02. 无标签标题\n────────────\n发送 #微博 1,2,3 获取详情"
	if got != want {
		t.Fatalf("formatWeiboHotReport() = %q, want %q", got, want)
	}
}

func TestFormatWeiboHotDetailIncludesTagsAndLink(t *testing.T) {
	got := formatWeiboHotDetail(top_list.WeiboHot{
		Rank:  1,
		Title: "标题",
		Tags:  []string{"新标签"},
		Url:   "https://example.test",
	})
	const want = "微博热搜 #1\n标题\n标签：新标签\n链接：https://example.test"
	if got != want {
		t.Fatalf("formatWeiboHotDetail() = %q, want %q", got, want)
	}
}
