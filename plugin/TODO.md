# Plugin System Performance TODO

## Current Architecture (Nov 2025)

**✅ Ports/Adapter Architecture Complete**
- Interfaces defined in `ports/` package
- Implementations in `adapters/` package
- Clean separation between domain and infrastructure
- Manager implements PluginManager and EventEmitter interfaces

**✅ Switched to Generated Protobuf Code**
- Using `protoc-gen-go` generated code instead of hand-written codec
- Reduces maintenance burden (~950 lines removed)
- Trade-off: Slightly slower but much easier to maintain and extend
- Can optimize specific hot paths later if profiling shows bottlenecks

## Critical (Must Do Before Production)

### 1. Add Player Movement Event
- [ ] Add `PlayerMoveEvent` message to `plugin.proto`
  ```protobuf
  message PlayerMoveEvent {
    string player_uuid = 1;
    double x = 2;
    double y = 3;
    double z = 4;
    float yaw = 5;
    float pitch = 6;
    bool on_ground = 7;
  }
  ```
- [ ] Add to `EventEnvelope.payload` oneof (field 16)
- [ ] Regenerate with `protoc --go_out=. --go_opt=paths=source_relative plugin/proto/plugin.proto`
- [ ] Wire up in player movement handler

### 2. Implement Event Batching
- [ ] Add `EventBatch` message to `plugin.proto`
  ```protobuf
  message EventBatch {
    repeated EventEnvelope events = 1;
  }
  ```
- [ ] Add to `HostToPlugin.payload` oneof (field 21)
- [ ] Regenerate protobuf code
- [ ] Update event dispatcher to batch high-frequency events
- [ ] Send batches once per tick instead of individual events

### 3. Buffer Pooling (Adapted for Generated Code)
- [ ] Create `sync.Pool` for marshal buffers in `manager.go`
- [ ] Typical size: 256-512 bytes for common events, 4KB for batches
- [ ] Pattern:
  ```go
  var msgBufPool = sync.Pool{
      New: func() any {
          b := make([]byte, 0, 512)
          return &b
      },
  }
  
  // In send path:
  bufPtr := msgBufPool.Get().(*[]byte)
  *bufPtr = proto.MarshalOptions{}.MarshalAppend((*bufPtr)[:0], msg)
  stream.Send(*bufPtr)
  msgBufPool.Put(bufPtr)
  ```

### 4. Message Reuse Pool
- [ ] Pool message structs themselves (not just buffers)
- [ ] Reduce GC pressure from event creation
- [ ] Use `proto.Reset()` to clear messages before returning to pool
- [ ] Pattern:
  ```go
  var playerMovePool = sync.Pool{
      New: func() any { return &pb.PlayerMoveEvent{} },
  }
  
  evt := playerMovePool.Get().(*pb.PlayerMoveEvent)
  evt.PlayerUuid = uuid
  // ... populate ...
  // after sending, reset and return
  proto.Reset(evt)
  playerMovePool.Put(evt)
  ```

## High Priority (Performance Optimization)

### 5. Event Filtering/Subscription (Already Implemented ✅)
- [x] Track which plugins subscribe to which events
- [x] Don't send events to plugins that don't care
- [ ] Add metrics to track filtered events
- [ ] Consider priority levels for event delivery

### 6. Optimize Marshal Performance
- [ ] Profile current `proto.Marshal()` performance under load
- [ ] Consider using `proto.MarshalOptions{UseCachedSize: true}` for hot paths
- [ ] Benchmark with 100+ players and high event frequency
- [ ] Target: <50μs per marshal for simple events

## Medium Priority (Nice to Have)

### 7. Fast Path for High-Frequency Events
- [ ] Consider dedicated stream/channel for movement updates
- [ ] Skip envelope overhead for position-only updates
- [ ] Pack multiple positions into single frame (binary-packed array)
- [ ] Pattern: `[player_count][uuid1][x][y][z][uuid2][x][y][z]...`

### 8. Metrics and Profiling
- [ ] Add prometheus metrics for:
  - Events sent per second (by type)
  - Marshal time (histogram)
  - Buffer pool hit/miss rates
  - Event batch sizes
- [ ] Add pprof endpoints for CPU/memory profiling
- [ ] Benchmark suite for marshal/unmarshal performance

### 9. Backpressure Handling (Partially Implemented ✅)
- [x] Event queue with bounded channels (64 events)
- [x] Log warnings when queue is full
- [ ] Detect slow plugins (measure event processing latency)
- [ ] Drop non-critical events (movement) if plugin is lagging
- [ ] Keep critical events (join/quit/commands)
- [ ] Add plugin health monitoring

## Low Priority (Future Improvements)

### 10. Alternative Encodings
- [ ] Benchmark current protobuf performance vs alternatives
- [ ] Consider MessagePack or Cap'n Proto if significant gains
- [ ] Would require rewriting plugin clients
- [ ] **Note:** Likely not worth the effort unless profiling shows major bottleneck

### 11. Compression
- [ ] For large batches (>1KB), consider LZ4/Snappy compression
- [ ] Trade CPU for network bandwidth
- [ ] Especially useful for remote plugins over network
- [ ] Protobuf is already fairly compact, compression may not help much

### 12. Custom Fast-Path Codec (If Needed)
- [ ] If profiling shows marshal performance is critical bottleneck
- [ ] Implement custom codec for specific high-frequency messages
- [ ] Keep generated code for everything else
- [ ] Hybrid approach: hand-written for hot paths, generated for maintainability

## Testing Requirements

- [ ] Load test with 100+ simulated players
- [ ] Verify no memory leaks (buffer pools working correctly when implemented)
- [ ] Benchmark marshal performance (target: <50μs for simple events with generated code)
- [ ] Profile with pprof during high load
- [ ] Test event batching reduces events/sec by 10-100x
- [ ] Memory profiling to track allocation rates

## Notes

**Current Status (Nov 2025):**
- ✅ Ports/adapter architecture implemented
- ✅ Generated protobuf code (`protoc-gen-go`)
- ✅ Raw proto transport (minimal overhead)
- ✅ Schema defined in `.proto` file
- ✅ Event subscription/filtering implemented
- ✅ Event response timeout handling (250ms)
- ✅ Clean package structure (ports, adapters, config)
- ❌ Missing buffer pooling
- ❌ Missing event batching
- ❌ No movement events yet
- ❌ No message reuse pools

**Performance Targets:**
- 100 players × 50 ticks/sec = 5,000 events/sec
- Marshal time budget: ~2ms/tick for all events (40μs per event)
- GC pause target: <10ms (requires minimal allocation)
- Current estimate: Generated protobuf can handle 10,000+ events/sec easily

**Design Philosophy (Updated):**
Use generated protobuf code for maintainability. This gives us:
1. ✅ Zero maintenance burden for schema changes
2. ✅ Full protobuf ecosystem compatibility
3. ✅ Optimized by Google's protobuf team
4. ✅ Wire-compatible with all protobuf implementations
5. ⚠️ Slightly more allocations than hand-written (can optimize with pools if needed)
6. ⚠️ Some reflection overhead (negligible for most use cases)

**If performance becomes an issue:**
- Profile first with pprof
- Implement buffer pooling (Task #3)
- Implement message pooling (Task #4)
- Consider hybrid approach: custom codec for hot paths only

**Migration Notes:**
- Switched from hand-written to generated code on Nov 9, 2025
- Removed ~950 lines of hand-written codec (`messages.go`, `codec.go`)
- Updated all consumers to use `google.golang.org/protobuf/proto` package
- Field names changed from PascalCase to generated format (e.g., `PluginID` → `PluginId`)

