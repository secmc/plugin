# Using Custom PHP Installation with Built-in gRPC

This guide covers using a pre-compiled PHP 8 build with gRPC built-in.

## Automatic Download (Easiest)

The setup script can automatically download PHP 8.4 with gRPC:

```bash
cd examples/plugins/php
./setup.sh

# Then start Dragonfly - it auto-starts the plugin!
cd ../../..
go run main.go
```

This downloads from [NetherGamesMC/php-build-scripts](https://github.com/NetherGamesMC/php-build-scripts) and sets everything up automatically.

**Supported:** Linux (x64/ARM64), macOS (x64/ARM64), Windows (x64)

**Windows users:** The PHP path is different. Update `plugins/plugins.yaml`:
```yaml
command: "examples/plugins/php/bin/php/php.exe"  # Windows uses bin/php/php.exe
```

**Note:** You don't need to manually run the plugin - Dragonfly starts it automatically based on `plugins/plugins.yaml`.

---

## Manual Setup with Your Own PHP Build

If you already have a PHP build in `bin/php7/` or want to use a custom one:

### 1. Install PHP Dependencies (using your custom PHP)

```bash
cd examples/plugins/php

# Use your custom PHP's composer
bin/php7/bin/php $(which composer) install

# Or if composer isn't installed, download it:
bin/php7/bin/php -r "copy('https://getcomposer.org/installer', 'composer-setup.php');"
bin/php7/bin/php composer-setup.php
bin/php7/bin/php composer.phar install
```

### 2. Verify gRPC Extension

```bash
bin/php7/bin/php -m | grep grpc
```

You should see `grpc` in the output.

### 3. Run the Plugin

```bash
# Option A: Use the wrapper script
./run-plugin.sh

# Option B: Run directly
bin/php7/bin/php src/HelloPlugin.php
```

### 4. Configure in plugins.yaml

Update `/plugins/plugins.yaml` to use your custom PHP:

```yaml
server_addr: "tcp://127.0.0.1:50050"

plugins:
  - id: example-php
    name: Example PHP Plugin
    command: "examples/plugins/php/bin/php7/bin/php"
    args: ["examples/plugins/php/src/HelloPlugin.php"]
    env:
      PHP_ENV: production
```

Or use the wrapper script:

```yaml
server_addr: "tcp://127.0.0.1:50050"

plugins:
  - id: example-php
    name: Example PHP Plugin
    command: "examples/plugins/php/run-plugin.sh"
    args: []
    work_dir: "examples/plugins/php"
```

## What's Included

Your custom PHP installation includes:

- **bin/** - PHP executables (php, php-config, phpize, etc.)
- **include/** - Header files for PHP extensions
- **lib/** - Shared libraries including gRPC
- **sbin/** - System binaries (if any)
- **share/** - Shared data and man pages

## Checking Available Extensions

```bash
# List all loaded extensions
bin/php7/bin/php -m

# Check specific extension
bin/php7/bin/php -r "echo extension_loaded('grpc') ? 'gRPC: YES' : 'gRPC: NO';"
bin/php7/bin/php -r "echo extension_loaded('protobuf') ? 'Protobuf: YES' : 'Protobuf: NO';"
```

## No Installation Needed!

Since gRPC is already compiled in, you **skip**:
- ❌ `pecl install grpc`
- ❌ `pecl install protobuf`
- ❌ Editing php.ini
- ❌ Building from source

Just install Composer dependencies and run! 🎉

## Troubleshooting

### Library not found errors

If you get errors about missing shared libraries:

```bash
# Check what libraries PHP needs
otool -L bin/php7/bin/php  # macOS
ldd bin/php7/bin/php       # Linux

# You may need to set library path
export DYLD_LIBRARY_PATH="$(pwd)/bin/php7/lib:$DYLD_LIBRARY_PATH"  # macOS
export LD_LIBRARY_PATH="$(pwd)/bin/php7/lib:$LD_LIBRARY_PATH"      # Linux
```

### PHP version mismatch

The folder is named `php7` but contains PHP 8. This is fine - just a naming thing:

```bash
bin/php7/bin/php -v
# Should show PHP 8.x.x
```

## Testing

Once running, test with:

```bash
# In Minecraft, run:
/cheers

# In chat, type:
!cheer Hello World

# The plugin should respond!
```

## Next Steps

- Edit `src/HelloPlugin.php` to add your custom logic
- See [README.md](README.md) for the full plugin API
- Check [../typescript/src/index.ts](../typescript/src/index.ts) for more examples

