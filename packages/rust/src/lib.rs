#[doc = include_str!("../README.md")]
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

pub mod command;
pub mod event;
#[path = "server/server.rs"]
pub mod server;

use std::error::Error;

pub use server::*;

// main usage stuff for plugin devs:
pub use dragonfly_plugin_macro::{event_handler, Command, Plugin};
pub use event::EventHandler;
use tokio::sync::mpsc;
use tokio_stream::{wrappers::ReceiverStream, StreamExt};

#[cfg(unix)]
use hyper_util::rt::TokioIo;
#[cfg(unix)]
use tokio::net::UnixStream;

use crate::command::CommandRegistry;

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

pub struct PluginRunner {}

impl PluginRunner {
    /// Runs the plugin, connecting to the server and starting the event loop.
    pub async fn run(plugin: impl Plugin + 'static, addr: &str) -> Result<(), Box<dyn Error>> {
        let mut raw_client = connect_to_server(addr).await?;

        let (tx, rx) = mpsc::channel(128);

        // Pre-buffer the hello message so it's sent immediately when stream opens.
        // This is required because the Go server blocks on Recv() waiting for the
        // hello before sending response headers.
        let hello_msg = types::PluginToHost {
            plugin_id: plugin.get_id().to_owned(),
            payload: Some(types::PluginPayload::Hello(types::PluginHello {
                name: plugin.get_name().to_owned(),
                version: plugin.get_version().to_owned(),
                api_version: plugin.get_api_version().to_owned(),
                commands: plugin.get_commands(),
                custom_items: vec![],
                custom_blocks: vec![],
            })),
        };
        tx.send(hello_msg).await?;

        let request_stream = ReceiverStream::new(rx);
        let mut event_stream = raw_client.event_stream(request_stream).await?.into_inner();

        let server = Server {
            plugin_id: plugin.get_id().to_owned(),
            sender: tx.clone(),
        };

        let mut events = plugin.get_subscriptions();

        // Auto-subscribe to Command if plugin has registered commands
        if !plugin.get_commands().is_empty() && !events.contains(&types::EventType::Command) {
            events.push(types::EventType::Command);
        }

        if !events.is_empty() {
            println!("Subscribing to {} event types...", events.len());
            server.subscribe(events).await?;
        }

        println!("Plugin '{}' connected and listening.", plugin.get_name());

        // 8. Run the main event loop
        while let Some(Ok(msg)) = event_stream.next().await {
            match msg.payload {
                // We received a game event
                Some(types::HostPayload::Event(envelope)) => {
                    event::dispatch_event(&server, &plugin, &envelope).await;
                }
                // The server is shutting us down
                Some(types::HostPayload::Shutdown(shutdown)) => {
                    println!("Server shutting down plugin: {}", shutdown.reason);
                    break; // Break the loop
                }
                _ => { /* Ignore other payloads */ }
            }
        }

        println!("Plugin '{}' disconnected.", plugin.get_name());
        Ok(())
    }
}

/// A trait that defines which events your plugin will receive.
///
/// You can implement this trait manually, or you can use the
/// `#[derive(Plugin)]` along with `#[events(Event1, Event2)`
/// implementation to generate it for you.
pub trait EventSubscriptions {
    fn get_subscriptions(&self) -> Vec<types::EventType>;
}

/// A struct that defines the details of your plugin.
pub struct PluginInfo<'a> {
    pub id: &'a str,
    pub name: &'a str,
    pub version: &'a str,
    pub api_version: &'a str,
}

/// The final trait required for our plugin to be runnable.
///
/// These functions get impled automatically by
/// `#[derive(Plugin)` like so:
/// ```rust
/// use dragonfly_plugin::{
///    PluginRunner,      // Our runtime, clearly named
///    Plugin,         // The derive macro
///    event::{EventContext, EventHandler},
///    event_handler,
///    types,
///    Server,
/// };
///
/// #[derive(Plugin, Default)]
/// #[plugin(
///    id = "example-rust",
///    name = "Example Rust Plugin",
///    version = "1.0.0",
///    api = "1.0.0"
/// )]
///struct MyPlugin {}
///
///#[event_handler]
///impl EventHandler for MyPlugin {
///    async fn on_player_join(
///        &self,
///        server: &Server,
///        event: &mut EventContext<'_, types::PlayerJoinEvent>,
///    ) { }
/// }
/// ```
pub trait Plugin: EventHandler + EventSubscriptions + CommandRegistry {
    fn get_info(&self) -> PluginInfo<'_>;

    fn get_id(&self) -> &str;

    fn get_name(&self) -> &str;

    fn get_version(&self) -> &str;

    fn get_api_version(&self) -> &str;
}
