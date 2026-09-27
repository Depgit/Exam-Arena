package ws

import (
	"encoding/json"
	"strings"
	"sync"
	"testing"
)

// A client whose send channel was closed by the hub (user reconnected) must
// silently drop later writes instead of panicking with "send on closed channel".
func TestSendAfterCloseDoesNotPanic(t *testing.T) {
	c := NewClient(NewHub(), nil, "u1", "alice")

	if !c.enqueue([]byte("before")) {
		t.Fatal("enqueue before close should succeed")
	}

	c.closeSend()
	c.closeSend() // idempotent

	if c.enqueue([]byte("after")) {
		t.Fatal("enqueue after close should report failure")
	}
	c.SendJSON(Message{Type: "x"}) // must not panic
}

// Writers and the closer race each other under -race.
func TestConcurrentSendAndClose(t *testing.T) {
	c := NewClient(NewHub(), nil, "u1", "alice")

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				c.SendJSON(Message{Type: "score_update"})
			}
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		c.closeSend()
	}()
	wg.Wait()
}

func TestHandleMessageRecoversFromPanic(t *testing.T) {
	h := NewHub()
	c := NewClient(h, nil, "u1", "alice")
	h.RegisterHandler("boom", func(*Client, json.RawMessage) { panic("handler exploded") })

	// Must not propagate the panic.
	h.handleMessage(&IncomingMessage{Client: c, Payload: []byte(`{"type":"boom","payload":{}}`)})

	select {
	case frame := <-c.send:
		if !strings.Contains(string(frame), `"type":"error"`) {
			t.Fatalf("expected an error frame, got %s", frame)
		}
	default:
		t.Fatal("expected an error frame to be queued for the client")
	}
}
