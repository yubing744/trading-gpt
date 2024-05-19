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
6、The returned JSON format does not support comments

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
