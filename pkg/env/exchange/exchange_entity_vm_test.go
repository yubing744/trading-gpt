package exchange

import (
	"sync"
	"testing"
	"time"

	"github.com/c9s/bbgo/pkg/fixedpoint"
	"github.com/dop251/goja"
	"github.com/stretchr/testify/assert"
	"github.com/yubing744/trading-gpt/pkg/utils"
)

// TestEvalWithVM_ConcurrentNoRace exercises the shared goja runtime from many
// goroutines through evalWithVM. Without the mutex, goja.Runtime access races
// (issue #93). Run with -race to catch regressions. With the lock, each
// goroutine also gets the correct, isolated result.
func TestEvalWithVM_ConcurrentNoRace(t *testing.T) {
	ent := &ExchangeEntity{vm: goja.New()}

	const workers = 32
	var wg sync.WaitGroup
	wg.Add(workers)

	errs := make([]error, workers)
	results := make([]float64, workers)

	for i := 0; i < workers; i++ {
		go func(idx int) {
			defer wg.Done()

			closePrice := fixedpoint.NewFromInt(int64(100 + idx))
			// ParsePrice sets variables on the shared vm and then evaluates,
			// so the whole set+eval sequence must be serialized by evalWithVM.
			// Use a non-integral multiplier so goja exports a float64 result.
			val, err := ent.evalWithVM(func(vm *goja.Runtime) (*fixedpoint.Value, error) {
				return utils.ParsePrice(vm, nil, closePrice, "last_close * 0.995")
			})
			if err != nil {
				errs[idx] = err
				return
			}
			if val != nil {
				results[idx] = val.Float64()
			}
		}(i)
	}

	wg.Wait()

	for i := 0; i < workers; i++ {
		assert.NoError(t, errs[i])
		expected := float64(100+i) * 0.995
		assert.InDelta(t, expected, results[i], 1e-6, "worker %d got wrong result", i)
	}
}

// TestEvalWithVM_TimeoutDoesNotHang verifies that a runaway script routed
// through the shared runtime is interrupted and does not hang the caller
// (issues #92 + #93 combined path).
func TestEvalWithVM_TimeoutDoesNotHang(t *testing.T) {
	ent := &ExchangeEntity{vm: goja.New()}

	done := make(chan struct{})
	var err error

	go func() {
		_, err = ent.evalWithVM(func(vm *goja.Runtime) (*fixedpoint.Value, error) {
			return utils.ArgToFixedpointWithTimeout(vm, "while (true) {}", 200*time.Millisecond)
		})
		close(done)
	}()

	select {
	case <-done:
		assert.Error(t, err, "runaway script must return an error, not hang")
	case <-time.After(5 * time.Second):
		t.Fatal("evalWithVM hung on a runaway script")
	}

	// The lock must be released after a timeout so the runtime stays usable.
	val, err := ent.evalWithVM(func(vm *goja.Runtime) (*fixedpoint.Value, error) {
		return utils.ArgToFixedpointWithTimeout(vm, "1.5 + 1", time.Second)
	})
	assert.NoError(t, err)
	if assert.NotNil(t, val) {
		assert.InDelta(t, 2.5, val.Float64(), 1e-9)
	}
}
