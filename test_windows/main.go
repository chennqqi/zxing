package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestWindowsEnvironment 测试Windows环境
func TestWindowsEnvironment(t *testing.T) {
	fmt.Printf("操作系统: %s\n", runtime.GOOS)
	fmt.Printf("架构: %s\n", runtime.GOARCH)
	fmt.Printf("Go版本: %s\n", runtime.Version())

	if runtime.GOOS != "windows" {
		t.Errorf("期望在Windows环境下运行，当前系统: %s", runtime.GOOS)
	}
}

// TestFileSystem 测试文件系统访问
func TestFileSystem(t *testing.T) {
	// 测试当前目录
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("获取当前目录失败: %v", err)
	}
	fmt.Printf("当前目录: %s\n", currentDir)

	// 测试目录列表
	files, err := os.ReadDir(currentDir)
	if err != nil {
		t.Fatalf("读取目录失败: %v", err)
	}

	fmt.Printf("目录中的文件数量: %d\n", len(files))
	for i, file := range files {
		if i < 5 { // 只显示前5个文件
			fmt.Printf("  %s\n", file.Name())
		}
	}
}

// TestPathSeparator 测试路径分隔符
func TestPathSeparator(t *testing.T) {
	expected := '\\'
	if filepath.Separator != expected {
		t.Errorf("期望路径分隔符为 %c，实际为 %c", expected, filepath.Separator)
	}
	fmt.Printf("路径分隔符: %c\n", filepath.Separator)
}

// TestEnvironmentVariables 测试环境变量
func TestEnvironmentVariables(t *testing.T) {
	envVars := []string{"PATH", "GOPATH", "GOROOT", "TEMP"}

	for _, envVar := range envVars {
		value := os.Getenv(envVar)
		if value != "" {
			fmt.Printf("%s: %s\n", envVar, value)
		} else {
			fmt.Printf("%s: (未设置)\n", envVar)
		}
	}
}

func main() {
	fmt.Println("=== Windows环境测试 ===")

	// 运行测试
	tests := []testing.InternalTest{
		{"TestWindowsEnvironment", TestWindowsEnvironment},
		{"TestFileSystem", TestFileSystem},
		{"TestPathSeparator", TestPathSeparator},
		{"TestEnvironmentVariables", TestEnvironmentVariables},
	}

	testing.Main(func(pat, str string) (bool, error) { return true, nil },
		tests,
		nil,
		nil)
}
