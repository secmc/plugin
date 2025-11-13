#!/bin/bash
set -e

echo "========================================="
echo "Dragonfly PHP Plugin Setup"
echo "========================================="
echo ""

# Detect OS and architecture
OS=$(uname -s 2>/dev/null || echo "Unknown")
ARCH=$(uname -m 2>/dev/null || echo "Unknown")

# Check if running in Git Bash or WSL on Windows
if [[ "$OS" == MINGW* ]] || [[ "$OS" == MSYS* ]] || [[ "$OS" == CYGWIN* ]]; then
    OS="Windows"
fi

# PHP build URL (PocketMine PHP 8.3 with gRPC built-in)
# Source: secmc/PHP-Binaries
if [ "$OS" = "Darwin" ]; then
    # macOS - both arm64 and x86_64 supported!
    if [ "$ARCH" = "arm64" ]; then
        PHP_BUILD_URL="https://github.com/secmc/PHP-Binaries/releases/download/pm5-php-8.3-latest/PHP-8.3-MacOS-arm64-PM5.tar.gz"
        PHP_BIN="bin/php7/bin/php"
        IS_WINDOWS=false
    else
        # x86_64 or other arch - use x86_64 build
        PHP_BUILD_URL="https://github.com/secmc/PHP-Binaries/releases/download/pm5-php-8.3-latest/PHP-8.3-MacOS-x86_64-PM5.tar.gz"
        PHP_BIN="bin/php7/bin/php"
        IS_WINDOWS=false
    fi
elif [ "$OS" = "Linux" ]; then
    if [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then
        echo "❌ Pre-built PHP binaries are not available for Linux ARM64"
        echo ""
        echo "Please use one of these alternatives:"
        echo ""
        echo "  Option 1: Build PHP with gRPC manually"
        echo "    - See CUSTOM_PHP.md for build instructions"
        echo ""
        echo "  Option 2: Use Docker"
        echo "    - See CUSTOM_PHP.md for Docker setup instructions"
        echo ""
        echo "  Option 3: Use system PHP with manual gRPC installation"
        echo "    - Install PHP 8.1+ and compile gRPC extension"
        echo ""
        exit 1
    else
        # Linux x86_64 - supported!
        PHP_BUILD_URL="https://github.com/secmc/PHP-Binaries/releases/download/pm5-php-8.3-latest/PHP-8.3-Linux-x86_64-PM5.tar.gz"
        PHP_BIN="bin/php7/bin/php"
        IS_WINDOWS=false
    fi
elif [ "$OS" = "Windows" ]; then
    # Windows x64 - supported!
    PHP_BUILD_URL="https://github.com/secmc/PHP-Binaries/releases/download/pm5-php-8.3-latest/PHP-8.3-Windows-x64-PM5.zip"
    PHP_BIN="bin/php/php.exe"
    IS_WINDOWS=true
else
    echo "❌ Unsupported OS: $OS"
    echo "   Pre-built binaries available for: macOS (arm64/x86_64), Linux x86_64, Windows x64"
    echo "   See CUSTOM_PHP.md for manual installation on other platforms"
    exit 1
fi

echo "🖥️  Detected: $OS $ARCH"

# Check if PHP binary already exists
if [ -f "$PHP_BIN" ]; then
    echo "✅ Custom PHP already installed"
    PHP_VERSION=$($PHP_BIN -v | head -n 1)
    echo "   $PHP_VERSION"
else
    echo "📥 Downloading pre-compiled PHP 8.3 with gRPC..."
    echo "   Source: secmc/PHP-Binaries"
    echo ""
    
    # Download PHP build
    if [ "$IS_WINDOWS" = true ]; then
        curl -L -o php-build.zip "$PHP_BUILD_URL"
        
        echo ""
        echo "📦 Extracting PHP build..."
        
        # Try to use unzip (Git Bash, WSL, etc.)
        if command -v unzip &> /dev/null; then
            unzip -q php-build.zip
        elif command -v powershell.exe &> /dev/null; then
            powershell.exe -Command "Expand-Archive -Path php-build.zip -DestinationPath . -Force"
        else
            echo "❌ No unzip tool found. Please install unzip or use WSL."
            exit 1
        fi
        rm php-build.zip
    else
        curl -L -o php-build.tar.gz "$PHP_BUILD_URL"
        
        echo ""
        echo "📦 Extracting PHP build..."
        tar -xzf php-build.tar.gz
        rm php-build.tar.gz
    fi
    
    if [ -f "$PHP_BIN" ]; then
        echo "✅ PHP extracted successfully"
        PHP_VERSION=$($PHP_BIN -v | head -n 1)
        echo "   $PHP_VERSION"
    else
        echo "❌ Failed to extract PHP binary"
        echo "   Expected: $PHP_BIN"
        exit 1
    fi
fi

echo ""

# Verify gRPC extension
echo "🔍 Verifying gRPC extension..."
$PHP_BIN -m | grep -q grpc
if [ $? -eq 0 ]; then
    echo "✅ gRPC extension found"
else
    echo "❌ gRPC extension not found in PHP build"
    exit 1
fi

# Check for Composer
echo ""
if command -v composer &> /dev/null; then
    COMPOSER_CMD="composer"
    echo "✅ System Composer found"
else
    # Download composer if not installed
    if [ ! -f "composer.phar" ]; then
        echo "📥 Downloading Composer..."
        $PHP_BIN -r "copy('https://getcomposer.org/installer', 'composer-setup.php');"
        $PHP_BIN composer-setup.php --quiet
        rm composer-setup.php
    fi
    COMPOSER_CMD="$PHP_BIN composer.phar"
    echo "✅ Local Composer ready"
fi

# Install PHP dependencies
echo ""
echo "📦 Installing PHP dependencies..."
$COMPOSER_CMD install

echo ""
echo "========================================="
echo "✅ Setup Complete!"
echo "========================================="
echo ""
echo "Your PHP plugin is ready to use!"
echo ""
echo "📁 PHP Installation: $PHP_BIN"
echo "📦 Extensions: gRPC ✓"
echo ""
echo "Next steps:"
echo ""
echo "  1️⃣  (Optional) Verify setup:"
echo "      ./verify-setup.sh"
echo ""
echo "  2️⃣  Start Dragonfly - it auto-starts the plugin!"
echo "      cd ../../.."
echo "      go run main.go"
echo ""
echo "  💡 The plugin is configured in plugins/plugins.yaml"
echo "     Dragonfly will automatically launch it!"
echo ""
echo "  🎮 In-game test commands:"
echo "      /cheers"
echo "      !cheer Hello"
echo ""

