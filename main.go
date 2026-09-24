package main

import (
	"fmt"
	"github.com/NooRMaseR/go-redis/stores"
)

func main() {
	store := stores.NewStore()
	store.Set("a", "val1")
	store.Set("b", "val2")
	val, err := store.Get("a")
	if err == nil {
		fmt.Printf("Value: %s\n", val)
	}
}
