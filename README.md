# 🛰️ Holoplan CLI

**Holoplan CLI** is a deterministic local-first CLI tool that generates wireframe layouts (in Draw.io XML format) from a YAML file of user stories. It uses a modular LLM pipeline to chunk stories into UI views, generate layouts, critique and correct them, and validate the result—without relying on a rigid DSL.

> ⚙️ **Goal:** Replace domain-specific UI compilers with zero-temperature LLMs for fast, reviewable wireframe generation.

---

## ✨ Features

- 📖 **YAML-based User Story Input**
- 🧠 **LLM-based Story Chunking**
- 🏗 **Hermes Model Layout Builder**
- 🔍 **Automated UI Critique + Correction Loop**
- 📏 **Go-based Spatial Validation**
- 🧾 **Draw.io XML Output + Critique Logs**
- 🧪 **Merge All Views into a Single File**

---

## 📦 Example

```bash
holoplan-cli --stories examples/user_stories.yaml
```

This will:

1. Load each user story
2. Chunk it into views
3. Generate `.drawio.xml` files for each
4. Auto-correct bad layouts
5. Validate spatial geometry
6. Output:

   * Per-view XML files in `output/`
   * Critique reports (if needed)
   * Merged `final.drawio.xml`

---

## 🔧 Usage

### ✅ Requirements

* Go 1.20+
* [Ollama](https://ollama.com/) installed and running locally
* At least one of the following models:

  * `hermes3:8b-llama3.1-q5_1` (for layout/correction)
  * `qwen3` or `llama3` (for critique and chunking)

### 🏁 Build & Run

```bash
# Run directly
go run src/main.go --stories examples/user_stories.yaml

# Or build binary
go build -o holoplan-cli
./holoplan-cli --stories examples/user_stories.yaml
```

### 📁 Output

* All generated views saved to `./output/`
* Final merged layout: `output/final.drawio.xml`
* Critique files: `output/[view_name].critique.txt` (if needed)

---

## 🧠 Architecture Overview

```
UserStory.yaml
   ↓
[Chunker Agent] → ViewPlan
   ↓
[Builder Agent] → XML layout
   ↓
[Auditor Agent] → Critique
   ↓ (resolve if needed)
[Validator] → Geometry check
   ↓
Save XML + critique logs
   ↓
Merge all → final.drawio.xml
```

* All LLMs are called locally through Ollama
* Temperature is `0` for deterministic output

---

## 🧪 Examples

```yaml
# examples/user_stories.yaml

- id: US-001
  title: Sign-in screen
  narrative: >
    As a user, I want to log in securely with a username and password, so I can access the app.
```

---

## 🧭 Roadmap Ideas

* [ ] CLI flags for output path, skip audit, etc.
* [ ] Add support for Figma export
* [ ] GUI wrapper for non-CLI users
* [ ] Multi-language prompt templates

---

## 🤖 Models + Prompts

* Uses LLMs to emulate deterministic UI compilers
* Prompts live in `src/prompts/`
* Builder uses Hermes; chunker/auditor use Llama/Qwen variants

---

## 📄 License

MIT




