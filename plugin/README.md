# Plugin System

Protobuf-based plugin system for Dragonfly server.

## Quick Start

```bash
# After editing proto/types/plugin.proto
cd proto/
npm install    # First time only
npm run generate  # Generate Go code
```

## Structure

```
proto/
├── types/         # .proto definitions
├── generated/     # Generated .pb.go (gitignored)
├── buf.yaml       # Buf config
└── package.json   # Build scripts
```

## Adding Events

Edit `proto/types/plugin.proto`, then run `npm run generate`.

