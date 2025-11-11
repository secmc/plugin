# Dragonfly External Plugin Architecture (Event/Action Stream Only)

This document describes the external plugin architecture implemented in Dragonfly. Plugins run as separate
processes (PHP, Node, Go, Rust, …) and interact with the server over a **single bidirectional gRPC stream** that
carries events (host → plugin) and actions (plugin → host). The Dragonfly server runs a gRPC server that plugins
connect to as clients, exchanging length-prefixed protobuf messages over HTTP/2.

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
* `EventResult` — optional response to an event that can cancel execution or mutate pointer-backed values such as
  chat messages, block break drops, or experience rewards.

Events and actions are wrapped in envelopes so that the protocol can evolve without breaking compatibility.
Unknown fields are ignored.

## 3. Host Architecture

The host side implementation resides in the [`plugin`](../plugin) package and revolves around the `Manager` type.

### Manager responsibilities

* Load plugin definitions from `plugins/plugins.yaml`.
* Start a gRPC server on a configurable port (default: 50050).
* Launch plugin processes (optional) and set standard environment variables:
  * `DF_PLUGIN_ID`
  * `DF_PLUGIN_SERVER_ADDRESS`
* Accept incoming connections from plugins and match them to configurations by plugin ID.
* Perform the initial handshake:
  1. Send `HostHello` after plugin connects.
  2. Wait for `PluginHello` (sent as first message by plugin) and register declared commands.
  3. Wait for `EventSubscribe` to activate event routing.
* Bridge Dragonfly events to plugins through `PluginPlayerHandler` / `PluginWorldHandler` wrappers.
* Consume `PluginToHost` messages, applying actions and logging output.
* Gracefully close plugins on shutdown by sending `HostShutdown` and stopping the gRPC server.

The `Manager` is constructed in `main.go` immediately after the server is created and attaches world and player
handlers. Player handlers surface join/quit/chat/command/block-break events. World handlers currently surface
`WORLD_CLOSE` notifications.

### Configuration (`plugins/plugins.yaml`)

Example configuration:

```yaml
server_port: 50050  # Port for Dragonfly's plugin gRPC server

plugins:
  - id: example-node
    name: Example Node Plugin
    command: "node"
    args: ["examples/plugins/node/hello.js"]
    work_dir: ""
    env:
      NODE_ENV: development
  - id: example-php
    name: Example PHP Plugin
    command: "php"
    args: ["examples/plugins/php/src/HelloPlugin.php"]
```

* `server_port`: Port where Dragonfly's gRPC server listens for plugin connections (default: 50050).
* `id`: Unique identifier; defaults to a generated slug if omitted.
* `name`: Friendly display name (logged only).
* `command`: Optional executable to launch. If omitted, Dragonfly assumes the plugin is already running.
* `args`: Arguments passed to `command`.
* `work_dir`: Optional working directory.
* `env`: Extra environment variables.

## 4. Event Routing

The manager sends events to plugins based on their subscriptions. Current events include values from the
`EventType` enum, such as:

* `PLAYER_JOIN` / `PLAYER_QUIT`
* `CHAT`
* `COMMAND`
* `PLAYER_BLOCK_BREAK`
* `WORLD_CLOSE`

Events carry minimal data required for action correlation (player UUID, name, coordinates). Plugins can correlate
responses with `EventEnvelope.event_id` if desired.

### Event cancellation and mutation

Some Dragonfly callbacks expose a `Context` (for cancellation) and pointer arguments (for mutation) — for example,
chat messages (`*string`), block break drops (`*[]item.Stack`), and XP values (`*int`). When such an event is routed
through the manager, the server waits for subscribed plugins to reply with `EventResult` messages. The host applies
each result in order:

* `cancel = true` triggers `ctx.Cancel()` on the Dragonfly handler, preventing the default behaviour.
* `chat` mutations replace the in-flight chat message so that later plugins and the base server see the updated value.
* `block_break` mutations may override the drop list (item name/meta/count pairs) and/or the XP reward.

Results are optional; plugins that do not need to influence the outcome can simply skip sending an `EventResult` for
that event.

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

1. Plugin connects to Dragonfly's gRPC server (`DF_PLUGIN_SERVER_ADDRESS`).
2. Plugin sends `PluginHello` as the first message containing:
   * `plugin_id` (from `DF_PLUGIN_ID` environment variable)
   * `name`, `version`
   * `api_version`
   * Optional command registrations (shown in `/help`).
3. Dragonfly identifies the plugin by `plugin_id` and sends `HostHello(api_version="v1")`.
4. Plugin sends `EventSubscribe` listing `EventType` values (for example, `[EventType.PLAYER_JOIN, EventType.COMMAND]`).
5. Stream enters steady state: host pushes events; plugin sends actions/logs as needed.

## 8. Backpressure & Fault Handling

Each plugin process owns a buffered send queue. If the queue fills (plugin not reading), events for that plugin are
dropped and a warning is logged. Connection failures trigger retries until the manager’s context is cancelled.

## 9. Examples

Reference implementations are provided under `examples/plugins`:

* [`examples/plugins/node/hello.js`](../examples/plugins/node/hello.js) — Node.js plugin using `@grpc/grpc-js` and
  `proto-loader`, demonstrating chat cancellation and mutation.
* [`examples/plugins/typescript/src/index.ts`](../examples/plugins/typescript/src/index.ts) — TypeScript plugin with
  full type safety using generated protobuf types.
* [`examples/plugins/php/src/HelloPlugin.php`](../examples/plugins/php/src/HelloPlugin.php) — PHP plugin using the
  standard PECL gRPC extension (client mode), including chat moderation and message rewriting via `EventResult`.

Each example reads `DF_PLUGIN_SERVER_ADDRESS` and `DF_PLUGIN_ID` from the environment, connects to the Dragonfly server
as a gRPC client, sends plugin hello, subscribes to events, and sends actions back to Dragonfly.

## 10. Versioning

The handshake contains `api_version` on both sides. Backwards-incompatible changes should increment this string and
only activate new behaviour when both sides agree. Unknown events/actions are safely ignored thanks to protobuf’s
forward-compatibility.
