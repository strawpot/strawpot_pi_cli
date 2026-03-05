---
name: strawpot-pi
description: Pi coding agent
metadata:
  version: "0.1.0"
  strawpot:
    bin:
      macos: strawpot_pi
      linux: strawpot_pi
    install:
      macos: curl -fsSL https://raw.githubusercontent.com/strawpot/strawpot_pi_cli/main/strawpot_pi/install.sh | sh
      linux: curl -fsSL https://raw.githubusercontent.com/strawpot/strawpot_pi_cli/main/strawpot_pi/install.sh | sh
    tools:
      pi:
        description: Pi Coding Agent
        install:
          macos: npm install -g @mariozechner/pi-coding-agent
          linux: npm install -g @mariozechner/pi-coding-agent
    params:
      model:
        type: string
        default: claude-sonnet-4-6
        description: Model to use for Pi coding agent
      dangerously_skip_permissions:
        type: boolean
        default: true
        description: Skip permission prompts (pi auto-approves in non-interactive mode)
    env:
      ANTHROPIC_API_KEY:
        required: false
        description: Anthropic API key (optional if logged in via OAuth)
---

# Pi Coding Agent

Runs Pi coding-agent as a subprocess. Supports interactive and non-interactive
modes, custom model selection, and skill-based prompt augmentation.
