# Holoplan CLI — System Overview

**Holoplan CLI** is a deterministic, multi-agent CLI tool that transforms plain user stories into complete Draw.io wireframes via local LLMs. It uses a modular pipeline of structured reasoning and validation to create production-ready UI blueprints.

---

## Pipeline Stages

Each user story flows through the following stages:

1. **Chunk** → Breaks a story into discrete UI views.
2. **Build** → Generates a Draw.io-compatible XML layout for each view.
3. **Validate** → Analyzes layout geometry for visual and semantic flaws.
4. **Audit** → Optionally critiques the layout for completeness or conformance to UX guidelines.

---

## StoryChunker (chunker.go)

- **Input**: Natural-language user story (e.g., "As a visitor, I want to view adoptable dogs.")
- **Output**: `ViewPlan` object containing:
  - A list of named views (e.g., `View Adoptable Dogs List`)
  - A reasoning string explaining the decomposition
- **LLM Used**: `huihui_ai/Hermes-3-Llama-3.2-abliterated:3b-q8_0`
- **Robustness**: Strips out `<think>` and extraneous LLM formatting to recover raw JSON

---

## Layout Builder (builder.go)

- **Input**: Each view definition (`ViewLayout`) from the chunking stage
- **Output**: XML layout (Draw.io-compatible) using defined view name and type
- **Key Design**: Uses consistent spatial rules for component placement

---

## Layout Validator (validator/layout.go)

Ensures spatial and semantic layout integrity before export.

### Validation Rules

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

### Technical Details

- Parses `mxGraphModel` XML used by Draw.io
- Extracts visible elements (`vertex="1"`)
- Coordinates (x, y, width, height) are used to enforce geometry

---

## File Structure Summary

| File/Dir | Purpose |
|----------|---------|
| `src/main.go` | Entry point and orchestration |
| `src/agents/` | Chunker and builder logic |
| `src/validator/` | Geometry-based layout rules |
| `examples/user_stories.yaml` | Input story corpus |
| `docs/overview.md` | System documentation |

---

## CLI Usage

```bash
go run src/main.go --stories examples/user_stories.yaml
````

