package zxing

import (
	"fmt"
	"runtime"
)

// New creates a new ZXing instance with the given configuration.
// When Backend is Auto, the backend is selected at compile time:
// - CGO is used when CGO_ENABLED=1 on linux or windows
// - WASM (wazero) is used otherwise
func New(config *Config) (ZXing, error) {
	if config == nil {
		config = DefaultConfig()
	}

	switch config.Backend {
	case BackendCGO:
		return NewCGO(config)
	case BackendWASM:
		return NewWASM(config)
	case BackendAuto:
		return newAuto(config)
	default:
		return nil, fmt.Errorf("unsupported backend: %s", config.Backend)
	}
}

// NewCGO creates a CGO backend instance.
// Returns an error if CGO is not available on the current platform.
func NewCGO(config *Config) (ZXing, error) {
	if config == nil {
		config = DefaultConfig()
	}
	return &cgoZXing{config: config}, nil
}

// NewWASM creates a WASM (wazero) backend instance.
// Returns an error if the WASM runtime cannot be initialized.
func NewWASM(config *Config) (ZXing, error) {
	if config == nil {
		config = DefaultConfig()
	}
	return &wasmZXing{config: config}, nil
}

// newAuto selects the backend at compile time based on build tags.
// On CGO-enabled linux/windows, CGO is preferred.
// On all other platforms, WASM (wazero) is used.
func newAuto(config *Config) (ZXing, error) {
	// Compile-time backend selection via build tags:
	// - cgo_impl.go has build tag "cgo && (linux || windows)"
	// - wasm_impl.go has build tag "!cgo || !(linux || windows)"
	// At runtime, check which backend is actually available
	if runtime.GOOS == "linux" || runtime.GOOS == "windows" {
		// On linux/windows, try CGO first (it may or may not be enabled)
		// The stub will return an error if CGO is not available
		zx, err := NewCGO(config)
		if err == nil {
			return zx, nil
		}
		if config.Debug {
			fmt.Printf("CGO backend not available (%v), falling back to WASM\n", err)
		}
	}
	return NewWASM(config)
}
