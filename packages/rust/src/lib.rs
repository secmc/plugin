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

pub use server::*;

// main usage stuff for plugin devs:
pub use event::PluginEventHandler;
pub use rust_plugin_macro::Handler;
use tokio::sync::mpsc;
use tokio_stream::{wrappers::ReceiverStream, StreamExt};

#[cfg(unix)]
use hyper_util::rt::TokioIo;
#[cfg(unix)]
use tokio::net::UnixStream;

/// Helper function to connect to the server, supporting both Unix sockets and TCP.
async fn connect_to_server(
    addr: &str,
) -> Result<types::PluginClient<tonic::transport::Channel>, Box<dyn Error>> {
    // Check if it's a Unix socket address (starts with "unix:" or is a path starting with "/")
    if addr.starts_with("unix:") || addr.starts_with('/') {
        #[cfg(unix)]
        {
            // Extract the path and convert to owned String for the closure
            let path: String = if addr.starts_with("unix://") {
                addr[7..].to_string()
            } else if addr.starts_with("unix:") {
                addr[5..].to_string()
            } else {
                addr.to_string()
            };
            // Create a lazy channel that uses Unix sockets.
            // Lazy is required so the hello message gets sent as part of stream
            // establishment, avoiding a deadlock with the Go server which waits
            // for the hello before sending response headers.
            let channel = tonic::transport::Endpoint::try_from("http://[::1]:50051")?
                .connect_with_connector_lazy(service_fn(move |_: tonic::transport::Uri| {
                    let path = path.clone();
                    async move {
                        let stream = UnixStream::connect(&path).await?;
                        Ok::<_, std::io::Error>(TokioIo::new(stream))
                    }
                }));
            Ok(types::PluginClient::new(channel))
        }
        #[cfg(not(unix))]
        {
            Err("Unix sockets are not supported on this platform".into())
        }
    } else {
        // Regular TCP connection
        Ok(types::PluginClient::connect(addr.to_string()).await?)
    }
}

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
    pub async fn run(
        self,
        handler: impl PluginEventHandler + PluginSubscriptions + 'static,
        addr: &str,
    ) -> Result<(), Box<dyn Error>> {
        let mut raw_client = connect_to_server(addr).await?;

        let (tx, rx) = mpsc::channel(128);

        // Pre-buffer the hello message so it's sent immediately when stream opens.
        // This is required because the Go server blocks on Recv() waiting for the
        // hello before sending response headers.
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

        let request_stream = ReceiverStream::new(rx);
        let mut event_stream = raw_client.event_stream(request_stream).await?.into_inner();

        let server = Server {
            plugin_id: self.id.clone(),
            sender: tx.clone(),
        };

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
