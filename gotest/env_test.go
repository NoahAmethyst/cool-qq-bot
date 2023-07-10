package gotest

import (
	"os"
	"testing"
)

func Test_ListEnv(t *testing.T) {
	env := os.Environ()
	for _, e := range env {
		t.Logf("%+v", e)
	}
}
