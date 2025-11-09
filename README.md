# Dragonfly Plugin System

A powerful, language-agnostic plugin system for [Dragonfly](https://github.com/df-mc/dragonfly) Minecraft Bedrock servers using gRPC and Protocol Buffers.

## Why Dragonfly's Plugin System?

| Benefit | Description | Use Case |
|---------|-------------|----------|
| 🌍 **Any Language** | Write plugins in JavaScript, TypeScript, PHP, Python, Rust, C++, or any language with gRPC support | Use the language your team knows best |
| 💰 **Sell Plugins** | Compile to binary (Rust, Go, C++) and distribute without source code | Create commercial plugins |
| 🔥 **Hot Reload** | Edit JS/TS/PHP plugins and see changes instantly - no server restart needed | Develop and debug plugins in real-time |
| 📱 **Remote Control** | Plugins connect over gRPC - run them anywhere (phone app, cloud service, discord bot) | Build mobile admin apps |
| 📦 **Use Any Library** | Import npm packages on a Go server, use Python ML libraries, etc. | Leverage entire ecosystems |
| ⚡ **Zero Performance Impact (COMING SOON)** | Plugins run in separate processes - slow/heavy plugin code doesn't affect server TPS | Run intensive tasks without lag |
| 🚀 **High Performance (COMING SOON)** | Optimized protobuf protocol with optional batching for low latency | Handle 100+ players with movement events |
| 🔒 **Sandboxing** | Control what plugins can access via gRPC permissions | Host untrusted plugins safely |

### Real-World Examples

```bash
# Hot reload: Edit plugin code while server is running
vim plugins/my-plugin.js   # Make changes
# Changes apply immediately - no restart!

# Remote plugin: Control server from your phone
# Plugin runs on your phone, connects to server over internet
phone-app → [gRPC] → Dragonfly Server

# Binary plugin: Sell without source code
rustc plugin.rs --release   # Compile to binary
# Distribute the binary - customers can't see your code
```

## Features

- **Multi-Language Support**: Write plugins in JavaScript, TypeScript, PHP, Python, Rust, C++, or any language with gRPC support
- **Event-Driven Architecture**: Subscribe to specific events (player join, chat, block break, etc.)
- **Type Safety**: Generated types for TypeScript and other statically typed languages

## Quick Start

### 1. Clone the Repository

```bash
git clone https://github.com/secmc/plugin.git
cd plugin
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Configure Plugins

Edit `plugins/plugins.yaml`:

```yaml
plugins:
  - id: my-plugin
    name: My First Plugin
    command: "node"
    args: ["examples/plugins/node/hello.js"]
    address: "127.0.0.1:50051"
```

### 4. Run the Server

```bash
go run main.go
```

Your plugin will automatically connect and start receiving events!

## Example Plugins

We provide complete working examples in multiple languages:

- **[TypeScript](examples/plugins/typescript/)** - Type-safe plugin with generated types (recommended for production)
- **[Node.js](examples/plugins/node/)** - Simple JavaScript plugin
- **[PHP](examples/plugins/php/)** - PHP plugin using gRPC extension

See [examples/plugins/README.md](examples/plugins/README.md) for detailed documentation and more examples.

## Creating Your First Plugin

### Minimal Example (PHP)

```php
<?php
// Example plugin showing command handling and block break event modification
require_once __DIR__ . '/vendor/autoload.php';

use Grpc\ChannelCredentials;

$pluginId = 'my-plugin';
$address = '127.0.0.1:50051';

$client = new \Df\Plugin\PluginClient($address, [
    'credentials' => ChannelCredentials::createInsecure(),
]);

$stream = $client->EventStream();

try {
    foreach ($stream->responses() as $message) {
        // Handle handshake
        if ($message->hasHello()) {
            $hello = new \DF\Plugin\PluginToHost();
            $hello->setPluginId($pluginId);
            $pluginHello = new \DF\Plugin\PluginHello();
            $pluginHello->setName('My Plugin');
            $pluginHello->setVersion('1.0.0');
            $pluginHello->setApiVersion($message->getHello()->getApiVersion());
            
            // Register /mine command
            $command = new \DF\Plugin\CommandSpec();
            $command->setName('/mine');
            $command->setDescription('Get mining boost');
            $pluginHello->setCommands([$command]);
            $hello->setHello($pluginHello);
            $stream->write($hello);

            // Subscribe to events
            $sub = new \DF\Plugin\PluginToHost();
            $sub->setPluginId($pluginId);
            $subscribe = new \DF\Plugin\EventSubscribe();
            $subscribe->setEvents(['PLAYER_JOIN', 'COMMAND', 'BLOCK_BREAK']);
            $sub->setSubscribe($subscribe);
            $stream->write($sub);
            continue;
        }

        if ($message->hasEvent()) {
            $event = $message->getEvent();
            
            // Handle /mine command
            if ($event->getType() === 'COMMAND' && $event->hasCommand()) {
                $cmd = $event->getCommand();
                if ($cmd->getCommand() === 'mine') {
                    // Send message to player
                    $action = new \DF\Plugin\Action();
                    $send = new \DF\Plugin\SendChatAction();
                    $send->setTargetUuid($cmd->getPlayerUuid());
                    $send->setMessage('§6⛏️ Mining boost activated! Break blocks for double drops!');
                    $action->setSendChat($send);
                    $batch = new \DF\Plugin\ActionBatch();
                    $batch->setActions([$action]);
                    $resp = new \DF\Plugin\PluginToHost();
                    $resp->setPluginId($pluginId);
                    $resp->setActions($batch);
                    $stream->write($resp);
                }
                
                // Acknowledge event
                $result = new \DF\Plugin\EventResult();
                $result->setEventId($event->getEventId());
                $result->setCancel(false);
                $resp = new \DF\Plugin\PluginToHost();
                $resp->setPluginId($pluginId);
                $resp->setEventResult($result);
                $stream->write($resp);
            }
            
            // Handle block break with double drops
            if ($event->getType() === 'BLOCK_BREAK' && $event->hasBlockBreak()) {
                $blockBreak = $event->getBlockBreak();
                echo "[php] {$blockBreak->getName()} broke block at ";
                echo "{$blockBreak->getX()},{$blockBreak->getY()},{$blockBreak->getZ()}\n";
                
                // Give double drops for every 10th block (X coordinate % 10 == 0)
                if ($blockBreak->getX() % 10 === 0) {
                    $drop = new \DF\Plugin\ItemStack();
                    $drop->setName('minecraft:diamond');
                    $drop->setCount(2);
                    $drop->setMeta(0);
                    
                    $mutation = new \DF\Plugin\BlockBreakMutation();
                    $mutation->setDrops([$drop]);
                    $mutation->setXp(10);
                    
                    $result = new \DF\Plugin\EventResult();
                    $result->setEventId($event->getEventId());
                    $result->setBlockBreak($mutation);
                    $resp = new \DF\Plugin\PluginToHost();
                    $resp->setPluginId($pluginId);
                    $resp->setEventResult($result);
                    $stream->write($resp);
                } else {
                    // Acknowledge normally
                    $result = new \DF\Plugin\EventResult();
                    $result->setEventId($event->getEventId());
                    $result->setCancel(false);
                    $resp = new \DF\Plugin\PluginToHost();
                    $resp->setPluginId($pluginId);
                    $resp->setEventResult($result);
                    $stream->write($resp);
                }
            }
        }
    }
} catch (Exception $e) {
    echo "[php] Error: " . $e->getMessage() . "\n";
} finally {
    $stream->writesDone();
}

echo "[php] plugin connected to {$address}\n";
```

## Project Structure

```
dragonfly-plugins/
├── dragonfly/              # Modified Dragonfly server with plugin support
├── plugin/                 # Plugin system core
│   ├── proto/             # Protocol Buffer definitions
│   ├── manager.go         # Plugin lifecycle management
│   └── README.md          # Plugin system documentation
├── examples/
│   └── plugins/           # Example plugins in various languages
├── plugins/
│   └── plugins.yaml       # Plugin configuration
└── main.go                # Server entry point
```

## How It Works

```
┌─────────────────┐         gRPC Stream          ┌──────────────────┐
│                 │ ←──────────────────────────→  │                  │
│  Dragonfly      │   Events: JOIN, CHAT, etc.   │  Your Plugin     │
│  Server         │   Actions: TELEPORT, etc.    │  (Any Language)  │
│  (Go)           │                               │                  │
└─────────────────┘                               └──────────────────┘
```

1. **Server starts** and loads plugin configuration from `plugins/plugins.yaml`
2. **Plugin process launches** via configured command (e.g., `node plugin.js`)
3. **Handshake** occurs where plugin registers capabilities
4. **Plugin subscribes** to events it wants to receive
5. **Events flow** from server to plugin in real-time
6. **Plugin executes actions** by sending messages back to server

## Documentation

- **[Plugin Examples](examples/plugins/README.md)** - Complete guide to example plugins
- **[Plugin System](plugin/README.md)** - Core plugin system documentation
- **[Protocol Buffer Definitions](plugin/proto/types/plugin.proto)** - API reference
- **[Plugin Architecture](docs/plugin-architecture.md)** - Design documentation

