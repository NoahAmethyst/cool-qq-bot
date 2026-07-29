# Daily Economic Summary Plain-Text Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the daily Jarvis economic summary use QQ-safe plain text, preserve emoji emphasis, and reject chapters that become empty after Markdown cleanup.

**Architecture:** Extract the weather reporter's existing Markdown cleanup into one package-level QQ text utility shared by weather and economic summaries. Strengthen the summary prompt, then wrap the existing four-marker parser with a sanitizer/validator that runs before message formatting and fan-out.

**Tech Stack:** Go 1.19, standard library `regexp` and `strings`, existing `testing` package, Jarvis summary prompt and QQ reporter flow.

---

## File Structure

- Create `coolq/qq_text.go`: own reusable Markdown-to-QQ-plain-text patterns and conversion.
- Create `coolq/qq_text_test.go`: verify Markdown cleanup and emoji preservation.
- Create `coolq/jarvis_summary_test.go`: verify the prompt contract and sanitized four-chapter parsing.
- Modify `coolq/weather.go`: remove local sanitizer definitions and call the shared utility.
- Modify `coolq/jarvis_summary.go`: strengthen the prompt and add sanitized chapter parsing.
- Modify `coolq/reporter.go`: use sanitized chapters before formatting and fan-out.
- Modify `docs/CODEX_PROJECT_CONTEXT.md`: record the daily summary's plain-text output contract.

### Task 1: Add the daily summary prompt contract

**Files:**
- Create: `coolq/jarvis_summary_test.go`
- Modify: `coolq/jarvis_summary.go:92-110`

- [ ] **Step 1: Write the failing prompt test**

Create `coolq/jarvis_summary_test.go`:

```go
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
```

- [ ] **Step 2: Run the focused test and verify RED**

Run:

```bash
go test ./coolq -run TestBuildWallStreetSummaryPromptRequiresPlainQQTextAndAllowsEmoji -count=1
```

Expected: FAIL because the current prompt does not contain the three plain-text and emoji requirements.

- [ ] **Step 3: Add the minimal prompt requirement**

In `buildWallStreetSummaryPrompt`, insert the new item before the marker protocol and renumber the marker item:

```go
5. 返回适合直接发送到QQ的普通文本。不要使用Markdown，不要使用Markdown标题、项目符号、加粗、斜体、删除线、代码块、表格、引用或Markdown链接。可以使用emoji和普通符号突出章节重点、重要结论、风险与机会。
6. 最终答案必须严格按以下顺序输出四章；每个 marker 必须恰好出现一次，不得改写、遗漏或重复 marker。每个 marker 后只输出该章正文，不要重复章节标题；marker 前可有少量说明但不会被发送：
```

- [ ] **Step 4: Run the focused test and verify GREEN**

Run:

```bash
go test ./coolq -run TestBuildWallStreetSummaryPromptRequiresPlainQQTextAndAllowsEmoji -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit the prompt contract**

```bash
git add coolq/jarvis_summary.go coolq/jarvis_summary_test.go
git commit -m "feat: require plain-text economic summaries"
```

### Task 2: Extract the shared QQ plain-text sanitizer

**Files:**
- Create: `coolq/qq_text.go`
- Create: `coolq/qq_text_test.go`
- Modify: `coolq/weather.go:3-27,108-145,207`

- [ ] **Step 1: Write the failing sanitizer test**

Create `coolq/qq_text_test.go`:

```go
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
```

- [ ] **Step 2: Run the focused test and verify RED**

Run:

```bash
go test ./coolq -run TestToPlainQQTextRemovesMarkdownAndPreservesEmoji -count=1
```

Expected: build FAIL with `undefined: toPlainQQText`.

- [ ] **Step 3: Create the shared utility**

Create `coolq/qq_text.go` by moving the existing regex variables and `plainQQText` behavior from `weather.go`, renaming the function to `toPlainQQText`:

```go
package coolq

import (
	"regexp"
	"strings"
)

var (
	markdownImagePattern  = regexp.MustCompile(`!\[([^\]]*)\]\([^)]+\)`)
	markdownLinkPattern   = regexp.MustCompile(`\[([^\]]+)\]\([^)]+\)`)
	markdownHeadingPrefix = regexp.MustCompile(`^\s{0,3}#{1,6}\s*`)
	markdownQuotePrefix   = regexp.MustCompile(`^\s{0,3}>\s?`)
	markdownListPrefix    = regexp.MustCompile(`^\s{0,3}(?:[-+*]|\d+[.)])\s+`)
	markdownTableDivider  = regexp.MustCompile(`^\s*\|?[\s:|-]+\|?\s*$`)
)

