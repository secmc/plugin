# Dragonfly Plugin System

This repository hosts the gRPC bridge that lets [Dragonfly](https://github.com/df-mc/dragonfly) talk to plugins written in any language. The host ships as a normal Dragonfly process (`cmd/main.go`), while plugins live in their own processes and communicate over protobuf streams.

## Highlights

- Works with any language that has a gRPC client (TypeScript, PHP, Rust, Go, Python, C++, etc.).
- Plugins stay isolated from the server process, so bad code cannot tank TPS.
- Hot‑reload JS/TS/PHP examples for tight feedback loops, or ship compiled binaries when you need to distribute closed-source plugins.
- Uses a strict protobuf schema with generated types under `proto/` for consistent APIs across languages.
- Ships with ready‑to‑run samples plus automation (`make proto`, `make run`) for day-one productivity.

## Prerequisites

- Go 1.22+ with `GOBIN` on your `PATH`.
- [buf](https://buf.build/docs/cli/installation/) and `protoc-gen-go` (`go install google.golang.org/protobuf/cmd/protoc-gen-go@latest`).

## Setup

```bash
git clone https://github.com/secmc/plugin.git
cd plugin
go mod download
make proto   # generates Go + TS stubs via buf and the post-gen script
```

## Running the Host

```bash
make run
```

`make run` boots the host pointed at `cmd/plugins/plugins.yaml`. Each entry launches a plugin command, sets the handshake metadata, and defines how the plugin connects back (Unix socket or TCP loopback). Minimal entry:

```yaml
plugins:
  - id: example-typescript
    name: Example TypeScript Plugin
    command: "npm"
    args: ["run", "dev", "--prefix", "examples/plugins/typescript"]
    address: "unix:///tmp/dragonfly_plugin.sock"
```

Keep plugin IDs kebab-case and match directory names so the examples and config stay in sync.

## Building Plugins

1. Start from one of the language samples in `examples/plugins/`.
2. Generate protobuf types as needed (`make proto` already covers TS/Go; other languages can run `buf generate` with their own templates).
3. Point your plugin command + args at the new entry in `cmd/plugins/plugins.yaml`.
4. Implement the handshake: respond to the `Hello` message with plugin metadata, then send an `EventSubscribe` specifying which events you care about.
5. React to streamed events and reply with `ActionBatch` or `EventResult` messages. Because everything flows over gRPC, you can deploy the plugin on the same machine or talk over loopback TCP for remote control tools.

## Development Workflow

```bash
make proto        # regenerate protobuf artifacts + lint through buf
go test ./...     # run all host-side suites
make run          # launch Dragonfly with the sample plugin config
npm run dev --prefix examples/plugins/typescript   # live dev loop for TS sample
examples/plugins/php/bin/php7/bin/php examples/plugins/php/src/HelloPlugin.php   # PHP sample
```
