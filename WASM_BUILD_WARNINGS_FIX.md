# WASM 构建警告修复说明

## 修复的警告

### 1. SHARED Library Warning
**警告信息**:
```
CMake Warning (dev) at zxing-cpp/core/CMakeLists.txt:493 (add_library):
  ADD_LIBRARY called with SHARED option but the target platform does not
  support dynamic linking.  Building a STATIC library instead.  This may lead
  to problems.
```

**原因**: WASM 平台不支持动态链接，但 zxing-cpp 的 CMakeLists.txt 默认尝试构建 SHARED 库。

**修复**:
- 在 `CMakeLists-wasm.txt` 中添加 `set(BUILD_SHARED_LIBS OFF ...)` 强制使用静态库
- 在所有构建脚本中添加 `-DBUILD_SHARED_LIBS=OFF` CMake 选项
- 添加 `-Wno-dev` 选项来抑制开发者警告

### 2. stb Module Not Found Warning
**警告信息**:
```
-- Checking for module 'stb'
--   Package 'stb', required by 'virtual:world', not found
```

**原因**: zxing-cpp 尝试通过 pkg-config 查找 stb 库，但 stb 通常是作为头文件包含的，不需要通过 pkg-config 查找。

**修复**:
- 在 `CMakeLists-wasm.txt` 中禁用 ZXingCPP 的示例、测试和 Python 绑定
- 这些选项会避免查找不必要的依赖（如 stb, Qt, OpenCV）
- 添加的选项：
  - `ZXING_EXAMPLES=OFF`
  - `ZXING_BLACKBOX_TESTS=OFF`
  - `ZXING_UNIT_TESTS=OFF`
  - `ZXING_INSTALL_TEST=OFF`
  - `ZXING_PYTHON_MODULE=OFF`

## 修改的文件

1. **CMakeLists-wasm.txt**
   - 添加 `BUILD_SHARED_LIBS=OFF` 设置
   - 添加 ZXingCPP 选项禁用配置

2. **build_wasm.sh** (Linux)
   - 添加 CMake 选项来消除警告

3. **build_wasm.ps1** (Windows)
   - 添加 CMake 选项来消除警告

4. **scripts/build_wasm_save.sh** (Linux)
   - 添加 CMake 选项来消除警告

5. **scripts/build_wasm_save.ps1** (Windows)
   - 添加 CMake 选项来消除警告

## 使用方法

现在运行构建脚本时，警告应该被消除或抑制：

```bash
# Linux
./build_wasm.sh

# 或使用脚本目录中的脚本
./scripts/build_wasm_save.sh
```

```powershell
# Windows
.\build_wasm.ps1

# 或使用脚本目录中的脚本
.\scripts\build_wasm_save.ps1
```

## 注意事项

- 这些警告不影响构建结果，只是信息性的
- `-Wno-dev` 选项会抑制所有开发者警告，包括其他可能的警告
- 如果需要在构建中看到所有警告，可以移除 `-Wno-dev` 选项
