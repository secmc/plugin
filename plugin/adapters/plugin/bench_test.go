package plugin

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"

	"github.com/secmc/plugin/plugin/config"
	pb "github.com/secmc/plugin/proto/generated/go"
)

// rawProtoCodec is a minimal client-side codec that passes raw protobuf bytes.
// It mirrors the server's codec so we can use the real gRPC transport.
type rawProtoCodec struct{}

func (rawProtoCodec) Name() string { return "proto" }
func (rawProtoCodec) Marshal(v any) ([]byte, error) {
	switch t := v.(type) {
	case []byte:
		return t, nil
	case *[]byte:
		return *t, nil
	default:
		return nil, nil
	}
}
func (rawProtoCodec) Unmarshal(data []byte, v any) error {
	switch t := v.(type) {
	case *[]byte:
		*t = append((*t)[:0], data...)
		return nil
	default:
		return nil
	}
}

var _benchStreamDesc = &grpc.StreamDesc{
	StreamName:    "EventStream",
	ServerStreams: true,
	ClientStreams: true,
}

func sendPluginToHost(t testing.TB, cs grpc.ClientStream, msg *pb.PluginToHost) {
	t.Helper()
	data, err := proto.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal PluginToHost: %v", err)
	}
	if err := cs.SendMsg(&data); err != nil {
		t.Fatalf("send PluginToHost: %v", err)
	}
}

func recvHostToPlugin(t testing.TB, cs grpc.ClientStream) *pb.HostToPlugin {
	t.Helper()
	var data []byte
	if err := cs.RecvMsg(&data); err != nil {
		t.Fatalf("recv HostToPlugin: %v", err)
	}
	msg := new(pb.HostToPlugin)
	if err := proto.Unmarshal(data, msg); err != nil {
		t.Fatalf("unmarshal HostToPlugin: %v", err)
	}
	return msg
}

func setupManagerAndPlugin(t testing.TB) (*Manager, grpc.ClientConnInterface, grpc.ClientStream, func()) {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
	m := NewManager(nil, logger, nil, nil)

	// Start manager on an ephemeral TCP port with one plugin slot (no process launch).
	const pluginID = "bench-plugin"
	cfg := config.Config{
		ServerAddr: "127.0.0.1:0",
		Plugins: []config.PluginConfig{{
			ID:   pluginID,
			Name: "Bench",
		}},
	}
	if err := m.StartWithConfig(cfg); err != nil {
		t.Fatalf("start manager: %v", err)
	}

	// Create a real gRPC client connection using matching raw codec.
	addr := m.grpcServer.Address()
	conn, err := grpc.Dial(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.ForceCodec(rawProtoCodec{})),
		grpc.WithBlock(),
		grpc.WithTimeout(3*time.Second),
	)
	if err != nil {
		t.Fatalf("dial plugin server: %v", err)
	}

	// Open the bidi stream.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	cs, err := grpc.NewClientStream(ctx, _benchStreamDesc, conn, "/df.plugin.Plugin/EventStream")
	if err != nil {
		cancel()
		t.Fatalf("open EventStream: %v", err)
	}

	// First message: Hello (must include PluginId).
	sendPluginToHost(t, cs, &pb.PluginToHost{
		PluginId: pluginID,
		Payload: &pb.PluginToHost_Hello{
			Hello: &pb.PluginHello{
				Name:       "Bench",
				Version:    "1.0.0",
				ApiVersion: "v1",
			},
		},
	})

	// Manager will attach stream and send HostHello; read and ignore.
	_ = recvHostToPlugin(t, cs)

	// Subscribe to COMMAND events so manager will dispatch to this plugin.
	sendPluginToHost(t, cs, &pb.PluginToHost{
		PluginId: pluginID,
		Payload: &pb.PluginToHost_Subscribe{
			Subscribe: &pb.EventSubscribe{
				Events: []pb.EventType{pb.EventType_COMMAND},
			},
		},
	})

	// Wait briefly for subscriptions to become active.
	time.Sleep(50 * time.Millisecond)

	cleanup := func() {
		_ = cs.CloseSend()
		cancel()
		_ = conn.Close()
		m.Close()
	}
	return m, conn, cs, cleanup
}

func runPluginResponder(t testing.TB, cs grpc.ClientStream, withActions bool) chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			// Block for incoming HostToPlugin messages.
			var data []byte
			if err := cs.RecvMsg(&data); err != nil {
				return
			}
			msg := new(pb.HostToPlugin)
			if err := proto.Unmarshal(data, msg); err != nil {
				return
			}
			switch payload := msg.GetPayload().(type) {
			case *pb.HostToPlugin_Event:
				evt := payload.Event
				// Immediately return an EventResult (no cancel).
				res := &pb.EventResult{EventId: evt.EventId}
				reply := &pb.PluginToHost{
					PluginId: msg.PluginId,
					Payload:  &pb.PluginToHost_EventResult{EventResult: res},
				}
				// Send EventResult
				buf, _ := proto.Marshal(reply)
				if err := cs.SendMsg(&buf); err != nil {
					return
				}
				// Optionally send a small actions batch to exercise the actions path.
				if withActions {
					// Use a world query action that is safe without an attached world (host will reply with an error).
					actions := &pb.ActionBatch{
						Actions: []*pb.Action{
							{
								Kind: &pb.Action_WorldQueryDefaultGameMode{
									WorldQueryDefaultGameMode: &pb.WorldQueryDefaultGameModeAction{
										World: &pb.WorldRef{Name: "overworld"},
									},
								},
							},
						},
					}
					actMsg := &pb.PluginToHost{
						PluginId: msg.PluginId,
						Payload:  &pb.PluginToHost_Actions{Actions: actions},
					}
					buf2, _ := proto.Marshal(actMsg)
					if err := cs.SendMsg(&buf2); err != nil {
						return
					}
				}
			default:
				// Ignore HostHello, ActionResult, etc.
			}
		}
	}()
	return done
}

func BenchmarkCommandRoundtrip(b *testing.B) {
	m, _, cs, cleanup := setupManagerAndPlugin(b)
	defer cleanup()

	// Start responder without actions for a pure event roundtrip.
	_ = runPluginResponder(b, cs, false)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		env := &pb.EventEnvelope{
			Type: pb.EventType_COMMAND,
			Payload: &pb.EventEnvelope_Command{
				Command: &pb.CommandEvent{
					PlayerUuid: "00000000-0000-0000-0000-000000000000",
					Name:       "Bench",
					Raw:        "/bench",
					Command:    "bench",
					Args:       []string{"a", "b", "c"},
				},
			},
		}
		results := m.emitCancellable(nil, env)
		if len(results) == 0 {
			b.Fatal("no event results")
		}
	}
}

func BenchmarkCommandWithActions(b *testing.B) {
	m, _, cs, cleanup := setupManagerAndPlugin(b)
	defer cleanup()

	// Start responder that also sends a small actions batch per event.
	_ = runPluginResponder(b, cs, true)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		env := &pb.EventEnvelope{
			Type: pb.EventType_COMMAND,
			Payload: &pb.EventEnvelope_Command{
				Command: &pb.CommandEvent{
					PlayerUuid: "00000000-0000-0000-0000-000000000000",
					Name:       "Bench",
					Raw:        "/bench",
					Command:    "bench",
					Args:       []string{"a", "b", "c"},
				},
			},
		}
		results := m.emitCancellable(nil, env)
		if len(results) == 0 {
			b.Fatal("no event results")
		}
	}
}
