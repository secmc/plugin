use dragonfly_plugin_macro::Command;

#[derive(Command)]
#[command(name = "ping", description = "Ping command", aliases("p"))]
pub struct Ping {
    times: i32,
}


