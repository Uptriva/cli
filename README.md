# uptriva

A command-line client for the [Uptriva](https://uptriva.com) monitoring API.
Create monitors, check incidents and manage status pages without opening a browser.

## Install

```sh
go install github.com/uptriva/cli/cmd/uptriva@latest
```

Or grab a binary from the [releases](https://github.com/uptriva/cli/releases) page.

## Getting started

```sh
uptriva auth login --token uk_your_token
```

Tokens are created in the dashboard under Security → API tokens. Scoped tokens
(`monitors:write`, `incidents:read`, page-restricted) work too — the API enforces
them server-side, so a read-only key can't accidentally write anything.

`UPTRIVA_TOKEN` and `UPTRIVA_API_URL` environment variables are also honoured.

## Commands

```
uptriva auth login --token <token>     store your API token
uptriva auth logout                    remove it
uptriva monitors list                  list monitors, table output
uptriva monitors list -o json          same, machine-readable
uptriva monitors list --status-page <id>
```

More subcommands land as the API grows. `-o json` works everywhere and pipes
straight into `jq`.

## Config

Stored at `~/.config/uptriva/config.yaml` (0600). Only the token and the API
base URL live there — nothing else leaves your machine.

## Development

```sh
go build ./cmd/uptriva
go test ./...
```

Releases are cut with GoReleaser on `v*` tags via GitHub Actions.

## License

MIT
