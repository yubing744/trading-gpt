package utils

import (
	"strings"
	"time"

	"github.com/c9s/bbgo/pkg/fixedpoint"
	"github.com/dop251/goja"
)

// DefaultEvalTimeout bounds how long a single goja expression evaluation may run
// before it is interrupted. It guards the decision loop against malicious,
// pathological, or infinite-loop scripts produced by the LLM (issue #92).
// It is a variable so it can be tuned by operators if needed.
var DefaultEvalTimeout = 5 * time.Second

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

// ArgToFixedpoint evaluates a JavaScript expression using the provided goja
// runtime and converts the result to a fixedpoint value. Evaluation is bounded
// by DefaultEvalTimeout so a runaway (e.g. infinite-loop) script cannot hang the
// caller (issue #92).
//
// NOTE: goja.Runtime is NOT safe for concurrent use. Callers that share a single
// runtime across goroutines must serialize access to it (see
// ExchangeEntity.vmMu / evalWithVM), issue #93.
func ArgToFixedpoint(vm *goja.Runtime, arg string) (*fixedpoint.Value, error) {
	return ArgToFixedpointWithTimeout(vm, arg, DefaultEvalTimeout)
}

// ArgToFixedpointWithTimeout is like ArgToFixedpoint but with an explicit
// evaluation timeout. A non-positive timeout disables the timeout guard.
//
// The timeout is enforced via goja's Interrupt mechanism: a timer fires
// vm.Interrupt from another goroutine, which causes the in-flight RunString to
// return an *goja.InterruptedError instead of hanging. Note that Interrupt only
// affects JavaScript execution, not native Go built-ins.
func ArgToFixedpointWithTimeout(vm *goja.Runtime, arg string, timeout time.Duration) (*fixedpoint.Value, error) {
	var timer *time.Timer
	if timeout > 0 {
		timer = time.AfterFunc(timeout, func() {
			vm.Interrupt("execution timeout")
		})
	}

	v, err := vm.RunString(arg)

	if timer != nil {
		timer.Stop()
		// Reset the interrupt flag so the (possibly shared) runtime is safe to
		// reuse, even if the timer fired right as RunString was returning.
		vm.ClearInterrupt()
	}

	if err != nil {
		return nil, err
	}

	num, ok := v.Export().(float64)
	if ok {
		val := fixedpoint.NewFromFloat(num)
		return &val, nil
	}

	return nil, nil
}
