package utils

import (
	"strings"
	"time"

	"github.com/c9s/bbgo/pkg/fixedpoint"
	"github.com/dop251/goja"
)

// argEvalTimeout bounds evaluation of LLM-derived expressions. Legitimate
// arguments are single arithmetic expressions that finish in microseconds;
// anything hitting this limit is degenerate (e.g. while(true){}) and must not
// hang the strategy goroutine.
const argEvalTimeout = 500 * time.Millisecond

func ExtractArgs(text string, cmd string) []string {
	rets := make([]string, 0)
	cmdIndex := strings.Index(text, cmd)

	if cmdIndex > 0 {
		subText := text[cmdIndex:]
		argStart := strings.Index(subText, "[")
		argEnd := strings.Index(subText, "]")

		if argStart > 0 && argEnd > 0 && argStart < argEnd {
			argText := subText[argStart+1 : argEnd]
			argTokens := strings.Split(argText, ",")
			for _, token := range argTokens {
				if strings.TrimSpace(token) != "" {
					rets = append(rets, strings.TrimSpace(token))
				}
			}
		}
	}

	return rets
}

// ArgToFixedpoint evaluates an LLM-derived expression on vm. The vm must not
// be shared across goroutines (goja.Runtime is not goroutine-safe) — callers
// should pass a fresh runtime per evaluation.
func ArgToFixedpoint(vm *goja.Runtime, arg string) (*fixedpoint.Value, error) {
	timer := time.AfterFunc(argEvalTimeout, func() {
		vm.Interrupt("arg evaluation timed out")
	})
	defer timer.Stop()

	v, err := vm.RunString(arg)
	if err != nil {
		return nil, err
	}

	// goja exports integer-valued results as int64, not float64.
	switch num := v.Export().(type) {
	case float64:
		val := fixedpoint.NewFromFloat(num)
		return &val, nil
	case int64:
		val := fixedpoint.NewFromInt(num)
		return &val, nil
	}

	return nil, nil
}
