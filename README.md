# mava-cli

`mava-cli` is a command-line tool for managing [Mava](https://mava.app) support tickets.

It supports:
- listing and filtering tickets
- viewing ticket details and message timelines (with attachments, tags, category, CSAT)
- searching messages by content, customer name, or custom attributes
- replying to tickets (including internal notes)
- assigning tickets to team members (by name, case-insensitive)
- updating ticket status
- listing team members dynamically from the API
- listing all tickets for a specific customer
- marking messages as read
- checking login state and org info

---

## Install

### Option A: Download from GitHub Releases

```bash
gh release list -R Leechael/mava-cli
TAG="vX.Y.Z"
gh release download "$TAG" -R Leechael/mava-cli --pattern "mava-cli-*.tar.gz"
```

Extract the archive for your platform and place `mava-cli` in your `PATH`.

### Option B: Build from source

```bash
git clone git@github.com:Leechael/mava-cli.git
cd mava-cli
make build
# binary is at bin/mava-cli
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

### Auth

- `mava-cli status` — check login state and show current user, org, and plan info

### Ticket management

- `mava-cli list` — list tickets with filters (status, priority, source, assigned-to, etc.)
- `mava-cli list --todo` — show tickets that need a human reply
- `mava-cli get <ticket-id>` — view ticket details and message timeline
- `mava-cli reply <ticket-id> [message]` — reply to a ticket (reads from stdin if message omitted)
- `mava-cli reply <ticket-id> --internal [message]` — send an internal note
- `mava-cli update-status <ticket-id> <status>` — update ticket status (Open, Pending, Waiting, Resolved, Spam)
- `mava-cli assign <ticket-id> <agent>` — assign ticket to an agent by name or ID
- `mava-cli mark-read <ticket-id> <message-id>...` — mark messages as read

### Search

- `mava-cli search <query>` — search by message content (default)
- `mava-cli search <query> --by customer` — search by customer name
- `mava-cli search <query> --by attributes` — search by custom attribute values
- `mava-cli search <query> --by customer --skip 10` — paginate results

### Customer

- `mava-cli list-customer-tickets <customer-id>` — list all tickets for a customer
- `mava-cli list-customer-tickets <customer-id> --skip <ticket-id>` — paginate (cursor is a ticket ID)

### Team

- `mava-cli list-members` — list all team members
- `mava-cli list-members --include-archived` — include archived members
- `mava-cli list-members --json` — output as JSON

### Output modes

- Default output is human-readable plain text
- `--json` — parseable JSON output
- `--jq <filter>` — apply jq filter to JSON output (on `list`, `get`, `search`, `list-customer-tickets`)

---

## Usage examples

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
