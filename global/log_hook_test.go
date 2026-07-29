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
