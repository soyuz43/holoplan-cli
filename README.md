# 🛰️ Holoplan CLI

**Holoplan CLI** is a deterministic local-first CLI tool that generates wireframe layouts (in Draw.io XML format) from a YAML file of user stories. It uses a modular LLM pipeline to chunk stories into UI views, generate layouts, critique and correct them, and validate the result—without relying on a rigid DSL.

> ⚙️ **Goal:** Replace domain-specific UI compilers with zero-temperature LLMs for fast, reviewable wireframe generation.

---

## ✨ Features

- 📖 **YAML-based User Story Input**
- 🧠 **LLM-based Story Chunking**
- 🏗 **Qwen2.5 Coder Model Layout Builder**
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
* And the following models:

  * `qwen2.5-coder:7b-instruct-q6_K` (for layout and correction)
  * `llama3.1:8b` and `qwen2.5-coder:3b-instruct-q8_0` (for critique and chunking)
  
Let’s integrate your `install.ps1` script into the `README.md` with clear instructions for Windows 10/11 users, including:

* Setup guidance for PowerShell users
* Notes about adding `~/bin` to their PATH if not already present
* Emphasis that this script builds and installs the binary

---

### 🏁 Installation & Running

#### 🪟 Windows 10/11 (PowerShell)

Run the included PowerShell script to build and install the CLI to your local `~/bin` directory:

```powershell
.\install.ps1
```

This will:

* Build `holoplan.exe` from `src/`
* Install it to `C:\Users\<you>\bin\holoplan.exe`
* ✅ You’ll see a success message if it worked

👉 Make sure `C:\Users\<you>\bin\` is added to your system's **PATH**:

* Search "Environment Variables"
* Edit your `Path` system/user variable
* Add: `C:\Users\<your-username>\bin\`

Once installed, you can run it from any terminal:

```powershell
holoplan --stories examples/user_stories.yaml
```

#### 🐧 Linux / macOS

You can build and run directly:

```bash
go run src/main.go --stories examples/user_stories.yaml
```

Or install the binary:

```bash
go build -o holoplan-cli
sudo mv holoplan-cli /usr/local/bin/holoplan
holoplan --stories examples/user_stories.yaml
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




