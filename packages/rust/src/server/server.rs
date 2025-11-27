//! Lightweight handle for sending actions and subscriptions to the host.
//!
//! Plugin authors receive a [`Server`] reference in every event handler.
//! It can be cloned freely and used to send actions like `send_chat`,
//! `teleport`, or `world_set_block` back to the Dragonfly host.

use tokio::sync::mpsc;

use crate::types::{self, PluginToHost};

#[derive(Clone)]
pub struct Server {
    pub plugin_id: String,
    pub sender: mpsc::Sender<PluginToHost>,
}

impl Server {
    /// Helper to build and send a single action.
    pub async fn send_action(
        &self,
        kind: types::action::Kind,
    ) -> Result<(), mpsc::error::SendError<PluginToHost>> {
        let action = types::Action {
            correlation_id: None,
            kind: Some(kind),
        };
        let batch = types::ActionBatch {
            actions: vec![action],
        };
        let msg = PluginToHost {
            plugin_id: self.plugin_id.clone(),
            payload: Some(types::PluginPayload::Actions(batch)),
        };
        self.sender.send(msg).await
    }

    /// Helper to send a batch of actions.
    pub async fn send_actions(
        &self,
        actions: Vec<types::Action>,
    ) -> Result<(), mpsc::error::SendError<PluginToHost>> {
        let batch = types::ActionBatch { actions };
        let msg = PluginToHost {
            plugin_id: self.plugin_id.clone(),
            payload: Some(types::PluginPayload::Actions(batch)),
        };
        self.sender.send(msg).await
    }

    /// Subscribe to a list of game events.
    pub async fn subscribe(
        &self,
        events: Vec<types::EventType>,
    ) -> Result<(), mpsc::error::SendError<PluginToHost>> {
        let sub = types::EventSubscribe {
            events: events.into_iter().map(|e| e.into()).collect(),
        };
        let msg = PluginToHost {
            plugin_id: self.plugin_id.clone(),
            payload: Some(types::PluginPayload::Subscribe(sub)),
        };
        self.sender.send(msg).await
    }
}

mod helpers;
