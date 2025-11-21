# Repository Guidelines

## Project Structure & Module Organization
Host runtime code lives in `cmd/`, which houses the entrypoint (`main.go`), player/world utilities, and the runtime config at `cmd/plugins/plugins.yaml`. Shared plugin lifecycle helpers live in `plugin/`. Protocol Buffers are defined in `proto/types`, generated artifacts land in `proto/generated`, and Buf config files sit beside them. Example clients reside under `examples/plugins/`, architecture docs in `docs/`, and automation (including `scripts/post_generation.sh`) in `scripts/`.

## Build, Test, and Development Commands
- `make run` – runs the Dragonfly host using `cmd/plugins/plugins.yaml`.
- `make proto` – executes `buf generate` plus the post-generation script for Go/JS stubs.
- `go test ./...` – runs every Go test; keep suites beside the code.
- `npm run dev --prefix examples/plugins/typescript` – hot reloads the TypeScript sample.
- `examples/plugins/php/bin/php7/bin/php examples/plugins/php/src/HelloPlugin.php` – quick PHP sample run.

## Coding Style & Naming Conventions
Go code must pass `gofmt`, keep package names lowercase, exports in PascalCase, and locals camelCase. Proto filenames stay snake_case, messages PascalCase, and enums SCREAMING_SNAKE. Generated TypeScript already applies ESLint/Prettier—mirror those settings and name async handlers after their event (`handlePlayerJoin`). Keep plugin IDs and directories kebab-case (e.g., `example-typescript`) and consistent with `cmd/plugins/plugins.yaml`.

## Testing Guidelines
Add `*_test.go` files per package with `Test<Name>` functions covering event decoding, handshake timeouts, and command routing. Run integration flows from each language example directory using the commands above and capture logs. Exercise every event touched by the change and note any intentionally skipped area (resource packs, external APIs) inside the PR.

## Commit & Pull Request Guidelines
Commits follow the terse, imperative style already in history (`feat: add makefile (#18)`). Group related work before review or rely on PR-level squash. Each PR should explain the motivation, list validation commands (`make run`, `go test ./...`, `npm run dev`, etc.), and attach logs/screens for gameplay-visible edits. Request reviewers familiar with the affected directories and verify regenerated files, configs, and docs are committed.

## Security & Configuration Tips
Keep the plugin server bound to `unix:///tmp/dragonfly_plugin.sock` or loopback TCP unless a remote plugin truly needs exposure. Inject API keys or RPC tokens through the plugin `env` block at runtime; never commit them. Vet third-party plugins before adding them to `required_plugins`, since startup stalls until each required plugin handshakes successfully.

## Dragonfly Dependency Context
This repository ships a plugin library meant to be embedded into a stock [Dragonfly](https://github.com/df-mc/dragonfly) server (see `cmd/main.go`). Many tasks need the upstream server sources for entity/world types, so the `dragonfly/` directory stays gitignored and must be recloned on demand. In any clean workspace—including Codex Web sessions—run `git clone https://github.com/df-mc/dragonfly` from the repo root before hacking on features that mirror Dragonfly behavior.
