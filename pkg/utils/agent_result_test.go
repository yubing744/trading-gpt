package utils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseResult(t *testing.T) {
	testDatas := []string{`
	Based on the provided data and the trading strategy, let's analyze the situation step-by-step:

	1. **Market Trend Analysis**:
		 - The current close price is 1.350, which is below the UpBand (1.353) but above the SMA (1.330) and well above the DownBand (1.308). This suggests that the price is in an uptrend but not in the overbought territory according to Bollinger Bands.
		 - The RSI value is 70.362, which is approaching the overbought threshold of 70 but not exceeding it. This indicates strong buying momentum but also suggests caution as the market may be overextended.
		 - The Fear and Greed Index is at 79, which indicates extreme greed in the market. This could signal that the market is due for a correction, as extreme sentiment often precedes market reversals.
	
	2. **Trading Strategy Application**:
		 - Since there are no open positions and the market is showing signs of an uptrend, a long position could be considered.
		 - According to the trading strategy, a trailing stop loss of 3% and a trailing stop profit of 10% should be set. This means if a long position is opened at the current price of 1.350, the stop-loss trigger price would be set at 1.350 * (1 - 0.03) = 1.3095, and the take-profit trigger price would be set at 1.350 * (1 + 0.10) = 1.4850.
	
	3. **Performance Evaluation**:
		 - Continuously monitoring the Bollinger Bands, RSI, and Fear and Greed Index will be crucial to adjust the strategy as needed.
		 - Reflecting on the decision to open a long position, considering the high Fear and Greed Index value, it's important to be cautious and ready to adjust the stop-loss if the market sentiment changes rapidly.
	
	4. **Constructive Criticism**:
		 - While the indicators suggest an uptrend, the high Fear and Greed Index value could be a sign of an impending correction. It's important to be vigilant and not overly optimistic due to the current greed in the market.
	
	5. **Summary**:
		 - The market data suggests a potential for profit by opening a long position, but caution is advised due to the high Fear and Greed Index value. The stop-loss and take-profit levels are set according to the trading strategy.
	
	Based on this analysis, the following JSON response is constructed:
	`,
		"```json",
		`
	{
			"thoughts": {
					"text": "Considering opening a long position based on the current market trend and indicators.",
					"analyze": "The close price is trending upwards, RSI is strong but not overbought, and the Fear and Greed Index is high, indicating greed.",
					"criticism": "Need to be cautious due to the high Fear and Greed Index, which could precede a correction.",
					"speak": "The market is showing signs of an uptrend, and we're considering a long position with a trailing stop loss and profit to maximize returns. However, we'll remain vigilant due to the current market sentiment."
			},
			"action": {
					"name": "exchange.open_long_position",
					"args": {
							"stop_loss_trigger_price": "1.3095",
							"take_profit_trigger_price": "1.4850"
					}
			}
	}`,
		"```",
		`Please note that the actual execution of this strategy should be done with caution and under the supervision of a financial advisor or with a thorough understanding of the risks involved in trading.`,
	}

	ret, err := ParseResult(strings.Join(testDatas, ""))
	assert.NoError(t, err)
	assert.Equal(t, "exchange.open_long_position", ret.Action.Name)
}

func TestParseResult2(t *testing.T) {
	testDatas := ` {
		"thoughts": {
				"text": "Considering opening a long position based on the current market trend and indicators.",
				"analyze": "The close price is trending upwards, RSI is strong but not overbought, and the Fear and Greed Index is high, indicating greed.",
				"criticism": "Need to be cautious due to the high Fear and Greed Index, which could precede a correction.",
				"speak": "The market is showing signs of an uptrend, and we're considering a long position with a trailing stop loss and profit to maximize returns. However, we'll remain vigilant due to the current market sentiment."
		},
		"action": {
				"name": "exchange.open_long_position",
				"args": {
						"stop_loss_trigger_price": "1.3095",
						"take_profit_trigger_price": "1.4850"
				}
		}
} `

	ret, err := ParseResult(testDatas)
	assert.NoError(t, err)
	assert.Equal(t, "exchange.open_long_position", ret.Action.Name)
}

func TestParseResult3(t *testing.T) {
	testDatas := `{
			"thoughts": {
			"text": "The current close price is 1.430, and the RSI value is 10.921, which indicates an oversold condition.",
			"analyze": "The Bollinger Bands UpBand is at 1.509, SMA is at 1.466, DownBand is at 1.423, and the Fear and Greed Index value is 77, which suggests extreme greed.",
			"criticism": "I should have considered closing the position earlier to avoid further losses due to the oversold condition of the asset.",
			"speak": "Based on the current market conditions, I recommend opening a long position with a stop loss at 1.423 and a take profit at 1.509."
			},
			"action": {
			"name": "exchange.open\_long\_position",
			"args": {
			"stop\_loss\_trigger\_price": "1.423",
			"take\_profit\_trigger\_price": "1.509"
			}
			}
			}`

	ret, err := ParseResult(testDatas)
	assert.NoError(t, err)
	assert.Equal(t, "exchange.open_long_position", ret.Action.Name)
}

func TestParseResult4(t *testing.T) {
	testDatas := `{
    "thoughts": {
        "text": "The entity cannot open a new long position because it already has an open long position with the same direction as the signal.",
        "analyze": "The entity cannot open a new long position because it already has an open long position with the same direction as the signal. This is because the entity is not allowed to open multiple long positions with the same direction at the same time.",
        "criticism": "The entity should not have opened a new long position because it already has an open long position with the same direction as the signal. This is a violation of the trading strategy.",
        "speak": "The entity cannot open a new long position because it already has an open long position with the same direction as the signal. Please try to fix this error by closing the existing long position before opening a new one."
    },
    "action": {"name": "exchange.close_position"}
}`

	ret, err := ParseResult(testDatas)
	assert.NoError(t, err)
	assert.Equal(t, "exchange.close_position", ret.Action.Name)
}
