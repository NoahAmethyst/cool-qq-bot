package coolq

import (
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/NoahAmethyst/go-cqhttp/util/content_util"
	"github.com/NoahAmethyst/go-cqhttp/util/top_list"
)

var jarvisSummaryMarkers = []string{
	"<<<CHAPTER_1>>>",
	"<<<CHAPTER_2>>>",
	"<<<CHAPTER_3>>>",
	"<<<CHAPTER_4>>>",
	"<<<CHAPTER_5>>>",
}

var (
	jarvisSummaryMarkerLinePattern = regexp.MustCompile(`(?im)^[ \t]*<<<[ \t]*chapter[ \t]*_[ \t]*[1-5][ \t]*>>>[ \t]*\r?\n?`)
	jarvisSummaryMarkerPattern     = regexp.MustCompile(`(?i)<<<\s*chapter\s*_\s*[1-5]\s*>>>`)
)

var jarvisSummaryChapterTitles = []string{
	"今日全球经济形势概览",
	"重点事件、原文补充与影响分析",
	"跨日时间线及因果传导",
	"未来展望、风险与机会",
	"投资建议/注意事项",
}

type wallStreetSummaryItem struct {
	collectedAt string
	title       string
	content     string
	sourceURL   string
	sortTime    time.Time
}

func isTrustedWallStreetURL(rawURL string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.User != nil || !strings.EqualFold(parsed.Scheme, "https") {
		return false
	}

	host := strings.ToLower(parsed.Hostname())
	return host == "wallstreetcn.com" || strings.HasSuffix(host, ".wallstreetcn.com")
}

func buildWallStreetSummaryPrompt(records map[string][]top_list.WallStreetNews) (string, error) {
	items := make([]wallStreetSummaryItem, 0)
	seenURLs := make(map[string]struct{})

	for recordTime, newsList := range records {
		for _, news := range newsList {
			if !isTrustedWallStreetURL(news.Url) {
				continue
			}
			if _, exists := seenURLs[news.Url]; exists {
				continue
			}
			seenURLs[news.Url] = struct{}{}

			collectedAt := strings.TrimSpace(news.CollectedAt)
			if collectedAt == "" {
				collectedAt = strings.TrimSpace(recordTime)
			}
			items = append(items, wallStreetSummaryItem{
				collectedAt: collectedAt,
				title:       news.Title,
				content:     news.Content,
				sourceURL:   news.Url,
				sortTime:    parseWallStreetNewsTime(collectedAt),
			})
		}
	}

	if len(items) == 0 {
		return buildWallStreetKnowledgeSummaryPrompt(), nil
	}

	sort.SliceStable(items, func(i, j int) bool {
		if items[i].sortTime.Equal(items[j].sortTime) {
			return items[i].collectedAt < items[j].collectedAt
		}
		return items[i].sortTime.Before(items[j].sortTime)
	})

	var prompt strings.Builder
	prompt.WriteString("以下是今日采集的华尔街见闻新闻材料。请使用与本次会话相同用户身份下已积累的知识库和过往新闻，将今日事件与历史事件串联分析。\n\n")
	for _, item := range items {
		prompt.WriteString(fmt.Sprintf(
			"新闻时间：%s\n标题：%s\n摘要：%s\n原文链接：%s\n\n",
			item.collectedAt,
			item.title,
			item.content,
			item.sourceURL,
		))
	}

	prompt.WriteString(`总结要求：
1. 只在本次总结阶段，选择性调用 web_scrape 读取与核心结论直接相关的上述原文链接；不要机械读取全部链接。
2. 必须检索同一用户身份知识库中的历史材料和过往新闻，写出明确日期，串联跨日时间线与因果传导，不得只罗列单日新闻。
3. 区分可核验事实与分析判断；事实注明事件日期和依据，分析、推测、风险或机会须明确标识。
4. 专业、严谨、客观，重点解释全球市场、政策、行业和资产价格之间的影响路径。
5. 最终答案必须严格按以下顺序输出五章；每个 marker 必须恰好出现一次，不得改写、遗漏或重复 marker。每个 marker 后只输出该章正文，不要重复章节标题；marker 前可有少量说明但不会被发送：
<<<CHAPTER_1>>> 对应“今日全球经济形势概览”
<<<CHAPTER_2>>> 对应“重点事件、原文补充与影响分析”
<<<CHAPTER_3>>> 对应“跨日时间线及因果传导”
<<<CHAPTER_4>>> 对应“未来展望、风险与机会”
<<<CHAPTER_5>>> 对应“投资建议/注意事项”
`)

	return prompt.String(), nil
}

