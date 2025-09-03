package zxing

import (
	"fmt"
	"runtime"
)

// New 创建新的 ZXing 实例
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

// NewCGO 创建 CGO 后端实例
func NewCGO(config *Config) (ZXing, error) {
	if config == nil {
		config = DefaultConfig()
	}
	
	impl := &universalZXing{
		backend: BackendCGO,
		config:  config,
	}
	
	return impl, nil
}

// NewWASM 创建 WASM 后端实例
func NewWASM(config *Config) (ZXing, error) {
	if config == nil {
		config = DefaultConfig()
	}
	
	impl := &universalZXing{
		backend: BackendWASM,
		config:  config,
	}
	
	return impl, nil
}

// newAuto 自动选择后端
func newAuto(config *Config) (ZXing, error) {
	// 根据运行环境自动选择后端
	if runtime.GOOS == "js" && runtime.GOARCH == "wasm" {
		// 在 WASM 环境中优先使用 WASM 后端
		return NewWASM(config)
	}
	
	// 在其他环境中优先使用 CGO 后端
	return NewCGO(config)
}