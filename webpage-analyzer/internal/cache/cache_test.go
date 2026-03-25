package cache

import "testing"

func TestRedisResultCache_Key(t *testing.T) {
	cache := NewRedisResultCache("localhost:6379", "", 0)

	keyA := cache.key("https://example.com")
	keyB := cache.key("https://example.com")
	keyC := cache.key("https://example.org")

	if keyA != keyB {
		t.Fatalf("expected same URL to produce identical key, got %q and %q", keyA, keyB)
	}
	if keyA == keyC {
		t.Fatalf("expected different URLs to produce different keys, got %q", keyA)
	}
}
