# Plugin System Performance TODO

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
- [ ] Implement Marshal/Unmarshal in `messages.go`
- [ ] Wire up in player movement handler

### 2. Implement Event Batching
- [ ] Add `EventBatch` message to `plugin.proto`
  ```protobuf
  message EventBatch {
    repeated EventEnvelope events = 1;
  }
  ```
- [ ] Add to `HostToPlugin.payload` oneof (field 21)
- [ ] Implement batch marshaling in `messages.go`
- [ ] Update event dispatcher to batch high-frequency events
- [ ] Send batches once per tick instead of individual events

### 3. Zero-Allocation Marshaling
- [ ] Add `MarshalTo(buf []byte) []byte` to all message types
- [ ] Keep existing `Marshal()` as wrapper calling `MarshalTo(nil)`
- [ ] Update `appendMessage` to use `MarshalTo` to avoid temporary buffers
- [ ] Pattern:
  ```go
  func (m *Type) Marshal() ([]byte, error) {
      return m.MarshalTo(nil), nil
  }
  
  func (m *Type) MarshalTo(buf []byte) []byte {
      buf = buf[:0]  // reset but keep capacity
      // ... append logic ...
      return buf
  }
  ```

### 4. Buffer Pooling
- [ ] Create `sync.Pool` for message buffers in `transport.go` or `manager.go`
- [ ] Typical size: 256-512 bytes for common events, 4KB for batches
- [ ] Update send paths to use pooled buffers
- [ ] Pattern:
  ```go
  var msgBufPool = sync.Pool{
      New: func() any { return make([]byte, 0, 512) },
  }
  
  buf := msgBufPool.Get().([]byte)[:0]
  buf = msg.MarshalTo(buf)
  stream.Send(buf)
  msgBufPool.Put(buf)
  ```

## High Priority (Performance Optimization)

### 5. Implement Size() Methods
- [ ] Add `Size() int` interface to codec.go
- [ ] Implement for all message types
- [ ] Optimize `appendMessage` to pre-calculate nested message sizes
- [ ] Avoids double-buffering when encoding nested messages

### 6. Message Reuse Pool
- [ ] Pool message structs themselves (not just buffers)
- [ ] Reduce GC pressure from event creation
- [ ] Pattern:
  ```go
  var playerMovePool = sync.Pool{
      New: func() any { return &PlayerMoveEvent{} },
  }
  
  evt := playerMovePool.Get().(*PlayerMoveEvent)
  evt.PlayerUUID = uuid
  // ... populate ...
  // after sending, clear and return
  *evt = PlayerMoveEvent{}  // zero out
  playerMovePool.Put(evt)
  ```

### 7. Event Filtering/Subscription
- [ ] Track which plugins subscribe to which events
- [ ] Don't send movement events to plugins that don't care
- [ ] Reduces network traffic and plugin processing overhead

## Medium Priority (Nice to Have)

### 8. Fast Path for High-Frequency Events
- [ ] Consider dedicated stream/channel for movement updates
- [ ] Skip envelope overhead for position-only updates
- [ ] Pack multiple positions into single frame (binary-packed array)
- [ ] Pattern: `[player_count][uuid1][x][y][z][uuid2][x][y][z]...`

### 9. Metrics and Profiling
- [ ] Add prometheus metrics for:
  - Events sent per second (by type)
  - Marshal time (histogram)
  - Buffer pool hit/miss rates
  - Event batch sizes
- [ ] Add pprof endpoints for CPU/memory profiling
- [ ] Benchmark suite for marshal/unmarshal performance

### 10. Backpressure Handling
- [ ] Detect slow plugins (event queue backing up)
- [ ] Drop non-critical events (movement) if plugin is lagging
- [ ] Keep critical events (join/quit/commands)
- [ ] Log warnings when dropping events

## Low Priority (Future Improvements)

### 11. Unsafe String Optimizations
- [ ] Consider `unsafe` for string→[]byte conversions in hot paths
- [ ] Only if profiling shows string allocation as bottleneck
- [ ] Document safety guarantees required

### 12. Alternative Encodings
- [ ] Benchmark protobuf vs MessagePack vs Cap'n Proto
- [ ] Consider switching if significant performance gain
- [ ] Would require rewriting plugin clients

### 13. Compression
- [ ] For large batches (>1KB), consider LZ4/Snappy compression
- [ ] Trade CPU for network bandwidth
- [ ] Especially useful for remote plugins over network

## Testing Requirements

- [ ] Load test with 100+ simulated players
- [ ] Verify no memory leaks (buffer pools working correctly)
- [ ] Benchmark marshal performance (target: <1μs for simple events)
- [ ] Profile with pprof during high load
- [ ] Test event batching reduces events/sec by 10-100x

## Notes

**Current Status:**
- ✅ Hand-written protobuf codec (zero reflection)
- ✅ Raw proto transport (minimal overhead)
- ✅ Schema defined in `.proto` file
- ❌ Missing buffer pooling
- ❌ Missing event batching
- ❌ No movement events yet

**Performance Targets:**
- 100 players × 50 ticks/sec = 5,000 events/sec
- Marshal time budget: ~2ms/tick for all events (40μs per event)
- GC pause target: <10ms (requires minimal allocation)

**Design Philosophy:**
Keep the hand-written codec approach. It gives us:
1. Zero reflection overhead
2. Full control over allocations
3. Buffer reuse capability
4. Wire-compatible with standard protobuf (plugins can use normal libraries)

