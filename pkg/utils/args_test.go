package utils

import (
	"testing"

	"github.com/dop251/goja"
	"github.com/stretchr/testify/assert"
)

func TestExtractArgs(t *testing.T) {
	text := `
	Based on the provided data, the current close price is lower than the average cost, indicating a loss on the long position. However, the market is showing signs of a potential uptrend as the current close price is higher than the previous close price and the high price is also increasing. 

Therefore, I recommend issuing the following command to the entity: 

/open_long_position [2.235, 2.285]

This command opens a new long position with a stop loss at 2.235 and a take profit at 2.285. This will allow the entity to potentially profit from the expected uptrend while also limiting potential losses if the trend does not continue.
	`
	args := ExtractArgs(text, "/open_long_position")
	assert.Equal(t, "2.235", args[0])
	assert.Equal(t, "2.285", args[1])
}

func TestExtractArgsWithoutArgs(t *testing.T) {
	text := `
	Based on the provided data, the current close price is lower than the average cost, indicating a loss on the long position. However, the market is showing signs of a potential uptrend as the current close price is higher than the previous close price and the high price is also increasing. 

Therefore, I recommend issuing the following command to the entity: 

/open_long_position

This command opens a new long position with a stop loss at 2.235 and a take profit at 2.285. This will allow the entity to potentially profit from the expected uptrend while also limiting potential losses if the trend does not continue.
	`
	args := ExtractArgs(text, "/open_long_position")
	assert.Len(t, args, 0)
}

func TestExtractArgsWithEmptyArgs(t *testing.T) {
	text := `
	Based on the provided data, the current close price is lower than the average cost, indicating a loss on the long position. However, the market is showing signs of a potential uptrend as the current close price is higher than the previous close price and the high price is also increasing. 

Therefore, I recommend issuing the following command to the entity: 

/open_long_position []

This command opens a new long position with a stop loss at 2.235 and a take profit at 2.285. This will allow the entity to potentially profit from the expected uptrend while also limiting potential losses if the trend does not continue.
	`
	args := ExtractArgs(text, "/open_long_position")
	assert.Len(t, args, 0)
}

func TestArgToFixedpoint(t *testing.T) {
	vm := goja.New()
	val, err := ArgToFixedpoint(vm, "1.22")
	assert.NoError(t, err)
	assert.Equal(t, 1.22, val.Float64())
}

func TestArgToFixedpointWithExpress(t *testing.T) {
	vm := goja.New()
	val, err := ArgToFixedpoint(vm, "1.22*2")
	assert.NoError(t, err)
	assert.Equal(t, 2.44, val.Float64())
}
