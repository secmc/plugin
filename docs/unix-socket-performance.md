# Unix Socket Performance Guide

## Overview

By default, plugins communicate with the Dragonfly server over TCP on localhost (`127.0.0.1:50050`). For **local plugins**, switching to Unix domain sockets can reduce latency from **1-5ms to under 1ms** (200-800μs typical).

## Why Unix Sockets Are Faster

| Transport | Round-trip Latency | Use Case |
|-----------|-------------------|----------|
| **TCP (localhost)** | 1-5ms | Default, works everywhere |
| **Unix Socket** | 200-800μs | Local plugins only, 5-10x faster |

Unix sockets bypass the TCP/IP network stack and use direct IPC (inter-process communication), making them ideal for plugins running on the same machine.

## Configuration

### Using Unix Sockets (Recommended for Performance)

Edit `cmd/plugins/plugins.yaml`:

```yaml
# Unix socket - faster for local plugins
server_addr: "unix:///tmp/dragonfly_plugin.sock"

plugins:
  - id: example-php
    name: Example PHP Plugin
    command: "../examples/plugins/php/bin/php7/bin/php"
    args: ["../examples/plugins/php/src/HelloPlugin.php"]
```

### TCP (Cross-platform)

```yaml
# TCP on localhost 
server_addr: "tcp://127.0.0.1:50050"
```

## Platform Support

### Linux / macOS

```yaml
server_addr: "unix:///tmp/dragonfly_plugin.sock"
```

### Windows 10+ / Windows 11

```yaml
server_addr: "tcp://127.0.0.1:50050"
```

### Older Windows (< Win 10 build 17063)

Use TCP mode:

```yaml
server_addr: "tcp://127.0.0.1:50050"
```

## Performance Impact by Event Type

Events that **block and wait for plugin responses** benefit the most:

| Event Type | Blocks? | Impact of Latency |
|------------|---------|-------------------|
| Player Move | ✅ Yes | High - happens every tick |
| Player Attack | ✅ Yes | High - combat feel |
| Player Chat | ✅ Yes | Medium - noticeable delay |
| Block Break | ✅ Yes | Medium |
| Player Join | ❌ No | None - fire-and-forget |
| Player Jump | ❌ No | None - fire-and-forget |

### Timeout Settings

The server waits up to **250ms** for plugin responses (defined in `manager.go`):

```go
const eventResponseTimeout = 250 * time.Millisecond
```

- **Fast plugins** (<1ms): No user-visible lag
- **Slow plugins** (>50ms): Noticeable delay in movement/combat
- **Hanging plugins** (>250ms): Event proceeds without plugin response

## Best Practices

1. **Use Unix sockets for local plugins** - Significantly reduces latency
2. **Respond quickly to cancellable events** - Movement and combat should return <1ms
3. **Defer heavy processing** - Use async tasks for database queries, API calls, etc.
4. **Monitor plugin performance** - Watch for timeout warnings in logs

## Switching from TCP to Unix Sockets

No code changes needed in your plugin! Just update the configuration:

**Before:**
```yaml
server_addr: "tcp://127.0.0.1:50050"
```

**After:**
```yaml
server_addr: "unix:///tmp/dragonfly_plugin.sock"
```

The PHP/TypeScript/Node plugins automatically use the `DF_PLUGIN_SERVER_ADDRESS` environment variable, which will be updated by Dragonfly.

## Example: PHP Plugin with Unix Socket

The PHP gRPC client automatically supports Unix sockets:

```php
// No changes needed! Works with both TCP and Unix sockets
$this->serverAddress = getenv('DF_PLUGIN_SERVER_ADDRESS') ?: '127.0.0.1:50050';
$this->client = new PluginClient($this->serverAddress, [
    'credentials' => ChannelCredentials::createInsecure(),
]);
```

## Troubleshooting

### Permission Denied

The socket file needs proper permissions. Dragonfly automatically sets `0666` permissions on Unix/Linux/macOS.

### Address Already in Use

Remove the old socket file:

```bash
rm /tmp/dragonfly_plugin.sock
```

Dragonfly automatically cleans up socket files on startup and shutdown.

### Plugin Cannot Connect

Check that:
1. The socket file exists: `ls -l /tmp/dragonfly_plugin.sock`
2. Dragonfly logs show: `plugin server listening on /tmp/dragonfly_plugin.sock`
3. Your plugin uses the correct address from `DF_PLUGIN_SERVER_ADDRESS`

## Advanced: Custom Socket Path

You can use any socket path:

```yaml
# Project-specific socket
server_addr: "unix:///var/run/dragonfly/plugin.sock"

# User-specific socket
server_addr: "unix://~/.dragonfly/plugin.sock"
```

Make sure the directory exists and has proper permissions.

