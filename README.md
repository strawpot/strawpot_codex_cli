# StrawPot Codex CLI

A Go wrapper that translates [StrawPot](https://github.com/strawpot) protocol arguments into [OpenAI Codex CLI](https://github.com/openai/codex) flags. It acts as a pure translation layer — process management, sessions, and infrastructure are handled by StrawPot core.

## Prerequisites

- [Codex CLI](https://github.com/openai/codex) (`brew install codex` or `npm install -g @openai/codex`)
- An OpenAI API key (or a ChatGPT Plus/Pro plan)

## Installation

```sh
curl -fsSL https://raw.githubusercontent.com/strawpot/strawpot_codex_cli/main/strawpot_codex/install.sh | sh
```

This downloads a pre-built binary for your platform (macOS/Linux, amd64/arm64) to `/usr/local/bin`. Override the install directory with `INSTALL_DIR`:

```sh
INSTALL_DIR=~/.local/bin curl -fsSL ... | sh
```

## Usage

The wrapper exposes two subcommands:

### `setup`

Runs `codex login` to authenticate with OpenAI.

```sh
strawpot_codex setup
```

### `build`

Translates StrawPot protocol flags into a Codex CLI command and outputs it as JSON.

```sh
strawpot_codex build \
  --agent-workspace-dir /path/to/workspace \
  --working-dir /path/to/project \
  --task "fix the bug" \
  --config '{"model":"o3"}'
```

Output:

```json
{
  "cmd": ["codex", "exec", "fix the bug", "-c", "model_instructions_file=\"/path/to/workspace/prompt.md\"", "-m", "o3", "--dangerously-bypass-approvals-and-sandbox", "-C", "/path/to/project", "--add-dir", "/path/to/workspace"],
  "cwd": "/path/to/project"
}
```

#### Build flags

| Flag | Required | Description |
|---|---|---|
| `--agent-workspace-dir` | Yes | Workspace directory (used as `--add-dir`) |
| `--working-dir` | No | Working directory for the command (`cwd` in output, passed as `-C`) |
| `--task` | No | Task prompt (positional arg to `codex exec`) |
| `--config` | No | JSON config object (default: `{}`) |
| `--role-prompt` | No | Role prompt text (written to `prompt.md`) |
| `--memory-prompt` | No | Memory/context prompt (appended to `prompt.md`) |
| `--skills-dir` | No | Directory with skill subdirectories (symlinked to `skills/`) |
| `--roles-dir` | No | Directory with role subdirectories (repeatable, symlinked to `roles/`) |
| `--agent-id` | No | Agent identifier |

## Configuration

### Config JSON

Pass via `--config`:

| Key | Type | Default | Description |
|---|---|---|---|
| `model` | string | `o3` | Model to use |
| `dangerously_skip_permissions` | boolean | `true` | Bypass all approvals and sandbox (`--dangerously-bypass-approvals-and-sandbox`). Set to `false` to require approval. |

### Environment variables

| Variable | Description |
|---|---|
| `OPENAI_API_KEY` | OpenAI API key (optional if logged in via `codex login`) |
| `SANDBOX_MODE` | Sandbox policy passed to `--sandbox` (e.g. `workspace-write`, `read-only`, `danger-full-access`) |

## Development

```sh
cd codex/wrapper
go test -v ./...
```

Releases are built with [GoReleaser](https://goreleaser.com/) and published automatically via GitHub Actions.

## License

See [LICENSE](LICENSE) for details.
