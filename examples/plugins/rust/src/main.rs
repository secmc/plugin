/// Rustic Economy: a small example plugin backed by SQLite.
///
/// This example demonstrates how to:
/// - Use `#[derive(Plugin)]` to declare plugin metadata and register commands.
/// - Use `#[derive(Command)]` to define a typed command enum.
/// - Hold state (a `SqlitePool`) inside your plugin struct.
/// - Use `Ctx` to reply to the invoking player.
/// - Use `#[event_handler]` for event subscriptions (even when you don't
///   implement any event methods yet).
use dragonfly_plugin::{
    Command, Plugin, PluginRunner, command::Ctx, event::EventHandler, event_handler, types,
};
use sqlx::{SqlitePool, sqlite::SqlitePoolOptions};

#[derive(Plugin)]
#[plugin(
    id = "rustic-economy",
    name = "Rustic Economy",
    version = "0.3.0",
    api = "1.0.0",
    commands(Eco)
)]
struct RusticEconomy {
    db: SqlitePool,
}

/// Database helpers for the Rustic Economy example.
impl RusticEconomy {
    async fn new() -> Result<Self, Box<dyn std::error::Error>> {
        // Create database connection
        let db = SqlitePoolOptions::new()
            .max_connections(5)
            .connect("sqlite:economy.db")
            .await?;

        // Create table if it doesn't exist.
        //
        // NOTE: This example stores balances as REAL/f64 for simplicity.
        // For real-world money you should use an integer representation
        // (e.g. cents as i64) to avoid floating point rounding issues.
        sqlx::query(
            "CREATE TABLE IF NOT EXISTS users (
                uuid TEXT PRIMARY KEY,
                balance REAL NOT NULL DEFAULT 0.0
            )",
        )
        .execute(&db)
        .await?;

        Ok(Self { db })
    }

    async fn get_balance(&self, uuid: &str) -> Result<f64, sqlx::Error> {
        let result: Option<(f64,)> = sqlx::query_as("SELECT balance FROM users WHERE uuid = ?")
            .bind(uuid)
            .fetch_optional(&self.db)
            .await?;

        Ok(result.map(|(bal,)| bal).unwrap_or(0.0))
    }

    async fn add_money(&self, uuid: &str, amount: f64) -> Result<f64, sqlx::Error> {
        // Insert or update user balance
        sqlx::query(
            "INSERT INTO users (uuid, balance) VALUES (?, ?)
             ON CONFLICT(uuid) DO UPDATE SET balance = balance + ?",
        )
        .bind(uuid)
        .bind(amount)
        .bind(amount)
        .execute(&self.db)
        .await?;

        self.get_balance(uuid).await
    }
}

#[derive(Command)]
#[command(
    name = "eco",
    description = "Rustic Economy commands.",
    aliases("economy", "rustic_eco")
)]
pub enum Eco {
    #[subcommand(aliases("donate"))]
    Pay { amount: f64 },
    #[subcommand(aliases("balance", "money"))]
    Bal,
}

impl EcoHandler for RusticEconomy {
    async fn pay(&self, ctx: Ctx<'_>, amount: f64) {
        match self.add_money(&ctx.sender, amount).await {
            Ok(new_balance) => {
                if let Err(e) = ctx
                    .reply(format!(
                        "Added ${:.2}! New balance: ${:.2}",
                        amount, new_balance
                    ))
                    .await
                {
                    eprintln!("Failed to send payment reply: {}", e);
                }
            }
            Err(e) => {
                eprintln!("Database error: {}", e);
                if let Err(send_err) = ctx
                    .reply("Error processing payment!".to_string())
                    .await
                {
                    eprintln!("Failed to send error reply: {}", send_err);
                }
            }
        }
    }

    async fn bal(&self, ctx: Ctx<'_>) {
        match self.get_balance(&ctx.sender).await {
            Ok(balance) => {
                if let Err(e) = ctx
                    .reply(format!("Your balance: ${:.2}", balance))
                    .await
                {
                    eprintln!("Failed to send balance reply: {}", e);
                }
            }
            Err(e) => {
                eprintln!("Database error: {}", e);
                if let Err(send_err) = ctx
                    .reply("Error checking balance!".to_string())
                    .await
                {
                    eprintln!("Failed to send error reply: {}", send_err);
                }
            }
        }
    }
}

#[event_handler]
impl EventHandler for RusticEconomy {}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    println!("Starting the plugin...");
    println!("Initializing database...");

    let plugin = RusticEconomy::new().await?;

    PluginRunner::run(plugin, "tcp://127.0.0.1:50050").await
}
