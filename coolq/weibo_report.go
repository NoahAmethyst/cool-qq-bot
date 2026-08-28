package coolq

import (
	"fmt"
	"strings"
	"time"

	"github.com/NoahAmethyst/go-cqhttp/util/top_list"
)

var weiboTagEmoji = map[string]string{
	"当前爆词":  "🔥",
	"好友正在看": "👀",
	"上午霸榜":  "🌅",
	"凌晨霸榜":  "🌙",
}

func formatWeiboTags(tags []string) string {
	formatted := make([]string, 0, len(tags))
	for _, tag := range tags {
		if emoji, ok := weiboTagEmoji[tag]; ok {
			formatted = append(formatted, emoji)
		} else if strings.Contains(tag, "登顶") {
			formatted = append(formatted, "👑")
		} else if tag != "" {
			formatted = append(formatted, tag)
		}
	}
	return strings.Join(formatted, " ")
}

func formatWeiboHotReport(hotList []top_list.WeiboHot, now time.Time) string {
	var content strings.Builder
	content.WriteString("╭─ 微博实时热搜\n")
	content.WriteString("│ ")
	content.WriteString(now.Format("2006-01-02 15:04"))
	content.WriteString("\n╰────────────\n")
	for _, hot := range hotList {
		content.WriteString(fmt.Sprintf("%02d. %s", hot.Rank, hot.Title))
		if tags := formatWeiboTags(hot.Tags); tags != "" {
			content.WriteString("  ")
			content.WriteString(tags)
		}
		content.WriteByte('\n')
	}
	content.WriteString("────────────\n发送 #微博 1,2,3 获取详情")
	return content.String()
}

func formatWeiboHotDetail(hot top_list.WeiboHot) string {
	var content strings.Builder
	content.WriteString(fmt.Sprintf("微博热搜 #%d\n%s", hot.Rank, hot.Title))
	if tags := formatWeiboTags(hot.Tags); tags != "" {
		content.WriteString("\n标签：")
		content.WriteString(tags)
	}
	content.WriteString("\n链接：")
	content.WriteString(hot.Url)
	return content.String()
}
