## Dragonfly Rust Plugin SDK

The `dragonfly-plugin` crate is the **Rust SDK for Dragonfly gRPC plugins**. It gives you:

- **Derive macros** to describe your plugin (`#[derive(Plugin)]`) and commands (`#[derive(Command)]`).
- A simple **event system** based on an `EventHandler` trait and an `#[event_handler]` macro.
- A `Server` handle with high‑level helpers (like `send_chat`, `teleport`, `world_set_block`, …).
- A `PluginRunner` that connects your process to the Dragonfly host and runs the event loop.

### Crate and directory layout

The Rust SDK lives under `packages/rust` as a small workspace:

- **`dragonfly-plugin` (this crate)**: Public SDK surface used by plugin authors.
  - `src/lib.rs`: Re-exports core modules and pulls in this README as crate-level docs.
  - `src/command.rs`: Command context (`Ctx`), parsing helpers, and `CommandRegistry` trait.
  - `src/event/`: Event system (`EventContext`, `EventHandler`, mutation helpers).
  - `src/server/`: `Server` handle and generated helpers for sending actions to the host.
  - `src/generated/df.plugin.rs`: Prost/tonic types generated from `proto/types/*.proto` (do not edit).
- **`macro/` (`dragonfly-plugin-macro`)**: Procedural macros for `#[derive(Plugin)]`, `#[derive(Command)]`,
  and `#[event_handler]`. This crate is re-exported by `dragonfly-plugin` and is not used directly by plugins.
- **`xtask/`**: Internal code generation tooling that reads `df.plugin.rs` and regenerates
  `event/handler.rs`, `event/mutations.rs`, and `server/helpers.rs`. It is not published.
- **`example/`**: A minimal example plugin crate showing recommended usage patterns for the SDK.
- **`tests/`**: Integration tests covering command derivation, event dispatch, server helpers,
  and the interaction between the runtime and macros.

All APIs in this README reflect the **0.3.x line**. Within 0.3.x we intend to keep:

- The `Plugin`, `EventHandler`, `EventSubscriptions`, and `CommandRegistry` trait shapes.
- The `event_handler`, `Plugin`, and `Command` macros and their attribute syntax.
- The `Server` helpers and `event::EventContext` semantics (including `cancel` and mutation helpers).

Breaking changes may still happen in a future 0.4.0, but not within 0.3.x.

---

## Quick start

### 1. Create a new plugin crate

```sh
cargo new my_plugin --bin
```

### 2. Add dependencies

```toml
[package]
name = "my_plugin"
version = "0.1.0"
edition = "2021"

[dependencies]
dragonfly-plugin = "0.3"
tokio = { version = "1", features = ["rt-multi-thread", "macros"] }
sqlx = { version = "0.8", features = ["runtime-tokio", "sqlite"] } # optional, for DB-backed examples
```

Only `dragonfly-plugin` and `tokio` are required; other crates (like `sqlx`) are up to your plugin.

### 3. Define your plugin

```rust,no_run
use dragonfly_plugin::{
    event::{EventContext, EventHandler},
    event_handler,
    types,
    Plugin,
    PluginRunner,
    Server,
};

#[derive(Plugin, Default)]
#[plugin(
    id = "example-rust",            // must match plugins.yaml
    name = "Example Rust Plugin",
    version = "0.3.0",
    api = "1.0.0"
)]
struct MyPlugin;

#[event_handler]
impl EventHandler for MyPlugin {
    async fn on_player_join(
        &self,
        server: &Server,
        event: &mut EventContext<'_, types::PlayerJoinEvent>,
    ) {
        let player_name = &event.data.name;
        println!("Player '{}' has joined.", player_name);

        let welcome = format!(
            "Welcome, {}! This server is running a Rust plugin.",
            player_name
        );

        // Ignore send errors; they usually mean the host shut down.
        let _ = server
            .send_chat(event.data.player_uuid.clone(), welcome)
            .await;
    }

    async fn on_chat(
        &self,
        _server: &Server,
        event: &mut EventContext<'_, types::ChatEvent>,
    ) {
        let new_message = format!("[Plugin] {}", event.data.message);
        event.set_message(new_message);
    }
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    println!("Starting example-rust plugin...");
    PluginRunner::run(MyPlugin, "tcp://127.0.0.1:50050").await
}
```

