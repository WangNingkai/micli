# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

MiCLI - 小米云服务 CLI 工具，用于控制 MIoT 设备、小爱音箱（Mina）和米家设备。支持设备属性读写、TTS 语音合成、AI 对话等功能。

## Build & Run

```bash
# 构建
go build -o micli

# 运行命令
go run main.go [command]

# 示例：列出设备
go run main.go list

# 示例：小爱音箱 TTS
go run main.go mina tts -d <device_id> --text "你好"

# 测试
go test ./...
go test -race ./...
```

## Architecture

```
main.go → cmd.Execute() (Cobra)
           │
           ├─ cmd/          # Cobra 命令定义
           │   ├─ root.go   # 初始化配置 + 服务实例
           │   ├─ mina*.go  # 小爱音箱命令
           │   └─ props_*.go # MIoT 属性操作
           │
           ├─ pkg/
           │   ├─ miservice/ # 小米 API 核心
           │   │   ├─ service.go  # 登录/认证/请求
           │   │   ├─ mina.go     # 小爱音箱 API
           │   │   └─ io.go       # MIoT/MiIO raw
           │   ├─ jarvis/    # ChatGPT 集成
           │   └─ tts/       # Edge TTS
           │
           └─ internal/
               ├─ app.go     # Gin 服务器 (mina serve)
               └─ conf/      # 配置管理
```

## Configuration

首次运行自动创建 `conf.ini`，交互式配置：
- MI_USER / MI_PASS - 小米账号
- REGION - 区域 (cn/de/us/i2/ru/sg/tw)
- MI_DID - 默认设备 DID
- OpenAI KEY/BASE_URL - AI 功能（可选）

Token 缓存：`~/.mi.token`

## Adding Commands

1. 在 `cmd/` 创建文件，定义 Cobra command
2. 在 `cmd/root.go` 的 `init()` 中注册：`rootCmd.AddCommand(newCmd)`
3. 使用全局服务：`ms`（通用）、`ioSrv`（MIoT）、`minaSrv`（小爱）

## Key Patterns

### Xiaomi API 认证流程
```
Service.login(sid) → 获取 serviceToken → Service.Request()
```

### 小爱设备选择
```go
// 交互式选择设备
deviceId, err := chooseMinaDevice(minaSrv)
// 或获取详细信息
device, err := chooseMinaDeviceDetail(minaSrv, deviceId)
```

### MIoT 属性操作
```go
// 获取属性
res, err := ioSrv.MiotGetProps(did, props)

// 设置属性
res, err := ioSrv.MiotSetProps(did, props, values)
```

## Dependencies

| Package | 用途 |
|---------|------|
| spf13/cobra | CLI 框架 |
| gin-gonic/gin | Web 服务器 |
| sashabaranov/go-openai | ChatGPT API |
| pterm/pterm | 终端交互 UI |
| gopkg/ini.v1 | 配置解析 |
| gorilla/websocket | 实时通信 |