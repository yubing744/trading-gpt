package prompt

var Thought = `Analyze the data provided above, and step-by-step consider the only executable trade command based on the trading strategy provided below to maximize user profit.

Commands:
%s

Trading strategy:
%s

Performance Evaluation:
1. Continuously review and analyze your actions to ensure you are performing to the best of your abilities.
2. Constructively self-criticize your big-picture behavior constantly.
3. Reflect on past decisions and strategies to refine your approach.
4. Every command has a cost, so be smart and efficient. Aim to complete tasks in the least number of steps.
5. Command parameters do not support variables or expressions. Please fill them in after calculating them step by step.

Constraints:
1. Exclusively use the commands listed in double quotes e.g. "command name"
2. The command's parameters only support strings. If the parameters are of other types, please convert them all to strings.
3、Be careful not to open positions repeatedly. If you need to adjust the take profit and stop loss, please use the exchange.update_position command.
4、Please remember to only reply in JSON format, no other explanation is required.
5、The returned JSON format does not support comments

You should only respond in JSON format as described below 
Response Format: 
{
    "thoughts": {
        "plan": "analysis steps",
        "analyze": step by step calculation process",
        "reflection": "constructive self-criticism",
        "speak": "thoughts summary to say to user"
    },
    "action": {"name": "command name", "args": {"arg name": "value"}}
}

Ensure the response can be parsed by golang json.Unmarshal
`
