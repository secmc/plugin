use std::convert::TryFrom;

use dragonfly_plugin::{command::CommandParseError, types, Command};

fn make_event(player_uuid: &str, command: &str, args: &[&str]) -> types::CommandEvent {
    types::CommandEvent {
        player_uuid: player_uuid.to_string(),
        name: format!("/{command} {}", args.join(" ")),
        raw: format!("/{command} {}", args.join(" ")),
        command: command.to_string(),
        args: args.iter().map(|s| s.to_string()).collect(),
    }
}

#[derive(Debug, Command)]
#[command(name = "ping", description = "Ping command", aliases("p"))]
struct Ping {
    times: i32,
}

#[derive(Debug, Command)]
#[command(name = "eco", description = "Economy command")]
enum Eco {
    #[subcommand(aliases("donate"))]
    Pay {
        amount: f64,
    },
    Bal,
}

#[derive(Debug, Command)]
#[command(name = "flaggy", description = "Flag-style command")]
enum Flaggy {
    Enable {
        #[allow(dead_code)]
        #[command(name = "value")]
        value: bool,
    },
}

#[test]
fn struct_command_parses_ok() {
    let event = make_event("player-uuid", "ping", &["3"]);
    let cmd = Ping::try_from(&event).expect("expected ping command to parse");
    assert_eq!(cmd.times, 3);
}

#[test]
fn struct_command_respects_aliases() {
    let event = make_event("player-uuid", "p", &["5"]);
    let cmd = Ping::try_from(&event).expect("expected alias to parse");
    assert_eq!(cmd.times, 5);
}

#[test]
fn struct_command_errors_when_name_does_not_match() {
    let event = make_event("player-uuid", "other", &["3"]);
    let err = Ping::try_from(&event).unwrap_err();
    assert!(matches!(err, CommandParseError::NoMatch));
}

#[test]
fn struct_command_reports_missing_and_invalid_args() {
    // Missing required arg.
    let event = make_event("player-uuid", "ping", &[]);
    let err = Ping::try_from(&event).unwrap_err();
    assert!(matches!(err, CommandParseError::Missing("times")));

    // Invalid arg type.
    let event = make_event("player-uuid", "ping", &["not-a-number"]);
    let err = Ping::try_from(&event).unwrap_err();
    assert!(matches!(err, CommandParseError::Invalid("times")));
}

#[test]
fn enum_command_parses_subcommands_and_args() {
    // canonical subcommand name
    let event = make_event("player-uuid", "eco", &["pay", "10.5"]);
    let cmd = Eco::try_from(&event).expect("expected eco pay to parse");
    match cmd {
        Eco::Pay { amount } => assert!((amount - 10.5).abs() < f64::EPSILON),
        other => panic!("expected Pay variant, got {other:?}"),
    }

    // alias subcommand
    let event = make_event("player-uuid", "eco", &["donate", "2"]);
    let cmd = Eco::try_from(&event).expect("expected eco donate to parse");
    match cmd {
        Eco::Pay { amount } => assert!((amount - 2.0).abs() < f64::EPSILON),
        other => panic!("expected Pay variant, got {other:?}"),
    }

    // unit-like subcommand
    let event = make_event("player-uuid", "eco", &["bal"]);
    let cmd = Eco::try_from(&event).expect("expected eco bal to parse");
    matches!(cmd, Eco::Bal);
}

#[test]
fn enum_command_reports_missing_or_unknown_subcommand() {
    // No args -> missing subcommand.
    let event = make_event("player-uuid", "eco", &[]);
    let err = Eco::try_from(&event).unwrap_err();
    assert!(matches!(err, CommandParseError::Missing("subcommand")));

    // Unrecognised subcommand string.
    let event = make_event("player-uuid", "eco", &["nope"]);
    let err = Eco::try_from(&event).unwrap_err();
    assert!(matches!(err, CommandParseError::UnknownSubcommand));
}

#[test]
fn bool_flags_parse_as_expected() {
    // NOTE: rust FromStr of bools is case sensitive to literally the word.
    // maybe later we add blanket impls for our own trait to parse from commands.
    // TODO: this enables a good amount of flexibility so 0.3.1
    // we could add that as its not a BC.

    // true-like values
    let event = make_event("player-uuid", "flaggy", &["enable", "true"]);
    let cmd = Flaggy::try_from(&event).expect("expected flaggy enable to parse");
    match cmd {
        Flaggy::Enable { value } => assert!(value, "expected true to parse as true"),
    }

    // false-like values
    let event = make_event("player-uuid", "flaggy", &["enable", "false"]);
    let cmd = Flaggy::try_from(&event).expect("expected flaggy enable to parse");
    match cmd {
        Flaggy::Enable { value } => assert!(!value, "expected false to parse as false"),
    }

    // invalid values should surface a parse error
    let event = make_event("player-uuid", "flaggy", &["enable", "not-a-bool"]);
    let err = Flaggy::try_from(&event).unwrap_err();
    assert!(matches!(err, CommandParseError::Invalid("value")));
}
