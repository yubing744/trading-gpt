package prompt

var ThoughtTpl = `{{if .MemoryEnabled}}
=== Trading Memory ===
{{if .Memory}}{{.Memory}}{{else}}No previous memory available.{{end}}

=== Memory Management ===
Memory word limit: {{.MaxWords}} words
Current memory usage: {{.CurrentWords}} words ({{.MemoryUsagePercent}}% of limit)
{{if .MemoryUsagePercent | lt 50}}
💡 Memory has plenty of space - you can add new insights and experiences to expand your knowledge base.
{{else if .MemoryUsagePercent | lt 80}}
⚠️ Memory is getting full - focus on adding only the most important new insights while keeping content concise.
{{else if .MemoryUsagePercent | lt 95}}
🚨 Memory is nearly full - prioritize consolidating and summarizing existing knowledge rather than adding new content.
{{else}}
🔥 Memory is at critical capacity - you MUST consolidate, summarize, and remove less critical information to make room for essential new insights.
{{end}}

IMPORTANT: The strategy runs in cycles, and your memory resets at the beginning of each cycle. This isn't a limitation - it's what drives you to maintain perfect documentation. After each reset, you rely ENTIRELY on your Memory Part to understand the project and continue work effectively. Each cycle, you must output complete memory within the word limit to maintain continuity.

{{end}}
Analyze the data provided above, and step-by-step consider the only executable trade command based on the trading strategy provided below to maximize user profit.

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

{{if .MemoryEnabled}}
You should only respond in JSON format as described below, no other explanation is required
Response Format: 
{
    "thoughts": {
        "plan": "analysis steps",
        "analyze": "step-by-step analysis", 
        "detail": "output detailed calculation process",
        "reflection": "comprehensive self-criticism including: 1) trade execution analysis, 2) strategy effectiveness evaluation, 3) market condition adaptation, 4) risk management review, 5) strategy improvement suggestions",
        "speak": "thoughts summary to say to user"
    },
    "action": {"name": "command name", "args": {"arg name": "value"}},
    "memory": {"content": "memory content to save, keep concise and within reasonable word limit"}
}
{{else}}
You should only respond in JSON format as described below, no other explanation is required
Response Format: 
{
    "thoughts": {
        "plan": "analysis steps",
        "analyze": "step-by-step analysis",
        "detail": "output detailed calculation process",
        "reflection": "comprehensive self-criticism including: 1) trade execution analysis, 2) strategy effectiveness evaluation, 3) market condition adaptation, 4) risk management review, 5) strategy improvement suggestions",
        "speak": "thoughts summary to say to user"
    },
    "action": {"name": "command name", "args": {"arg name": "value"}}
}
{{end}}

Ensure the response can be parsed by golang json.Unmarshal
`

// TradeReflectionTpl is a template for generating reflections on closed trading positions
var TradeReflectionTpl = `You are an expert trading advisor analyzing a closed trading position. Please provide a thoughtful reflection on this trade, focusing on:
1. Analysis of the entry and exit points
2. Performance evaluation (profit/loss analysis)
3. What went well in this trade
4. What could have been improved
5. Lessons learned and recommendations for future similar trades

Trade details:
- Symbol: {{.Symbol}}
- Strategy ID: {{.StrategyID}}
- Entry Price: {{.EntryPrice}}
- Exit Price: {{.ExitPrice}}
- Quantity: {{.Quantity}}
- Profit/Loss: {{.ProfitAndLoss}}
- Close Reason: {{.CloseReason}}
- Close Time: {{.Timestamp}}

Please format your response as a structured markdown document with clear headings and bullet points. This reflection will be saved to the memory bank for future reference in trading decisions.
`
