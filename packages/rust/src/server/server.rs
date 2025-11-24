use tokio::sync::mpsc;

use crate::{
    event::{EventContext, EventResultUpdate},
    types::{self, PluginToHost},
};

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

    /// Internal helper to send an event result (cancel/mutate)
    /// This is called by the auto-generated `dispatch_event` function.
    #[doc(hidden)]
    pub(crate) async fn send_event_result(
        &self,
        context: EventContext<'_, impl Sized>,
    ) -> Result<(), mpsc::error::SendError<types::PluginToHost>> {
        let (event_id, result) = context.into_result();

        let payload = match result {
            // Do nothing if the handler didn't mutate or cancel
            EventResultUpdate::None => return Ok(()),
            EventResultUpdate::Cancelled => types::EventResult {
                event_id,
                cancel: Some(true),
                update: None,
            },
            EventResultUpdate::Mutated(update) => types::EventResult {
                event_id,
                cancel: None,
                update: Some(update),
            },
        };

        let msg = types::PluginToHost {
            plugin_id: self.plugin_id.clone(),
            payload: Some(types::PluginPayload::EventResult(payload)),
        };
        self.sender.send(msg).await
    }
}

mod helpers;
