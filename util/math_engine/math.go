package math_engine

import (
	"github.com/dengsgo/math-engine/engine"
	log "github.com/sirupsen/logrus"
	"strings"
)

func CheckMathExpression(expression string) (string, bool) {
	if len(expression) == 0 {
		return "", false
	}
	match := true
	if strings.Contains(expression, "（") || strings.Contains(expression, "）") {
		expression = strings.ReplaceAll(expression, "（", "(")
		expression = strings.ReplaceAll(expression, "）", ")")
	}
	if strings.Contains(expression, "×") || strings.Contains(expression, "÷") {
		expression = strings.ReplaceAll(expression, "×", "*")
		expression = strings.ReplaceAll(expression, "÷", "/")
	}
	if tokens, err := engine.Parse(expression); err != nil {
		return "", false
	} else if len(tokens) == 0 {
		match = false
	}

	return expression, match
}

func Calculate(expression string) (float64, error) {

	r, err := engine.ParseAndExec(expression)
	if err != nil {
		log.Error(err)
	}
	return r, err
}
