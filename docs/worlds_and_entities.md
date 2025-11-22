# Worlds and Entities Implementation Notes

This document captures the world- and entity-related behaviours the plugin host mirrors from Dragonfly and that external plugins need to support.

## World identity and lookup
- World references sent to plugins include both the configured world name and its dimension string (lower-cased) to disambiguate lookups across dimensions. The helper populates `WorldRef` with `Name` and `Dimension` derived from the `world.World` instance. 【F:plugin/adapters/plugin/event_helpers.go†L164-L173】
- Incoming `WorldRef` values from plugins are resolved by name first, then by dimension string if the name is empty or unknown. Registered worlds are stored in a lower-cased map keyed by name; dimension matching also lower-cases. 【F:plugin/adapters/plugin/manager.go†L378-L427】
- Worlds are registered when the manager attaches to them and unregistered after a `WORLD_CLOSE` event is emitted so stale references are not reused. 【F:plugin/adapters/plugin/world_events.go†L192-L203】

## World configuration and range management
- Dragonfly exposes runtime setters such as `SetDefaultGameMode`, `SetDifficulty`, and `SetTickRange` to adjust core world behaviour. The plugin bridge needs to surface these so plugins can match the server’s world state and react when these settings change.
- World `Range` queries control which chunks are loaded or targeted for tick updates. Any plugin-facing API that performs block mutations or entity searches should obey the configured range.
- World-level effects such as `PlaySound`, `AddParticle`, and `SetBlock` require positional parameters relative to the world’s chunk range and dimension to avoid inconsistencies when multiple worlds are present.

## World event surface area
- The plugin world handler hooks into Dragonfly’s `world.Handler` and forwards each server callback into the event stream (`WORLD_*` event types). Implementations must wire the handler for close, liquid flow/decay/hardening, sound playback, fire spread, block burn, crop trample, leaves decay, entity spawn/despawn, and explosions. 【F:plugin/adapters/handlers/world.go†L12-L67】
- Each event payload contains a `WorldRef` plus context-specific data:
  - Liquid flow, decay, and harden events include source/target block positions and liquid or block states. 【F:plugin/adapters/plugin/world_events.go†L13-L55】
  - World sounds include the emitted sound type (stringified Go type) and 3D position. 【F:plugin/adapters/plugin/world_events.go†L57-L68】
  - Fire spread, block burn, crop trample, and leaves decay send the relevant block positions. 【F:plugin/adapters/plugin/world_events.go†L70-L117】
  - Entity spawn/despawn events bundle the entity reference alongside the world reference. 【F:plugin/adapters/plugin/world_events.go†L119-L141】
  - Explosions surface the epicenter, affected entities/blocks, and the calculated item-drop chance and spawn-fire flags, which plugins may mutate. 【F:plugin/adapters/plugin/world_events.go†L143-L190】【F:proto/types/mutations.proto†L9-L100】
  - World close broadcasts the world reference and immediately unregisters the world so further lookups fail until it is reattached. 【F:plugin/adapters/plugin/world_events.go†L192-L203】

## Entity representation
- Entities are serialized to `EntityRef` with their Go type name, UUID (when available from the entity handle), position, and rotation. This data allows plugins to recognize entities and correlate later mutations (such as explosion filtering). 【F:plugin/adapters/plugin/event_helpers.go†L199-L226】【F:proto/types/common.proto†L109-L120】
- Entity refs appear in spawn/despawn events and as the optional `affected_entities` list in explosions for selective mutation by plugins. Filtering uses plugin-supplied UUID lists to drop entities from the explosion impact set. 【F:plugin/adapters/plugin/world_events.go†L119-L190】【F:proto/types/world_events.proto†L59-L76】

## Entity lifecycle and querying
- World transactions expose `AddEntity`/`RemoveEntity` to add or detach entities; removal invalidates the handle returned by spawn to mirror Dragonfly’s semantics. Plugins should receive enough context to track these lifecycle changes and avoid reusing stale references.
- Entity iteration helpers (`Entities`, `Players`, `EntitiesWithin`) must be mirrored so plugins can enumerate everything in a world or filter by bounding box. Bounding boxes follow Dragonfly’s `cube.BBox` conventions and should consider the world’s tick range when evaluating membership.
- Viewer lookups (`Viewers`) are used to target updates (sounds, particles, block updates) to interested clients. Even if the viewer interface itself is not directly exposed, the plugin layer should provide enough information to deliver per-viewer or per-position updates consistently.

## World and block state serialization
- Block positions are converted into `BlockPos` tuples; block and liquid states encode the block name and property map or liquid depth/falling flags and type. These representations show up across world events (liquid changes, explosion block lists). 【F:plugin/adapters/plugin/event_helpers.go†L53-L92】【F:proto/types/common.proto†L85-L120】
- Liquid/block conversions accept both liquids and non-liquid blocks when populating liquid hardening events, ensuring plugins see the before/after composition. 【F:plugin/adapters/plugin/world_events.go†L42-L55】【F:plugin/adapters/plugin/event_helpers.go†L72-L92】

## World explosion mutation semantics
- Explosion events are cancellable and support mutations to the affected entity list, block positions, item drop chance, and whether fire is spawned. Mutation handlers apply plugin-specified UUID filters and block lists back into Dragonfly structures, or clear them when plugins supply empty/invalid data. 【F:plugin/adapters/plugin/world_events.go†L143-L190】
- The protocol exposes the mutable fields through `WorldExplosionMutation` so plugins can remove specific entities/blocks or adjust drop probability and fire. Ensure plugin responses are processed in order so multiple plugins can compose mutations. 【F:proto/types/mutations.proto†L9-L100】

## Required proto coverage
- The `world_events.proto` schema lists every world-facing event type that must be mirrored over gRPC. Implementations should verify outgoing events populate the fields defined there and that incoming mutations respect optionality. 【F:proto/types/world_events.proto†L9-L80】
- `common.proto` defines the reusable structures (world refs, entity refs, block/liquid state, vectors) that need consistent encoding/decoding across host and plugin runtimes. 【F:proto/types/common.proto†L85-L120】

## Structures and world composition
- Dragonfly supports loading and placing structures at runtime via `world/structure.go`. Plugins that author or paste structures need hooks to schedule placement, manage rotation/mirroring, and resolve collisions with existing blocks or entities within the world’s configured range and dimension.
