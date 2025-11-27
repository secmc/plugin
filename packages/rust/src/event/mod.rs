//! Core event types and helpers for the Rust plugin SDK.
//!
//! Most plugin authors will work with:
//! - [`EventContext`], which wraps each incoming event and lets you cancel
//!   or mutate it before the host processes it.
//! - [`EventHandler`], a trait with an async method per event type. You
//!   typically implement this inside an `#[event_handler]` block.
//!
//! The concrete event structs (`ChatEvent`, `PlayerJoinEvent`, …) live in
//! [`crate::types`], generated from the protobuf definitions.

pub mod context;
pub mod handler;
pub mod mutations;

pub use context::*;
pub use handler::*;