The `#[event_handler]` macro:

- Detects which `on_*` methods you implement.
- Generates an `EventSubscriptions` impl that subscribes to the corresponding `types::EventType` variants.
- Wires those events into `event::dispatch_event`.

---

## Commands

The 0.3.x series introduces a **first‑class command system**.

### Declaring a command

```rust,no_run
use dragonfly_plugin::{command::Ctx, Command};

#[derive(Command)]
#[command(
    name = "eco",
    description = "Economy commands.",
    aliases("economy", "rustic_eco")
)]
pub enum Eco {
    #[subcommand(aliases("donate"))]
    Pay { amount: f64 },

    #[subcommand(aliases("balance", "money"))]
    Bal,
}
```

This generates:

- A static `Eco::spec() -> types::CommandSpec`.
- A `TryFrom<&types::CommandEvent>` impl that parses args into `Eco`.
- An `EcoHandler` trait with async methods (`pay`, `bal`) and an `__execute` helper.

### Handling commands in your plugin

Add the command type to your plugin’s `#[plugin]` attribute, and implement the generated handler trait for your plugin type:

```rust,ignore
use dragonfly_plugin::{command::Ctx, Command, Plugin};

#[derive(Plugin)]
#[plugin(
    id = "rustic-economy",
    name = "Rustic Economy",
    version = "0.1.0",
    api = "1.0.0",
    commands(Eco)
)]
struct RusticEconomy {
    // your state here, e.g. DB pools
}

impl EcoHandler for RusticEconomy {
    async fn pay(&self, ctx: Ctx<'_>, amount: f64) {
        // ...
        let _ = ctx
            .reply(format!("You paid yourself ${:.2}.", amount))
            .await;
    }

    async fn bal(&self, ctx: Ctx<'_>) {
        // ...
        let _ = ctx.reply("Your balance is $0.00".to_string()).await;
    }
}
```

The `#[derive(Plugin)]` macro then:

- Reports the command specs in the initial hello handshake.
- Generates a `CommandRegistry` impl that:
  - Parses `CommandEvent`s into your command types.
  - Cancels the event if a command matches.
  - Dispatches into your `EcoHandler` implementation.

Within 0.3.x the **shape of the command API** (`Ctx`, `CommandRegistry`, `CommandParseError`, and the `Command` derive attributes) is considered stable.

---

## Events, context, and mutations

- `event::EventContext<'_, T>` wraps each incoming event:
  - `data: &T` gives read‑only access.
  - `cancel().await` marks the event as cancelled and immediately sends a response.
  - Event‑specific methods (like `set_message` for `ChatEvent`) live in generated extensions.
- `event::EventHandler` is a trait with an async method per event type; you usually never write `impl EventHandler` by hand except inside an `#[event_handler]` block.

You generally do not construct `EventContext` yourself; the runtime does it for you.

---

## Connection and runtime

Use `PluginRunner::run(plugin, addr)` from your `main` function:

- For TCP, pass e.g. `"tcp://127.0.0.1:50050"` or `"127.0.0.1:50050"`.
- On Unix hosts you may also pass:
  - `"unix:///tmp/dragonfly_plugin.sock"` or
  - an absolute path (`"/tmp/dragonfly_plugin.sock"`).

On non‑Unix platforms, Unix socket addresses will return an error.

`PluginRunner`:

- Connects to the host.
- Sends an initial hello (`PluginHello`) with your plugin ID, name, version, API version and commands.
- Subscribes to your `EventSubscriptions`.
- Drives the main event loop until the host sends a shutdown message or closes the stream.

---

## Stability policy for 0.3.x

Within the 0.3.x series we aim to keep:

- Trait surfaces for `Plugin`, `EventHandler`, `EventSubscriptions`, `CommandRegistry`.
- Macro names and high‑level attribute syntax (`#[plugin(...)]`, `#[event_handler]`, `#[derive(Command)]`, `#[subcommand(...)]`).
- `Server` helper method names and argument shapes.
- `EventContext` behavior for `cancel`, mutation helpers, and double‑send (panic in debug, log in release).

We may still:

- Add new events and actions.
- Add new helpers or mutation methods.
- Improve error messages and diagnostics.

For details on how the code is generated and how to maintain it, see `MAINTAINING.md`.
