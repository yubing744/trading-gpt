# Active Context: Trading-AI

## Current Focus
Enhanced README.md documentation to better showcase Trading-GPT's capabilities. Updated Features section and added comprehensive Architecture diagram to improve project presentation and user understanding.

## Current Mode
**Documentation Mode** - Completed README.md enhancement with simplified, layered architecture visualization and streamlined feature presentation.

## Technical Documentation Requirements
The technical specification document (tech_spec.md) must include:
1. **需求描述** (Requirements Description)
   - Project background and core needs
   - Project goals and scope

2. **需求分析** (Requirements Analysis)
   - Role analysis: User roles and system roles
   - Use case analysis: Core use cases and auxiliary use cases

3. **概要设计** (Conceptual Design)
   - Architecture diagram using Mermaid
   - Domain class diagram using Mermaid
   - Module diagram using Mermaid
   - Database design (if applicable)
   - Key technical challenges

4. **详细设计** (Detailed Design)
   - Module-by-module explanation of interfaces and core business logic
   - UML diagrams for interfaces using Mermaid
   - Sequence diagrams for core business logic using Mermaid
   - Textual descriptions accompanying all diagrams

5. **测试用例** (Test Cases)
   - Testing strategy and approach
   - Key test scenarios

All diagrams must be created using Mermaid syntax, and each section should be concise and clear.

## Recent Changes

### README.md Enhancement - 2025-11-06
- **Updated Features section** - Simplified from 3-category 18-item structure to single-level 9-item list
- **Added Architecture diagram** - Created layered Mermaid diagram showing 5 architectural layers (Exchange → Core Engine → Environment → AI → User)
- **Improved visual clarity** - Used `graph BT` (bottom-to-top) layout for clear layer separation without misalignment
- **Streamlined presentation** - Removed color styling and emoji categories for cleaner, more professional appearance
- **Better documentation links** - Consolidated to single documentation link
- **Key design decisions**:
  - Chose balanced approach over comprehensive (avoided information overload) or minimal (met all requirements)
  - Bottom-to-top layer visualization provides intuitive understanding of data flow
  - Single-level feature list improves scannability while preserving key capabilities
  - Maintained all existing content (examples, configuration, usage instructions)

### Dynamic Indicator Query System (Issue #62) - 2025-11-01
- **Implemented `get_indicator` command** for ExchangeEntity enabling zero-config dynamic indicator queries
- **Support for 9 indicator types**: RSI, BOLL, SMA, EWMA, VWMA, ATR, ATRP, VR, EMV
- **Any timeframe support**: 1m, 5m, 15m, 30m, 1h, 4h, 1d, etc. - unlimited flexibility
- **Flexible parameters**: window_size, band_width with intelligent defaults
- **AI-specified custom names**: Optional `name` parameter for better readability
- **Comprehensive documentation**: Added detailed usage scenarios and examples to `docs/features/next-commands-feature.md`
- **Multi-timeframe analysis**: AI can compare same indicator across different timeframes for trend confirmation
- **Adaptive strategies**: AI can adjust indicator parameters based on market conditions

### Code Review Fixes (PR #65) - 2025-11-01
Addressed all critical issues from PR #65 code review with 5 major fixes:

1. **Duplicate Indicator Detection** (High Priority)
   - Check pre-configured indicators before creating new ones
   - Match by type, interval, window_size, and band_width
   - Reuse existing indicators to avoid resource waste
   - Prevents data inconsistency and memory bloat

2. **Enhanced Parameter Validation** (Medium Priority)
   - Strict validation for window_size and band_width
   - Verify positive values and valid ranges
   - Type whitelist validation for security
   - Clear error messages for AI troubleshooting

3. **Data Sufficiency Check** (Medium Priority)
   - Verify sufficient kline data before calculation
   - Prevents crashes from insufficient historical data
   - Critical for new trading pairs with limited history

4. **Frequency Limiting** (Medium Priority)
   - Limit to 5 dynamic indicator requests per 15-minute cycle
   - Auto-reset counter each cycle
   - Prevents resource exhaustion from excessive requests

5. **Custom Name Support** (User Request)
   - Optional `name` parameter for AI to specify indicator names
   - Auto-generated names include all parameters: "rsi_5m_w14_dynamic"
   - Improves debugging and prevents naming conflicts

### Thread Safety Improvements (PR #65) - 2025-11-01
- **FearAndGreedEntity**: Added thread-safe event channel using `atomic.Value`
- **TwitterAPIEntity**: Added thread-safe event channel using `atomic.Value`
- **ExchangeEntity**: Added thread-safe event channel using `atomic.Value`
- **Command count limits**: Added `MaxCommandsPerCycle = 10` to prevent resource exhaustion
- **Context cancellation**: Added proper context checking in command execution loops
- **File permissions**: Restricted to 0600/0700 for enhanced security
- **Integer parsing**: Replaced `fmt.Sscanf` with `strconv.Atoi` for consistency