func buildWallStreetKnowledgeSummaryPrompt() string {
	return `今日本地缓存没有新的华尔街见闻经济新闻材料。请不要因此停止总结。

请检索同一用户身份知识库中的历史经济材料和过往新闻，基于最近一段时间已经保存到知识库中的经济内容，生成一份面向今天的经济总结。需要说明本次总结主要依据历史知识库材料，而不是今日新增本地缓存。

总结要求：
1. 必须检索同一用户身份知识库中的历史经济材料和过往新闻，写出明确日期，串联跨日时间线与因果传导，不得只罗列单日新闻。
2. 区分可核验事实与分析判断；事实注明事件日期和依据，分析、推测、风险或机会须明确标识。
3. 专业、严谨、客观，重点解释全球市场、政策、行业和资产价格之间的影响路径。
4. 最终答案必须严格按以下顺序输出五章；每个 marker 必须恰好出现一次，不得改写、遗漏或重复 marker。每个 marker 后只输出该章正文，不要重复章节标题；marker 前可有少量说明但不会被发送：
<<<CHAPTER_1>>> 对应“今日全球经济形势概览”
<<<CHAPTER_2>>> 对应“重点事件、原文补充与影响分析”
<<<CHAPTER_3>>> 对应“跨日时间线及因果传导”
<<<CHAPTER_4>>> 对应“未来展望、风险与机会”
<<<CHAPTER_5>>> 对应“投资建议/注意事项”
`
}

func parseWallStreetNewsTime(value string) time.Time {
	for _, layout := range []string{time.RFC3339, "2006-01-02 15:04"} {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed
		}
	}
	return time.Time{}
}

func parseJarvisSummary(answer string) ([]string, error) {
	chapters := make([]string, len(jarvisSummaryMarkers))
	positions := make([]int, len(jarvisSummaryMarkers))
	for i, marker := range jarvisSummaryMarkers {
		if strings.Count(answer, marker) != 1 {
			return chapters, fmt.Errorf("summary marker %d count is invalid", i+1)
		}
		positions[i] = strings.Index(answer, marker)
		if i > 0 && positions[i] <= positions[i-1] {
			return chapters, fmt.Errorf("summary marker order is invalid")
		}
	}

	for i, marker := range jarvisSummaryMarkers {
		start := positions[i] + len(marker)
		end := len(answer)
		if i+1 < len(jarvisSummaryMarkers) {
			end = positions[i+1]
		}
		chapters[i] = strings.TrimSpace(answer[start:end])
		if chapters[i] == "" {
			return nil, fmt.Errorf("summary chapter %d is empty", i+1)
		}
	}
	return chapters, nil
}

func parsePlainJarvisSummary(answer string) ([]string, error) {
	chapters, err := parseJarvisSummary(answer)
	if err != nil {
		return nil, err
	}

	for index, chapter := range chapters {
		chapters[index] = toPlainQQText(chapter)
		if chapters[index] == "" {
			return nil, fmt.Errorf("summary chapter %d is empty after sanitization", index+1)
		}
	}
	return chapters, nil
}

func jarvisSummaryMarkerCounts(answer string) string {
	counts := make([]string, 0, len(jarvisSummaryMarkers))
	for i, marker := range jarvisSummaryMarkers {
		counts = append(counts, fmt.Sprintf("%d:%d", i+1, strings.Count(answer, marker)))
	}
	return strings.Join(counts, ",")
}

