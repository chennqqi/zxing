package zxing

import (
	"os"
	"strconv"
)

// Backend 后端类型
type Backend string

const (
	// BackendCGO CGO 后端
	BackendCGO Backend = "cgo"

	// BackendWASM WASM 后端
	BackendWASM Backend = "wasm"

	// BackendAuto 自动选择后端
	BackendAuto Backend = "auto"
)

// Config ZXing 配置
type Config struct {
	// Backend 指定使用的后端
	Backend Backend

	// WASMPath WASM 文件路径（仅 WASM 后端使用）
	WASMPath string

	// Debug 是否启用调试模式
	Debug bool

	// Timeout 操作超时时间（秒）
	Timeout int
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Backend:  BackendAuto,
		WASMPath: "wasm/zxingwrapper.wasm",
		Debug:    false,
		Timeout:  30,
	}
}

// LoadConfigFromEnv 从环境变量加载配置
func LoadConfigFromEnv() *Config {
	config := DefaultConfig()

	// 读取后端配置
	if backend := os.Getenv("ZXING_BACKEND"); backend != "" {
		config.Backend = Backend(backend)
	}

	// 读取 WASM 路径
	if wasmPath := os.Getenv("ZXING_WASM_PATH"); wasmPath != "" {
		config.WASMPath = wasmPath
	}

	// 读取调试模式
	if debug := os.Getenv("ZXING_DEBUG"); debug != "" {
		if debugBool, err := strconv.ParseBool(debug); err == nil {
			config.Debug = debugBool
		}
	}

	// 读取超时时间
	if timeout := os.Getenv("ZXING_TIMEOUT"); timeout != "" {
		if timeoutInt, err := strconv.Atoi(timeout); err == nil {
			config.Timeout = timeoutInt
		}
	}

	return config
}
