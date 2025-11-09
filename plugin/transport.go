package plugin

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

// grpcStream wraps a gRPC client stream for bidirectional communication
type grpcStream struct {
	conn   *grpc.ClientConn
	stream grpc.ClientStream
	cancel context.CancelFunc
	mu     sync.Mutex
}

// rawProtoCodec is a pass-through codec that treats payloads as already
// protobuf-marshalled bytes. We name it "proto" so the Content-Type
// remains application/grpc+proto to interoperate with grpc-js.
type rawProtoCodec struct{}

func (rawProtoCodec) Name() string { return "proto" }

func (rawProtoCodec) Marshal(v any) ([]byte, error) {
	switch t := v.(type) {
	case []byte:
		return t, nil
	case *[]byte:
		return *t, nil
	default:
		return nil, fmt.Errorf("rawProtoCodec: unsupported marshal type %T", v)
	}
}

func (rawProtoCodec) Unmarshal(data []byte, v any) error {
	switch t := v.(type) {
	case *[]byte:
		*t = append((*t)[:0], data...)
		return nil
	default:
		return fmt.Errorf("rawProtoCodec: unsupported unmarshal target %T (need *[]byte)", v)
	}
}

func dialEventStream(parent context.Context, address string, connectTimeout time.Duration) (*grpcStream, error) {
	conn, err := grpc.NewClient(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("dial failed: %w", err)
	}

	conn.Connect()
	if connectTimeout > 0 {
		waitCtx, cancel := context.WithTimeout(parent, connectTimeout)
		defer cancel()
		for {
			state := conn.GetState()
			if state == connectivity.Ready {
				break
			}
			if !conn.WaitForStateChange(waitCtx, state) {
				_ = conn.Close()
				return nil, fmt.Errorf("connect timeout: %w", waitCtx.Err())
			}
		}
	}

	// Create stream context
	ctx, cancel := context.WithCancel(parent)

	// Create the bidirectional stream using the generic streaming API
	streamDesc := &grpc.StreamDesc{
		StreamName:    "EventStream",
		ServerStreams: true,
		ClientStreams: true,
	}

	stream, err := conn.NewStream(ctx, streamDesc, "/df.plugin.Plugin/EventStream", grpc.ForceCodec(rawProtoCodec{}))
	if err != nil {
		cancel()
		conn.Close()
		return nil, fmt.Errorf("create stream failed: %w", err)
	}

	return &grpcStream{
		conn:   conn,
		stream: stream,
		cancel: cancel,
	}, nil
}

func (s *grpcStream) Send(data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.stream == nil {
		return errors.New("stream closed")
	}

	// Send the raw protobuf message bytes
	if err := s.stream.SendMsg(&data); err != nil {
		return err
	}
	return nil
}

func (s *grpcStream) Recv() ([]byte, error) {
	var data []byte
	if err := s.stream.RecvMsg(&data); err != nil {
		return nil, err
	}
	return data, nil
}

func (s *grpcStream) CloseSend() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.stream == nil {
		return nil
	}
	return s.stream.CloseSend()
}

func (s *grpcStream) Close() error {
	s.cancel()
	s.CloseSend()
	if s.conn != nil {
		return s.conn.Close()
	}
	return nil
}
