package browser

import (
	"testing"
	"time"
)

func TestNewChromeRendererWithConfig_InvalidCommand(t *testing.T) {
	renderer := NewChromeRendererWithConfig("command-that-does-not-exist", 5*time.Second)
	if renderer != nil {
		t.Fatal("expected nil renderer for invalid command")
	}
}

func TestNewChromeRendererWithConfig_DefaultTimeout(t *testing.T) {
	renderer := NewChromeRendererWithConfig("sh", 0)
	if renderer == nil {
		t.Fatal("expected renderer for valid command")
	}
	if renderer.timeout != 15*time.Second {
		t.Fatalf("expected default timeout 15s, got %v", renderer.timeout)
	}
}
