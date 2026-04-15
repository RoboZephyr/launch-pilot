# Launch Pilot

Visual control console for macOS launchd services. View status, read logs, diagnose issues, and manage user-domain launch agents from your browser.

## Install

```bash
brew install A404coder/tap/launch-pilot
```

Or build from source:

```bash
make build
```

## Usage

```bash
launch-pilot              # random port, opens browser
launch-pilot --port 8080  # explicit port
launch-pilot --no-open    # skip auto-opening browser
launch-pilot --version    # print version
```

## License

MIT
