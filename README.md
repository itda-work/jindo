# jd - Claude Code Configuration Manager

<p align="center">
  <img src="assets/logo.png" alt="jd logo" width="200">
</p>

A CLI tool for managing Claude Code configurations including skills, commands, and agents.

## Features

- **Skills Management**: Create, edit, list, and delete Claude Code skills
- **Commands Management**: Manage slash commands for Claude Code
- **Agents Management**: Configure and manage Claude Code agents
- **Search**: Search across all resources by keyword
- **Validation**: Validate format and content of all configurations
- **AI-Assisted Creation**: Use Claude CLI for interactive skill/command/agent creation

## Installation

### Quick Install (Recommended)

#### macOS / Linux

```bash
curl -fsSL https://cdn.jsdelivr.net/gh/itda-work/itda-jindo@main/install.sh | bash
```

#### Windows (PowerShell)

```powershell
irm https://cdn.jsdelivr.net/gh/itda-work/itda-jindo@main/install.ps1 | iex
```

### Custom Installation

#### macOS / Linux

```bash
# Install to a custom directory
JD_INSTALL_DIR=~/bin curl -fsSL https://cdn.jsdelivr.net/gh/itda-work/itda-jindo@main/install.sh | bash

# Install a specific version
VERSION=v0.1.0 curl -fsSL https://cdn.jsdelivr.net/gh/itda-work/itda-jindo@main/install.sh | bash
```

#### Windows (PowerShell)

```powershell
# Install to a custom directory
$env:JD_INSTALL_DIR = "C:\tools"; irm https://cdn.jsdelivr.net/gh/itda-work/itda-jindo@main/install.ps1 | iex

# Install a specific version
$env:VERSION = "v0.1.0"; irm https://cdn.jsdelivr.net/gh/itda-work/itda-jindo@main/install.ps1 | iex
```

### Build from Source

```bash
git clone https://github.com/itda-work/itda-jindo.git
cd itda-jindo
make build
```

## Usage

### Subcommand Aliases

For convenience, short aliases are available:

- `skills` → `s`
- `commands` → `c`
- `agents` → `a`

### Skills

Skills are reusable prompts stored in `~/.claude/skills/<name>/SKILL.md`.

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

Commands are slash commands stored in `~/.claude/commands/`.

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

Agents are AI configurations stored in `~/.claude/agents/`.

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
└── agents/
    └── <agent>.md
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

## Development

### Prerequisites

- Go 1.21+
- Python 3.x (for pre-commit)

### Setup

```bash
# Clone the repository
git clone https://github.com/itda-work/itda-jindo.git
cd itda-jindo

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
