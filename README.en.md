# AI Agents Transformer

[English Version](README.en.md) | [中文版本](README.md)

Cross‑platform AI agent workflow DSL converter. Uses iFlytek (Spark Agent) as the hub: source files are normalized into a unified DSL, then generated to target‑specific DSL for Dify and Coze. Supports auto detection, concurrent batch, and multi‑stage validation.

---

## Table of Contents
- [Overview](#overview)
- [Supported Paths](#supported-paths)
- [Tolerance & Placeholders](#tolerance)
- [Quick Start (Windows / macOS / Linux)](#quick-start)
- [CLI Reference (Complete)](#cli)
- [Coze YAML Import/Export](#coze-yaml)
- [Build](#build)
- [Dev & Test](#dev)
- [FAQ](#faq)
- [License & Credits](#license)

---

<a id="overview"></a>
## Overview
- Bidirectional conversion: Dify ↔ iFlytek, Coze ↔ iFlytek
- Coze ZIP: official Coze ZIP → iFlytek
- Auto detection: YAML/ZIP (ZIP is detected as Coze)
- Concurrent batch: `batch` picks workers by CPU, supports pattern & overwrite
- Validation pipeline: structural / semantic / platform checks with user‑friendly messages
- Node coverage: start / end / llm / code / condition / classifier / iteration

<a id="supported-paths"></a>
## Supported Paths
- Dify ↔ iFlytek (bidirectional)
- Coze ↔ iFlytek (bidirectional, YAML)
- Dify → iFlytek → Coze (recommended)
- Coze → iFlytek → Dify
- Coze ZIP → iFlytek (native)

Not supported:
- Dify ↔ Coze direct conversion (use iFlytek as the hub)
- iFlytek → Coze ZIP (ZIP as target is not supported yet)

<a id="tolerance"></a>
## Tolerance & Placeholders
To keep workflow structure intact when encountering unsupported node types on the target platform:
- Replace the node with a “code node” placeholder
- Put the original node type in the code node title for quick follow‑up
- Preserve incoming/outgoing edges to keep the flow executable
- With `--verbose`, print details and summary, e.g.:
  - Converting unsupported node type '4' (ID: 133604) to code node placeholder
  - 25 unsupported nodes were converted to code node placeholders

<a id="quick-start"></a>
## Quick Start
All commands below are executed at the project root.

### Windows (PowerShell)
```powershell
# iFlytek → Dify
./ai-agent-converter.exe convert --from iflytek --to dify --input agent.yml --output dify.yml

# Dify → iFlytek
./ai-agent-converter.exe convert --from dify --to iflytek --input dify.yml --output agent.yml

# iFlytek → Coze (YAML)
./ai-agent-converter.exe convert --from iflytek --to coze --input agent.yml --output coze.yml

# Coze ZIP → iFlytek (ZIP auto‑detected as Coze)
./ai-agent-converter.exe convert --to iflytek --input workflow.zip --output agent.yml --verbose

# Batch (concurrent, overwrite)
./ai-agent-converter.exe batch --from iflytek --to dify --input-dir .\tests\fixtures\iflytek --pattern 'iflytek*.yml' --output-dir .\out --workers 4 --overwrite

# Validate DSL
./ai-agent-converter.exe validate --input agent.yml

# Quiet mode (errors only)
./ai-agent-converter.exe convert --from iflytek --to dify --input agent.yml --output dify.yml --quiet
```
Tip: use `.exe`; prefer single quotes for `--pattern`.

### macOS / Linux (Terminal)
```bash
# iFlytek → Coze (YAML)
./ai-agent-converter convert --from iflytek --to coze --input agent.yml --output coze.yml

# Dify → iFlytek
./ai-agent-converter convert --from dify --to iflytek --input dify.yml --output agent.yml

# Auto‑detect (YAML)
./ai-agent-converter convert --to dify --input agent.yml --output dify.yml --verbose

# Batch (default concurrency)
./ai-agent-converter batch --from iflytek --to dify --input-dir ./workflows --output-dir ./converted --workers 8 --overwrite

# Validate
./ai-agent-converter validate --input agent.yml
```

<a id="cli"></a>
## CLI Reference (Complete)
Run at project root; use `./ai-agent-converter.exe` on Windows and `./ai-agent-converter` on macOS/Linux.

### convert
- Purpose: convert DSL across platforms
- Required: `--to`, `--input/-i`, `--output/-o`
- Optional: `--from` (auto‑detect when omitted; ZIP → Coze)
- Limits: Dify↔Coze direct is not supported; iFlytek→Coze ZIP is not supported

### validate
- Purpose: validate DSL (structural/semantic/platform)
- Required: `--input/-i`
- Optional: `--from` (auto‑detect when omitted)

### batch
- Purpose: concurrent batch conversion
- Required: `--from`, `--to`, `--input-dir`, `--output-dir`
- Optional: `--pattern` (default `*.yml`), `--workers` (CPU‑based by default), `--overwrite`, global `--quiet/--verbose`

### info
- Purpose: show capabilities
- Flags: `--nodes`, `--types`, `--all`

### platforms
- Purpose: list supported platforms
- Flags: `--detailed`

### completion (optional)
- Purpose: generate shell completion scripts
- PowerShell (current session):
  - `ai-agent-converter.exe completion powershell | Out-String | Invoke-Expression`
- PowerShell (persist to Profile):
  - `ai-agent-converter.exe completion powershell | Out-File -Encoding UTF8 $PROFILE`
- Bash: `ai-agent-converter completion bash > /etc/bash_completion.d/ai-agent-converter`
- Zsh: `ai-agent-converter completion zsh > "${fpath[1]}/_ai-agent-converter"`

<a id="coze-yaml"></a>
## Coze YAML Import/Export
Coze official workflow currently does not support YAML I/O. We maintain a fork that enables it:
- Repo: https://github.com/2064968308github/coze_transformer
- Usage: convert to/from YAML with that repo, then use this tool for cross‑platform conversion (e.g., Coze YAML → iFlytek, or iFlytek → Coze YAML).

<a id="build"></a>
## Build
### Windows (recommended)
```powershell
go build -o ai-agent-converter.exe ./cmd/
./ai-agent-converter.exe --help
```
If a same‑name file without extension exists, PowerShell may show “Choose an app to open”. Keep the `.exe` only.

### macOS / Linux
```bash
go build -o ai-agent-converter ./cmd/
chmod +x ./ai-agent-converter
./ai-agent-converter --help
```

<a id="dev"></a>
## Dev & Test
```bash
go fmt ./...
go vet ./...
go test ./... -cover
```

<a id="faq"></a>
## FAQ
- PowerShell opens a “Choose an app” dialog: keep and run `ai-agent-converter.exe` only
- Coze ZIP auto‑detection: ZIP is detected as Coze; `--from` is not required
- Dify ↔ Coze direct: not supported; use iFlytek as the hub
- Coze ZIP as output: not supported yet (use YAML or the YAML fork above)
- Batch `--pattern`: single quotes on PowerShell; single/double quotes on macOS/Linux
- Quiet mode: `--quiet` prints errors only

<a id="license"></a>
## License & Credits
- License: see LICENSE
- Coze YAML I/O fork: https://github.com/2064968308github/coze_transformer
