# Epoch - AI-Native Git Client

A pure Go implementation of Git with AI-powered features for intelligent commit management.

## Features

- **Auto-Commit Generation**: Generate meaningful commit messages from diffs using Cerebras LLM
- **Commit Rewording**: AI-powered commit message improvement following Conventional Commits
- **Auto-Squash**: Automatically detect and squash similar commits based on temporal proximity
- **Pure Implementation**: Custom Git object database, no dependency on system git

## Requirements

- Go 1.25+
- Cerebras API key (free at [cloud.cerebras.ai](https://cloud.cerebras.ai))

## Installation

```bash
go build -o epoch .
```

## Configuration

Create `epoch.toml` in your project root:

```toml
[ai]
provider = "cerebras"
model = "llama3.1-8b"
temperature = 0.3

[squash]
window_minutes = 10
always_preview = true

[commits]
enforce_conventional = true
allowed_types = ["feat", "fix", "docs", "style", "refactor", "perf", "test", "chore", "build", "ci", "revert"]
```

## Environment

Set your Cerebras API key:

```bash
export CEREBRAS_API_KEY=your-api-key
```

## Usage

### Initialize Repository
```bash
epoch init
```

### Stage Files
```bash
epoch add .
epoch add file1.txt file2.txt
```

### Show Status
```bash
epoch status
epoch status --branch
```

### Show Diff
```bash
epoch diff           # unstaged changes
epoch diff --staged  # staged changes
```

### Create Commit
```bash
epoch commit -m "feat: add new feature"
epoch commit --auto  # AI-generated commit message
epoch commit --auto --yes  # skip confirmation
```

### Reword Commit
```bash
epoch reword HEAD              # edit manually
epoch reword --auto HEAD       # AI-improved message
epoch reword --auto HEAD --yes # skip confirmation
```

### Auto-Squash
```bash
epoch squash              # auto-squash with preview
epoch squash --window=5  # 5-minute window
epoch squash --dry-run   # preview only
```

## Commands

| Command | Description |
|---------|-------------|
| `init` | Initialize a new repository |
| `add` | Add files to staging area |
| `status` | Show working tree status |
| `diff` | Show changes |
| `commit` | Create a commit |
| `reword` | Rewrite commit message |
| `squash` | Auto-squash commits |

## Architecture

```
epoch/
├── main.go           # CLI entry point
├── config/           # Configuration parsing
├── git/              # Git object model
│   ├── blob.go      # Blob objects
│   ├── tree.go      # Tree objects
│   ├── commit.go    # Commit objects
│   ├── index.go     # Staging area
│   ├── object_store.go  # Object storage
│   ├── ref.go       # Ref management
│   └── diff.go      # Diff generation
└── llm/             # AI integration
    ├── client.go    # Cerebras client
    └── prompts.go   # Prompt templates
```

## License

MIT