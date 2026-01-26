# jd - Claude Code Configuration Manager

<p align="center">
  <img src="assets/logo.png" alt="jd logo" width="200">
</p>

A CLI tool for managing Claude Code configurations including skills, commands, agents, and hooks.

## Features

- **Skills Management**: Create, edit, list, and delete Claude Code skills
- **Commands Management**: Manage slash commands for Claude Code
- **Agents Management**: Configure and manage Claude Code agents
- **Hooks Management**: Manage hooks in settings.json with wizard-style creation
- **Package Manager**: Install skills/commands/agents from GitHub repositories
- **Search**: Search across all resources by keyword
- **Validation**: Validate format and content of all configurations
- **AI-Assisted Creation**: Use Claude CLI for interactive skill/command/agent creation

## Installation

### Quick Install (Recommended)

#### macOS / Linux

```bash
curl -fsSL https://cdn.jsdelivr.net/gh/itda-skills/jindo@main/install.sh | bash
```

#### Windows (PowerShell)

```powershell
irm https://cdn.jsdelivr.net/gh/itda-skills/jindo@main/install.ps1 | iex
```

### Custom Installation

#### macOS / Linux

```bash
# Install to a custom directory
JD_INSTALL_DIR=~/bin curl -fsSL https://cdn.jsdelivr.net/gh/itda-skills/jindo@main/install.sh | bash

# Install a specific version
VERSION=v0.1.0 curl -fsSL https://cdn.jsdelivr.net/gh/itda-skills/jindo@main/install.sh | bash
```

#### Windows (PowerShell)

```powershell
# Install to a custom directory
$env:JD_INSTALL_DIR = "C:\tools"; irm https://cdn.jsdelivr.net/gh/itda-skills/jindo@main/install.ps1 | iex

# Install a specific version
$env:VERSION = "v0.1.0"; irm https://cdn.jsdelivr.net/gh/itda-skills/jindo@main/install.ps1 | iex
```

### Build from Source

```bash
git clone https://github.com/itda-skills/jindo.git
cd jindo
make build
```

## Usage

### Subcommand Aliases

For convenience, short aliases are available:

- `skills` → `s`
- `commands` → `c`
- `agents` → `a`
- `hooks` → `h`
- `pkg` → `p`
- `list` → `l`, `ls`

Sub-command aliases (common across all resource types):

- `list` → `l`, `ls`
- `new` → `n`, `add`, `create`
- `show` → `s`, `get`, `view`
- `edit` → `e`, `update`, `modify`
- `delete` → `d`, `rm`, `remove`

### Default Scope

If a `.claude/` directory exists in your current working directory, `jd` commands default to **local** scope (`.claude/`).
Otherwise they default to **global** scope (`~/.claude/`).

Use `--local` or `--global` to override.

### List All

Quickly list all skills, agents, commands, and hooks.

```bash
jd list                # List all
jd l                   # short form
jd ls                  # alias
jd list --json         # JSON output
```

### Skills

Skills are reusable prompts stored in `~/.claude/skills/<name>/SKILL.md` (global) or `.claude/skills/<name>/SKILL.md` (local).

```bash
# List all skills
jd skills list
jd s list              # short form
jd s list --json       # JSON output

# Show skill details
jd s show <skill-name>
jd s show <skill-name> --brief

# Create a new skill (AI-assisted)
jd s new my-skill

# Create a new skill (template only)
jd s new my-skill --no-ai
jd s new my-skill --no-ai -d "Description" -t "Bash, Read, Write"

# Edit a skill (AI-assisted)
jd s edit my-skill

# Edit a skill in editor
jd s edit my-skill --editor

# Delete a skill
jd s delete my-skill
jd s rm my-skill -f    # skip confirmation
```

### Commands

Commands are slash commands stored in `~/.claude/commands/` (global) or `.claude/commands/` (local).

```bash
# List all commands
jd commands list
jd c list

# Show command details
jd c show <command-name>
jd c show game:asset   # subdirectory command

# Create a new command
jd c new my-command
jd c new my-command --no-ai

# Edit a command
jd c edit my-command
jd c edit my-command --editor

# Delete a command
jd c delete my-command
jd c rm my-command -f
```

### Agents

Agents are AI configurations stored in `~/.claude/agents/` (global) or `.claude/agents/` (local).

```bash
# List all agents
jd agents list
jd a list

# Show agent details
jd a show <agent-name>

# Create a new agent
jd a new my-agent
jd a new my-agent --no-ai -d "Description" -m "claude-sonnet-4-20250514"

# Edit an agent
jd a edit my-agent
jd a edit my-agent --editor

# Delete an agent
jd a delete my-agent
jd a rm my-agent -f
```

### Hooks

Hooks are event-driven scripts configured in `~/.claude/settings.json` (global) or `.claude/settings.json` (local).

```bash
# List all hooks
jd hooks list
jd h list
jd h l --json          # JSON output

# Show hook details
jd h show <hook-name>
jd h s PreToolUse-Bash-0 --json

# Create a new hook (wizard mode)
jd h new

# Create a hook with flags
jd h new -e pre -m "Bash" -c "echo 'Running bash'"
jd h new -e post -m "Bash|Write" -c "~/.claude/hooks/log.sh"
jd h new -e post -m "Bash" --script   # Auto-create script file

# Edit a hook
jd h edit <hook-name>
jd h edit PreToolUse-Bash-0 -m "Bash|Edit"
jd h edit PreToolUse-Bash-0 -c "new-command.sh"

# Delete a hook
jd h delete <hook-name>
jd h rm PreToolUse-Bash-0 -f   # skip confirmation
```

