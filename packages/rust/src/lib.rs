#![allow(clippy::all)]

#[path = "generated/df.plugin.rs"]
mod df_plugin;

pub mod types {
    pub use super::df_plugin::plugin_client::PluginClient;
    pub use super::df_plugin::*;
    pub use super::df_plugin::{
        action::Kind as ActionKind, event_envelope::Payload as EventPayload,
        event_result::Update as EventResultUpdate, host_to_plugin::Payload as HostPayload,
        plugin_to_host::Payload as PluginPayload,
    };
}

pub mod event;
#[path = "server/server.rs"]
pub mod server;

use std::error::Error;

// internal uses.
pub(crate) use server::*;

// main usage stuff for plugin devs:
pub use async_trait::async_trait;
pub use event::PluginEventHandler;
use tokio::sync::mpsc;
use tokio_stream::{wrappers::ReceiverStream, StreamExt};
// TODO: pub use rust_plugin_macro::bedrock_plugin;

pub struct Plugin {
    id: String,
    name: String,
    version: String,
    api_version: String,
}

impl Plugin {
    pub fn new(id: &str, name: &str, version: &str, api_version: &str) -> Self {
        Self {
            id: id.to_string(),
            name: name.to_string(),
            version: version.to_string(),
            api_version: api_version.to_string(),
        }
    }

    /// Runs the plugin, connecting to the server and starting the event loop.
    pub async fn run<A>(
        self,
        handler: impl PluginEventHandler + PluginSubscriptions + 'static,
        addr: A,
    ) -> Result<(), Box<dyn Error>>
    where
        // Yeah this was AI, but holy hell is it good at doing dynamic type stuff.
        A: TryInto<tonic::transport::Endpoint>,
        A::Error: Into<Box<dyn Error + Send + Sync>>,
    {
        let mut raw_client = types::PluginClient::connect(addr)
            .await
            .map_err(|e| Box::new(e) as Box<dyn Error>)?;

        let (tx, rx) = mpsc::channel(128);

        let request_stream = ReceiverStream::new(rx);

        let mut event_stream = raw_client.event_stream(request_stream).await?.into_inner();

        let server = Server {
            plugin_id: self.id.clone(),
            sender: tx.clone(),
        };

        let hello_msg = types::PluginToHost {
            plugin_id: self.id.clone(),
            payload: Some(types::PluginPayload::Hello(types::PluginHello {
                name: self.name.clone(),
                version: self.version.clone(),
                api_version: self.api_version.clone(),
                commands: vec![],
                custom_items: vec![],
            })),
        };
        tx.send(hello_msg).await?;

        let events = handler.get_subscriptions();
        if !events.is_empty() {
            println!("Subscribing to {} event types...", events.len());
            server.subscribe(events).await?;
        }

        println!("Plugin '{}' connected and listening.", self.name);

        // 8. Run the main event loop
        while let Some(Ok(msg)) = event_stream.next().await {
            match msg.payload {
                // We received a game event
                Some(types::HostPayload::Event(envelope)) => {
                    event::dispatch_event(&server, &handler, &envelope).await;
                }
                // The server is shutting us down
                Some(types::HostPayload::Shutdown(shutdown)) => {
                    println!("Server shutting down plugin: {}", shutdown.reason);
                    break; // Break the loop
                }
                _ => { /* Ignore other payloads */ }
            }
        }

        println!("Plugin '{}' disconnected.", self.name);
        Ok(())
    }
}

/// A trait that defines which events your plugin will receive.
///
/// You can implement this trait manually, or you can use the
/// `#[bedrock_plugin]` macro on your `PluginEventHandler`
/// implementation to generate it for you.
pub trait PluginSubscriptions {
    fn get_subscriptions(&self) -> Vec<types::EventType>;
}
