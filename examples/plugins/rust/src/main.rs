/// This is a semi advanced example of a simple economy plugin.
/// we are gonna use sqlite, to store user money.
/// two commands:
/// pay: pay yourself money
/// bal: view your balance / money
use dragonfly_plugin::{
    Plugin, PluginRunner,
    command::{Command, Ctx, command_handlers},
    event::EventHandler,
    event_handler, types,
};
use sqlx::{SqlitePool, sqlite::SqlitePoolOptions};

#[derive(Plugin)]
#[plugin(
    id = "rustic-economy",
    name = "Rustic Economy",
    version = "0.1.0",
    api = "1.0.0",
    commands(Eco)
)]
struct RusticEconomy {
    db: SqlitePool,
}

/// This impl is just a helper for dealing with our SQL stuff.
impl RusticEconomy {
    async fn new() -> Result<Self, Box<dyn std::error::Error>> {
        // Create database connection
        let db = SqlitePoolOptions::new()
            .max_connections(5)
            .connect("sqlite:economy.db")
            .await?;

        // Create table if it doesn't exist
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
#[command(name = "eco", description = "Rustic Economy commands.")]
pub enum Eco {
    Pay { amount: f64 },
    Bal,
}

#[command_handlers]
impl Eco {
    async fn pay(state: &RusticEconomy, ctx: Ctx<'_>, amount: f64) {
        match state.add_money(&ctx.sender, amount).await {
            Ok(new_balance) => ctx
                .reply(format!(
                    "Added ${:.2}! New balance: ${:.2}",
                    amount, new_balance
                ))
                .await
                .unwrap(),
            Err(e) => {
                eprintln!("Database error: {}", e);
                ctx.reply("Error processing payment!".to_string())
                    .await
                    .unwrap()
            }
        }
    }

    async fn bal(state: &RusticEconomy, ctx: Ctx<'_>) {
        match state.get_balance(&ctx.sender).await {
            Ok(balance) => ctx
                .reply(format!("Your balance: ${:.2}", balance))
                .await
                .unwrap(),
            Err(e) => {
                eprintln!("Database error: {}", e);
                ctx.reply("Error checking balance!".to_string())
                    .await
                    .unwrap()
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
