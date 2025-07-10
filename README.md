# tg-spec-cli

[![CI](https://github.com/superboomer/tg-spec-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/superboomer/tg-spec-cli/actions/workflows/ci.yml)
[![Coverage Status](https://coveralls.io/repos/github/superboomer/tg-spec-cli/badge.svg?branch=master)](https://coveralls.io/github/superboomer/tg-spec-cli?branch=master)

CLI utility for generating OpenAPI specifications from the official Telegram Bot API documentation.

## Features
- Fetches and parses the latest Telegram Bot API documentation
- Generates OpenAPI (Swagger) specifications

## Installation

Clone the repository and build the CLI:

```sh
git clone https://github.com/superboomer/tg-spec-cli.git
cd tg-spec-cli
go build ./...
```

## Usage

```sh
./tg-spec-cli generate [flags]
```

### Flags
- `-o`, `--output`   Output path for the OpenAPI specification. You can specify a directory or a full file path. If the path contains `%v`, it will be replaced with the API version (e.g., `./specs/bot-api-%v.json`). If a directory does not exist, it will be created automatically. If `%v` is not present, the file will be saved with the exact name you provide.
- `-l`, `--log-level`  Log level: `silent`, `debug`, `info`, `warn`, `error`, `fatal` (default: `info`). Use `debug` for maximum details about the generation process. Use `silent` to disable all log output.
- `-u`, `--url`      URL of the Telegram Bot API documentation (default: `https://core.telegram.org/bots/api`).
- `-t`, `--type`     API type: `botapi` (default) for the standard Telegram Bot API, or `gateway` for the Telegram Gateway API (experimental, uses https://core.telegram.org/gateway/api).

### Example

```sh
# Save to ./specs/bot-api-9.1.json (version will be inserted)
./tg-spec-cli generate -o ./specs/bot-api-%v.json -l debug

# Save to ./bot-api.json (no version in filename)
./tg-spec-cli generate -o ./bot-api.json

# Run with no log output
./tg-spec-cli generate -l silent

# Generate OpenAPI spec for the Telegram Gateway API
./tg-spec-cli generate -t gateway -o ./specs/gateway-api-%v.json

# Save to current directory with default name and version (standard Bot API)
./tg-spec-cli generate
```

## Project Structure
- `cmd/cli/` — CLI entrypoint and commands
- `internal/app/` — Application logic
- `internal/generator/` — OpenAPI generator
- `internal/telegram/` — Telegram API parsing
- `internal/logger/` — Logging setup

## License
MIT