func toPlainQQText(answer string) string {
	answer = strings.ReplaceAll(answer, "\r\n", "\n")
	answer = strings.ReplaceAll(answer, "\r", "\n")
	answer = markdownImagePattern.ReplaceAllString(answer, "$1")
	answer = markdownLinkPattern.ReplaceAllString(answer, "$1")

	lines := strings.Split(answer, "\n")
	result := make([]string, 0, len(lines))
	previousBlank := false
	replacer := strings.NewReplacer(
		"```", "",
		"**", "",
		"__", "",
		"~~", "",
		"`", "",
		"*", "",
		"|", " ",
	)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") || markdownTableDivider.MatchString(trimmed) {
			continue
		}
		line = markdownHeadingPrefix.ReplaceAllString(line, "")
		line = markdownQuotePrefix.ReplaceAllString(line, "")
		line = markdownListPrefix.ReplaceAllString(line, "")
		line = strings.TrimSpace(replacer.Replace(line))

		if line == "" {
			if previousBlank || len(result) == 0 {
				continue
			}
			result = append(result, "")
			previousBlank = true
			continue
		}

		result = append(result, line)
		previousBlank = false
	}

	for len(result) > 0 && result[len(result)-1] == "" {
		result = result[:len(result)-1]
	}
	return strings.TrimSpace(strings.Join(result, "\n"))
}
```

Remove `regexp` and the Markdown regex variables from `weather.go`, delete its old `plainQQText` function, and change the weather call site to:

```go
content := toPlainQQText(response.GetAnswer())
```

- [ ] **Step 4: Format and run sanitizer plus weather package tests**

Run:

```bash
gofmt -w coolq/qq_text.go coolq/qq_text_test.go coolq/weather.go
go test ./coolq -run 'TestToPlainQQTextRemovesMarkdownAndPreservesEmoji|TestBuildWallStreetSummaryPromptRequiresPlainQQTextAndAllowsEmoji' -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit the shared sanitizer**

```bash
git add coolq/qq_text.go coolq/qq_text_test.go coolq/weather.go
git commit -m "refactor: share QQ plain-text sanitizer"
```

### Task 3: Sanitize and validate all economic summary chapters

**Files:**
- Modify: `coolq/jarvis_summary_test.go`
- Modify: `coolq/jarvis_summary.go:123-148`
- Modify: `coolq/reporter.go:431`

- [ ] **Step 1: Write failing sanitized-chapter tests**

Append to `coolq/jarvis_summary_test.go`:

```go
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
```

- [ ] **Step 2: Run the focused tests and verify RED**

Run:

```bash
go test ./coolq -run TestParsePlainJarvisSummary -count=1
```

Expected: build FAIL with `undefined: parsePlainJarvisSummary`.

- [ ] **Step 3: Add the sanitized parser**

Add to `coolq/jarvis_summary.go` immediately after `parseJarvisSummary`:

```go
func parsePlainJarvisSummary(answer string) ([4]string, error) {
	chapters, err := parseJarvisSummary(answer)
	if err != nil {
		return [4]string{}, err
	}

	for index, chapter := range chapters {
		chapters[index] = toPlainQQText(chapter)
		if chapters[index] == "" {
			return [4]string{}, fmt.Errorf("summary chapter %d is empty after sanitization", index+1)
		}
	}
	return chapters, nil
}
```

Change `ReportSummary` in `coolq/reporter.go` to use the sanitized parser:

```go
chapters, err := parsePlainJarvisSummary(response.GetAnswer())
```

Keep the existing `operation=parse_chapters` failure path so no partial summary is sent.

- [ ] **Step 4: Format and run all new summary tests**

Run:

```bash
gofmt -w coolq/jarvis_summary.go coolq/jarvis_summary_test.go coolq/reporter.go
go test ./coolq -run 'TestBuildWallStreetSummaryPrompt|TestParsePlainJarvisSummary|TestToPlainQQText' -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit the sanitized summary pipeline**

```bash
git add coolq/jarvis_summary.go coolq/jarvis_summary_test.go coolq/reporter.go
git commit -m "feat: sanitize daily economic summaries"
```

### Task 4: Synchronize project context and verify the complete change

**Files:**
- Modify: `docs/CODEX_PROJECT_CONTEXT.md:66-82`

- [ ] **Step 1: Update the Wall Street summary flow documentation**

Add this bullet after the fixed four-chapter list:

```markdown
- Summary chapters are sent as QQ-safe plain text rather than Markdown. The prompt permits emoji and ordinary symbols for emphasis, and the bot removes common Markdown headings, list markers, emphasis, code fences, tables, quotes, and links before formatting each chapter. If any sanitized chapter is empty, the whole summary uses the existing failure notification path.
```

- [ ] **Step 2: Run formatting and diff validation**

Run:

```bash
gofmt -w coolq/qq_text.go coolq/qq_text_test.go coolq/weather.go coolq/jarvis_summary.go coolq/jarvis_summary_test.go coolq/reporter.go
git diff --check
```

Expected: no output from `git diff --check`.

- [ ] **Step 3: Run the focused package suite**

Run:

```bash
go test ./coolq -count=1
```

Expected: PASS.

- [ ] **Step 4: Run the broadest practical local suite**

Run:

```bash
go test ./... -count=1
```

Expected: PASS for self-contained packages. If a test under `gotest/` requires network access, cloud services, or credentials, record that specific environmental limitation without exposing sensitive values, then retain the passing `./coolq` result as the focused verification.

- [ ] **Step 5: Review the final diff and repository status**

Run:

```bash
git diff --check
git diff -- coolq/qq_text.go coolq/qq_text_test.go coolq/weather.go coolq/jarvis_summary.go coolq/jarvis_summary_test.go coolq/reporter.go docs/CODEX_PROJECT_CONTEXT.md
git status --short
```

Expected: only files from this plan plus the user's pre-existing `config.yml`, `qq_bot.yml`, and `qrcode.png` changes appear; the user-owned files remain untouched.

- [ ] **Step 6: Commit documentation and verification state**

```bash
git add docs/CODEX_PROJECT_CONTEXT.md
git commit -m "docs: document plain-text economic summaries"
```

This plan does not push, deploy, modify Kubernetes, or touch sensitive runtime files.
