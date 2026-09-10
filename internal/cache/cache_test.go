package cache

import (
	"testing"
	"time"
)

func TestGetSetJSON(t *testing.T) {
	c := New(time.Minute, 8)
	type payload struct {
		N int `json:"n"`
	}
	c.SetJSON("a", payload{N: 3})
	var got payload
	if !c.GetJSON("a", &got) || got.N != 3 {
		t.Fatalf("got=%+v", got)
	}
}

func TestExpiry(t *testing.T) {
	c := New(10*time.Millisecond, 8)
	c.SetJSON("a", map[string]int{"n": 1})
	time.Sleep(20 * time.Millisecond)
	var dest map[string]int
	if c.GetJSON("a", &dest) {
		t.Fatal("expected miss after expiry")
	}
}

func TestMaxEvicts(t *testing.T) {
	c := New(time.Minute, 2)
	c.SetJSON("a", 1)
	c.SetJSON("b", 2)
	c.SetJSON("c", 3)
	if c.Size() != 2 {
		t.Fatalf("size=%d", c.Size())
	}
}

func TestDisabled(t *testing.T) {
	c := New(0, 8)
	c.SetJSON("a", 1)
	var dest int
	if c.GetJSON("a", &dest) {
		t.Fatal("ttl 0 should disable cache")
	}
}
