use crate::{server::Server, types};

pub use dragonfly_plugin_macro::{command_handlers, Command};
use tokio::sync::mpsc;

/// Per-command execution context.
pub struct Ctx<'a> {
    pub server: &'a Server,
    pub sender: String,
}

impl<'a> Ctx<'a> {
    pub fn new(server: &'a Server, player_uuid: String) -> Self {
        Self {
            server,
            sender: player_uuid,
        }
    }

    pub async fn reply(
        &self,
        msg: impl Into<String>,
    ) -> Result<(), mpsc::error::SendError<types::PluginToHost>> {
        self.server.send_chat(self.sender.clone(), msg.into()).await
    }
}

/// Trait plugins use to expose commands to the host.
pub trait CommandRegistry {
    fn get_commands(&self) -> Vec<types::CommandSpec> {
        Vec::new()
    }

    /// Dispatch to registered commands. Returns true if a command was handled.
    #[allow(async_fn_in_trait)]
    async fn dispatch_commands(
        &self,
        _server: &crate::Server,
        _event: &mut crate::event::EventContext<'_, types::CommandEvent>,
    ) -> bool {
        false
    }
}

#[derive(Debug)]
pub enum CommandParseError {
    NoMatch,
    Missing(&'static str),
    Invalid(&'static str),
    UnknownSubcommand,
}

/// Parse a required argument at the given index.
pub fn parse_required_arg<T>(
    args: &[String],
    index: usize,
    name: &'static str,
) -> Result<T, CommandParseError>
where
    T: std::str::FromStr,
{
    let s = args.get(index).ok_or(CommandParseError::Missing(name))?;
    s.parse().map_err(|_| CommandParseError::Invalid(name))
}

/// Parse an optional argument at the given index.
/// Returns Ok(None) if the argument is missing.
/// Returns Ok(Some(value)) if present and parseable.
/// Returns Err if present but invalid.
pub fn parse_optional_arg<T>(
    args: &[String],
    index: usize,
    name: &'static str,
) -> Result<Option<T>, CommandParseError>
where
    T: std::str::FromStr,
{
    match args.get(index) {
        None => Ok(None),
        Some(s) if s.is_empty() => Ok(None),
        Some(s) => s
            .parse()
            .map(Some)
            .map_err(|_| CommandParseError::Invalid(name)),
    }
}
