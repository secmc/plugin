# Dragonfly External Plugin Architecture (Event/Action Stream Only)

This document describes the external plugin architecture implemented in Dragonfly. Plugins run as separate
processes (PHP, Node, Go, Rust, …) and interact with the server over a **single bidirectional gRPC stream** that
carries events (host → plugin) and actions (plugin → host). The transport is negotiated by the Dragonfly process,
which connects to each plugin over HTTP/2 and exchanges length-prefixed protobuf messages.

## 1. Goals

* Support plugins written in any language with gRPC/protobuf bindings.
* Avoid in-process `plugin.Open` style ABIs — every plugin is an external process.
* Keep server hot paths (movement, physics, ticking) inside Dragonfly; plugins layer higher-level game logic.
* Rely on **one streaming RPC** per plugin for everything: registration, logging, events, actions, and future
  extensions.

## 2. Protobuf Schema

The schema lives in [`proto/df/plugin.proto`](../plugin/proto/messages.go) and is mirrored manually in Go. The
core service:

```proto
service Plugin {
  rpc EventStream(stream HostToPlugin) returns (stream PluginToHost);
}
```

### Host → Plugin (`HostToPlugin`)

* `HostHello` — announces API version.
* `HostShutdown` — tells a plugin to terminate gracefully.
* `EventEnvelope` — carries runtime events (player join, quit, chat, command, block break, world shutdown).

### Plugin → Host (`PluginToHost`)

* `PluginHello` — identifies the plugin, version, supported API version, and command registrations.
* `EventSubscribe` — declares the event types the plugin wants to receive.
* `ActionBatch` — one or more actions for the server to execute (send chat, teleport, kick).
* `LogMessage` — plugin side logging surfaced in the server logs.

Events and actions are wrapped in envelopes so that the protocol can evolve without breaking compatibility.
Unknown fields are ignored.

## 3. Host Architecture

The host side implementation resides in the [`plugin`](../plugin) package and revolves around the `Manager` type.

### Manager responsibilities

* Load plugin definitions from `plugins/plugins.yaml`.
* Launch plugin processes (optional) and set standard environment variables:
  * `DF_PLUGIN_ID`
  * `DF_PLUGIN_GRPC_ADDRESS`
* Dial each plugin’s gRPC endpoint (using HTTP/2 prior knowledge) and open the bidirectional `EventStream`.
* Perform the initial handshake:
  1. Send `HostHello`.
  2. Wait for `PluginHello` and register declared commands.
  3. Wait for `EventSubscribe` to activate event routing.
* Bridge Dragonfly events to plugins through `PluginPlayerHandler` / `PluginWorldHandler` wrappers.
* Consume `PluginToHost` messages, applying actions and logging output.
* Gracefully close plugins on shutdown by sending `HostShutdown` and cancelling the stream context.

The `Manager` is constructed in `main.go` immediately after the server is created and attaches world and player
handlers. Player handlers surface join/quit/chat/command/block-break events. World handlers currently surface
`WORLD_CLOSE` notifications.

### Configuration (`plugins/plugins.yaml`)

Example configuration:

```yaml
plugins:
  - id: example-node
    name: Example Node Plugin
    command: "node"
    args: ["examples/plugins/node/hello.js"]
    work_dir: ""
    address: "127.0.0.1:50051"
    env:
      NODE_ENV: development
```

* `id`: Unique identifier; defaults to a generated slug if omitted.
* `name`: Friendly display name (logged only).
* `command`: Optional executable to launch. If omitted, Dragonfly assumes the plugin is already running and just
  attempts to connect to `address`.
* `args`: Arguments passed to `command`.
* `work_dir`: Optional working directory.
* `env`: Extra environment variables.
* `address`: gRPC endpoint that the plugin listens on. Using `:0` instructs Dragonfly to pick a free port and
  communicate it via `DF_PLUGIN_GRPC_ADDRESS`.

## 4. Event Routing

The manager sends events to plugins based on their subscriptions. Current events include:

* `PLAYER_JOIN` / `PLAYER_QUIT`
* `CHAT`
* `COMMAND`
* `BLOCK_BREAK`
* `WORLD_CLOSE`

Events carry minimal data required for action correlation (player UUID, name, coordinates). Plugins can correlate
responses with `EventEnvelope.event_id` if desired.

## 5. Actions

Plugins can request server side changes by sending an `ActionBatch`:

* `SendChatAction` — message to a specific player or broadcast if `target_uuid` is empty.
* `TeleportAction` — move a player to coordinates and adjust rotation.
* `KickAction` — disconnect a player with a reason.

Actions are executed on the proper game goroutines through entity handles (`world.EntityHandle.ExecWorld`) to
respect Dragonfly’s threading model.

## 6. Logging

`PluginToHost.LogMessage` entries are forwarded to the server log with `info`, `warn`, or `error` severity based on
`log.level`.

## 7. Handshake Flow (Plugin Side)

1. Dragonfly connects and sends `HostHello(api_version="v1")`.
2. Plugin responds with `PluginHello` containing:
   * `name`, `version`
   * `api_version`
   * Optional command registrations (shown in `/help`).
3. Plugin sends `EventSubscribe` listing uppercase event names (`["PLAYER_JOIN", "COMMAND"]`).
4. Stream enters steady state: host pushes events; plugin sends actions/logs as needed.

## 8. Backpressure & Fault Handling

Each plugin process owns a buffered send queue. If the queue fills (plugin not reading), events for that plugin are
dropped and a warning is logged. Connection failures trigger retries until the manager’s context is cancelled.

## 9. Examples

Reference implementations are provided under `examples/plugins`:

* [`examples/plugins/node/hello.js`](../examples/plugins/node/hello.js) — Node.js plugin using `@grpc/grpc-js` and
  `protobufjs`.
* [`examples/plugins/php/HelloPlugin.php`](../examples/plugins/php/HelloPlugin.php) — PHP plugin built with the
  official gRPC extension.

Each example reads `DF_PLUGIN_GRPC_ADDRESS` and `DF_PLUGIN_ID` from the environment, starts a gRPC server, registers
commands, subscribes to events, and echoes actions back to Dragonfly.

## 10. Versioning

The handshake contains `api_version` on both sides. Backwards-incompatible changes should increment this string and
only activate new behaviour when both sides agree. Unknown events/actions are safely ignored thanks to protobuf’s
forward-compatibility.
