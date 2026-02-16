package console

import (
	"bufio"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestBroadcasterSendReceive(t *testing.T) {
	b := NewBroadcaster()

	ch := b.Subscribe()
	defer b.Unsubscribe(ch)

	b.Send(Event{Type: "observation", Data: `{"id":1}`})

	select {
	case e := <-ch:
		if e.Type != "observation" {
			t.Errorf("Type = %q, want %q", e.Type, "observation")
		}
		if e.Data != `{"id":1}` {
			t.Errorf("Data = %q, want %q", e.Data, `{"id":1}`)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for event")
	}
}

func TestBroadcasterMultipleSubscribers(t *testing.T) {
	b := NewBroadcaster()

	ch1 := b.Subscribe()
	defer b.Unsubscribe(ch1)
	ch2 := b.Subscribe()
	defer b.Unsubscribe(ch2)

	b.Send(Event{Type: "test", Data: "hello"})

	for _, ch := range []<-chan Event{ch1, ch2} {
		select {
		case e := <-ch:
			if e.Data != "hello" {
				t.Errorf("Data = %q, want %q", e.Data, "hello")
			}
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for event")
		}
	}
}

func TestBroadcasterUnsubscribe(t *testing.T) {
	b := NewBroadcaster()

	ch := b.Subscribe()
	b.Unsubscribe(ch)

	// Should not block or panic
	b.Send(Event{Type: "test", Data: "after-unsub"})

	select {
	case _, ok := <-ch:
		if ok {
			t.Error("expected channel to be closed after unsubscribe")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("channel should be closed")
	}
}

func TestSSEHandler(t *testing.T) {
	srv := testServer(t)

	// Send an event before connecting (should not block)
	srv.sse.Send(Event{Type: "early", Data: "miss"})

	// Create a request with a cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	req := httptest.NewRequest("GET", "/api/events", nil).WithContext(ctx)
	rr := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		srv.Handler().ServeHTTP(rr, req)
		close(done)
	}()

	// Give the handler time to subscribe
	time.Sleep(50 * time.Millisecond)

	// Send an event while connected
	srv.sse.Send(Event{Type: "observation", Data: `{"id":42}`})

	// Give time for the write
	time.Sleep(50 * time.Millisecond)

	// Cancel to close the connection
	cancel()
	<-done

	// Check response headers
	if ct := rr.Header().Get("Content-Type"); ct != "text/event-stream" {
		t.Errorf("Content-Type = %q, want text/event-stream", ct)
	}
	if cc := rr.Header().Get("Cache-Control"); cc != "no-cache" {
		t.Errorf("Cache-Control = %q, want no-cache", cc)
	}

	// Parse the SSE output
	body := rr.Body.String()
	scanner := bufio.NewScanner(strings.NewReader(body))
	var foundEvent, foundData bool
	for scanner.Scan() {
		line := scanner.Text()
		if line == "event: observation" {
			foundEvent = true
		}
		if line == `data: {"id":42}` {
			foundData = true
		}
	}
	if !foundEvent {
		t.Errorf("missing event line in SSE output: %s", body)
	}
	if !foundData {
		t.Errorf("missing data line in SSE output: %s", body)
	}
}

func TestSSEHandlerClientCount(t *testing.T) {
	b := NewBroadcaster()

	if got := b.ClientCount(); got != 0 {
		t.Errorf("ClientCount() = %d, want 0", got)
	}

	ch1 := b.Subscribe()
	ch2 := b.Subscribe()

	if got := b.ClientCount(); got != 2 {
		t.Errorf("ClientCount() = %d, want 2", got)
	}

	b.Unsubscribe(ch1)
	if got := b.ClientCount(); got != 1 {
		t.Errorf("ClientCount() = %d, want 1", got)
	}

	b.Unsubscribe(ch2)
	if got := b.ClientCount(); got != 0 {
		t.Errorf("ClientCount() = %d, want 0", got)
	}
}

func TestSSEEndpointRegistered(t *testing.T) {
	srv := testServer(t)

	// Verify the endpoint exists (should return 200 with SSE headers)
	ctx, cancel := context.WithCancel(context.Background())
	req := httptest.NewRequest("GET", "/api/events", nil).WithContext(ctx)
	rr := httptest.NewRecorder()

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}
