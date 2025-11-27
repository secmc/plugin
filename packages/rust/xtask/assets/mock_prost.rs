pub struct Vec3 {
    x: f64,
    y: f64,
    z: f64,
}

// We need `Clone` for the test helpers
#[derive(Clone)]
pub struct ItemStack {
    name: String,
    count: i32,
}

#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord)]
pub enum GameMode {
    Survival = 0,
    Creative = 1,
}

// A mock action struct
#[derive(Clone, PartialEq)]
pub struct SetGameModeAction {
    #[prost(string, tag = "1")]
    pub player_uuid: String,
    #[prost(enumeration = "GameMode", tag = "2")]
    pub game_mode: i32,
}

// Another mock action struct
#[derive(Clone, PartialEq)]
pub struct GiveItemAction {
    #[prost(string, tag = "1")]
    pub player_uuid: String,
    #[prost(message, optional, tag = "2")]
    pub item: ::core::option::Option<ItemStack>,
}

// Mock action enum
#[allow(dead_code)]
mod action {
    pub enum Kind {
        SetGameMode(super::SetGameModeAction),
        GiveItem(super::GiveItemAction),
    }
}

#[derive(Clone, PartialEq)]
pub struct ChatEvent {
    #[prost(string, tag = "1")]
    pub player_uuid: String,
    #[prost(string, tag = "2")]
    pub message: String,
}

#[derive(Clone, PartialEq)]
pub struct BlockBreakEvent {
    #[prost(string, tag = "1")]
    pub player_uuid: String,
}

#[allow(dead_code)]
mod event_envelope {
    pub enum Payload {
        Chat(super::ChatEvent),
        BlockBreak(super::BlockBreakEvent),
    }
}

#[derive(Clone, PartialEq)]
pub struct ChatMutation {
    #[prost(message, optional, tag = "1")]
    pub message: ::core::option::Option<String>,
}

#[derive(Clone, PartialEq)]
pub struct BlockBreakMutation {
    #[prost(message, optional, tag = "1")]
    pub drops: ::core::option::Option<ItemStackList>,
}

// The other missing enum your test is looking for!
#[allow(dead_code)]
mod event_result {
    pub enum Update {
        Chat(super::ChatMutation),
        BlockBreak(super::BlockBreakMutation),
    }
}
