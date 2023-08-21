package math_engine

import (
	"github.com/dengsgo/math-engine/engine"
	"github.com/rs/zerolog/log"
	"regexp"
	"strings"
)

func IsMathExpression(expression string) bool {
	if strings.Contains(expression, "（") || strings.Contains(expression, "）") {
		expression = strings.ReplaceAll(expression, "（", "(")
		expression = strings.ReplaceAll(expression, "）", ")")
	}
	pattern := `^[-+*×÷/()\d\s]+$`
	match, _ := regexp.MatchString(pattern, expression)
	return match
}

func Calculate(expression string) (float64, error) {
	if strings.Contains(expression, "×") || strings.Contains(expression, "÷") {
		expression = strings.ReplaceAll(expression, "×", "*")
		expression = strings.ReplaceAll(expression, "÷", "/")
	}

	r, err := engine.ParseAndExec(expression)
	if err != nil {
		log.Err(err)
	}
	return r, err
}
