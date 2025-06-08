package proxy

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"
)

func explosiveProxy(t *testing.T) Proxy {
	return func(ctx context.Context, _ *Request) (*Response, error) {
		t.Error("This proxy shouldn't been executed!")
		return &Response{}, nil
	}
}

func newDummyReadCloser(content string) io.ReadCloser {
	return dummyReadCloser{strings.NewReader(content)}
}

type dummyReadCloser struct {
	reader io.Reader
}

func (d dummyReadCloser) Read(p []byte) (int, error) {
	return d.reader.Read(p)
}

func (dummyReadCloser) Close() error {
	return nil
}

type dummyRC struct {
	r      io.Reader
	closed bool
	mu     *sync.Mutex
}

func (d *dummyRC) Read(b []byte) (int, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return -1, fmt.Errorf("Reading from a closed source")
	}
	return d.r.Read(b)
}

func (d *dummyRC) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.closed = true
	return nil
}

func (d *dummyRC) IsClosed() bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	res := d.closed
	return res
}
