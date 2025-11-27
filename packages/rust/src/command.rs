//! Command helpers and traits used by the `#[derive(Command)]` macro.
//!
//! Plugin authors usually interact with:
//! - `Ctx`, the per-command execution context (for replying to the sender).
//! - `CommandRegistry`, which is implemented for you by `#[derive(Plugin)]`.
//! - `CommandParseError`, surfaced as friendly messages to players.

use crate::{server::Server, types};

use tokio::sync::mpsc;

/// Per-command execution context.
///
/// This context is constructed by the runtime when a command matches,
/// and exposes the `Server` handle plus the UUID of the player that
/// issued the command.
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

    /// Sends a chat message back to the command sender.
    ///
    /// This is a convenience wrapper around `Server::send_chat`.
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

impl std::fmt::Display for CommandParseError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            CommandParseError::NoMatch => {
                write!(f, "command did not match")
            }
            CommandParseError::Missing(name) => {
                write!(f, "missing required argument `{name}`")
            }
            CommandParseError::Invalid(name) => {
                write!(f, "invalid value for argument `{name}`")
            }
            CommandParseError::UnknownSubcommand => {
                write!(f, "unknown subcommand")
            }
        }
    }
}

impl std::error::Error for CommandParseError {}

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

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn parse_required_arg_ok() {
        let args = vec!["42".to_string()];
        let value: i32 = parse_required_arg(&args, 0, "amount").unwrap();
        assert_eq!(value, 42);
    }

    #[test]
    fn parse_required_arg_missing() {
        let args: Vec<String> = Vec::new();
        let err = parse_required_arg::<i32>(&args, 0, "amount").unwrap_err();
        match err {
            CommandParseError::Missing(name) => assert_eq!(name, "amount"),
            e => panic!("expected Missing, got {e:?}"),
        }
    }

    #[test]
    fn parse_required_arg_invalid() {
        let args = vec!["not-a-number".to_string()];
        let err = parse_required_arg::<i32>(&args, 0, "amount").unwrap_err();
        match err {
            CommandParseError::Invalid(name) => assert_eq!(name, "amount"),
            e => panic!("expected Invalid, got {e:?}"),
        }
    }

    #[test]
    fn parse_optional_arg_none_when_missing_or_empty() {
        // Missing index
        let args: Vec<String> = Vec::new();
        let value: Option<i32> = parse_optional_arg(&args, 0, "amount").unwrap();
        assert!(value.is_none());

        // Present but empty string
        let args = vec!["".to_string()];
        let value: Option<i32> = parse_optional_arg(&args, 0, "amount").unwrap();
        assert!(value.is_none());
    }

    #[test]
    fn parse_optional_arg_some_when_valid() {
        let args = vec!["7".to_string()];
        let value: Option<i32> = parse_optional_arg(&args, 0, "amount").unwrap();
        assert_eq!(value, Some(7));
    }

    #[test]
    fn parse_optional_arg_error_when_invalid() {
        let args = vec!["nope".to_string()];
        let err = parse_optional_arg::<i32>(&args, 0, "amount").unwrap_err();
        match err {
            CommandParseError::Invalid(name) => assert_eq!(name, "amount"),
            e => panic!("expected Invalid, got {e:?}"),
        }
    }

    #[test]
    fn display_messages_are_human_friendly() {
        let err = CommandParseError::Missing("amount");
        assert!(err.to_string().contains("missing required argument"));
        assert!(err.to_string().contains("amount"));

        let err = CommandParseError::Invalid("amount");
        assert!(err.to_string().contains("invalid value for argument"));

        let err = CommandParseError::UnknownSubcommand;
        assert!(err.to_string().contains("unknown subcommand"));
    }
}
