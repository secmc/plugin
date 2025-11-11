# Dragonfly Plugin Examples

This directory contains example plugins demonstrating how to create Dragonfly plugins in different languages.

## Available Examples

### 1. Node.js Plugin (`node/`)

Simple JavaScript plugin using `@grpc/grpc-js` and `@grpc/proto-loader`.

```bash
cd node/
npm install
```

---

### 2. TypeScript Plugin (`typescript/`)

Type-safe plugin with generated types

```bash
cd typescript/
npm install
npm run generate
```

---

### 3. PHP Plugin (`php/`)

PHP plugin using gRPC extension.

```bash
cd php/
# Requires: php-grpc extension installed
php src/HelloPlugin.php
```

**Features:**
- ✅ Use existing PHP libraries
- ⚠️ Requires gRPC extension

---
## Quick Start

1. **Choose a language** based on your needs (TypeScript recommended for production)
2. **Follow the setup** in that example's directory
3. **Enable in config** - Edit `plugins/plugins.yaml`:
   ```yaml
   plugins:
     - id: my-plugin
       name: My Plugin
       command: "node"
       args: ["examples/plugins/typescript/dist/index.js"]
       address: "127.0.0.1:50051"
   ```
4. **Run Dragonfly** - The plugin will connect automatically

## Plugin Configuration

Edit `plugins/plugins.yaml` to enable/configure plugins:

```yaml
server_port: 50050

plugins:
  # Node.js example
  - id: example-node
    name: Example Node Plugin
    command: "node"
    args: ["examples/plugins/node/hello.js"]
    env:
      NODE_ENV: development

  # TypeScript example
  - id: example-typescript
    name: Example TypeScript Plugin
    command: "node"
    args: ["examples/plugins/typescript/dist/index.js"]

  # PHP example
  - id: example-php
    name: Example PHP Plugin
    command: "php"
    args: ["examples/plugins/php/src/HelloPlugin.php"]
```

## Protocol Documentation

All plugins communicate using the same protobuf protocol defined in `plugin/proto/plugin.proto`.

**Key concepts:**

### 1. Bidirectional Stream

Plugins use a single bidirectional gRPC stream for all communication:

```
Host ←→ Plugin (EventStream)
```

### 2. Message Types

**Host → Plugin:**
- `HostHello` - Initial handshake
- `EventEnvelope` - Game events (join, quit, chat, commands, etc.)
- `HostShutdown` - Server shutting down

**Plugin → Host:**
- `PluginHello` - Register plugin capabilities
- `EventSubscribe` - Subscribe to specific events
- `ActionBatch` - Execute actions (teleport, chat, kick, etc.)
- `EventResult` - Cancel or mutate events

### 3. Event Flow

```
1. Plugin connects and sends PluginHello (includes plugin ID & commands)
2. Plugin sends EventSubscribe (which events to receive)
3. Host responds with HostHello (API version handshake)
4. Host sends events as they occur
5. Plugin can respond with:
   - Actions (do something)
   - EventResult (cancel/modify event)
```

### 4. Example Event Types

Values come from the `EventType` enum:

- `PLAYER_JOIN` - Player connected
- `PLAYER_QUIT` - Player disconnected
- `CHAT` - Player sent chat message
- `COMMAND` - Player executed command
- `PLAYER_BLOCK_BREAK` - Player broke a block
- `WORLD_CLOSE` - World is closing

### 5. Example Actions

- `SendChatAction` - Send message to player
- `TeleportAction` - Teleport player
- `KickAction` - Kick player

## Creating Your Own Plugin

### Minimal Plugin (Node.js)

```javascript
import grpc from '@grpc/grpc-js';
import protoLoader from '@grpc/proto-loader';

const packageDef = protoLoader.loadSync('plugin/proto/plugin.proto');
const proto = grpc.loadPackageDefinition(packageDef).df.plugin;

const server = new grpc.Server();
server.addService(proto.Plugin.service, {
  EventStream: (call) => {
    call.on('data', (msg) => {
      if (msg.hello) {
        call.write({
          pluginId: 'my-plugin',
          hello: {
            name: 'My Plugin',
            version: '1.0.0',
            apiVersion: msg.hello.apiVersion,
          }
        });
      }
    });
  }
});

server.bindAsync('127.0.0.1:50051', 
  grpc.ServerCredentials.createInsecure(), 
  () => console.log('Plugin ready'));
```
