package grpc

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"runtime"
	"strings"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// GrpcStream wraps a bidirectional stream for a connected plugin
type GrpcStream struct {
	stream grpc.ServerStream
	mu     sync.Mutex
}

// GrpcServer manages the gRPC server that plugins connect to
type GrpcServer struct {
	server   *grpc.Server
	listener net.Listener
	handler  StreamHandler
	mu       sync.Mutex
}

// StreamHandler is called when a new plugin connects
type StreamHandler func(stream *GrpcStream) error

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

// pluginService implements the gRPC service
type pluginService struct {
	handler StreamHandler
}

// EventStream handles the bidirectional stream from plugins
func (s *pluginService) EventStream(stream grpc.ServerStream) error {
	if s.handler == nil {
		return errors.New("no handler registered")
	}
	return s.handler(&GrpcStream{stream: stream})
}

// NewServer creates a new gRPC server that plugins will connect to
func NewServer(address string, handler StreamHandler) (*GrpcServer, error) {
	network := "tcp"
	addr := address
	if strings.Contains(address, "://") {
		u, err := url.Parse(address)
		if err != nil {
			return nil, fmt.Errorf("invalid server address %q: %w", address, err)
		}
		switch u.Scheme {
		case "unix":
			network = "unix"
			addr = u.Path
		case "tcp":
			network = "tcp"
			addr = u.Host
		default:
			return nil, fmt.Errorf("unsupported address scheme %q", u.Scheme)
		}
	} else {
		if strings.HasPrefix(address, "/") {
			network = "unix"
		} else {
			network = "tcp"
		}
		addr = address
	}

	if network == "unix" {
		// Clean up old socket file
		_ = os.Remove(addr)
	}

	listener, err := net.Listen(network, addr)
	if err != nil {
		return nil, fmt.Errorf("listen failed: %w", err)
	}

	// Set permissions on unix for sockets
	if network == "unix" && runtime.GOOS != "windows" {
		_ = os.Chmod(addr, 0666)
	}

	server := grpc.NewServer(
		grpc.ForceServerCodec(rawProtoCodec{}),
		grpc.Creds(insecure.NewCredentials()),
	)

	service := &pluginService{handler: handler}

	// Register the service manually
	streamDesc := grpc.StreamDesc{
		StreamName: "EventStream",
		Handler: func(srv any, stream grpc.ServerStream) error {
			return service.EventStream(stream)
		},
		ServerStreams: true,
		ClientStreams: true,
	}

	server.RegisterService(&grpc.ServiceDesc{
		ServiceName: "df.plugin.Plugin",
		HandlerType: (*any)(nil),
		Streams:     []grpc.StreamDesc{streamDesc},
	}, service)

	return &GrpcServer{
		server:   server,
		listener: listener,
		handler:  handler,
	}, nil
}

// Serve starts accepting plugin connections
func (s *GrpcServer) Serve() error {
	return s.server.Serve(s.listener)
}

// Stop gracefully stops the server
func (s *GrpcServer) Stop() {
	s.server.GracefulStop()

	// Remove Unix socket file if it was used
	if addr := s.listener.Addr(); addr.Network() == "unix" {
		os.Remove(addr.String())
	}
}

// Address returns the address the server is listening on
func (s *GrpcServer) Address() string {
	return s.listener.Addr().String()
}

func (s *GrpcStream) Send(data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.stream == nil {
		return errors.New("stream closed")
	}

	if err := s.stream.SendMsg(&data); err != nil {
		return err
	}
	return nil
}

func (s *GrpcStream) Recv() ([]byte, error) {
	var data []byte
	if err := s.stream.RecvMsg(&data); err != nil {
		if errors.Is(err, io.EOF) {
			return nil, err
		}
		return nil, err
	}
	return data, nil
}

func (s *GrpcStream) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stream = nil
	return nil
}
