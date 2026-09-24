package tests

import (
	"reflect"
	"slices"
	"testing"
	"time"
	"github.com/NooRMaseR/go-redis/stores"
)

func TestTTLKeys_getKeys(t *testing.T) {
	store := stores.NewTTLStore(0)
	store.Set("a", stores.TTL{Value: "val1", Expiration: time.Now().Add(10 * time.Second)})
	store.Set("b", stores.TTL{Value: "val2", Expiration: time.Now().Add(20 * time.Second)})
	keys := store.Keys()

	slices.Sort(keys)

	expected := []string{"a","b"}
	if !reflect.DeepEqual(keys, expected) {
		t.Errorf("Expected keys %v, got %v", expected, keys)
	}
}

func TestKeysTTL_checkValues(t *testing.T) {
	store := stores.NewTTLStore(0)
	store.Set("a", stores.TTL{Value: "val1", Expiration: time.Now().Add(10 * time.Second)})
	store.Set("b", stores.TTL{Value: "val2", Expiration: time.Now().Add(10 * time.Second)})

	if val1, err1 := store.Get("a"); err1 != nil || val1 != "val1" {
		t.Errorf("Expected value: \"val1\", found: %v", val1)
	}
	if val2, err2 := store.Get("b"); err2 != nil || val2 != "val2" {
		t.Errorf("Expected value: \"val2\", found: %v", val2)
	}
	if val3, err3 := store.Get("c"); err3 == nil || val3 != "" {
		t.Errorf("Expected value: \"\" and exists: false, found: %v, %v", val3, err3)
	}
}

func TestKTTLeys_checkLen(t *testing.T) {
	store := stores.NewTTLStore(0)
	store.Set("a", stores.TTL{Value: "val1", Expiration: time.Now().Add(10 * time.Second)})
	store.Set("b", stores.TTL{Value: "val2", Expiration: time.Now().Add(10 * time.Second)})
	if length := store.Len(); length != 2 {
		t.Errorf("Expected 2 elements, Found %v", length)
	}
}

func TestKTTLeys_checkPop(t *testing.T) {
	store := stores.NewTTLStore(0)
	store.Set("a", stores.TTL{Value: "val1", Expiration: time.Now().Add(10 * time.Second)})
	poped, err := store.Pop("a")
	if err != nil || poped != "val1" {
		t.Errorf("Expected val1, found %v", poped)
	}
	if val, getErr := store.Get("a"); getErr == nil || val != "" {
		t.Errorf("Expected value: \"\", found: %v,", val)
	}
	if _, popErr := store.Pop("a"); popErr == nil {
		t.Errorf("Should't be Poped")
	}

}

func TestKeysTTL_checkRename(t *testing.T) {
	store := stores.NewTTLStore(0)
	store.Set("a", stores.TTL{Value: "val", Expiration: time.Now().Add(10 * time.Second)})
	if err := store.Rename("a", "b"); err != nil {
		t.Errorf("%s", err.Error())
	}
	if store.Rename("s", "a") == nil {
		t.Errorf("Expected Rename to return an error")
	}
}

func TestTTLClear(t *testing.T) {
	store := stores.NewTTLStore(0)
	store.Set("a", stores.TTL{Value: "val", Expiration: time.Now().Add(10 * time.Second)})
	store.Clear()
	if length := store.Len(); length != 0 {
		t.Errorf("Expected Len to be 0 because it's cleared, found len = %v", length)
	}
}

func TestBackgroundCleanup(t *testing.T) {
	// Janitor runs every 1 second
	store := stores.NewTTLStore(time.Second)
	defer store.Close()

	// Insert a key that expires in 30ms
	store.Set("ephemeral", stores.TTL{
		Value:      "temp",
		Expiration: time.Now().Add(40 * time.Millisecond),
	})

	// Wait 60ms so both expiration and ticker pass
	time.Sleep(2 * time.Second)

	// Check underlying raw map directly (bypass lazy Get)
	_, err := store.Get("ephemeral")

	if err == nil {
		t.Errorf("expected background janitor to delete expired key from map")
	}
}
