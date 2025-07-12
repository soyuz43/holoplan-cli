## 📐 Holoplan Pipeline Overview (`docs/pipeline.md`)

---

### 🧭 Purpose

The Holoplan CLI takes structured **user stories** written in YAML and transforms them into **Draw\.io XML wireframes** using a modular, LLM-powered pipeline.

Each stage in the pipeline uses deterministic or guided language model calls to progressively synthesize, audit, and refine UI layouts based on story-driven design intent.

---

### 🔄 Pipeline Flow

```plaintext
[YAML User Stories]
       ↓
   Chunker Agent
  (Extract Views)
       ↓
   Builder Agent
(Generate Layout XML)
       ↓
   Auditor Agent
(LLM Verifies XML vs. Story)
       ↓
   Resolver Agent
(LLM Fixes XML if Audit Fails)
       ↓
  Validator Module
(Spatial + Semantic Rules)
       ↓
   Final Draw.io XML
      per View
```

Each story may go through multiple correction cycles (`MaxCorrections = 3`) until the audit passes or retry limit is reached.

---

### 📦 Inputs

**User Stories (`.yaml`)**
Each story must include:

* `id`, `title`, and `narrative`
* Either a single `view` or list of `views`
* Optional: `interaction_origin`, `resulting_view`, `shared_components`

---

### 🧠 Agents & Responsibilities

| Agent      | Responsibility                                     | LLM Model       | Output Format           |
| ---------- | -------------------------------------------------- | --------------- | ----------------------- |
| `Chunker`  | Breaks story into view layouts                     | `qwen2.5-coder` | `types.ViewPlan` (JSON) |
| `Builder`  | Generates raw Draw\.io XML for each view           | `qwen2.5-coder` | `string` (XML)          |
| `Auditor`  | Compares user story to layout and finds mismatches | `llama3.1:8b`   | `{"issues": [...]}`     |
| `Resolver` | Fixes XML layout based on audit issues             | `qwen2.5-coder` | `{"xml": "<...>"}`      |

All LLM calls are made to `localhost:11434` via the Ollama API.

---

### ✅ Validation Rules

The `validator` module performs checks after all LLM corrections:

* **No Collisions:** UI elements must not overlap
* **Vertical Flow:** Components should flow top-to-bottom within vertical bands
* **Semantic Zones:** Navbar should be top, footer bottom, modals centered
* **Attribute Sanity:** All XML attributes must be quoted and valid

---

### 📁 Output Structure

```plaintext
output/
├── <storyID>_<viewName>.drawio         # Final layout XML
├── <storyID>_<viewName>.critique.txt   # If audit failed, shows LLM critique
└── final.drawio                        # Combined <mxfile> with all diagrams
```

---

### 🛠️ Error Handling

Each stage uses `defer recover()` to catch panics and continue the pipeline. If a stage fails:

* It logs the error
* Skips to the next view or story
* Fall back to previous good output (if any)

---

### 🔍 Example Story Flow

```yaml
- id: USR-001
  title: Login Flow
  narrative: As a user, I want a centered login button on the login screen.
  view: LoginScreen
```

1. **Chunker:** produces `LoginScreen` view with components like `"Navbar"`, `"Login Button"`, `"Footer"`.
2. **Builder:** generates `<mxGraphModel>` XML with those components.
3. **Auditor:** compares story vs XML and may report `"Login button is not centered"`.
4. **Resolver:** fixes layout and resubmits for re-audit.
5. **Validator:** ensures spatial rules are satisfied.
6. **Output:** saved as `output/usr-001_loginscreen.drawio`.

---

### 🚀 To Run the Pipeline

```bash
go run main.go run --stories path/to/user_stories.yaml
```

If no path is given, the CLI prompts for it interactively.

---

### 🧪 Testing + Debugging

* Use `logger.Printf` inside each agent to trace prompt and response
* All raw LLM responses are logged before and after JSON parsing
* Validation failures print clear error messages with element IDs

---

### 📝 Prompt Engineering Tips

All agents use **low temperature (0.0)** and **explicit JSON format instructions** to encourage deterministic outputs.

* Builder prompts are embedded via `//go:embed`
* Audit prompts include several examples for consistency
* Resolver prompts format issues as bullet lists to simplify parsing

---
