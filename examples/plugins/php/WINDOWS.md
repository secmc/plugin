# Windows Setup Guide

Running PHP plugins on Windows with Dragonfly.

## Quick Setup

```bash
# In Git Bash, PowerShell, or WSL
cd examples/plugins/php
./setup.sh
```

The script will:
1. Download PHP 8.4 for Windows (x64)
2. Extract to `bin/php/` directory
3. Install Composer dependencies

## Important: Update plugins.yaml

Windows uses a different directory structure. After running `./setup.sh`, edit `plugins/plugins.yaml`:

```yaml
server_addr: "tcp://127.0.0.1:50050"

plugins:
  - id: example-php
    name: Example PHP Plugin
    command: "examples/plugins/php/bin/php/php.exe"  # ← Change this line
    args: ["examples/plugins/php/src/HelloPlugin.php"]
```

## Running on Windows

### Option 1: Git Bash (Recommended)

Git Bash provides a Unix-like environment:

```bash
cd examples/plugins/php
./setup.sh
./verify-setup.sh
```

### Option 2: PowerShell

```powershell
cd examples\plugins\php

# Run setup
bash setup.sh

# Or manually download and extract
Invoke-WebRequest -Uri "https://github.com/NetherGamesMC/php-build-scripts/releases/download/pm5-php-8.4-latest/PHP-8.4-Windows-x64-PM5.zip" -OutFile php-build.zip
Expand-Archive -Path php-build.zip -DestinationPath . -Force
Remove-Item php-build.zip

# Install dependencies
bin\php\php.exe composer.phar install
```

### Option 3: WSL (Windows Subsystem for Linux)

If you have WSL installed, you can use the Linux setup:

```bash
cd examples/plugins/php
./setup.sh  # Downloads Linux version
```

## Verify Setup

```bash
./verify-setup.sh
```

Should show:
```
🖥️  Platform: Windows
✓ Check 1: PHP Binary
  ✅ Found: bin/php/php.exe
  📦 PHP 8.4.x ...
✓ Check 2: gRPC Extension
  ✅ gRPC extension loaded
```

## Starting Dragonfly

Once configured, start Dragonfly normally:

```bash
cd ../../..
go run main.go
```

Dragonfly will automatically start the PHP plugin!

## Troubleshooting

### "php.exe not found"

Make sure you updated `plugins/plugins.yaml` with the Windows path:
```yaml
command: "examples/plugins/php/bin/php/php.exe"
```

### "DLL not found" errors

The PHP build includes all necessary DLLs in `bin/php/`. Make sure you're running from the correct directory.

### Line ending issues

If you get errors like `'\r': command not found`, convert line endings:

```bash
# In Git Bash
dos2unix setup.sh
dos2unix run-plugin.sh
dos2unix verify-setup.sh
```

Or configure Git to handle line endings:
```bash
git config core.autocrlf true
```

## Testing

In Minecraft, try:
- `/cheers`
- `!cheer Hello from Windows!`

You should see responses from the PHP plugin running on Windows!

## Notes

- The Windows build includes **grpc**, **protobuf**, and other extensions pre-compiled
- No need to install PECL or compile extensions
- The build is from [NetherGamesMC/php-build-scripts](https://github.com/NetherGamesMC/php-build-scripts)
- If you need a different PHP version, check their releases page

## Next Steps

- Edit `src/HelloPlugin.php` to customize your plugin
- See [README.md](README.md) for the full plugin API
- Check [../typescript/src/index.ts](../typescript/src/index.ts) for more examples

