# MiCLI

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**Take XiaoMi Cloud Service to the command line.**

MiCLI is a powerful CLI tool for controlling MIoT devices, XiaoAi speakers (Mina), and MiHome devices. It supports device property reading/writing, TTS synthesis, AI conversations, and more.

## Features

- 🏠 **MIoT Device Control** - Read and write device properties
- 🔊 **XiaoAi Speaker** - TTS, playback control, voice records
- 🤖 **AI Integration** - ChatGPT integration for smart conversations
- 🎙️ **Edge TTS** - Microsoft Edge TTS without API key
- 🌐 **Web Server** - Built-in Gin server for extended functionality

## Installation

### Build from Source

```bash
git clone https://github.com/WangNingkai/micli.git
cd micli
go build -o micli
```

### Prerequisites

- Go 1.22 or higher
- XiaoMi account

## Quick Start

### 1. Initialize Configuration

Run any command for the first time, it will create `conf.ini` interactively:

```bash
./micli list
```

Configuration options:
- `MI_USER` / `MI_PASS` - XiaoMi account credentials
- `REGION` - Region (cn/de/us/i2/ru/sg/tw)
- `MI_DID` - Default device DID
- OpenAI settings (optional)

### 2. List Devices

```bash
./micli list
```

### 3. Device Properties

```bash
# Get device properties
./micli get -d <device_id> --props <prop1>,<prop2>

# Set device properties
./micli set -d <device_id> --props <prop1>=<value1>
```

### 4. XiaoAi Speaker

```bash
# TTS synthesis
./micli mina tts -d <device_id> --text "Hello, World!"

# List XiaoAi devices
./micli mina list

# Start serve mode
./micli mina serve
```

### 5. Text-to-Speech

```bash
# Interactive voice selection
./micli tts -t "你好，世界"

# Specify voice
./micli tts -t "Hello" -v en-US-JennyNeural
```

## Commands Reference

| Command | Description |
|---------|-------------|
| `list` | List all devices |
| `get` | Get MIoT device properties |
| `set` | Set MIoT device properties |
| `action` | Execute MIoT action |
| `spec` | Show MIoT specification |
| `decode` | Decode MIoT data |
| `mina` | XiaoAi speaker commands |
| `mina list` | List XiaoAi devices |
| `mina tts` | TTS for XiaoAi |
| `mina player` | Player control |
| `mina serve` | Start web server |
| `tts` | Edge TTS synthesis |
| `miot_raw` | Raw MIoT request |
| `miio_raw` | Raw MiIO request |
| `reset` | Reset configuration |

## MIoT Specification

Find your device specification at [miot-spec](https://home.miot-spec.com/).

## Configuration

Configuration file: `conf.ini`

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

Token cache: `~/.mi.token`

## Architecture

```
main.go → cmd.Execute() (Cobra)
           │
           ├─ cmd/          # Cobra commands
           │   ├─ root.go   # Config & service init
           │   ├─ mina*.go  # XiaoAi commands
           │   └─ props_*.go # MIoT operations
           │
           ├─ pkg/
           │   ├─ miservice/ # XiaoMi API core
           │   ├─ jarvis/    # ChatGPT integration
           │   └─ tts/       # Edge TTS
           │
           └─ internal/
               ├─ app.go     # Gin server
               └─ conf/      # Configuration
```

## Development

```bash
# Run
go run main.go [command]

# Build
go build -o micli

# Test
go test ./...

# Test with race detection
go test -race ./...
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Gin](https://github.com/gin-gonic/gin) - Web framework
- [go-openai](https://github.com/sashabaranov/go-openai) - OpenAI API client
- [pterm](https://github.com/pterm/pterm) - Terminal UI

---

**Disclaimer:** This project is for educational purposes only. Use at your own risk.