func formatJarvisSummaryChapter(index int, content string) []string {
	const maxMessageRunes = 1000

	title := jarvisSummaryChapterTitles[index]
	shortHeader := fmt.Sprintf("第%s章：%s\n", chineseChapterNumber(index), title)
	if utf8.RuneCountInString(shortHeader+content) <= maxMessageRunes {
		return []string{shortHeader + content}
	}

	const pageHeaderBudget = 64
	fragmentLimit := maxMessageRunes - pageHeaderBudget
	for {
		fragments := content_util.SplitLongMessage(content, fragmentLimit)
		messages := make([]string, 0, len(fragments))
		nextFragmentLimit := fragmentLimit
		for fragmentIndex, fragment := range fragments {
			header := fmt.Sprintf(
				"【第%s章 %d/%d】\n",
				chineseChapterNumber(index),
				fragmentIndex+1,
				len(fragments),
			)
			pageMessage := header + fragment
			if utf8.RuneCountInString(pageMessage) > maxMessageRunes {
				allowedRunes := maxMessageRunes - utf8.RuneCountInString(header)
				if allowedRunes < nextFragmentLimit {
					nextFragmentLimit = allowedRunes
				}
			}
			messages = append(messages, pageMessage)
		}
		if nextFragmentLimit >= fragmentLimit {
			return messages
		}
		fragmentLimit = nextFragmentLimit
	}
}

// formatJarvisSummaryFallback preserves a non-empty Jarvis response whose chapter
// markers cannot be parsed, while keeping every outbound QQ message within its limit.
func formatJarvisSummaryFallback(answer string) []string {
	for _, marker := range jarvisSummaryMarkers {
		answer = strings.ReplaceAll(answer, marker+"\r\n", "")
		answer = strings.ReplaceAll(answer, marker+"\n", "")
		answer = strings.ReplaceAll(answer, marker, "")
	}
	answer = jarvisSummaryMarkerLinePattern.ReplaceAllString(answer, "")
	answer = jarvisSummaryMarkerPattern.ReplaceAllString(answer, "")

	content := toPlainQQText(answer)
	if content == "" {
		return nil
	}

	const maxMessageRunes = 1000
	const pageContentRunes = maxMessageRunes - 64
	const singleHeader = "经济总结\n"
	if utf8.RuneCountInString(singleHeader+content) <= maxMessageRunes {
		return []string{singleHeader + content}
	}

	fragments := splitJarvisSummaryFallbackContent(content, pageContentRunes)
	messages := make([]string, 0, len(fragments))
	for index, fragment := range fragments {
		messages = append(messages, fmt.Sprintf("【经济总结 %d/%d】\n%s", index+1, len(fragments), fragment))
	}
	return messages
}

// formatJarvisSummaryForDelivery returns chapter-formatted messages when the
// response meets the optional five-chapter convention. Otherwise it preserves
// the sanitized response through the plain-text fallback path.
func formatJarvisSummaryForDelivery(answer string) ([]string, bool, error) {
	chapters, err := parsePlainJarvisSummary(answer)
	if err != nil {
		return formatJarvisSummaryFallback(answer), false, err
	}

	messages := make([]string, 0, len(chapters))
	for index, chapter := range chapters {
		messages = append(messages, formatJarvisSummaryChapter(index, chapter)...)
	}
	return messages, true, nil
}

func splitJarvisSummaryFallbackContent(content string, maxRunes int) []string {
	runes := []rune(content)
	fragments := make([]string, 0, (len(runes)+maxRunes-1)/maxRunes)
	for start := 0; start < len(runes); start += maxRunes {
		end := start + maxRunes
		if end > len(runes) {
			end = len(runes)
		}
		fragments = append(fragments, string(runes[start:end]))
	}
	return fragments
}

func chineseChapterNumber(index int) string {
	return [...]string{"一", "二", "三", "四", "五"}[index]
}
