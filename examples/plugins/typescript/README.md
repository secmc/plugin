# TypeScript Dragonfly Plugin Example

This example demonstrates a Dragonfly plugin written in **TypeScript** with full type safety and excellent IDE support.

## Features

- ✅ **Full Type Safety** - TypeScript interfaces for all protocol messages
- ✅ **Modern TypeScript** - ES2022 modules, strict mode
- ✅ **Developer Experience** - Hot reload with `tsx`

## Quick Start

```bash
# Install dependencies (includes @bufbuild/protobuf for proto runtime)
npm install
```

## TypeScript Types

This plugin imports types from `@dragonfly/proto`:

```typescript
import { HostToPlugin, PluginToHost, GameMode } from '@dragonfly/proto';
```

**Current:** Path alias pointing to `../../../proto/generated/ts` (local development)  
**Future:** Published npm package - no changes needed, just `npm install @dragonfly/proto`

Types are generated using [buf](https://buf.build) with `ts-proto`. See `proto/buf.gen.ts.yaml` for configuration.

## Plugin Capabilities

This example plugin demonstrates:

1. **Player Join** - Sends welcome message with color codes
2. **Commands**:
   - `/greet` - Send a colorful greeting
   - `/tp` - Teleport to spawn (0, 100, 0)
3. **Chat Filtering** - Blocks messages with bad words
4. **Chat Mutations**:
   - `!shout <message>` - Makes text uppercase
   - `!rainbow <message>` - Applies rainbow colors
5. **Block Break** - Example of modifying drops and XP

## Environment Variables

- `DF_PLUGIN_ID` - Plugin identifier (default: `typescript-plugin`)
- `DF_PLUGIN_GRPC_ADDRESS` - gRPC server address (default: `127.0.0.1:50052`)
