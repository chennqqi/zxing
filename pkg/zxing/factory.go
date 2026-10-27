package zxing

import (
	"fmt"
)

// New creates a new ZXing instance with the given configuration.
// When Backend is Auto, the backend is selected at compile time:
// - CGO is used when CGO_ENABLED=1 on linux, windows or darwin
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
// CGO requires CGO_ENABLED=1 on linux, windows or darwin with precompiled
// static libraries in lib/{os}-{arch}/. Run 'go run ./cmd/build build-lib'
// or use Docker ('docker/Dockerfile.linux-build' or 'docker/Dockerfile.win-build')
// to rebuild libraries if missing.
func NewCGO(config *Config) (ZXing, error) {
	if config == nil {
		config = DefaultConfig()
	}
	if !cgoAvailable {
		return nil, fmt.Errorf("CGO backend is not available (requires CGO_ENABLED=1 on linux, windows or darwin with precompiled static libraries in lib/{os}-{arch}/)")
	}
	return &cgoZXing{config: config}, nil
}

// NewWASM creates a WASM (wazero) backend instance.
// The wazero backend is available on all platforms except GOOS=js GOARCH=wasm,
// including when CGO is enabled. Returns an error if the WASM runtime cannot
// be initialized.
func NewWASM(config *Config) (ZXing, error) {
	if config == nil {
		config = DefaultConfig()
	}
	return &wasmZXing{config: config}, nil
}

// newAuto selects the backend at compile time based on the cgoAvailable constant.
// cgoAvailable is true when CGO is enabled on linux/windows/darwin (see cgo_impl.go),
// and false otherwise (see cgo_stub.go).
func newAuto(config *Config) (ZXing, error) {
	if cgoAvailable {
		return NewCGO(config)
	}
	return NewWASM(config)
}
