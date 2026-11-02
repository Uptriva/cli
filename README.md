# uptriva

A command-line client for the [Uptriva](https://uptriva.com) monitoring API.
Manage monitors, incidents, contacts and status pages without opening a browser.

## Install

```sh
go install github.com/uptriva/cli/cmd/uptriva@latest
```

Or grab a binary from the [releases](https://github.com/uptriva/cli/releases) page.

## Getting started

```sh
uptriva auth login
```

It prompts for your token — no secrets in your shell history. Tokens come from
the dashboard under Security → API tokens. Scoped tokens (`monitors:write`,
page-restricted) work as you'd expect; the API enforces them server-side.

`UPTRIVA_TOKEN` and `UPTRIVA_API_URL` also work if you'd rather not store a config.

## Commands

```
uptriva monitors list                      monitors as a table
uptriva monitors list --status up --type http
uptriva monitors create --type http --target https://example.com --status-page <id>
uptriva monitors show Website              details — accepts name or id
uptriva monitors pause|resume|delete Website
uptriva monitors update Website --interval 60

uptriva watch                              live status view, refreshes every 10s
uptriva heartbeat "Pinger Falkenstein"     push a heartbeat ping

uptriva status-pages list|show|create|delete
uptriva incidents list|create|ack|resolve|update|publish
uptriva maintenance list|create|delete
uptriva contacts list|create|delete        email, slack, telegram, sms
uptriva webhooks list|create|delete|test
uptriva groups list|create|delete

uptriva agents                             fleet status
uptriva agents enroll-token --name ci      enrollment tokens
uptriva agent-events acknowledge|resolve|reopen <id>

uptriva sla get|targets                    SLA stats
uptriva overview                           account overview
uptriva auth whoami, uptriva auth logout
```

Most commands that take an id also take a name — `uptriva monitors pause Website`
works the same as the UUID. `-o json` works everywhere and pipes into `jq`.

## Config

Stored at `~/.config/uptriva/config.yaml` (0600). Token and API base URL only —
nothing else leaves your machine.

## Development

```sh
go build ./cmd/uptriva
go test ./...
```

Releases build with GoReleaser on `v*` tags via GitHub Actions.

## License

MIT
