package prompt

var ThoughtTpl = `Analyze the data provided above, and step-by-step consider the only executable trade command based on the trading strategy provided below to maximize user profit.

Commands:
{{- range $index, $item := .ActionTips}}
{{add $index 1}}. {{$item}}
{{- end}}

Trading strategy:
{{.Strategy}}

{{if .StrategyAttentionPoints}}
Strategy points of attention:
{{- range $index, $item := .StrategyAttentionPoints}}
{{add $index 1}}. {{$item}}
{{- end}}
{{end}}

Constraints:
1. Exclusively use the commands listed in double quotes e.g. "command name"
2. The command's parameters only support strings. If the parameters are of other types, please convert them all to strings.
3. Command parameters do not support variables or expressions. Please fill them in after calculating them step by step.
4、Be careful not to open positions repeatedly. If you need to adjust the take profit and stop loss, please use the exchange.update_position command.
5、The analyze statement can be very long to ensure that the reasoning process of the analysis is rigorous.
6、When comparing two numbers, if a digit in the decimal part is already greater, there's no need to compare the subsequent digits.
7、The returned JSON format does not support comments

You should only respond in JSON format as described below, no other explanation is required
Response Format: 
{
    "thoughts": {
        "plan": "analysis steps",
        "analyze": "step-by-step analysis",
        "detail": "output detailed calculation process",
        "reflection": "constructive self-criticism",
        "speak": "thoughts summary to say to user"
    },
    "action": {"name": "command name", "args": {"arg name": "value"}}
}

Ensure the response can be parsed by golang json.Unmarshal
`

// ReflectionTpl provides a template for structured self-reflection and criticism
var ReflectionTpl = `Reflect critically on the decision-making process and the chosen actions based on the trading strategy.

Review your recent trading actions and evaluate:
1. Adherence to the trading strategy principles
2. Quality of data analysis and interpretation
3. Risk management effectiveness
4. Emotional control in decision making
5. Lessons learned and areas for improvement

Constraints:
1. Be honest and constructive in your self-criticism
2. Focus on identifying specific patterns or biases that may affect decision quality
3. Consider both successful decisions and potential mistakes
4. Highlight alternative approaches that might have been more effective
5. Provide actionable insights for improving future trades

You should only respond in JSON format as described below
Response Format:
{
    "reflection": {
        "strategy_adherence": "analysis of how well decisions aligned with strategy",
        "data_analysis": "evaluation of data interpretation quality",
        "risk_management": "assessment of position sizing and risk controls",
        "emotional_factors": "identification of any emotional biases in decisions",
        "improvement_areas": "specific aspects to improve",
        "action_items": "concrete steps to enhance trading performance"
    }
}

Ensure the response can be parsed by golang json.Unmarshal
`
