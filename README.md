# strawpot_codex_cli

StrawPot wrapper for [OpenAI Codex CLI](https://github.com/openai/codex). Translates StrawPot's agent protocol into Codex CLI flags.

## Overview

This wrapper provides two subcommands:

- **`setup`** — Runs `codex login` for interactive authentication
- **`build`** — Translates StrawPot protocol args to a Codex CLI command, returning JSON: `{"cmd": [...], "cwd": "..."}`

## Install

```sh
curl -fsSL https://raw.githubusercontent.com/strawpot/strawpot_codex_cli/main/strawpot_codex/install.sh | sh
```

## Development

```sh
cd codex/wrapper
go test -v ./...
go build -o strawpot_codex .
```

## License

MIT
