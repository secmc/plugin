# Plugin System

Protobuf-based plugin system for Dragonfly server using ports/adapter architecture.

## Usage

```go
mgr := plugin.NewManager(srv, logger)
if err := mgr.Start(""); err != nil {
    log.Fatal(err)
}
defer mgr.Close()

mgr.AttachWorld(srv.World())
mgr.AttachPlayer(player)
```

## Configuration

Create `plugins/plugins.yaml`:

```yaml
plugins:
  - id: my-plugin
    name: My Plugin
    command: node
    args: [dist/index.js]
    work_dir: ./plugins/my-plugin
    address: 127.0.0.1:0
```

## Protobuf

```bash
cd proto/
npm install
npm run generate
```

## Events

- PLAYER_JOIN / PLAYER_QUIT
- CHAT
- COMMAND
- BLOCK_BREAK
- WORLD_CLOSE

## Actions

- SendChat
- Teleport
- Kick
- SetGameMode