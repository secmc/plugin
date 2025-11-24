use crate::types;

/// This enum is used internally by `dispatch_event` to
/// determine what action to take after an event handler runs.
#[doc(hidden)]
pub enum EventResultUpdate {
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

    event_id: &'a str,
    result: EventResultUpdate,
}

impl<'a, T> EventContext<'a, T> {
    #[doc(hidden)]
    pub fn new(event_id: &'a str, data: &'a T) -> Self {
        Self {
            event_id,
            data,
            result: EventResultUpdate::None,
        }
    }

    /// Consumes the context and returns the final result.
    #[doc(hidden)]
    pub fn into_result(self) -> (String, EventResultUpdate) {
        (self.event_id.to_string(), self.result)
    }

    /// Cancels the event.
    ///
    /// The server's default handler will not run.
    pub fn cancel(&mut self) {
        self.result = EventResultUpdate::Cancelled;
    }

    /// Internal helper to set a mutation.
    /// This is called by the auto-generated helper methods.
    #[doc(hidden)]
    pub fn set_mutation(&mut self, update: types::event_result::Update) {
        self.result = EventResultUpdate::Mutated(update);
    }
}
