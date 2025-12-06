# Dragonfly Plugin System

Write plugins for your Minecraft Bedrock server in whatever language you love. This gRPC bridge sits on top of the [Dragonfly](https://github.com/df-mc/dragonfly) server and lets external processes stream protobuf events and actions without touching the core runtime.

## Why Dragonfly Plugins?

| Benefit | Description |
| --- | --- |
| 🌍 **Any Language** | JavaScript, TypeScript, PHP, Python, Rust, C++, Go—if it can speak gRPC, it can be a plugin. |
| 💰 **Sell Plugins** | Compile to a binary (Rust, Go, C++) and ship closed-source builds. |
| 🔥 **Hot Reload** | Edit JS/TS/PHP plugins while the server runs; changes apply immediately. |
| 📱 **Remote Control** | Plugins connect over gRPC, so you can run them on your phone, a web app, or a remote service. |
| 📦 **Use Any Library** | Mix npm packages, Python ML libs, or anything else your runtime supports. |
| ⚡ **Zero Performance Impact** | Plugins live in separate processes, so heavy work never blocks Dragonfly’s TPS. |
| 🚀 **High Performance (SOON)** | The protocol is optimized protobuf with room for batching. |
| 🔒 **Sandboxing** | Grant only the permissions each plugin needs over the gRPC interface. |

### Real-World Examples

```bash
# Hot reload: Edit plugin code while server is running
vim plugins/my-plugin.js   # Make changes
# Changes apply immediately - no restart!

# Remote plugin: Control server from your phone
# Plugin runs on your phone, connects to server over internet
phone-app → [gRPC] → Dragonfly Server

# Binary plugin: Sell without source code
rustc plugin.rs --release   # Compile to binary
# Distribute the binary - customers can't see your code
```

## Key Features

- **Event-driven API**: Subscribe to joins, chat, commands, block events, and more.
- **Generated types**: Proto definitions live in `proto/types/` with generated Go + TypeScript stubs under `proto/generated/`.
- **Language samples**: TypeScript, Node, PHP, and more under `examples/plugins/` to kick-start new plugins.
- **Automation ready**: `make proto` (buf + scripts) and `make run` wire up the host for you.

## Quick Start

1. **Clone & bootstrap**
   ```bash
   git clone https://github.com/secmc/plugin.git
   cd plugin
   go mod download
   make proto
   ```
2. **Configure a plugin** in `cmd/plugins/plugins.yaml`:
   ```yaml
   plugins:
     - id: example-typescript
       name: Example TypeScript Plugin
       command: "npm"
       args: ["run", "dev", "--prefix", "examples/plugins/typescript"]
       address: "unix:///tmp/dragonfly_plugin.sock"
   ```
3. **Run the host**
   ```bash
   make run
   ```
4. **Iterate in your language** – edit the example plugin, or point the config at your own command/binary.

## How It Works

```
┌─────────────────┐         gRPC Stream          ┌──────────────────┐
│                 │ ←──────────────────────────→ │                  │
│  Dragonfly      │   Events: JOIN, CHAT, etc.   │  Your Plugin     │
│  Server (Go)    │   Actions: TELEPORT, etc.    │  (Any Language)  │
└─────────────────┘                              └──────────────────┘
```

1. **Server starts** and loads plugin configuration from `cmd/plugins/plugins.yaml`.
2. **Plugin process launches** via the configured command (for example `node plugin.js`).
3. **Handshake** occurs where the plugin registers its metadata and commands.
4. **Plugin subscribes** to the events it wants.
5. **Events flow** from Dragonfly to the plugin in real time.
6. **Plugin executes actions** by sending protobuf messages back to the host.

## Building Plugins

1. Copy an example from `examples/plugins/` or start fresh with `proto/types/plugin.proto`.
2. Run `make proto` (or `buf generate` with your template) to refresh client stubs.
3. Add your command + args + socket info to `cmd/plugins/plugins.yaml`.
4. Implement the handshake: reply to `PluginHello`, register commands, then send `EventSubscribe`.
5. Handle streamed events and reply with `ActionBatch` or `EventResult` messages. Because plugins speak gRPC, they can run locally, over loopback TCP, or on a remote machine.

## Development Workflow

```bash
make proto        # regenerate protobuf artifacts + post-gen scripts
go test ./...     # run all Go suites
make run          # launch Dragonfly host with sample config
npm run dev --prefix examples/plugins/typescript   # TypeScript live dev
examples/plugins/php/bin/php7/bin/php examples/plugins/php/src/HelloPlugin.php   # PHP sample
```

## Prerequisites

- Go 1.22+ with `GOBIN` on your `PATH`.
- [buf](https://buf.build/docs/cli/installation/) and `protoc-gen-go` (`go install google.golang.org/protobuf/cmd/protoc-gen-go@latest`).
