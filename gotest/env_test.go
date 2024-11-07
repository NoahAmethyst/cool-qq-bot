package gotest

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func Test_ListEnv(t *testing.T) {
	env := os.Environ()
	for _, e := range env {
		t.Logf("%+v", e)
	}
}

func Test_Trim(t *testing.T) {
	content := "#模式"
	v := strings.TrimSpace(strings.ReplaceAll(content, "#模式", ""))
	fmt.Print(v)
}
