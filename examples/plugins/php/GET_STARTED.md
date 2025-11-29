# Get Started with PHP Plugins - 2 Minutes! ⚡

The **easiest** way to run PHP plugins on Dragonfly.

## 🚀 Quick Setup

```bash
cd examples/plugins/php
./setup.sh
```

**Done!** This single command:
- ✅ Downloads PHP 8.4 with gRPC (~50MB)
- ✅ Installs Composer dependencies
- ✅ Works on Linux, macOS & Windows

**Windows users:** After setup, update this line in `plugins/plugins.yaml`:
```yaml
command: "examples/plugins/php/bin/php/php.exe"
```

## 🎮 Run It

```bash
# Start Dragonfly - it automatically starts the plugin!
cd ../../..
go run main.go
```

The plugin is configured in `plugins/plugins.yaml`, so Dragonfly will automatically launch it.

**Optional:** Test the plugin standalone (without Dragonfly):
```bash
cd examples/plugins/php
./run-plugin.sh
```

## 🧪 Test It

Join your Minecraft server and try:

```
/cheers
```

Or in chat:

```
!cheer Hello World
```

You should see responses from the PHP plugin! 🎉

## 📚 Learn More

- **[QUICKSTART.md](QUICKSTART.md)** - All setup options
- **[README.md](README.md)** - Complete documentation
- **[CUSTOM_PHP.md](CUSTOM_PHP.md)** - Using your own PHP build
- **[HelloPlugin.php](src/HelloPlugin.php)** - Plugin source code

## 🔧 Troubleshooting

### Verify setup

```bash
./verify-setup.sh
```

### Plugin not connecting?

Check that:
1. Dragonfly's plugin server (`server_addr` in `plugins.yaml`) is running
2. PHP binary exists: `ls bin/php7/bin/php`
3. Dragonfly config has the plugin enabled in `plugins/plugins.yaml`

### Need help?

- Check [README.md](README.md) troubleshooting section
- Review TypeScript example at `../typescript/src/index.ts`
- Check plugin proto at `../../plugin/proto/types/plugin.proto`

## 🎯 What's Included

The example PHP plugin demonstrates:

- ✅ **Commands** - `/cheers` command
- ✅ **Chat Events** - Listen and filter chat messages
- ✅ **Chat Mutations** - Transform `!cheer` messages
- ✅ **Event Cancellation** - Block messages with "spoiler"
- ✅ **Player Events** - Join/quit notifications
- ✅ **Actions** - Send messages to players

## 📝 Next Steps

Edit `src/HelloPlugin.php` to add your own logic!

See the [TypeScript example](../typescript/src/index.ts) for more advanced features:
- Teleportation
- Game mode changes
- Block break events
- Command arguments parsing
- Multiple action batching

Happy coding! 🚀

