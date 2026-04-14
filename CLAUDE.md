# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

MiCLI - 小米云服务 CLI 工具，用于控制 MIoT 设备、小爱音箱（Mina）和米家设备。支持设备属性读写、TTS 语音合成、AI 对话等功能。

- **Language**: Go 1.25.0
- **CLI Framework**: Cobra
- **Web Framework**: Gin
- **License**: MIT

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

# 静态检查
golangci-lint run
```

## Architecture

```
main.go → cmd.Execute() (Cobra)
           │
           ├─ cmd/               # Cobra 命令定义
           │   ├─ root.go        # 初始化配置 + 全局服务实例 (ms, ioSrv, minaSrv)
           │   ├─ mina*.go       # 小爱音箱命令 (serve/tts/player/records/list/set_did)
           │   ├─ props_*.go     # MIoT 属性操作 (get/set)
           │   ├─ action.go      # MIoT 动作执行
           │   ├─ spec.go        # MIoT 规范查看
           │   ├─ miot_raw.go    # MIoT 原始 API
           │   ├─ miio_raw.go    # MiIO 原始 API
           │   ├─ tts.go         # Edge TTS 独立命令
           │   ├─ decode.go      # MIoT 数据解码
           │   ├─ set_did.go     # 设置默认 MIoT DID
           │   ├─ qr_login.go    # 二维码登录
           │   └─ reset.go       # 重置配置
           │
           ├─ pkg/
           │   ├─ miservice/     # 小米 API 核心
           │   │   ├─ service.go     # 登录/认证/请求
           │   │   ├─ mina.go        # 小爱音箱 API
           │   │   ├─ io.go          # MIoT/MiIO 操作
           │   │   ├─ token.go       # Token 存储
           │   │   └─ qrlogin.go     # 二维码登录
           │   ├─ jarvis/        # ChatGPT 集成
           │   │   ├─ jarvis.go      # 接口定义
           │   │   └─ chatgpt.go     # OpenAI 客户端
           │   ├─ tts/           # Edge TTS
           │   │   ├─ edge_tts.go    # 封装
           │   │   └─ edgetts/       # 核心实现
           │   └─ util/          # 工具函数
           │       ├─ util.go        # 加签/加密/字符串处理
           │       └─ log.go         # Logrus 日志
           │
           ├─ internal/
           │   ├─ app.go         # Gin 服务器 (mina serve)
           │   ├─ conf/          # 配置管理
           │   ├─ handlers/      # HTTP 处理器
           │   ├─ middleware/    # CORS 中间件
           │   └─ static/        # 静态文件
           │
           ├─ public/            # 嵌入式前端 (embed.FS)
           │   └─ dist/          # 构建产物
           │
           └─ data/              # 本地缓存
               ├─ devices.json   # 设备列表缓存
               └─ miot-spec.json # MIoT 规范缓存
```

## Configuration

首次运行自动创建 `conf.ini`，交互式配置：
- MI_USER / MI_PASS - 小米账号
- REGION - 区域 (cn/de/us/i2/ru/sg/tw)
- MI_DID - 默认设备 DID
- OpenAI KEY/BASE_URL - AI 功能（可选）

Token 缓存：`~/.mi.token`

自动化命令：`commands.json`（基于关键词触发 TTS/HTTP/Action/Chat 步骤链）

## Adding Commands

1. 在 `cmd/` 创建文件，定义 Cobra command
2. 在 `cmd/root.go` 的 `init()` 中注册：`rootCmd.AddCommand(newCmd)`
3. 使用全局服务：`ms`（通用）、`ioSrv`（MIoT）、`minaSrv`（小爱）
4. 使用 `handleResult(res, err)` 统一处理输出

## Key Patterns

### Xiaomi API 认证流程
```
QRLogin() 或 passwordLogin() → serviceToken → 缓存到 ~/.mi.token → Service.Request() 带签名
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

// 执行动作
res, err := ioSrv.MiotAction(did, []int{siid, aiid}, inList)
```

### MIoT 属性格式
- 属性: `siid-piid`（如 `2-1` 表示 service 2, property 1）
- 动作: `siid-aiid`（如 `5-4`）
- 空值: `#NA`，布尔: `#1`/`#0`

### Mina Serve 模式（AI 对话）
`mina serve` 是最复杂的功能：轮询小爱语音记录 → 匹配 `commands.json` 关键词 → 执行步骤链（tts/request/action/chat）。支持 ChatGPT 流式响应 + Edge TTS 语音播报。

### 硬件命令映射
不同小爱型号有不同的 MIoT 动作 IID，定义在 `cmd/root.go` 的 `HardwareCommandDict` 中。新增型号时在此处添加映射。

### 请求签名
`pkg/util/util.go` 中的 `SignData()` 使用 HMAC-SHA256 生成 `_nonce`、`data`、`signature`。

## Dependencies

| Package | 用途 |
|---------|------|
| spf13/cobra | CLI 框架 |
| gin-gonic/gin | Web 服务器 |
| sashabaranov/go-openai | ChatGPT API |
| pterm/pterm | 终端交互 UI |
| gopkg/ini.v1 | 配置解析 |
| gorilla/websocket | 实时通信 |
| imroc/req/v3 | HTTP 客户端 |
| json-iterator/go | JSON 序列化 |
| tidwall/gjson | JSON 路径查询 |
| samber/lo | 切片工具 |
| mdp/qrterminal/v3 | 终端二维码 |

## CI/CD

GitHub Actions 按标签触发（排除 `-alpha*`），构建 Linux (amd64/arm64)、macOS、Windows 三平台二进制，使用 `softprops/action-gh-release` 发布 draft release。