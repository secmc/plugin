## Rustic Economy – Rust example plugin

`rustic-economy` is a **Rust example plugin** for Dragonfly that demonstrates:

- A simple **SQLite-backed economy** using `sqlx`.
- The Rust SDK macros `#[derive(Plugin)]` and `#[derive(Command)]`.
- The generated **command system** (`Eco` enum + `EcoHandler` trait).
- Using `Ctx` to reply to the invoking player.

It is meant as a learning/reference plugin, not a production-ready economy.

### What this plugin does

- Stores each player’s balance in a local `economy.db` SQLite database.
- Exposes one command, `/eco` (with aliases `/economy` and `/rustic_eco`):
  - `/eco pay <amount>` (`/eco donate <amount>`): add money to your own balance.
  - `/eco bal` (aliases `/eco balance`, `/eco money`): show your current balance.

Balances are stored as `REAL`/`f64` for simplicity. For real money, you should use
an integer representation (e.g. cents as `i64`) to avoid floating‑point issues.

### Files and structure

- `Cargo.toml`: Rust crate metadata for the example plugin.
- `src/main.rs`: The entire plugin implementation:
  - `RusticEconomy` struct holding a `SqlitePool`.
  - `impl RusticEconomy { new, get_balance, add_money }` – DB helpers.
  - `Eco` command enum + `EcoHandler` impl with `pay` and `bal` handlers.
  - `main` function that initialises the DB and runs `PluginRunner`.

### Requirements

- Rust (stable) and `cargo`.
- A Dragonfly host that has the Rust SDK wired in (this repo’s Go host).
- SQLite available on the host machine (the plugin writes `economy.db`
  next to where it is run).

### Building the plugin

From the repo root:

```bash
cd examples/plugins/rust
cargo build --release
```

The compiled binary will be in `target/release/rustic-economy` (or `.exe` on Windows).
Point your Dragonfly `plugins.yaml` at that binary.

### Example `plugins.yaml` entry

```yaml
plugins:
  - id: rustic-economy
    name: Rustic Economy
    command: "./examples/plugins/rust/target/release/rustic-economy"
    address: "tcp://127.0.0.1:50050"
```

Ensure the `id` matches the `#[plugin(id = "rustic-economy", ...)]` attribute in
`src/main.rs`.

### Running and testing

1. Start Dragonfly with the plugin enabled via `plugins.yaml`.
2. Join the server as a player.
3. Run economy commands in chat:
   - `/eco pay 10` – adds 10 to your balance and shows the new total.
   - `/eco bal` – prints your current balance.
4. Check that `economy.db` is created and populated in the working directory.

If any DB or send‑chat errors occur, the plugin logs them to stderr and replies
with a generic error message so players aren’t exposed to internals.

### How it uses the Rust SDK

- `#[derive(Plugin)]` + `#[plugin(...)]` describe plugin metadata and register
  the `Eco` command with the host.
- `#[derive(Command)]` generates a `EcoHandler` trait and argument parsing from
  `types::CommandEvent` into the `Eco` enum.
- `Ctx<'_>` is used to send replies: `ctx.reply("...".to_string()).await`.
- `PluginRunner::run(plugin, "tcp://127.0.0.1:50050")` connects the plugin
  process to the Dragonfly host and runs the event loop.

Use this example as a starting point when building stateful Rust plugins that
compose the SDK’s command and event systems with your own storage layer.