**Event Types (with aliases):**

| Event        | Alias    | Description                    |
| ------------ | -------- | ------------------------------ |
| PreToolUse   | `pre`    | Runs before a tool is executed |
| PostToolUse  | `post`   | Runs after a tool is executed  |
| Notification | `notify` | Runs on notifications          |
| Stop         | -        | Runs when Claude stops         |
| SubagentStop | `sub`    | Runs when a subagent stops     |

**Matcher Patterns:**

- Single tool: `"Bash"`, `"Write"`, `"Edit"`
- Multiple tools: `"Bash|Write|Edit"` (regex OR)
- All tools: `"*"`

**Environment Variables (available in hook scripts):**

- `$TOOL_NAME` - Name of the tool being called
- `$TOOL_INPUT` - JSON input to the tool
- `$TOOL_OUTPUT` - JSON output from the tool (PostToolUse only)

### Package Manager

Install and manage skills, commands, and agents from GitHub repositories.

```bash
# Register a repository
jd pkg repo add gh:owner/repo
jd p r add gh:affaan-m/everything-claude-code
jd p r add gh:user/claude-skills --namespace mysk

# List registered repositories
jd p r list
jd p r ls --json

# Update repository index
jd p r update
jd p r up my-namespace

# Remove a repository
jd p r remove <namespace>

# Browse packages (TUI)
jd p browse
jd p b

# Browse a specific repository
jd p b <namespace>
jd p b affa-ever --type skills

# Search packages
jd p search <query>
jd p se web --json

# Install a package
jd p install <namespace>:<path>
jd p i affa-ever:skills/web-fetch
jd p i affa-ever:commands/commit.md
jd p i affa-ever:skills/web-fetch@v1.2.0   # specific version

# List installed packages
jd p list
jd p ls --json

# Show package info
jd p info <name>
jd p in affa-ever--web-fetch --json

# Update packages
jd p update                      # Check all packages
jd p up affa-ever--web-fetch     # Check specific package
jd p up --apply                  # Apply all updates

# Uninstall a package
jd p uninstall <name>
jd p un affa-ever--web-fetch
```

### Search

Search across all skills, commands, and agents.

```bash
# Search all resources
jd search <keyword>

# Search specific resource types
jd search <keyword> -s    # skills only
jd search <keyword> -c    # commands only
jd search <keyword> -a    # agents only

# Search names only (not content)
jd search <keyword> -n
```

### Validate

Validate the format and content of all configurations.

```bash
# Validate all
jd validate

# Validate specific types
jd validate -s    # skills only
jd validate -c    # commands only
jd validate -a    # agents only

# Verbose output
jd validate -v
```

### Update

Update jd to the latest version.

```bash
# Check and update interactively
jd update
jd up

# Check for updates only
jd up --check
jd up -c

# Update without confirmation
jd up -y
jd up --force

# Update to a specific version
jd up -v v0.2.0

# Update using OS install script (curl/PowerShell)
jd up --script
```

## File Structure

```text
~/.claude/
├── skills/
│   └── <skill-name>/
│       └── SKILL.md
├── commands/
│   ├── <command>.md
│   └── <subdir>/
│       └── <command>.md
├── agents/
│   └── <agent>.md
├── hooks/                    # Hook scripts (auto-created by jd hooks new --script)
│   └── <event>-<matcher>.sh
└── settings.json             # Contains hooks configuration

~/.itda-skills/                # Package manager data
├── repos.json                # Registered repositories
├── packages.json             # Installed packages metadata
├── cache/                    # Repository cache
└── skills/                   # Installed skills (with namespace prefix)
    └── <namespace>--<name>/
```

### Skill File Format (SKILL.md)

```markdown
---
name: my-skill
description: A brief description of the skill
allowed-tools: Bash, Read, Write, Edit, Glob, Grep
---

# My Skill

## Overview

Describe what this skill does.

## Usage

Explain how to use this skill.
```

### Command File Format

```markdown
---
description: A brief description of the command
---

# My Command

Instructions for the command...
```

### Agent File Format

```markdown
---
name: my-agent
description: A brief description of the agent
model: claude-sonnet-4-20250514
---

# My Agent

Agent system prompt and instructions...
```

### Hook Configuration (settings.json)

Hooks are configured in `~/.claude/settings.json` (global) or `.claude/settings.json` (local):

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "echo 'Running Bash command'"
          }
        ]
      },
      {
        "matcher": "Bash|Write|Edit",
        "hooks": [
          {
            "type": "command",
            "command": "~/.claude/hooks/pretooluse-bash-write-edit.sh"
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "~/.claude/hooks/log-all.sh"
          }
        ]
      }
    ]
  }
}
```

## Development

### Prerequisites

- Go 1.21+
- Python 3.x (for pre-commit)

### Setup

```bash
# Clone the repository
git clone https://github.com/itda-skills/jindo.git
cd jindo

# Install pre-commit hooks
pip install pre-commit
pre-commit install

# Build
make build

# Run linter
make lint

# Run tests
make test
```

### Makefile Targets

- `make build` - Build the binary
- `make lint` - Run golangci-lint
- `make test` - Run tests
- `make clean` - Remove build artifacts

## License

MIT License