### Limit Order Feature (Issue #58) - 2025-01-22
- **Implemented limit order support** with price expression parsing
- **Added automatic cleanup mechanism** - unfilled limit orders are canceled at each decision cycle
- **Created price expression parser** (`pkg/utils/price.go`) supporting variables like `last_close`, `last_high`, etc.
- **Extended trading commands** with 4 new parameters: order_type, limit_price, time_in_force, post_only
- **Comprehensive testing** - 9 unit tests all passing
- **Developer documentation** - Created `docs/limit-order-feature.md` with usage guide and technical details
- **Key design decision**: Auto-cleanup approach keeps AI decision-making simple while preventing order accumulation

### Memory System Implementation (Issue #40)
- **Successfully implemented file-based memory system** for trading-gpt
- Extended `pkg/types/result.go` to include Memory field in Result structure
- Modified `pkg/prompt/prompt.go` to integrate memory prompts with conditional rendering and cycle reset information
- Added memory configuration in `pkg/config/config.go` with MemoryConfig struct
- **Created independent memory package** `pkg/memory/memory_manager.go` for better code organization
- Integrated memory functionality into `pkg/jarvis.go` using the independent memory package
- Implemented word limit enforcement with AI feedback for memory truncation
- **Added strategy cycle reset information** to AI prompts to explain memory behavior
- **Updated all comments to English** for consistency, including `pkg/types/result.go` and `pkg/chat/feishu/feishu_chat_provider.go`
- **Fixed directory creation issue** in `pkg/memory/memory_manager.go` - now automatically creates directories if they don't exist
- **Fixed memory processing logic** in `pkg/jarvis.go` - AI outputs complete memory content, so we now replace instead of merge
- **Fixed memory prompt logic** in `pkg/prompt/prompt.go` - now uses `MemoryEnabled` instead of `.Memory` to control memory prompts, ensuring AI outputs memory even on first run
- **Enhanced memory usage feedback** in `pkg/prompt/prompt.go` - added current word count, usage percentage, and intelligent suggestions to help AI decide whether to expand or consolidate memory
- Created example configuration file `bbgo-memory-example.yaml`
- Created sample memory file `memory-bank/trading-memory.md`
- Updated progress documentation to reflect memory system implementation

## Next Steps
1. Consider adding README badges (build status, version, license)
2. Monitor user feedback on new documentation structure
3. Continue feature development with updated documentation standards
4. Update README when major features are added
5. Consider adding screenshots or quick-start guide

## Active Decisions and Considerations
- **Documentation Approach**: Using structured Markdown files organized in a clear hierarchy to maintain project knowledge.
- **Documentation Scope**: Focus on both technical implementation details and higher-level product/project context.
- **Documentation Maintenance**: Regular updates to keep documentation in sync with implementation.
- **Memory Feature Implementation**: The core logic for reflection and memory management will reside in `pkg/jarvis.go` to maintain central control, interacting with the `Trading Agent` as needed for context or decision execution.

## Important Patterns and Preferences
- **Code Organization**: Modular structure with clear separation of concerns
- **Interface Design**: Clear interfaces for agents, environments, and other components
- **Error Handling**: Consistent error handling and reporting throughout the system
- **Testing**: Appropriate test coverage for critical components
- **Configuration**: Environment-based configuration with sensible defaults

## Learnings and Project Insights
- The integration of LLMs with trading systems creates a powerful tool for democratizing algorithmic trading
- Natural language interfaces can significantly reduce the barrier to entry for algorithmic trading
- Careful consideration of risk management is essential for any trading system
- The modular architecture facilitates extension and maintenance
- **Simplicity in AI decision-making is crucial** - Auto-cleanup approach for limit orders prevents AI from needing to manage order state, keeping decision complexity low
- **Expression-based parameters provide flexibility** - Price expressions like "last_close * 0.995" allow dynamic pricing without complex AI logic
- **Cycle-based cleanup aligns with decision patterns** - Periodic cleanup matches the natural rhythm of strategy evaluation cycles
- **Zero-config dynamic queries unlock AI potential** - Removing pre-configuration requirements enables true adaptive strategies
- **Duplicate detection is critical for resource efficiency** - Reusing pre-configured indicators prevents waste and ensures consistency
- **Comprehensive validation prevents silent failures** - Strict parameter checking with clear error messages helps AI learn and adapt
- **Thread safety cannot be assumed** - `atomic.Value` for shared state access prevents race conditions in concurrent environments
- **Frequency limits protect system stability** - Rate limiting prevents resource exhaustion from AI exploration behaviors
- **Documentation clarity matters for adoption** - Simplified, layered presentation improves understanding without sacrificing completeness
- **Bottom-to-top architecture visualization is intuitive** - Showing data flow from exchanges upward to user helps readers understand system structure naturally
- **Less is more in feature presentation** - Streamlined single-level list (9 items) is more scannable than multi-category structure (18 items)
- **Iterative refinement based on feedback** - User review led to three successive improvements (remove colors, simplify categories, fix layer alignment)
