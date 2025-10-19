package sse

import (
	"testing"
)

func TestSafeMapBasicOperations(t *testing.T) {
	sm := NewSafeMap()
	client := &UserClient{UID: "user1"}

	// Test Set and Has
	sm.Set("user1", client)
	if !sm.Has("user1") {
		t.Error("SafeMap should contain key 'user1'")
	}

	// Test Get
	c, ok := sm.Get("user1")
	if !ok || c.UID != "user1" {
		t.Error("SafeMap Get failed to retrieve correct value")
	}

	// Test Len
	if sm.Len() != 1 {
		t.Errorf("Expected Len=1, got %d", sm.Len())
	}

	// Test Keys
	keys := sm.Keys()
	if len(keys) != 1 || keys[0] != "user1" {
		t.Errorf("Expected keys ['user1'], got %v", keys)
	}

	// Test Values
	values := sm.Values()
	if len(values) != 1 || values[0].UID != "user1" {
		t.Errorf("Expected values with UID 'user1', got %v", values)
	}

	// Test Delete
	sm.Delete("user1")
	if sm.Has("user1") {
		t.Error("SafeMap key 'user1' should have been deleted")
	}

	// Test Clear
	sm.Set("a", client)
	sm.Set("b", client)
	sm.Clear()
	if sm.Len() != 0 {
		t.Error("SafeMap should be empty after Clear()")
	}
}

func TestSafeMapRange(t *testing.T) {
	sm := NewSafeMap()
	c1 := &UserClient{UID: "u1"}
	c2 := &UserClient{UID: "u2"}
	sm.Set("u1", c1)
	sm.Set("u2", c2)

	found := make(map[string]bool)
	sm.Range(func(k, v interface{}) bool {
		key := k.(string)
		found[key] = true
		return true
	})

	if !(found["u1"] && found["u2"]) {
		t.Errorf("SafeMap Range did not iterate all keys, got %v", found)
	}
}
