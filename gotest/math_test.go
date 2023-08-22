package gotest

import (
	"fmt"
	"github.com/Mrs4s/go-cqhttp/util/math_engine"
	"github.com/dengsgo/math-engine/engine"
	"testing"
)

func Test_Math(t *testing.T) {
	s := "1 + 2 * 6 / 4 + (456 - 8 * 9.2) - (2 + 4 ^ 5)"
	// call top level function
	r, err := math_engine.Calculate(s)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%s = %v", s, r)

}

func Test_CheckExpression(t *testing.T) {
	s1 := "20*22+20*8+3.5*4+14*22+100+34*8+40*4"
	s1, _ = math_engine.IsMathExpression(s1)
	t.Logf("%+v", s1)

	if tokens, err := engine.Parse(s1); err != nil {
		t.Error(err)
	} else {
		t.Logf("%+v", tokens)
	}

	s2 := "1 + 2 * 6 / 4 + (456 - 8 * 9.2) - (2 + 4 ^ 5)"
	s2, _ = math_engine.IsMathExpression(s2)
	t.Logf("%+v", s2)

	if tokens, err := engine.Parse(s2); err != nil {
		t.Error(err)
	} else {
		t.Logf("%+v", tokens)
	}
}
