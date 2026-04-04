# MiCLI

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**将小米云服务带入命令行。**

MiCLI 是一个强大的 CLI 工具，用于控制 MIoT 设备、小爱音箱（Mina）和米家设备。支持设备属性读写、TTS 语音合成、AI 对话等功能。

## 功能特性

- 🏠 **MIoT 设备控制** - 读写设备属性
- 🔊 **小爱音箱** - TTS 语音合成、播放控制、语音记录
- 🤖 **AI 集成** - ChatGPT 智能对话
- 🎙️ **Edge TTS** - 微软 Edge TTS，无需 API Key
- 🌐 **Web 服务** - 内置 Gin 服务器，扩展更多功能

## 安装

### 从源码构建

```bash
git clone https://github.com/WangNingkai/micli.git
cd micli
go build -o micli
```

### 环境要求

- Go 1.22 或更高版本
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

### 4. 小爱音箱

```bash
# TTS 语音合成
./micli mina tts -d <device_id> --text "你好世界"

# 列出小爱设备
./micli mina list

# 启动服务模式
./micli mina serve
```

### 5. 文本转语音

```bash
# 交互式选择语音
./micli tts -t "你好，世界"

# 指定语音
./micli tts -t "Hello" -v en-US-JennyNeural
```

## 命令参考

| 命令 | 说明 |
|------|------|
| `list` | 列出所有设备 |
| `get` | 获取 MIoT 设备属性 |
| `set` | 设置 MIoT 设备属性 |
| `action` | 执行 MIoT 动作 |
| `spec` | 查看 MIoT 规范 |
| `decode` | 解码 MIoT 数据 |
| `mina` | 小爱音箱命令 |
| `mina list` | 列出小爱设备 |
| `mina tts` | 小爱 TTS |
| `mina player` | 播放控制 |
| `mina serve` | 启动 Web 服务 |
| `tts` | Edge TTS 合成 |
| `miot_raw` | 原始 MIoT 请求 |
| `miio_raw` | 原始 MiIO 请求 |
| `reset` | 重置配置 |

## MIoT 规范

在 [miot-spec](https://home.miot-spec.com/) 查找设备规范。

## 配置文件

配置文件路径：`conf.ini`

```ini
[account]
MI_USER = your_email@example.com
MI_PASS = your_password
REGION = cn

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
           ├─ cmd/          # Cobra 命令
           │   ├─ root.go   # 配置和服务初始化
           │   ├─ mina*.go  # 小爱音箱命令
           │   └─ props_*.go # MIoT 属性操作
           │
           ├─ pkg/
           │   ├─ miservice/ # 小米 API 核心
           │   ├─ jarvis/    # ChatGPT 集成
           │   └─ tts/       # Edge TTS
           │
           └─ internal/
               ├─ app.go     # Gin 服务器
               └─ conf/      # 配置管理
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
```

## 贡献

欢迎提交 Pull Request！

## 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件。

## 致谢

- [Cobra](https://github.com/spf13/cobra) - CLI 框架
- [Gin](https://github.com/gin-gonic/gin) - Web 框架
- [go-openai](https://github.com/sashabaranov/go-openai) - OpenAI API 客户端
- [pterm](https://github.com/pterm/pterm) - 终端 UI

---

**免责声明：** 本项目仅供学习交流使用，请勿用于非法用途，后果自负。