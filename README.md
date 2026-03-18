# mava

`mava` is a command-line tool for managing [Mava](https://mava.app) support tickets.

It supports:
- listing and filtering tickets
- viewing ticket details and message timelines
- searching messages by content
- replying to tickets (including internal notes)
- assigning tickets to team members (by name, case-insensitive)
- updating ticket status
- listing team members dynamically from the API

---

## Install

### Option A: Download from GitHub Releases

```bash
gh release list -R Leechael/mava-cli
TAG="vX.Y.Z"
gh release download "$TAG" -R Leechael/mava-cli --pattern "mava-*.tar.gz"
```

Extract the archive for your platform and place `mava` in your `PATH`.

### Option B: Build from source

```bash
git clone git@github.com:Leechael/mava-cli.git
cd mava-cli
make build
# binary is at bin/mava
```

---

## Configuration

Set your Mava auth token via environment variable:

```bash
export MAVA_TOKEN="<your-jwt-token>"
```

To obtain a token, log in to [dashboard.mava.app](https://dashboard.mava.app), open browser DevTools, and copy the `x-auth-token` cookie value.

---

## Commands

### Ticket management

- `mava list` — list tickets with filters (status, priority, source, assigned-to, etc.)
- `mava list --todo` — show tickets that need a human reply
- `mava get <ticket-id>` — view ticket details and message timeline
- `mava search <query>` — search messages by content
- `mava reply <ticket-id> [message]` — reply to a ticket (reads from stdin if message omitted)
- `mava reply <ticket-id> --internal [message]` — send an internal note
- `mava update-status <ticket-id> <status>` — update ticket status (Open, Pending, Waiting, Resolved, Spam)
- `mava assign <ticket-id> <agent>` — assign ticket to an agent by name or ID

### Team

- `mava list-members` — list all team members
- `mava list-members --include-archived` — include archived members
- `mava list-members --json` — output as JSON

### Output modes

- Default output is human-readable plain text
- `--json` — parseable JSON output
- `--jq <filter>` — apply jq filter to JSON output (on `list`, `get`, `search`)

---

## Usage examples

```bash
# list open tickets
mava list --status Open

# list tickets assigned to you that need reply
mava list --todo

# view a ticket
mava get 69a5592c9927182b6142cff2

# search for messages
mava search "API key"

# reply to a ticket
mava reply 69a5592c9927182b6142cff2 "Thanks for reaching out!"

# reply from stdin (pipe in a file, editor output, etc.)
cat response.md | mava reply 69a5592c9927182b6142cff2

# send internal note
mava reply 69a5592c9927182b6142cff2 --internal "Escalating to eng team"

# assign ticket (case-insensitive name matching)
mava assign 69a5592c9927182b6142cff2 paco
mava assign 69a5592c9927182b6142cff2 Hugo

# update status
mava update-status 69a5592c9927182b6142cff2 Resolved

# list team members
mava list-members

# JSON output with jq
mava list --json --jq '.tickets[0].customer'
mava get 69a5592c9927182b6142cff2 --json --jq '.messages | length'
```
