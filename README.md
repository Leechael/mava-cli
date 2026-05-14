# mava-cli

> mava-cli is an open-source command-line tool for managing [Mava](https://mava.app) customer support tickets — built for support teams and developers who prefer terminal workflows over browser tabs.

[![GitHub stars](https://img.shields.io/github/stars/Leechael/mava-cli?style=social)](https://github.com/Leechael/mava-cli)
[![Build](https://img.shields.io/github/actions/workflow/status/Leechael/mava-cli/go-ci.yml?style=flat-square&label=build)](https://github.com/Leechael/mava-cli/actions)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/Leechael/mava-cli?style=flat-square)](https://github.com/Leechael/mava-cli/releases)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.22-blue?style=flat-square)](https://go.dev)

[![Star History Chart](https://api.star-history.com/svg?repos=Leechael/mava-cli&type=Date)](https://star-history.com/#Leechael/mava-cli&Date)

## ✨ Features

| Feature | Description |
|---------|-------------|
| 🎫 List & Filter Tickets | View tickets by status, priority, source, assignment, and more |
| 🔍 Full-Text Search | Search messages, customers, and custom attributes |
| 💬 Reply from Terminal | Send public replies and internal notes without opening a browser |
| 👤 Customer Timeline | View all tickets for a specific customer at a glance |
| 👥 Team Management | List members and assign tickets by name (case-insensitive) |
| 📊 JSON + jq Support | Pipe output to `jq` for scripting and automation |

## 🚀 Quick Start

### 1. Install

```bash
# macOS / Linux — auto-detect platform
TAG=$(gh release view -R Leechael/mava-cli --json tagName -q .tagName)
gh release download "$TAG" -R Leechael/mava-cli --pattern "mava-cli-*-$(uname -s)-$(uname -m).tar.gz"
tar -xzf mava-cli-*.tar.gz && mv mava-cli /usr/local/bin/
```

Or download manually from the [Releases](https://github.com/Leechael/mava-cli/releases) page.

### 2. Configure

```bash
export MAVA_TOKEN="<your-jwt-token>"
```

Get your token from [dashboard.mava.app](https://dashboard.mava.app) → DevTools → `x-auth-token` cookie.

### 3. Run

```bash
mava-cli status
mava-cli list --todo
```

⚡ Done! You're ready to manage tickets from the terminal.

## 📖 Commands

### Ticket Management

| Command | Description |
|---------|-------------|
| `mava-cli list` | List tickets with filters (status, priority, source, assigned-to, etc.) |
| `mava-cli list --todo` | Show tickets that need a human reply |
| `mava-cli get <ticket-id>` | View ticket details, message timeline, attachments, tags, category, CSAT |
| `mava-cli reply <ticket-id> [message]` | Reply to a ticket (reads from stdin if message omitted) |
| `mava-cli reply <ticket-id> --internal [message]` | Send an internal note |
| `mava-cli edit <ticket-id> <message-id> [message]` | Edit a message (reads from stdin if message omitted) |
| `mava-cli delete <ticket-id> <message-id>` | Delete a message |
| `mava-cli update-status <ticket-id> <status>` | Update status: Open, Pending, Waiting, Resolved, Spam |
| `mava-cli assign <ticket-id> <agent>` | Assign ticket to an agent by name or ID |
| `mava-cli mark-read <ticket-id> <message-id>...` | Mark messages as read |

### Account & Configuration

| Command | Description |
|---------|-------------|
| `mava-cli integrations` | List connected channels (Discord, Telegram, WebChat, Email, etc.) |
| `mava-cli attributes` | List custom attribute definitions (the schema for `search --by attributes`) |

### Search & Customer

| Command | Description |
|---------|-------------|
| `mava-cli search <query>` | Search by message content (default), customer name, or custom attributes |
| `mava-cli list-customer-tickets <customer-id>` | List all tickets for a customer |

### Team & Auth

| Command | Description |
|---------|-------------|
| `mava-cli status` | Check login state and show user, org, and plan info |
| `mava-cli list-members` | List all team members (supports `--include-archived`, `--json`) |

### Output Modes

- Default: human-readable plain text
- `--json`: parseable JSON output
- `--jq <filter>`: apply jq filter to JSON output

## 💡 Usage Examples

```bash
# check login state
mava-cli status

# list open tickets
mava-cli list --status Open

# list tickets assigned to you that need reply
mava-cli list --todo

# view a ticket (shows tags, category, CSAT, and attachments if present)
mava-cli get 69a5592c9927182b6142cff2

# search messages by content
mava-cli search "API key"

# search by customer name
mava-cli search "hugo" --by customer

# search by custom attributes
mava-cli search "Diamond" --by attributes

# reply to a ticket
mava-cli reply 69a5592c9927182b6142cff2 "Thanks for reaching out!"

# reply from stdin (pipe in a file, editor output, etc.)
cat response.md | mava-cli reply 69a5592c9927182b6142cff2

# send internal note
mava-cli reply 69a5592c9927182b6142cff2 --internal "Escalating to eng team"

# edit a message
mava-cli edit 69a5592c9927182b6142cff2 <message-id> "Updated content"

# edit from stdin
cat revised.md | mava-cli edit 69a5592c9927182b6142cff2 <message-id>

# delete a message
mava-cli delete 69a5592c9927182b6142cff2 <message-id>

# list connected integrations
mava-cli integrations

# list custom attribute definitions
mava-cli attributes

# assign ticket (case-insensitive name matching)
mava-cli assign 69a5592c9927182b6142cff2 paco

# update status
mava-cli update-status 69a5592c9927182b6142cff2 Resolved

# mark messages as read
mava-cli mark-read 69a5592c9927182b6142cff2 <message-id> <message-id>

# list all tickets for a customer
mava-cli list-customer-tickets 69c6435b1480f431002df5c7

# list team members
mava-cli list-members

# JSON output with jq
mava-cli list --json --jq '.tickets[0].customer'
mava-cli get 69a5592c9927182b6142cff2 --json --jq '.messages | length'
mava-cli search "hugo" --by customer --json --jq '.[].status'
```

## 🤝 Contributing

1. Fork the repo
2. Create a branch: `git checkout -b feature/your-feature`
3. Run tests: `make test`
4. Submit a PR with a clear description of your change

We respond to all PRs within 48 hours.

## 📄 License

MIT © [Leechael](https://github.com/Leechael)
