use dragonfly_plugin::{server::Server, types};
use tokio::sync::mpsc;

#[tokio::test]
async fn send_action_wraps_single_action_in_batch() {
    let (tx, mut rx) = mpsc::channel(1);

    let server = Server {
        plugin_id: "plugin-id".to_string(),
        sender: tx,
    };

    let kind = types::action::Kind::SendChat(types::SendChatAction {
        target_uuid: "player-uuid".to_string(),
        message: "hello".to_string(),
    });

    server.send_action(kind).await.expect("send_action failed");

    let msg = rx.recv().await.expect("expected PluginToHost message");
    assert_eq!(msg.plugin_id, "plugin-id");

    match msg.payload.expect("missing payload") {
        types::PluginPayload::Actions(batch) => {
            assert_eq!(batch.actions.len(), 1);
            let action = &batch.actions[0];
            assert!(action.correlation_id.is_none());
            match action.kind.as_ref().expect("missing action kind") {
                types::ActionKind::SendChat(chat) => {
                    assert_eq!(chat.target_uuid, "player-uuid");
                    assert_eq!(chat.message, "hello");
                }
                other => panic!("unexpected action kind: {:?}", other),
            }
        }
        other => panic!("unexpected payload: {:?}", other),
    }
}

#[tokio::test]
async fn send_chat_helper_builds_correct_action() {
    let (tx, mut rx) = mpsc::channel(1);

    let server = Server {
        plugin_id: "plugin-id".to_string(),
        sender: tx,
    };

    server
        .send_chat("player-uuid".to_string(), "hi there".to_string())
        .await
        .expect("send_chat failed");

    let msg = rx.recv().await.expect("expected PluginToHost message");
    match msg.payload.expect("missing payload") {
        types::PluginPayload::Actions(batch) => {
            assert_eq!(batch.actions.len(), 1);
            let action = &batch.actions[0];
            match action.kind.as_ref().expect("missing action kind") {
                types::ActionKind::SendChat(chat) => {
                    assert_eq!(chat.target_uuid, "player-uuid");
                    assert_eq!(chat.message, "hi there");
                }
                other => panic!("unexpected action kind: {:?}", other),
            }
        }
        other => panic!("unexpected payload: {:?}", other),
    }
}

#[tokio::test]
async fn subscribe_sends_subscribe_payload() {
    let (tx, mut rx) = mpsc::channel(1);

    let server = Server {
        plugin_id: "plugin-id".to_string(),
        sender: tx,
    };

    server
        .subscribe(vec![types::EventType::Chat, types::EventType::Command])
        .await
        .expect("subscribe failed");

    let msg = rx.recv().await.expect("expected PluginToHost message");
    assert_eq!(msg.plugin_id, "plugin-id");

    match msg.payload.expect("missing payload") {
        types::PluginPayload::Subscribe(sub) => {
            // Order is preserved from the vec we passed in.
            assert_eq!(sub.events.len(), 2);
            assert_eq!(sub.events[0], types::EventType::Chat as i32);
            assert_eq!(sub.events[1], types::EventType::Command as i32);
        }
        other => panic!("unexpected payload: {:?}", other),
    }
}


