use tokio::sync::mpsc;

use crate::types::{self, PluginToHost};

/// This enum is used internally by `dispatch_event` to
/// determine what action to take after an event handler runs.
#[doc(hidden)]
#[derive(Debug)]
pub enum EventResult {
    /// Do nothing, let the default server behavior happen. just sends ack.
    None,
    /// Cancel the event, stopping default server behavior.
    Cancelled,
    /// Mutate the event, which is sent back to the server.
    Mutated(types::event_result::Update),
}

/// A smart wrapper for a server event.
///
/// This struct provides read-only access to the event's data
/// and methods to mutate or cancel it.
pub struct EventContext<'a, T> {
    pub data: &'a T,
    pub result: EventResult,

    event_id: &'a str,
    sender: mpsc::Sender<PluginToHost>,
    plugin_id: String,
    sent: bool,
}

impl<'a, T> EventContext<'a, T> {
    #[doc(hidden)]
    pub fn new(
        event_id: &'a str,
        data: &'a T,
        sender: mpsc::Sender<PluginToHost>,
        plugin_id: String,
    ) -> Self {
        Self {
            event_id,
            data,
            result: EventResult::None,
            sender,
            plugin_id,
            sent: false,
        }
    }

    /// Consumes the context and returns the final result.
    #[doc(hidden)]
    pub fn into_result(self) -> (String, EventResult) {
        (self.event_id.to_string(), self.result)
    }

    /// Cancels the event.
    ///
    pub async fn cancel(&mut self) {
        self.result = EventResult::Cancelled;
        self.send().await
    }

    pub(crate) async fn send_ack_if_needed(&mut self) {
        if self.sent {
            return;
        }
        // result is still EventResultUpdate::None, which sends ack
        self.send().await;
    }

    pub async fn send(&mut self) {
        if self.sent {
            #[cfg(debug_assertions)]
            panic!("Attempted to respond twice to the same event!");

            #[cfg(not(debug_assertions))]
            {
                eprintln!("Warning: send() called after response already sent");
                return;
            }
        }

        self.sent = true;

        let event_id = self.event_id.to_owned();

        let payload = match &self.result {
            // If nothing was changed just send ack.
            EventResult::None => types::EventResult {
                event_id,
                cancel: None,
                update: None,
            },
            EventResult::Cancelled => types::EventResult {
                event_id,
                cancel: Some(true),
                update: None,
            },
            EventResult::Mutated(update) => types::EventResult {
                event_id,
                cancel: None,
                // TODO: later try to fix this clone.
                // this gives us best API usage but is memory semantically wrong.
                // calling this func or like .cancel should consume event.
                //
                // but for newbies thats hard to understand.
                update: Some(update.clone()),
            },
        };

        let msg = types::PluginToHost {
            plugin_id: self.plugin_id.clone(),
            payload: Some(types::PluginPayload::EventResult(payload)),
        };

        if let Err(e) = self.sender.send(msg).await {
            eprintln!("Failed to send event response: {}", e);
        }
    }
}
