package gotest

import (
	"fmt"
	"github.com/dengsgo/math-engine/engine"
	"testing"
)

func Test_Math(t *testing.T) {
	s := "1 + 2 * 6 / 4 + (456 - 8 * 9.2) - (2 + 4 ^ 5)"
	// call top level function
	r, err := engine.ParseAndExec(s)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%s = %v", s, r)

}
