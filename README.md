# StrawPot Pi CLI

A Go wrapper that translates [StrawPot](https://github.com/strawpot) protocol arguments into [Pi coding-agent](https://github.com/badlogic/pi-mono/tree/main/packages/coding-agent) CLI flags. It acts as a pure translation layer — process management, sessions, and infrastructure are handled by StrawPot core.

## Prerequisites

- [Pi coding-agent](https://www.npmjs.com/package/@mariozechner/pi-coding-agent) (`npm install -g @mariozechner/pi-coding-agent`)
- An Anthropic API key (or OAuth login)

## Installation

```sh
curl -fsSL https://raw.githubusercontent.com/strawpot/strawpot_pi_cli/main/strawpot_pi/install.sh | sh
```

This downloads a pre-built binary for your platform (macOS/Linux, amd64/arm64) to `/usr/local/bin`. Override the install directory with `INSTALL_DIR`:

```sh
INSTALL_DIR=~/.local/bin curl -fsSL ... | sh
```

## Usage

The wrapper exposes two subcommands:

### `setup`

Runs `pi login` to authenticate.

```sh
strawpot_pi setup
```

### `build`

Translates StrawPot protocol flags into a Pi CLI command and outputs it as JSON.

```sh
strawpot_pi build \
  --agent-workspace-dir /path/to/workspace \
  --working-dir /path/to/project \
  --task "fix the bug" \
  --config '{"model":"claude-sonnet-4-6"}'
```

Output:

```json
{
  "cmd": ["pi", "-p", "fix the bug", "--model", "claude-sonnet-4-6"],
  "cwd": "/path/to/project"
}
```

#### Build flags

| Flag | Required | Description |
|---|---|---|
| `--agent-workspace-dir` | Yes | Workspace directory for prompts and symlinks |
| `--working-dir` | No | Working directory for the command (`cwd` in output) |
| `--task` | No | Task prompt (passed as `pi -p`) |
| `--config` | No | JSON config object (default: `{}`) |
| `--role-prompt` | No | Role prompt text (written to `AGENTS.md`) |
| `--memory-prompt` | No | Memory/context prompt (appended to `AGENTS.md`) |
| `--skills-dir` | No | Directory with skill subdirectories (symlinked to `skills/`) |
| `--roles-dir` | No | Directory with role subdirectories (repeatable, symlinked to `roles/`) |
| `--agent-id` | No | Agent identifier |

## Configuration

### Config JSON

Pass via `--config`:

| Key | Type | Default | Description |
|---|---|---|---|
| `model` | string | `claude-sonnet-4-6` | Model to use |
| `dangerously_skip_permissions` | boolean | `true` | Pi auto-approves in non-interactive (`-p`) mode — this flag is accepted for protocol compatibility. |

### Environment variables

| Variable | Description |
|---|---|
| `ANTHROPIC_API_KEY` | Anthropic API key (optional if logged in via OAuth) |

### Notes

- Pi coding-agent automatically discovers `AGENTS.md` files in the workspace directory and uses them as system instructions.
- In non-interactive mode (`-p`), Pi auto-approves all tool calls — there is no separate permission flag.

## Development

```sh
cd pi/wrapper
go test -v ./...
```

Releases are built with [GoReleaser](https://goreleaser.com/) and published automatically via GitHub Actions.

## Related

- [Pi coding-agent](https://github.com/badlogic/pi-mono/tree/main/packages/coding-agent) — the upstream coding agent this wrapper integrates
- [Pi on npm](https://www.npmjs.com/package/@mariozechner/pi-coding-agent) — npm package for installation
- [pi-mono](https://github.com/badlogic/pi-mono) — the monorepo containing Pi and related packages

## License

See [LICENSE](LICENSE) for details.
