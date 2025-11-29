# Dragonfly PHP Plugin Example

Example PHP plugin for Dragonfly using gRPC and Protocol Buffers.

## 🚀 Quick Start (Recommended)

**Automated setup** - downloads PHP 8.4 with gRPC built-in from [NetherGamesMC](https://github.com/NetherGamesMC/php-build-scripts):

```bash
cd examples/plugins/php

# One command setup! Downloads PHP 8.4 + gRPC, installs dependencies
./setup.sh

# Run the plugin
./run-plugin.sh
```

**That's it!** The setup script:
- ✅ Auto-detects your OS (Linux/macOS/Windows) and architecture
- ✅ Downloads pre-compiled PHP 8.4 with gRPC extension (~50MB)
- ✅ Installs Composer dependencies
- ✅ Verifies everything works

**Windows users:** After running `./setup.sh`, update `plugins/plugins.yaml`:
```yaml
command: "examples/plugins/php/bin/php/php.exe"  # Windows path
```

The plugin is **already configured** in `plugins/plugins.yaml`.

**Documentation:**
- 🪟 [WINDOWS.md](WINDOWS.md) - Complete Windows setup guide
- 📦 [CUSTOM_PHP.md](CUSTOM_PHP.md) - Using your own PHP build
- ⚡ [GET_STARTED.md](GET_STARTED.md) - 2-minute quick start

---

## Prerequisites (For Manual Setup)

- **PHP 8.1+** with the following extensions:
  - `grpc` (required)
  - `protobuf` (optional but recommended for performance)
- **Composer** for dependency management
- **protoc** (Protocol Buffer compiler)
- **grpc_php_plugin** for gRPC code generation

## Installation

### 1. Install PHP and Extensions

#### macOS (with Homebrew)
```bash
brew install php
pecl install grpc
pecl install protobuf
```

Add to your `php.ini`:
```ini
extension=grpc.so
extension=protobuf.so
```

#### Ubuntu/Debian
```bash
sudo apt install php php-dev php-pear
sudo pecl install grpc
sudo pecl install protobuf
```

### 2. Install protobuf compiler
```bash
# macOS
brew install protobuf

# Ubuntu/Debian
sudo apt install protobuf-compiler
```

### 3. Install gRPC PHP plugin
```bash
# Clone gRPC repository
git clone -b v1.57.0 https://github.com/grpc/grpc
cd grpc

# Build the PHP plugin
mkdir -p cmake/build
cd cmake/build
cmake ../..
make grpc_php_plugin

# Copy to system path
sudo cp grpc_php_plugin /usr/local/bin/
```

### 4. Run Setup Script
```bash
chmod +x setup.sh generate-proto.sh
./setup.sh
```

### 5. Generate Protobuf Files
```bash
./generate-proto.sh
```


## Configuration

The PHP plugin is already configured in `plugins/plugins.yaml`:

```yaml
server_addr: "tcp://127.0.0.1:50050"

plugins:
  - id: example-php
    name: Example PHP Plugin
    command: "examples/plugins/php/bin/php7/bin/php"
    args: ["examples/plugins/php/src/HelloPlugin.php"]
```

## Running the Plugin

### With Dragonfly (Automatic - Recommended)

The plugin is already configured in `plugins/plugins.yaml`. Just start Dragonfly:

```bash
cd ../../..  # Back to project root
go run main.go
```

Dragonfly will automatically launch the PHP plugin process!

### Standalone Testing (Optional)

To test the plugin without Dragonfly:

```bash
# Using the wrapper script
./run-plugin.sh

# Or directly
bin/php7/bin/php src/HelloPlugin.php
```

### With Docker
```dockerfile
FROM php:8.2-cli

RUN apt-get update && apt-get install -y \
    git \
    unzip \
    libzip-dev \
    && docker-php-ext-install zip

RUN pecl install grpc protobuf \
    && docker-php-ext-enable grpc protobuf

COPY --from=composer:latest /usr/bin/composer /usr/bin/composer

WORKDIR /app
COPY . .

RUN composer install

CMD ["php", "src/HelloPlugin.php"]
```

## Features

The example plugin demonstrates:

- ✅ Bidirectional gRPC streaming
- ✅ Event subscription (PLAYER_JOIN, COMMAND, CHAT)
- ✅ Command registration (`/cheers`)
- ✅ Chat event filtering and mutation
- ✅ Action dispatch (sending messages)
- ✅ Event cancellation

## Troubleshooting

### "grpc extension not found"
Install the gRPC PHP extension:
```bash
pecl install grpc
```

### "protoc: command not found"
Install protobuf compiler:
```bash
# macOS
brew install protobuf

# Linux
sudo apt install protobuf-compiler
```

### "grpc_php_plugin not found"
Either build it from source (see installation steps) or adjust the path in `generate-proto.sh`

### Alternative: Use protobuf-php/protobuf
If you can't build `grpc_php_plugin`, you can use the pure PHP alternative:
```bash
composer require protobuf-php/protobuf
```

## Development

Edit `src/HelloPlugin.php` to add your custom plugin logic. The plugin uses the same protobuf messages as the TypeScript and Node.js examples, so you can reference those for additional examples.

## Learn More

- [gRPC PHP Documentation](https://grpc.io/docs/languages/php/)
- [Protocol Buffers PHP](https://developers.google.com/protocol-buffers/docs/reference/php-generated)
- [Dragonfly Plugin Architecture](../../../docs/plugin-architecture.md)

