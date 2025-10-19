package sse

import (
	"go.uber.org/zap"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHub(t *testing.T) {
	hub := NewHub(
		WithStore(NewMemoryStore()),
		WithEnableResendEvents(),
		WithLogger(zap.NewExample()),
		WithPushBufferSize(200),
		WithPushFailedHandleFn(func(uid string, event *Event) {}),
		WithWorkerNum(10),
	)
	defer hub.Close()

	// build client and register to hub
	client := &UserClient{
		UID:  "u1",
		Send: make(chan *Event, 1),
	}
	hub.register <- client
	time.Sleep(10 * time.Millisecond)

	// test push event
	event := &Event{
		ID:    "e1",
		Event: "test",
		Data:  "data",
	}
	err := hub.Push([]string{"u1"}, event)
	assert.NoError(t, err)

	select {
	case e := <-client.Send:
		assert.Equal(t, "e1", e.ID)
		assert.Equal(t, "data", e.Data)
	case <-time.After(100 * time.Millisecond):
		t.Error("expected event pushed to client but got timeout")
	}

	// unregister client
	hub.unregister <- client
	time.Sleep(100 * time.Millisecond)

	// push again, because the client has been offline, will be ignored push
	err = hub.Push([]string{"u1"}, event)
	assert.NoError(t, err)
	time.Sleep(100 * time.Millisecond)

	// check stats
	total, _, failed, _ := hub.PushStats.Snapshot()
	assert.Equal(t, 1, int(total))
	assert.Equal(t, 0, int(failed))
}

func TestHubPushToAll(t *testing.T) {
	hub := NewHub()
	defer hub.Close()

	c1 := &UserClient{UID: "a", Send: make(chan *Event, 1)}
	c2 := &UserClient{UID: "b", Send: make(chan *Event, 1)}

	hub.register <- c1
	hub.register <- c2
	time.Sleep(20 * time.Millisecond)

	event := &Event{ID: "e2", Event: "broadcast", Data: "hello"}
	err := hub.Push(nil, event)
	assert.NoError(t, err)

	for _, cli := range []*UserClient{c1, c2} {
		select {
		case e := <-cli.Send:
			assert.Equal(t, "hello", e.Data)
		case <-time.After(100 * time.Millisecond):
			t.Errorf("expected broadcast event to client %s, got timeout", cli.UID)
		}
	}
}
