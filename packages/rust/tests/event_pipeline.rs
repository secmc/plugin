use dragonfly_plugin::{
    command::Ctx,
    event::{EventContext, EventHandler},
    event_handler,
    server::Server,
    types, Command, EventSubscriptions, Plugin,
};
use tokio::sync::{mpsc, Mutex};

#[tokio::test]
async fn event_context_cancel_sends_cancelled_result() {
    let (tx, mut rx) = mpsc::channel(1);

    let chat = types::ChatEvent {
        player_uuid: "player-uuid".to_string(),
        name: "Player".to_string(),
        message: "hello".to_string(),
    };

    let mut ctx = EventContext::new("event-1", &chat, tx, "plugin-id".to_string());
    ctx.cancel().await;

    let msg = rx.recv().await.expect("expected event result message");
    assert_eq!(msg.plugin_id, "plugin-id");

    match msg.payload.expect("missing payload") {
        types::PluginPayload::EventResult(result) => {
            assert_eq!(result.event_id, "event-1");
            assert_eq!(result.cancel, Some(true));
            assert!(result.update.is_none());
        }
        other => panic!("unexpected payload: {:?}", other),
    }
}

#[tokio::test]
async fn event_context_mutation_helper_sets_update() {
    let (tx, mut rx) = mpsc::channel(1);

    let chat = types::ChatEvent {
        player_uuid: "player-uuid".to_string(),
        name: "Player".to_string(),
        message: "before".to_string(),
    };

    let mut ctx = EventContext::new("event-mutate", &chat, tx, "plugin-id".to_string());
    ctx.set_message("after".to_string());
    ctx.send().await;

    let msg = rx.recv().await.expect("expected mutation result");

    match msg.payload.expect("missing payload") {
        types::PluginPayload::EventResult(result) => {
            assert_eq!(result.event_id, "event-mutate");
            assert!(result.cancel.is_none());

            let update = result.update.expect("missing update");
            match update {
                types::EventResultUpdate::Chat(mutation) => {
                    assert_eq!(mutation.message.as_deref(), Some("after"));
                }
                other => panic!("unexpected update variant: {:?}", other),
            }
        }
        other => panic!("unexpected payload: {:?}", other),
    }
}

#[derive(Default)]
struct RecordingPlugin {
    calls: Mutex<Vec<&'static str>>,
}

impl EventSubscriptions for RecordingPlugin {
    fn get_subscriptions(&self) -> Vec<types::EventType> {
        vec![types::EventType::Chat]
    }
}

impl dragonfly_plugin::command::CommandRegistry for RecordingPlugin {}

impl EventHandler for RecordingPlugin {
    async fn on_chat(&self, _server: &Server, _event: &mut EventContext<'_, types::ChatEvent>) {
        self.calls.lock().await.push("chat");
    }
}

#[tokio::test]
async fn dispatch_event_routes_chat_to_handler() {
    let (tx, mut rx) = mpsc::channel(1);

    let server = Server {
        plugin_id: "plugin-id".to_string(),
        sender: tx,
    };

    let plugin = RecordingPlugin::default();

    let chat = types::ChatEvent {
        player_uuid: "player-uuid".to_string(),
        name: "Player".to_string(),
        message: "hello".to_string(),
    };

    let envelope = types::EventEnvelope {
        event_id: "chat-event".to_string(),
        r#type: types::EventType::Chat as i32,
        expects_response: true,
        payload: Some(types::EventPayload::Chat(chat)),
    };

    dragonfly_plugin::event::dispatch_event(&server, &plugin, &envelope).await;

    // Handler was called.
    let calls = plugin.calls.lock().await;
    assert_eq!(calls.as_slice(), &["chat"]);
    drop(calls);

    // Ack was sent.
    let msg = rx.recv().await.expect("expected ack from dispatch_event");
    assert_eq!(msg.plugin_id, "plugin-id");
}

#[tokio::test]
#[should_panic(expected = "Attempted to respond twice to the same event!")]
async fn event_context_double_send_panics_in_debug() {
    let (tx, _rx) = mpsc::channel(1);

    let chat = types::ChatEvent {
        player_uuid: "player-uuid".to_string(),
        name: "Player".to_string(),
        message: "hello".to_string(),
    };

    let mut ctx = EventContext::new("event-double", &chat, tx, "plugin-id".to_string());

    // First send is fine.
    ctx.send().await;
    // Second send should panic in debug builds.
    ctx.send().await;
}

#[derive(Default, Plugin)]
#[plugin(
    id = "test-plugin",
    name = "Test Plugin",
    version = "0.0.0",
    api = "1.0.0",
    commands(PingCommand)
)]
struct CommandPlugin {
    calls: Mutex<Vec<String>>,
}

#[derive(Debug, Command)]
#[command(name = "ping", description = "Ping command")]
struct PingCommand {
    value: i32,
}

#[event_handler]
impl EventHandler for CommandPlugin {
    async fn on_command(
        &self,
        _server: &Server,
        _event: &mut EventContext<'_, types::CommandEvent>,
    ) {
        self.calls
            .lock()
            .await
            .push("on_command_fallback".to_string());
    }
}

impl PingCommandHandler for CommandPlugin {
    async fn ping_command(&self, ctx: Ctx<'_>, value: i32) {
        self.calls
            .lock()
            .await
            .push(format!("handled:{}:{value}", ctx.sender));
    }
}

#[tokio::test]
async fn dispatch_event_dispatches_command_before_on_command() {
    let (tx, mut rx) = mpsc::channel(1);

    let server = Server {
        plugin_id: "plugin-id".to_string(),
        sender: tx,
    };

    let plugin = CommandPlugin::default();

    let cmd_event = types::CommandEvent {
        player_uuid: "player-uuid".to_string(),
        name: "/ping 5".to_string(),
        raw: "/ping 5".to_string(),
        command: "ping".to_string(),
        args: vec!["5".to_string()],
    };

    let envelope = types::EventEnvelope {
        event_id: "cmd-event".to_string(),
        r#type: types::EventType::Command as i32,
        expects_response: true,
        payload: Some(types::EventPayload::Command(cmd_event)),
    };

    dragonfly_plugin::event::dispatch_event(&server, &plugin, &envelope).await;

    // Command handler should have run, but on_command fallback should not.
    let calls = plugin.calls.lock().await;
    assert_eq!(calls.len(), 1);
    assert!(calls[0].starts_with("handled:player-uuid:5"));
    drop(calls);

    // An EventResult ack should have been sent.
    let msg = rx.recv().await.expect("expected command EventResult");
    assert_eq!(msg.plugin_id, "plugin-id");
}
