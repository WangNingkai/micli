# MiCLI

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**将小米云服务带入命令行。**

MiCLI 是一个强大的 CLI 工具，用于控制 MIoT 设备、小爱音箱（Mina）和米家设备。支持设备属性读写、TTS 语音合成、AI 对话等功能。

## 功能特性

- 🏠 **MIoT 设备控制** - 读写设备属性
- 🔊 **小爱音箱** - TTS 语音合成、播放控制、语音记录
- 🤖 **AI 集成** - ChatGPT 智能对话
- 🎙️ **Edge TTS** - 微软 Edge TTS，无需 API Key
- 📱 **二维码登录** - 使用米家 App 扫码认证
- 🌐 **Web 服务** - 内置 Gin 服务器，扩展更多功能
- ⚙️ **自动化命令** - 基于关键词的命令链，通过 `commands.json` 配置

## 安装

### 从源码构建

```bash
git clone https://github.com/WangNingkai/micli.git
cd micli
go build -o micli
```

### 环境要求

- Go 1.25 或更高版本
- 小米账号

## 快速开始

### 1. 初始化配置

首次运行任意命令，将自动创建 `conf.ini` 配置文件：

```bash
./micli list
```

配置项说明：
- `MI_USER` / `MI_PASS` - 小米账号密码
- `REGION` - 区域 (cn/de/us/i2/ru/sg/tw)
- `MI_DID` - 默认设备 DID
- OpenAI 配置（可选）

### 2. 列出设备

```bash
./micli list
```

### 3. 设备属性操作

```bash
# 获取设备属性
./micli get -d <device_id> --props <属性1>,<属性2>

# 设置设备属性
./micli set -d <device_id> --props <属性1>=<值>
```

属性格式：`siid-piid`（如 `2-1` 表示 service 2, property 1）

### 4. MIoT 动作

```bash
# 执行 MIoT 动作
./micli action <device_id> <siid-aiid> [参数...]
```

### 5. 小爱音箱

```bash
# TTS 语音合成
./micli mina tts -d <device_id> --text "你好世界"

# 列出小爱设备
./micli mina list

# 播放控制（播放/暂停/音量）
./micli mina player <command>

# 获取语音记录
./micli mina records [limit]

# 设置默认小爱设备
./micli mina set_did

# 启动 AI 对话模式
./micli mina serve
```

### 6. 文本转语音

```bash
# 交互式选择语音
./micli tts -t "你好，世界"

# 指定语音
./micli tts -t "Hello" -v en-US-JennyNeural
```

### 7. 二维码登录

```bash
# 生成二维码，使用米家 App 扫码登录
./micli qr-login
```

## 命令参考

| 命令 | 说明 |
|------|------|
| `list` | 列出所有设备 |
| `get` | 获取 MIoT 设备属性 |
| `set` | 设置 MIoT 设备属性 |
| `action <iid> [args]` | 执行 MIoT 动作 |
| `spec [model]` | 查看 MIoT 规范 |
| `decode` | 解码 MIoT 加密数据 |
| `mina` | 小爱音箱命令 |
| `mina list` | 列出小爱设备 |
| `mina tts` | 小爱 TTS |
| `mina player` | 播放控制（播放/暂停/音量） |
| `mina records [limit]` | 获取语音对话记录 |
| `mina serve` | 启动 Web 服务 + AI 对话 |
| `mina set_did` | 设置默认小爱设备 DID |
| `tts -t <text>` | Edge TTS 合成 |
| `miot_raw <cmd> <params>` | 原始 MIoT API 调用 |
| `miio_raw <uri> <data>` | 原始 MiIO API 调用 |
| `set_did` | 设置默认 MIoT 设备 DID |
| `qr-login` | 二维码认证登录 |
| `reset` | 重置配置 |

## 自动化

创建 `commands.json` 文件（参考 `commands.sample.json`）定义基于关键词的自动化命令链：

