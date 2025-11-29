# PHP Plugin Quick Start Guide

Three ways to run PHP plugins with Dragonfly.

## Option 1: Automated Setup (Easiest! ⭐)

**Just run the setup script!** It auto-downloads PHP 8.4 with gRPC built-in:

```bash
cd examples/plugins/php

# One command - downloads PHP, installs dependencies
./setup.sh

# Start Dragonfly - it automatically starts the plugin!
cd ../../..
go run main.go
```

The plugin is configured in `plugins/plugins.yaml`, so Dragonfly will automatically launch it when the server starts.

**Supported platforms:**
- macOS (Intel x64 / Apple Silicon ARM64)
- Linux (x64 / ARM64)
- Windows (x64)

**Windows users:** After setup, edit `plugins/plugins.yaml` and change:
```yaml
command: "examples/plugins/php/bin/php/php.exe"  # Change to Windows path
```

That's it! Skip to [Testing](#testing) below.

---

## Option 2: Using Your Own PHP Build

If you have your own pre-compiled PHP build (`.tar.gz`) with gRPC:

```bash
cd examples/plugins/php

# Extract your PHP build
tar -xzf /path/to/PHP-8.4-Linux-x86_64-PM5.tar.gz

# This should create: bin/php7/bin/php

# Install dependencies
bin/php7/bin/php $(which composer) install

# Run it
./run-plugin.sh
```

The plugin is already configured in `plugins/plugins.yaml`.

---

## Testing

Once Dragonfly is running you should see logs like:

```
[plugin-manager] plugin server listening on 127.0.0.1:50050
[plugin-manager] plugin connected: example-php
```

## Testing Commands

Join the Minecraft server and try:

- `/cheers` - Get a greeting from the PHP plugin
- Type `!cheer Hello` in chat - Message will be transformed with 🥂 emoji
- Type a message with "spoiler" - Message will be blocked

## Troubleshooting

### PHP binary not found
Run the verify script to check your setup:
```bash
./verify-setup.sh
```

Or manually check:
```bash
ls bin/php7/bin/php
```

### Plugin cannot connect
- Ensure Dragonfly is running and `server_addr` in `plugins.yaml` matches `DF_PLUGIN_SERVER_ADDRESS`
- Verify `bin/php7/bin/php` exists and has the gRPC extension enabled

### PHP version issues
Check your PHP version:
```bash
bin/php7/bin/php -v
```

Should show PHP 8.4.

## Next Steps

- Read [README.md](README.md) for detailed PHP setup
- Check [plugin.proto](../../../plugin/proto/types/plugin.proto) for available events and actions
- Look at [index.ts](../typescript/src/index.ts) for more complex examples

