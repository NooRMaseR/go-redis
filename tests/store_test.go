package tests

import (
	"fmt"
	"reflect"
	"sync"
	"testing"
	"slices"
	"github.com/NooRMaseR/go-redis/stores"
)

func TestKeys_getKeys(t *testing.T) {
	store := stores.NewStore()
	store.Set("a", "val", 0)
	store.Set("b", "val", 0)
	keys := store.Keys()

	slices.Sort(keys)
	expected := []string{"a","b"}
	if !reflect.DeepEqual(keys, expected) {
		t.Errorf("Expected keys %v, got %v", expected, keys)
	}
}

func TestKeys_checkValues(t *testing.T) {
	store := stores.NewStore()
	store.Set("a", "val1", 0)
	store.Set("b", "val2", 0)

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

func TestKeys_checkLen(t *testing.T) {
	store := stores.NewStore()
	store.Set("a", "val1", 0)
	store.Set("b", "val2", 0)
	if length := store.Len(); length != 2 {
		t.Errorf("Expected 2 elements, Found %v", length)
	}
}

func TestKeys_checkPop(t *testing.T) {
	store := stores.NewStore()
	store.Set("a", "val1", 0)
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

func TestKeys_checkRename(t *testing.T) {
	store := stores.NewStore()
	store.Set("a", "val", 0)
	if err := store.Rename("a", "b"); err != nil {
		t.Errorf("%s", err.Error())
	}
	if store.Rename("s", "a") == nil {
		t.Errorf("Expected Rename to return an error")
	}
}

func TestClear(t *testing.T) {
	store := stores.NewStore()
	store.Set("a", "val", 0)
	store.Clear()
	if length := store.Len(); length != 0 {
		t.Errorf("Expected Len to be 0 because it's cleared, found len = %v", length)
	}
}

func TestDelete(t *testing.T) {
	store := stores.NewStore()
	store.Set("key1", "val1", 0)
	store.Set("key2", "val2", 0)

	if store.Len() != 2 {
		t.Fatalf("expected store length to be 2, got %d", store.Len())
	}

	store.Delete("key1")

	// Verify key1 is gone
	if val, err := store.Get("key1"); err == nil || val != "" {
		t.Errorf("expected key1 to be deleted, but found val: %q, err: %v", val, err)
	}

	// Verify the length decreased
	if store.Len() != 1 {
		t.Errorf("expected length to be 1 after deletion, got %d", store.Len())
	}

	// Verify other keys were not touched
	if val, err := store.Get("key2"); err != nil || val != "val2" {
		t.Errorf("expected key2 to remain unaffected, got val: %q, err: %v", val, err)
	}

	// 3. Delete a non-existent key (should not panic or alter store)
	store.Delete("non_existent_key")

	if store.Len() != 1 {
		t.Errorf("store length should remain 1 after deleting non-existent key, got %d", store.Len())
	}
}

func TestConcurrentAccess(t *testing.T) {
    store := stores.NewStore()
    var wg sync.WaitGroup
    iterations := 500

    // Concurrently write
    for i := range iterations {
        wg.Add(1)
        go func(n int) {
            defer wg.Done()
            k := fmt.Sprintf("k%d", n)
            store.Set(k, "val", 0)
        }(i)
    }

    // Concurrently read & check
    for i := range iterations {
        wg.Add(1)
        go func(n int) {
            defer wg.Done()
            k := fmt.Sprintf("k%d", n)
            _, _ = store.Get(k)
            _ = store.Len()
            _ = store.Keys()
        }(i)
    }

    // Concurrently pop & delete
    for i := 0; i < iterations/2; i++ {
        wg.Add(1)
        go func(n int) {
            defer wg.Done()
            k := fmt.Sprintf("k%d", n)
            _, _ = store.Pop(k)
        }(i)
    }

    wg.Wait()
}