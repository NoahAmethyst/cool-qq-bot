# Log Level Color Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Render only the `INFO` label in green while keeping complete `WARN` lines yellow and complete `ERROR` lines red.

**Architecture:** Keep the existing `global.LogFormat` and make ANSI sequence placement conditional on the log level. `INFO` receives a green sequence immediately before its label and a reset immediately after it; every other colored level retains the existing whole-line sequence placement. The existing `EnableColor` boundary continues to keep file logs and color-disabled console logs free of ANSI sequences.

**Tech Stack:** Go 1.19, Logrus formatter API, standard `testing` package

---

## File structure

- Create `global/log_hook_test.go`: deterministic formatter tests for label-only, whole-line, and disabled coloring.
- Modify `global/log_hook.go`: change the INFO color to green and place INFO color codes around only its level label.

### Task 1: Implement mixed log-level color scopes

**Files:**
- Create: `global/log_hook_test.go`
- Modify: `global/log_hook.go:179-233`

- [ ] **Step 1: Write the failing INFO label test**

Create `global/log_hook_test.go` with the first regression test:

```go
package global

import (
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestLogFormatColorsOnlyInfoLabel(t *testing.T) {
	entry := &logrus.Entry{
		Level:   logrus.InfoLevel,
		Time:    time.Date(2026, time.July, 29, 12, 34, 56, 0, time.UTC),
		Message: "ready",
	}

	got, err := (LogFormat{EnableColor: true}).Format(entry)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	want := "[2026-07-29 12:34:56] [\x1b[32mINFO\x1b[0m]: ready \n"
	if string(got) != want {
		t.Fatalf("Format() = %q, want %q", got, want)
	}
}
```

- [ ] **Step 2: Run the INFO test and verify RED**

Run:

```bash
go test ./global -run '^TestLogFormatColorsOnlyInfoLabel$' -count=1 -v
```

Expected: FAIL because current output starts with the white sequence
`\x1b[37m` and keeps that color active for the complete line instead of
placing green around only `INFO`.

- [ ] **Step 3: Implement the minimal formatter change**

In `global/log_hook.go`, replace `LogFormat.Format` with:

```go
// Format implements logrus.Formatter
func (f LogFormat) Format(entry *logrus.Entry) ([]byte, error) {
	buf := NewBuffer()
	defer PutBuffer(buf)

	colorInfoLabelOnly := f.EnableColor && entry.Level == logrus.InfoLevel
	if f.EnableColor && !colorInfoLabelOnly {
		buf.WriteString(GetLogLevelColorCode(entry.Level))
	}

	buf.WriteByte('[')
	buf.WriteString(entry.Time.Format("2006-01-02 15:04:05"))
	buf.WriteString("] [")
	if colorInfoLabelOnly {
		buf.WriteString(GetLogLevelColorCode(entry.Level))
	}
	buf.WriteString(strings.ToUpper(entry.Level.String()))
	if colorInfoLabelOnly {
		buf.WriteString(colorReset)
	}
	buf.WriteString("]: ")
	buf.WriteString(entry.Message)
	buf.WriteString(" \n")

	if f.EnableColor && !colorInfoLabelOnly {
		buf.WriteString(colorReset)
	}

	ret := make([]byte, len(buf.Bytes()))
	copy(ret, buf.Bytes()) // copy buffer
	return ret, nil
}
```

In the same file, change only the INFO constant:

```go
colorCodeInfo = "\x1b[32m" // color.Style{color.Green}.String()
```

- [ ] **Step 4: Run the INFO test and verify GREEN**

Run:

```bash
go test ./global -run '^TestLogFormatColorsOnlyInfoLabel$' -count=1 -v
```

Expected: PASS.

- [ ] **Step 5: Add preservation tests for WARN, ERROR, and disabled colors**

Append these independent tests to `global/log_hook_test.go`:

```go
func TestLogFormatColorsEntireWarnLine(t *testing.T) {
	entry := &logrus.Entry{
		Level:   logrus.WarnLevel,
		Time:    time.Date(2026, time.July, 29, 12, 34, 56, 0, time.UTC),
		Message: "caution",
	}

	got, err := (LogFormat{EnableColor: true}).Format(entry)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	want := "\x1b[33m[2026-07-29 12:34:56] [WARNING]: caution \n\x1b[0m"
	if string(got) != want {
		t.Fatalf("Format() = %q, want %q", got, want)
	}
}

func TestLogFormatColorsEntireErrorLine(t *testing.T) {
	entry := &logrus.Entry{
		Level:   logrus.ErrorLevel,
		Time:    time.Date(2026, time.July, 29, 12, 34, 56, 0, time.UTC),
		Message: "failed",
	}

	got, err := (LogFormat{EnableColor: true}).Format(entry)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	want := "\x1b[31m[2026-07-29 12:34:56] [ERROR]: failed \n\x1b[0m"
	if string(got) != want {
		t.Fatalf("Format() = %q, want %q", got, want)
	}
}

func TestLogFormatWithoutColorContainsNoANSI(t *testing.T) {
	entry := &logrus.Entry{
		Level:   logrus.InfoLevel,
		Time:    time.Date(2026, time.July, 29, 12, 34, 56, 0, time.UTC),
		Message: "ready",
	}

	got, err := (LogFormat{EnableColor: false}).Format(entry)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	want := "[2026-07-29 12:34:56] [INFO]: ready \n"
	if string(got) != want {
		t.Fatalf("Format() = %q, want %q", got, want)
	}
}
```

Logrus renders `WarnLevel.String()` as `warning`, so the exact expected WARN
label remains `[WARNING]`, matching current formatter behavior.

- [ ] **Step 6: Format the changed Go files**

Run:

```bash
gofmt -w global/log_hook.go global/log_hook_test.go
```

Expected: both files are formatted without output.

- [ ] **Step 7: Run focused formatter tests**

Run:

```bash
go test ./global -run '^TestLogFormat' -count=1 -v
```

Expected: all four `TestLogFormat...` tests PASS.

- [ ] **Step 8: Run the complete affected-package test suite**

Run:

```bash
go test ./global -count=1
```

Expected: PASS. This is the broadest deterministic suite for the changed
package; repository context records unrelated blockers in `go test ./...`.

- [ ] **Step 9: Inspect the final patch**

Run:

```bash
git diff --check
git diff -- global/log_hook.go global/log_hook_test.go
```

Expected: `git diff --check` exits successfully; the diff contains only the
mixed-scope formatter change and its four focused tests.

- [ ] **Step 10: Commit the implementation**

Run:

```bash
git add global/log_hook.go global/log_hook_test.go
git diff --cached --check
git diff --cached --name-only
git commit -m "feat: color log levels by severity"
```

Expected: the cached file list contains only `global/log_hook.go` and
`global/log_hook_test.go`, and the commit succeeds without including the
user's pre-existing workspace changes.
