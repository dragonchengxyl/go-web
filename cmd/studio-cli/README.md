# Studio CLI

命令行工具，用于管理游戏发布和上传到 Studio 平台。

## 功能特性

- 🔐 API 认证管理
- 📦 游戏包分片上传（支持断点续传）
- 📊 实时上传进度显示
- ✅ SHA256 完整性校验
- 🚀 自动发布游戏版本

## 安装

### 从源码构建

```bash
cd cmd/studio-cli
go build -o studio-cli
```

### 安装到系统

```bash
go install github.com/studio/platform/cmd/studio-cli@latest
```

## 使用指南

### 1. 登录

首次使用需要登录并保存凭证：

```bash
studio-cli login \
  --email admin@studio.com \
  --password your-password \
  --api-url http://localhost:8080
```

凭证将保存在 `~/.studio-cli/credentials.json`

### 2. 发布游戏版本

```bash
studio-cli publish \
  --game thunder \
  --branch main \
  --version v1.2.3 \
  --title "Thunder - Daily Update" \
  --changelog ./CHANGELOG.md \
  --package ./dist/thunder-v1.2.3-windows.zip \
  --platform windows \
  --auto-publish
```

#### 参数说明

- `--game`: 游戏 slug 标识符（必需）
- `--branch`: 分支名称（main/beta/experimental，默认 main）
- `--version`: 版本号（必需，如 v1.2.3）
- `--title`: 发布标题（可选）
- `--changelog`: 更新日志文件路径（可选）
- `--package`: 游戏包文件路径（必需）
- `--platform`: 目标平台（windows/macos/linux，默认 windows）
- `--auto-publish`: 上传后自动发布（可选）

### 3. 登出

```bash
studio-cli logout
```

## 上传机制

### 分片上传

- 文件自动分割为 5MB 的块
- 最多 3 个并发上传
- 支持断点续传（上传状态保存在 `~/.studio-cli/uploads/`）
- 自动重试失败的块

### 进度显示

```
📦 File size: 245.30 MB
📊 Total chunks: 50
Uploading [████████████████████] 100% (245.3 MB / 245.3 MB)
```

### 完整性校验

上传前自动计算 SHA256 校验和，确保文件完整性。

## 示例工作流

### 发布新版本

```bash
# 1. 构建游戏
./build-game.sh

# 2. 打包
zip -r thunder-v1.2.3-windows.zip dist/

# 3. 发布
studio-cli publish \
  --game thunder \
  --version v1.2.3 \
  --package thunder-v1.2.3-windows.zip \
  --auto-publish
```

### 多平台发布

```bash
# Windows
studio-cli publish --game thunder --version v1.2.3 \
  --package thunder-v1.2.3-windows.zip --platform windows

# macOS
studio-cli publish --game thunder --version v1.2.3 \
  --package thunder-v1.2.3-macos.zip --platform macos

# Linux
studio-cli publish --game thunder --version v1.2.3 \
  --package thunder-v1.2.3-linux.tar.gz --platform linux
```

## 配置文件

### 凭证文件

位置：`~/.studio-cli/credentials.json`

```json
{
  "api_url": "http://localhost:8080",
  "access_token": "eyJhbGc...",
  "email": "admin@studio.com"
}
```

### 上传缓存

位置：`~/.studio-cli/uploads/`

存储上传进度，支持断点续传。

## 故障排除

### 登录失败

```bash
# 检查 API 地址是否正确
curl http://localhost:8080/health

# 检查凭证
cat ~/.studio-cli/credentials.json
```

### 上传失败

```bash
# 清除上传缓存重试
rm -rf ~/.studio-cli/uploads/*

# 检查文件是否存在
ls -lh ./dist/thunder-v1.2.3-windows.zip
```

### 权限错误

确保使用管理员账号登录，普通用户无法发布游戏版本。

## 开发

### 运行测试

```bash
go test ./internal/cli/...
```

### 构建所有平台

```bash
# Windows
GOOS=windows GOARCH=amd64 go build -o studio-cli.exe

# macOS
GOOS=darwin GOARCH=amd64 go build -o studio-cli-macos

# Linux
GOOS=linux GOARCH=amd64 go build -o studio-cli-linux
```

## 许可证

MIT License
