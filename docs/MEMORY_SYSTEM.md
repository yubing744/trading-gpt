# Trading-GPT Memory System

## Overview

The file-based memory system allows the trading AI to learn from past trading experiences and maintain persistent memory across sessions. This enables continuous improvement in trading decisions based on historical performance and market insights.

## Features

- **Persistent Memory**: AI maintains memory across trading sessions
- **Word Limit Control**: Configurable memory size with automatic truncation
- **AI Feedback**: System provides feedback when memory is truncated to help AI learn
- **Automatic Integration**: Memory is automatically loaded and included in AI prompts
- **English-Only Prompts**: All memory prompts are in English for consistency

## Configuration

Add the following to your `bbgo.yaml` configuration file:

```yaml
# Memory configuration
memory:
  enabled: true
  memory_path: "memory-bank/trading-memory.md"
  max_words: 1000
```

### Configuration Options

- `enabled`: Enable/disable the memory system (default: false)
- `memory_path`: Path to the memory file (default: "memory-bank/trading-memory.md")
- `max_words`: Maximum word limit for memory content (default: 1000)

## How It Works

1. **Memory Loading**: At the start of each strategy run, the system loads existing memory from the file
2. **AI Integration**: Memory is automatically included in AI prompts when making trading decisions
3. **Memory Generation**: AI can output new memory content in the response JSON
4. **Memory Saving**: New memory is automatically saved to the file with timestamp
5. **Word Limit Enforcement**: If memory exceeds the word limit, it's truncated and AI receives feedback
6. **Cycle Reset**: The strategy runs in cycles, and AI memory resets at the beginning of each cycle. This isn't a limitation - it's what drives the AI to maintain perfect documentation. After each reset, the AI relies ENTIRELY on the Memory Part to understand the project and continue work effectively.

## Memory File Format

The memory file uses a simple markdown format:

```markdown
# Trading Memory Document

## 2024-01-15 14:30:25
In volatile markets, RSI overbought/oversold signals are more reliable, especially near key support and resistance levels.

## 2024-01-15 14:25:10
Single trade risk should not exceed 2% of total capital. Pause trading after 3 consecutive losses.

## 2024-01-15 14:20:05
Moving average crossover strategies perform better in trending markets, while Bollinger Bands strategies are suitable for ranging markets.
```

## AI Response Format

When memory is enabled, the AI will include a memory field in its JSON response:

```json
{
    "thoughts": {
        "plan": "analysis steps",
        "analyze": "step-by-step analysis", 
        "detail": "output detailed calculation process",
        "reflection": "constructive self-criticism",
        "speak": "thoughts summary to say to user"
    },
    "action": {"name": "command name", "args": {"arg name": "value"}},
    "memory": {"content": "memory content to save, keep concise and within reasonable word limit"}
}
```

## Memory Truncation Feedback

When memory content exceeds the word limit:

- The system automatically truncates the memory to fit within the limit
- AI receives a warning message about the truncation
- The warning includes current word count and limit information
- This helps the AI learn to keep future memory content more concise

## Example Usage

1. **Enable Memory**: Set `memory.enabled: true` in your configuration
2. **Run Trading**: Start your trading strategy as usual
3. **AI Learning**: The AI will automatically learn from trading experiences
4. **Memory Growth**: Check the memory file to see accumulated insights
5. **Continuous Improvement**: AI decisions improve over time based on memory

## Best Practices

- **Keep Memory Concise**: Encourage AI to write brief, actionable insights
- **Monitor Word Limit**: Adjust `max_words` based on your needs
- **Regular Review**: Periodically review memory content for quality
- **Backup Memory**: Consider backing up important memory files
- **Test Scenarios**: Test the system with various trading scenarios

## Troubleshooting

- **Memory Not Loading**: Check file path and permissions
- **Memory Not Saving**: Verify write permissions to memory file directory
- **Truncation Warnings**: Reduce `max_words` or encourage more concise AI output
- **Performance Issues**: Consider reducing `max_words` for faster processing

## Implementation Details

The memory system is integrated into:
- `pkg/types/result.go`: Memory field in Result structure
- `pkg/prompt/prompt.go`: Memory prompts in templates with cycle reset information
- `pkg/config/config.go`: Memory configuration
- `pkg/memory/memory_manager.go`: Independent memory management package
- `pkg/jarvis.go`: Memory integration and processing logic

This implementation follows the design principles of simplicity, directness, and seamless integration with existing trading workflows. The independent memory package ensures better code organization and maintainability.
