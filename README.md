# openv

A CLI tool for managing environment variables with 1Password.

## Features

- Import environment variables from .env files into 1Password
- Export environment variables from 1Password to .env files
- Run commands with environment variables from 1Password
- Supports multiple environments (production, staging, etc.)
- Integrates with various deployment platforms (GitHub, Netlify, Vercel, etc.)

## Installation

```bash
go install github.com/hinterland-software/openv@latest
```

## Development

### Prerequisites

- Go 1.21 or later
- Make

### Building

```bash
make build
```

## Usage

### Authentication

openv requires a [1Password service account token](https://my.1password.com/developer-tools/infrastructure-secrets/serviceaccount/). You can provide it in three ways:

1. Via the `--token` flag
2. Via the `OPENV_TOKEN` environment variable
3. Via interactive prompt

### Commands

#### Export Variables

```bash
openv export --url github.com/org/repo --env production -f .env
```

#### Import Variables

```bash
openv import --url github.com/org/repo --env staging --file .env.staging
```

#### Run Commands

```bash
openv run --url github.com/org/repo --env production -- npm start
```

### Configuration

Create a config file at `~/.openv.yaml` to set default values:

```yaml
token: op-service-account-token
vault: my-vault-name
```

### Environment Variables

Environment variables can be set in the config file or passed as flags.

```bash
export OPENV_TOKEN=op-service-account-token
```

### Logging

openv supports different verbosity levels for logging:

- `--verbose` or `-v`: Enable debug logging with detailed information
- `--quiet` or `-q`: Suppress all output except errors
- Default: Normal information level logging

Example with verbose logging:

```bash
openv export --url github.com/org/repo --env production -v
```

Example with quiet logging:

```bash
openv export --url github.com/org/repo --env production -q
```

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.

## License

MIT

## Security Considerations

- Store your 1Password service account token securely
- Use environment variables or config files with appropriate permissions
- Avoid logging sensitive data when using verbose mode
- Review environment variables before syncing with deployment platforms
- Keep the CLI and its dependencies updated

## Support & Security

For security issues, please email security@hinterland-software.com or use GitHub's security advisory feature.
Do not report security vulnerabilities through public GitHub issues.
