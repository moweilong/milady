package sse

import (
	"context"
	"sync"
	"testing"
)

func TestEvent_CheckValid(t *testing.T) {
	tests := []struct {
		name    string
		event   *Event
		wantErr bool
	}{
		{
			name: "valid event",
			event: &Event{
				ID:    "1",
				Event: "test",
				Data:  "test data",
			},
			wantErr: false,
		},
		{
			name: "invalid event",
			event: &Event{
				ID:    "1",
				Event: "",
				Data:  "test data",
			},
			wantErr: true,
		},
		{
			name: "invalid data",
			event: &Event{
				ID:    "1",
				Event: "test",
				Data:  nil,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.event.CheckValid()
			if (err != nil) != tt.wantErr {
				t.Errorf("Event.CheckValid() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type mockStore struct {
	mu     sync.RWMutex
	events map[string][]*Event
}

func newMockStore() *mockStore {
	return &mockStore{
		events: make(map[string][]*Event),
	}
}

func (m *mockStore) Save(_ context.Context, e *Event) error {
	m.mu.Lock()
	if _, ok := m.events[e.Event]; !ok {
		m.events[e.Event] = make([]*Event, 0)
	}
	m.events[e.Event] = append(m.events[e.Event], e)
	m.mu.Unlock()
	return nil
}

func (m *mockStore) GetSince(_ context.Context, eventType string, lastID string) ([]*Event, error) {
	m.mu.RLock()
	var result []*Event
	if events, ok := m.events[eventType]; ok {
		for _, e := range events {
			if e.ID > lastID {
				result = append(result, e)
			}
		}
	}
	m.mu.RUnlock()
	return result, nil
}

func TestStore(t *testing.T) {
	store := newMockStore()
	event1 := &Event{
		ID:    "1",
		Event: "test",
		Data:  "test data",
	}
	event2 := &Event{
		ID:    "2",
		Event: "test",
		Data:  "test data",
	}
	event3 := &Event{
		ID:    "3",
		Event: "test",
		Data:  "test data",
	}

	ctx := context.Background()
	_ = store.Save(ctx, event1)
	_ = store.Save(ctx, event2)
	_ = store.Save(ctx, event3)
	events, _ := store.GetSince(ctx, "test", "0")
	if len(events) != 3 {
		t.Errorf("Expected 3 events, got %d", len(events))
	}
}

// -------------------------------------------------------------------------------------

type memoryStore struct {
	mu     sync.RWMutex
	events map[string][]*Event
}

func NewMemoryStore() Store {
	return &memoryStore{
		events: make(map[string][]*Event),
	}
}

func (m *memoryStore) Save(_ context.Context, e *Event) error {
	m.mu.Lock()
	m.events[e.Event] = append(m.events[e.Event], e)
	m.mu.Unlock()
	return nil
}

func (m *memoryStore) ListByLastID(_ context.Context, eventType string, lastID string, pageSize int) ([]*Event, string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*Event
	var nextID string

	if events, ok := m.events[eventType]; ok {
		// find the starting position
		start := 0
		if lastID != "" {
			for i, e := range events {
				if e.ID > lastID {
					start = i
					break
				}
			}
		}

		// get pagination data
		end := start + pageSize
		if end > len(events) {
			end = len(events)
		} else if end < len(events) {
			nextID = events[end].ID
		}

		result = events[start:end]
	}

	return result, nextID, nil
}

// -------------------------------------------------------------------------------------

// redis store example

//type redisStore struct {
//	client *redis.Client
//	expire time.Duration
//}
//
//func NewRedisStore(addr, password string, expire time.Duration) Store {
//	rdb := redis.NewClient(&redis.Options{
//		Addr:     addr,
//		Password: password,
//	})
//	return &redisStore{client: rdb, expire: expire}
//}
//
//func (s *redisStore) Save(ctx context.Context, e *Event) error {
//	key := fmt.Sprintf("sse:%s", e.Event)
//	b, _ := json.Marshal(e)
//	s.client.RPush(ctx, key, b)
//	s.client.Expire(ctx, key, s.expire)
//	return nil
//}
//
//func (s *redisStore) ListByLastID(ctx context.Context, eventType string, lastID string, pageSize int) ([]*Event, string, error) {
//	key := fmt.Sprintf("sse:%s", eventType)
//
//	// get the total number of events
//	total, err := s.client.LLen(ctx, key).Result()
//	if err != nil {
//		return nil, "", err
//	}
//
//	// if no lastID is specified, start from 0
//	start := int64(0)
//	if lastID != "" {
//		// use binary search to locate the position of lastID
//		low := int64(0)
//		high := total - 1
//		found := false
//
//		for low <= high {
//			mid := (low + high) / 2
//			val, err := s.client.LIndex(ctx, key, mid).Result()
//			if err != nil {
//				return nil, "", err
//			}
//
//			e := &Event{}
//			if err = json.Unmarshal([]byte(val), e); err != nil {
//				return nil, "", err
//			}
//
//			cmp := strings.Compare(e.ID, lastID)
//			if cmp == 0 {
//				start = mid + 1
//				found = true
//				break
//			} else if cmp < 0 {
//				low = mid + 1
//			} else {
//				high = mid - 1
//			}
//		}
//
//		if !found {
//			// if not found lastID, return empty
//			return nil, "", nil
//		}
//	}
//
//	end := start + int64(pageSize) - 1
//	if end >= total {
//		end = total - 1
//	}
//
//	// get the pagination data
//	values, err := s.client.LRange(ctx, key, start, end).Result()
//	if err != nil {
//		return nil, "", err
//	}
//
//	var result []*Event
//	for _, v := range values {
//		e := &Event{}
//		if err = json.Unmarshal([]byte(v), e); err == nil {
//			result = append(result, e)
//		}
//	}
//
//	// calculate the next pagination start ID
//	var nextID string
//	if end+1 < total {
//		nextVal, err := s.client.LIndex(ctx, key, end+1).Result()
//		if err == nil {
//			e := &Event{}
//			if err = json.Unmarshal([]byte(nextVal), e); err == nil {
//				nextID = e.ID
//			}
//		}
//	}
//
//	return result, nextID, nil
//}
