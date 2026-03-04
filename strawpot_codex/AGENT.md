---
name: strawpot-codex
description: OpenAI Codex CLI agent
metadata:
  version: "0.1.0"
  strawpot:
    bin:
      macos: strawpot_codex
      linux: strawpot_codex
    install:
      macos: curl -fsSL https://raw.githubusercontent.com/strawpot/strawpot_codex_cli/main/strawpot_codex/install.sh | sh
      linux: curl -fsSL https://raw.githubusercontent.com/strawpot/strawpot_codex_cli/main/strawpot_codex/install.sh | sh
    tools:
      codex:
        description: OpenAI Codex CLI
        install:
          macos: npm install -g @openai/codex
          linux: npm install -g @openai/codex
    params:
      model:
        type: string
        default: o3
        description: Model to use for Codex CLI
      dangerously_skip_permissions:
        type: boolean
        default: true
        description: Bypass all approvals and sandbox (enabled by default, set to false to require approval)
    env:
      OPENAI_API_KEY:
        required: false
        description: OpenAI API key (optional if logged in via codex login)
---

# Codex CLI Agent

Runs OpenAI Codex CLI as a subprocess. Supports non-interactive execution,
custom model selection, and instruction-based prompt augmentation.
