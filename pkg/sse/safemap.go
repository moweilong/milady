package sse

import "sync"

// SafeMap goroutine security Map structure encapsulating sync.Map
type SafeMap struct {
	m sync.Map // userID -> []*Client
}

// NewSafeMap creates a new SafeMap
func NewSafeMap() *SafeMap {
	return &SafeMap{m: sync.Map{}}
}

// Set store key-value pairs
func (sm *SafeMap) Set(uid string, client *UserClient) {
	sm.m.Store(uid, client)
}

// Get value by key
func (sm *SafeMap) Get(uid string) (*UserClient, bool) {
	value, ok := sm.m.Load(uid)
	if !ok {
		return nil, false
	}
	client, ok := value.(*UserClient)
	if !ok {
		return nil, false
	}
	return client, true
}

// Delete key-value pair
func (sm *SafeMap) Delete(uid string) {
	sm.m.Delete(uid)
}

// Has checked if key exists
func (sm *SafeMap) Has(uid string) bool {
	_, ok := sm.m.Load(uid)
	return ok
}

// Range all key-value pairs
func (sm *SafeMap) Range(f func(key, value interface{}) bool) {
	sm.m.Range(f)
}

// Keys get all keys
func (sm *SafeMap) Keys() []string {
	keys := make([]string, 0)
	sm.m.Range(func(key, _ interface{}) bool {
		k, ok := key.(string)
		if ok {
			keys = append(keys, k)
		}
		return true
	})
	return keys
}

// Values get all values
func (sm *SafeMap) Values() []*UserClient {
	var values []*UserClient
	sm.m.Range(func(_, value interface{}) bool {
		v, ok := value.(*UserClient)
		if !ok {
			return true
		}
		values = append(values, v)
		return true
	})
	return values
}

// Len Gets the number of elements in Map
// Note: Due to the nature of sync.Map, this operation is O(n) complex
func (sm *SafeMap) Len() int {
	length := 0
	sm.m.Range(func(_, _ interface{}) bool {
		length++
		return true
	})
	return length
}

// Clear all key-value pairs
func (sm *SafeMap) Clear() {
	sm.m = sync.Map{}
}
