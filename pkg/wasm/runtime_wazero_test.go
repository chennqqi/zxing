//go:build !cgo || !(linux || windows)

package wasm

import (
	"context"
	"testing"
)

func TestWazeroLoadAndDecode(t *testing.T) {
	rt := NewRuntime()
	err := rt.Initialize(context.Background(), "../../wasm/zxingwrapper.wasm")
	if err != nil {
		t.Fatalf("Failed to initialize wazero runtime: %v", err)
	}
	defer rt.Close()

	if !rt.IsReady() {
		t.Fatal("Runtime should be ready after Initialize")
	}
}
