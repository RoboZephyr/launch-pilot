# Launchboard

Visual dashboard for macOS launchd services. View status, read logs, diagnose issues, and manage user-domain launch agents from your browser.

## Install

```bash
brew install A404coder/tap/launchboard
```

Or build from source:

```bash
make build
```

## Usage

```bash
launchboard              # random port, opens browser
launchboard --port 8080  # explicit port
launchboard --no-open    # skip auto-opening browser
launchboard --version    # print version
```

## License

MIT
