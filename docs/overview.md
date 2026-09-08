# Holoplan CLI — System Overview

**Holoplan CLI** is a deterministic, multi-agent CLI tool that transforms plain user stories into complete Draw.io wireframes and/or Figma JSON documents via local LLMs. It uses a modular pipeline of structured reasoning and validation to create production-ready UI blueprints.

---

## Pipeline Stages

Each user story flows through the following stages:

1. **Chunk** → Breaks a story into discrete UI views.
2. **Build** → Generates a Draw.io-compatible XML layout or Figma JSON document for each view.
3. **Validate** → Analyzes layout geometry for visual and semantic flaws (format-specific).
4. **Audit** → Optionally critiques the layout for completeness or conformance to UX guidelines.

---

## StoryChunker (chunker.go)

- **Input**: Natural-language user story (e.g., "As a visitor, I want to view adoptable dogs.")
- **Output**: `ViewPlan` object containing:
  - A list of named views (e.g., `View Adoptable Dogs List`)
  - A reasoning string explaining the decomposition
- **LLM Used**: `huihui_ai/Hermes-3-Llama-3.2-abliterated:3b-q8_0`
- **Robustness**: Strips out `reasoning` and extraneous LLM formatting to recover raw JSON

---

## Layout Builder (builder.go)

- **Input**: Each view definition (`ViewLayout`) from the chunking stage
- **Output**: XML layout (Draw.io-compatible) or JSON (Figma-compatible) using defined view name and type
- **Key Design**: Uses consistent spatial rules for component placement

---

## Layout Validator (validator/)

Ensures spatial and semantic layout integrity before export. Two format-specific validators exist:

### Draw.io Validator (`validator/layout.go`, `validator/zones.go`)

**Validation Rules:**

1. **Collision Detection**  
   Ensures UI elements do not visually overlap (`checkCollisions`)

2. **Vertical Flow Order**  
   Verifies that elements follow a top-down reading order (`checkVerticalFlow`)

3. **Semantic Zone Conformance**  
   Ensures that common UI elements appear in conventional areas:
   - `nav` elements near the top
   - `modal` elements in the center
   - `footer` elements near the bottom  
   (`checkSemanticZones`)

**Technical Details:**
- Parses `mxGraphModel` XML used by Draw.io
- Extracts visible elements (`vertex="1"`)
- Coordinates (x, y, width, height) are used to enforce geometry

### Figma Validator (`validator/figma.go`)

**Validation Rules:**

1. **Schema Conformance**  
   Validates required fields on every node: ID, name, type, `absoluteBoundingBox`, `visible`

2. **Node Type Allowlist**  
   Only `FRAME`, `RECTANGLE`, `TEXT`, `GROUP`, `COMPONENT` types allowed

3. **Root Node Constraints**  
   Root must be `FRAME` with ID `"0:1"`, have `absoluteBoundingBox`, and be `visible`

4. **Unique IDs**  
   No duplicate node IDs across the entire document

5. **Collision Detection** (`checkFigmaCollisions`)  
   Accurate sibling overlap detection using tree ancestry (not geometric approximation):
   - Parent-child containment is allowed (not a collision)
   - Children at negative coordinates or extending beyond parent bounds are allowed
   - Unrelated nodes with geometric overlap are correctly flagged as collisions
   - Uses ancestry map built during tree traversal for O(n) ancestor checks

**Technical Details:**
- Parses Figma JSON document structure (`types.FigmaDocument`)
- Builds ancestry map (`nodeID → parentID`) during single tree traversal
- Walks parent chain for accurate ancestor/descendant checks

---

## File Structure Summary

| File/Dir | Purpose |
|----------|---------|
| `src/main.go` | Entry point and orchestration |
| `src/agents/` | Chunker, Builder, Auditor, Resolver logic |
| `src/validator/layout.go` | Draw.io geometry validation |
| `src/validator/zones.go` | Draw.io semantic zone validation |
| `src/validator/figma.go` | Figma JSON schema & geometry validation |
| `src/validator/figma_test.go` | Figma validator test suite (26 tests) |
| `src/types/` | Pure data structures (UserStory, ViewPlan, FigmaDocument, etc.) |
| `src/runner/` | Pipeline coordination and error handling |
| `src/shared/` | Cross-cutting utilities (XML sanitization, JSON extraction) |
| `examples/user_stories.yaml` | Input story corpus |
| `docs/overview.md` | System documentation |
| `docs/pipeline.md` | Pipeline flow documentation |

---

## CLI Usage

```bash
go run src/main.go --stories examples/user_stories.yaml
```
