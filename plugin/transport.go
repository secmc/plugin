package plugin

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/net/http2"
)

// grpcStream implements a minimal HTTP/2-based bidirectional gRPC stream client
// specialised for the EventStream method used by plugins.
type grpcStream struct {
	cancel   context.CancelFunc
	reqBody  *io.PipeWriter
	respBody io.ReadCloser
	resp     *http.Response
	mu       sync.Mutex
}

func dialEventStream(parent context.Context, address string, timeout time.Duration) (*grpcStream, error) {
	baseCtx := parent
	var cancelTimeout context.CancelFunc
	if timeout > 0 {
		baseCtx, cancelTimeout = context.WithTimeout(parent, timeout)
	}

	dialer := &net.Dialer{}
	ctx, cancel := context.WithCancel(baseCtx)
	tr := &http2.Transport{
		AllowHTTP: true,
		DialTLS: func(network, addr string, _ *tls.Config) (net.Conn, error) {
			return dialer.DialContext(ctx, network, addr)
		},
	}
	pr, pw := io.Pipe()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("http://%s/df.plugin.Plugin/EventStream", address), pr)
	if err != nil {
		cancel()
		return nil, err
	}
	req.Header.Set("Content-Type", "application/grpc+proto")
	req.Header.Set("TE", "trailers")

	resp, err := tr.RoundTrip(req)
	if err != nil {
		cancel()
		pw.CloseWithError(err)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		cancel()
		pw.Close()
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	streamCancel := func() {
		cancel()
		if cancelTimeout != nil {
			cancelTimeout()
		}
	}

	return &grpcStream{cancel: streamCancel, reqBody: pw, respBody: resp.Body, resp: resp}, nil
}

func (s *grpcStream) Send(data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.reqBody == nil {
		return errors.New("stream closed")
	}
	var header [5]byte
	copy(header[:], []byte{0, 0, 0, 0, 0})
	length := len(data)
	header[1] = byte(length >> 24)
	header[2] = byte(length >> 16)
	header[3] = byte(length >> 8)
	header[4] = byte(length)
	if _, err := s.reqBody.Write(header[:]); err != nil {
		return err
	}
	if _, err := s.reqBody.Write(data); err != nil {
		return err
	}
	return nil
}

func (s *grpcStream) Recv() ([]byte, error) {
	header := make([]byte, 5)
	if _, err := io.ReadFull(s.respBody, header); err != nil {
		return nil, err
	}
	if header[0] != 0 {
		return nil, fmt.Errorf("unsupported compression: %d", header[0])
	}
	length := int(header[1])<<24 | int(header[2])<<16 | int(header[3])<<8 | int(header[4])
	if length < 0 {
		return nil, errors.New("negative message length")
	}
	data := make([]byte, length)
	if _, err := io.ReadFull(s.respBody, data); err != nil {
		return nil, err
	}
	return data, nil
}

func (s *grpcStream) CloseSend() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.reqBody == nil {
		return nil
	}
	err := s.reqBody.Close()
	s.reqBody = nil
	return err
}

func (s *grpcStream) Close() error {
	s.cancel()
	s.CloseSend()
	if s.respBody != nil {
		return s.respBody.Close()
	}
	return nil
}
