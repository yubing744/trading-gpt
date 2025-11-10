# Active Context: Trading-AI

## Current Focus
Current efforts center on (1) the file-based memory function that lets the AI learn from trading experiences across sessions and (2) hardening the agent result parser so malformed or richly formatted LLM JSON (including multi-line thoughts and memory payloads) can still be recovered reliably.

## Current Mode
**Act Mode** - Successfully implemented the file-based memory system as requested in GitHub issue #40.

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
- **Successfully implemented file-based memory system** for trading-gpt as requested in GitHub issue #40
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
- **Hardened agent result parsing** in `pkg/utils/agent_result.go` by adding a byte-wise sanitizer that escapes raw newlines/tabs within quoted strings so long-form LLM responses (thoughts + memory) can be parsed without data loss
- **Expanded parser test suite** in `pkg/utils/agent_result_test.go` to cover multi-paragraph responses, CRLF/tab handling, and Result memory deserialization

## Next Steps
1. **Test the memory system** with real trading scenarios to ensure proper functionality
2. **Monitor memory file growth** and adjust word limits as needed
3. **Enhance memory content quality** by refining AI prompts for better memory generation
4. **Add memory analytics** to track learning effectiveness over time
5. **Consider memory backup and versioning** for important trading insights
6. **Document memory system usage** and best practices for users
7. **Monitor LLM output formats** to identify additional parser edge cases worth codifying in tests

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