```json
[
  {
    "keyword": "天气怎么样",
    "step": [
      {
        "type": "request",
        "request": {
          "url": "http://www.weather.com.cn/data/sk/101010100.html",
          "out": "res.data",
          "wait": true
        }
      }
    ]
  },
  {
    "keyword": "开启高级对话",
    "type": "chat",
    "chat": {
      "endFlag": "关闭高级对话",
      "stream": true,
      "useEdgeTTS": false
    }
  }
]
```

步骤类型：`tts`、`request`、`action`、`chat`（ChatGPT）

## MIoT 规范

在 [miot-spec](https://home.miot-spec.com/) 查找设备规范。

## 配置文件

配置文件路径：`conf.ini`

```ini
[app]
DEBUG = false
PORT = :8080

[account]
MI_USER = your_email@example.com
MI_PASS = your_password
REGION = cn
MI_DID =

[mina]
DID = your_device_did

[openai]
KEY = sk-xxx
BASE_URL = https://api.openai.com/v1

[file]
TRANSFER_SH = https://transfer.sh
```

Token 缓存：`~/.mi.token`

## 项目结构

```
main.go → cmd.Execute() (Cobra)
           │
           ├─ cmd/               # Cobra 命令
           │   ├─ root.go        # 配置和服务初始化
           │   ├─ mina*.go       # 小爱音箱命令
           │   ├─ props_*.go     # MIoT 属性操作
           │   ├─ action.go      # MIoT 动作执行
           │   ├─ spec.go        # MIoT 规范查看
           │   ├─ miot_raw.go    # 原始 MIoT API
           │   ├─ miio_raw.go    # 原始 MiIO API
           │   ├─ tts.go         # Edge TTS 命令
           │   ├─ decode.go      # MIoT 数据解码
           │   ├─ set_did.go     # 设置默认 DID
           │   ├─ qr_login.go    # 二维码登录
           │   └─ reset.go       # 重置配置
           │
           ├─ pkg/
           │   ├─ miservice/     # 小米 API 核心
           │   │   ├─ service.go # 登录/认证/请求
           │   │   ├─ mina.go    # 小爱音箱 API
           │   │   ├─ io.go      # MIoT/MiIO 操作
           │   │   ├─ token.go   # Token 存储
           │   │   └─ qrlogin.go # 二维码登录
           │   ├─ jarvis/        # ChatGPT 集成
           │   ├─ tts/           # Edge TTS
           │   └─ util/          # 工具函数（加签/加密）
           │
           ├─ internal/
           │   ├─ app.go         # Gin 服务器
           │   ├─ conf/          # 配置管理
           │   ├─ handlers/      # HTTP 处理器
           │   ├─ middleware/    # CORS 中间件
           │   └─ static/        # 静态文件
           │
           ├─ public/            # 嵌入式前端 (embed.FS)
           │   └─ dist/          # 前端构建产物
           │
           └─ data/              # 本地缓存
               ├─ devices.json   # 设备列表缓存
               └─ miot-spec.json # MIoT 规范缓存
```

## 开发

```bash
# 运行
go run main.go [command]

# 构建
go build -o micli

# 测试
go test ./...

# 竞态检测
go test -race ./...

# 静态检查
golangci-lint run
```

## CI/CD

GitHub Actions 在版本标签触发时（排除 `-alpha*`）构建 Linux (amd64/arm64)、macOS、Windows 三平台二进制文件。

## 贡献

欢迎提交 Pull Request！

## 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件。

## 致谢

- [Cobra](https://github.com/spf13/cobra) - CLI 框架
- [Gin](https://github.com/gin-gonic/gin) - Web 框架
- [go-openai](https://github.com/sashabaranov/go-openai) - OpenAI API 客户端
- [pterm](https://github.com/pterm/pterm) - 终端 UI
- [req](https://github.com/imroc/req) - HTTP 客户端
- [Edge TTS](https://github.com/microsoft/edge-tts) - 微软 Edge TTS

---

**免责声明：** 本项目仅供学习交流使用，请勿用于非法用途，后果自负。
