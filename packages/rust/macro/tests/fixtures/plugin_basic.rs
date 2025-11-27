use dragonfly_plugin_macro::{event_handler, Plugin};

struct MyPlugin;

#[event_handler]
impl dragonfly_plugin::EventHandler for MyPlugin {
    async fn on_chat(
        &self,
        _server: &dragonfly_plugin::Server,
        _event: &mut dragonfly_plugin::event::EventContext<
            '_,
            dragonfly_plugin::types::ChatEvent,
        >,
    ) {
    }
}

#[derive(Plugin)]
#[plugin(id = "example-rust", name = "Example Rust Plugin", version = "0.3.0", api = "1.0.0")]
struct PluginInfoPlugin;


