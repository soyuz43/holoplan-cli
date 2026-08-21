# Holoplan CLI Agent Guidelines

## Core Principles

### Local-First Determinism
- Prioritize solutions that operate entirely locally without external API dependencies
- Ensure deterministic behavior through fixed LLM temperature (0.0) and seed values
- Maintain offline capability as a core requirement for all features

### Modular Agent Architecture
- Follow the established pattern of specialized agents (Chunker, Builder, Auditor, Resolver)
- Each agent must have a single, well-defined responsibility
- Avoid creating monolithic components that handle multiple pipeline stages
- Preserve clear interfaces between agents using structured data types

### User Story-Driven Design
- All functionality must serve the transformation of user stories to visual layouts
- Respect and extend the existing YAML schema thoughtfully
- Changes should enhance the expression of user intents rather than constrain them

### Output Format Integrity
- For Draw.io XML output: maintain the full audit/correction/validation pipeline
- For alternative formats (Figma, etc.): implement format-appropriate validation layers
- Never skip spatial validation for geometry-sensitive outputs without explicit justification

### Dependency Minimalism
- Leverage existing tooling (Ollama, Go standard library) before adding new dependencies
- Evaluate maintenance burden, security implications, and operational complexity
- Prefer solutions that reuse current LLM infrastructure over introducing new model types

### Separation of Concerns
- Types: `src/types` - pure data structures
- Agents: `src/agents` - LLM-interacting components with single responsibilities
- Orchestration: `src/runner` - pipeline coordination and error handling
- Utilities: `src/shared` - cross-cutting concerns like XML processing
- Validation: `src/validator` - domain-specific constraint checking

### Prompt-Centric Behavior
- View LLM prompts in `src/agents/prompts` as the primary behavioral specification
- Modify agent behavior through prompt refinement before altering code
- Treat prompts as immutable contracts once established for a model version

### Robust Error Handling
- Apply the `recoverLLM` pattern for isolating LLM agent failures
- Implement graceful degradation (skip failed stories, continue processing)
- Log errors with sufficient context for debugging without exposing internals

### Spatial and Structural Awareness
- Respect geometric constraints validated by the validator package
- Consider layout semantics when generating or modifying components
- Preserve mergeability of multiple views into single output files

## Practical Implementation Guidelines

### Adding New Features
1. Determine which pipeline stage(s) are affected
2. For LLM behavior changes: first iterate on relevant prompt files
3. For structural changes: modify appropriate data types in `src/types`
4. For validation needs: extend `src/validator` with focused constraint checks
5. Update orchestration logic in `src/runner/pipeline.go` only when necessary

### Modifying Existing Agents
- Follow the established pattern: context preparation → HTTP call to Ollama → JSON parsing
- Maintain temperature at 0.0 and seed at 42 for consistency
- Use the `extractCleanJSON` and `escapeLineBreaks` utilities for LLM response handling
- Preserve the `recoverLLM` defer pattern for panic recovery

### Output Format Considerations
- When adding new formats:
  - Evaluate if audit/correction/validation steps apply
  - For XML-based formats: apply Draw.io-like processing pipeline
  - For non-XML formats: implement format-specific validation if needed
- Update `saveOutput` function with appropriate file extensions and content handling
- Modify `mergeDrawio` or create equivalent merger for new formats as needed

### Validation Principles
- Add validation checks that prevent entire classes of layout issues
- Focus on geometric constraints (overlaps, spacing, alignment) rather than pixel-perfect positioning
- Design validation to be deterministic and fast
- Consider whether issues are better fixed through generation prompts vs. post-processing

### Error Recovery Patterns
- Use `defer recoverLLM("AgentName")` in all LLM-calling functions
- Return meaningful boolean success indicators from agent functions
- Log failures at appropriate levels (debug for expected variabilities, error for systemic issues)
- Allow pipeline to continue processing other stories when individual items fail

## Things to Avoid

### Architectural Violations
- Don't create agents that handle multiple pipeline responsibilities
- Don't bypass the chunking stage by hardcoding view structures
- Don't embed LLM prompts directly in agent code (use prompt files)
- Don't add dependencies that require external network calls during operation

### Determinism Breakers
- Don't alter LLM temperature without strong justification and testing
- Don't introduce non-deterministic elements in prompt processing
- Don't rely on external services for core functionality

### Validation Circumvention
- Don't skip XML syntax validation for Draw.io output
- Don't ignore spatial validation failures without investigation
- Don't modify validated output in ways that could reintroduce violations

### Dependency Creep
- Don't add JSON/YAML parsing alternatives when existing libraries suffice
- Don't introduce new HTTP clients when the standard library implementation works
- Don't add UI framework dependencies for a CLI tool

### Output Inconsistency
- Don't change file naming conventions without updating all related code
- Don't alter output directory structure without coordination
- Don't break merge compatibility when modifying individual view generation

### Communication Principles
- Do make assumptions explicit when evidence is incomplete
- Do prioritize correctness over performance optimization prematurely
- Do consider long-term maintenance implications of clever solutions
- Do verify assumptions against actual code behavior before implementation
- Do distinguish between factual code observations and inferred behaviors
- Do accept uncertainty when evidence is insufficient rather than guessing
