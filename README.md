# MiCLI

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**Take XiaoMi Cloud Service to the command line.**

MiCLI is a powerful CLI tool for controlling MIoT devices, XiaoAi speakers (Mina), and MiHome devices. It supports device property reading/writing, TTS synthesis, AI conversations, and more.

## Features

- 🏠 **MIoT Device Control** - Read and write device properties
- 🔊 **XiaoAi Speaker** - TTS, playback control, voice records
- 🤖 **AI Integration** - ChatGPT integration for smart conversations
- 🎙️ **Edge TTS** - Microsoft Edge TTS without API key
- 📱 **QR Code Login** - Scan QR code with Mi Home app to authenticate
- 🔄 **Token Auto-Refresh** - Automatic token renewal before expiry (25-day threshold)
- 🏷️ **Device Aliases** - Custom short names with fuzzy matching for quick device access
- 🎬 **Scene Control** - List and execute MiHome smart scenes/automations
- 🏡 **Home Filtering** - Filter devices by home/room name
- 📊 **Device Statistics** - View power consumption and usage history
- 🔧 **Consumables Management** - Track filters, batteries, and other consumable items
- 🌐 **Web Server** - Built-in Gin server for extended functionality
- ⚙️ **Automation** - Keyword-based command chaining via `commands.json`

## Installation

### Build from Source

```bash
git clone https://github.com/WangNingkai/micli.git
cd micli
go build -o micli
```

### Prerequisites

- Go 1.25 or higher
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
# List all devices
./micli list

# Filter by home name
./micli list --home "Living Room"

# Reload from cloud
./micli list -r
```

### 3. Device Aliases

```bash
# List all aliases
./micli alias list

# Add an alias
./micli alias add bedroom-light <device_name_or_did>

# Remove an alias
./micli alias rm bedroom-light
```

### 4. Device Properties

```bash
# Get device properties (by DID, name, or alias)
./micli get -d <device_id> --props <prop1>,<prop2>

# Set device properties
./micli set -d <device_id> --props <prop1>=<value1>
```

Property format: `siid-piid` (e.g., `2-1` for service 2, property 1)

### 5. MIoT Actions

```bash
# Execute MIoT action
./micli action <device_id> <siid-aiid> [args...]
```

### 6. Scene Control

```bash
# List all scenes
./micli scene list

# Execute a scene by ID or name
./micli scene run <scene_id_or_name>
```

### 7. Device Statistics

```bash
# View daily power consumption for the last 7 days
./micli stats <device_id> --key 7.1 --type day

# Use specific device via --did flag
./micli stats <device_name> -k 7.1 -t day -d <device_id>

# View hourly data for the last 24 hours
./micli stats <device_id> --key 7.1 --type hour

# View weekly/monthly statistics
./micli stats <device_id> --key 7.1 --type week
./micli stats <device_id> --key 7.1 --type month
```

### 8. Consumables Management

```bash
# List consumable items (filters, batteries, etc.)
./micli consumables

# Filter by home name
./micli consumables --home "My Home"
```

### 9. XiaoAi Speaker

```bash
# TTS synthesis
./micli mina tts -d <device_id> --text "Hello, World!"

# List XiaoAi devices
./micli mina list

# Player control (play/pause/volume)
./micli mina player <command>

# Get voice records
./micli mina records [limit]

# Set default XiaoAi device
./micli mina set_did

# Start AI conversation mode
./micli mina serve
```

### 10. Text-to-Speech

```bash
# Interactive voice selection
./micli tts -t "你好，世界"

# Specify voice
./micli tts -t "Hello" -v en-US-JennyNeural
```

### 11. QR Code Login

```bash
# Generate QR code for Mi Home app scan
./micli qr-login
```

## Commands Reference

| Command | Description |
|---------|-------------|
| `list` | List all devices |
| `list --home <name>` | Filter devices by home name |
| `get` | Get MIoT device properties |
| `set` | Set MIoT device properties |
| `action <iid> [args]` | Execute MIoT action |
| `alias` | Manage device name aliases |
| `alias list` | List all custom aliases |
| `alias add <alias> <device>` | Add an alias |
| `alias rm <alias>` | Remove an alias |
| `scene` | Manage smart scenes |
| `scene list` | List all scenes |
| `scene run <id/name>` | Execute a scene |
| `stats <did/name> [-d <device>]` | View device statistics |
| `consumables [--home <name>]` | List consumable items |
| `spec [model]` | Show MIoT specification |
| `decode` | Decode MIoT encrypted data |
| `mina` | XiaoAi speaker commands |
| `mina list` | List XiaoAi devices |
| `mina tts` | TTS for XiaoAi |
| `mina player` | Player control (play/pause/volume) |
| `mina records [limit]` | Get voice conversation records |
| `mina serve` | Start web server + AI conversation |
| `mina set_did` | Set default XiaoAi DID |
| `tts -t <text>` | Edge TTS synthesis |
| `miot_raw <cmd> <params>` | Raw MIoT API call |
| `miio_raw <uri> <data>` | Raw MiIO API call |
| `set_did` | Set default MIoT DID |
| `qr-login` | QR code authentication |
| `reset` | Reset configuration |

## Automation

Create `commands.json` (see `commands.sample.json` for examples) to define keyword-triggered automation chains:

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

Step types: `tts`, `request`, `action`, `chat` (ChatGPT)

## MIoT Specification

Find your device specification at [miot-spec](https://home.miot-spec.com/).

## Configuration

Configuration file: `conf.ini`

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

Token cache: `~/.mi.token`

## Architecture

```
main.go → cmd.Execute() (Cobra)
           │
           ├─ cmd/               # Cobra commands
           │   ├─ root.go        # Config & service init
           │   ├─ mina*.go       # XiaoAi commands
           │   ├─ props_*.go     # MIoT operations
           │   ├─ action.go      # MIoT actions
           │   ├─ alias.go       # Device alias commands
           │   ├─ scene.go       # Scene control commands
           │   ├─ stats.go       # Device statistics
           │   ├─ consumables.go # Consumables management
           │   ├─ spec.go        # MIoT spec viewer
           │   ├─ miot_raw.go    # Raw MIoT API
           │   ├─ miio_raw.go    # Raw MiIO API
           │   ├─ tts.go         # Edge TTS command
           │   ├─ decode.go      # MIoT decoder
           │   ├─ set_did.go     # Set default DID
           │   ├─ qr_login.go    # QR authentication
           │   └─ reset.go       # Reset config
           │
           ├─ pkg/
           │   ├─ miservice/     # XiaoMi API core
           │   │   ├─ service.go # Auth & requests
           │   │   ├─ mina.go    # XiaoAi API
           │   │   ├─ io.go      # MIoT/MiIO ops
           │   │   ├─ token.go   # Token storage
           │   │   └─ qrlogin.go # QR login
           │   ├─ jarvis/        # ChatGPT integration
           │   ├─ tts/           # Edge TTS
           │   └─ util/          # Utilities (signing, crypto)
           │
           ├─ internal/
           │   ├─ app.go         # Gin server
           │   ├─ conf/          # Configuration
           │   ├─ handlers/      # HTTP handlers
           │   ├─ middleware/    # CORS middleware
           │   └─ static/        # Static files
           │
           ├─ public/            # Embedded frontend (embed.FS)
           │   └─ dist/          # Frontend build output
           │
           └─ data/              # Local cache
               ├─ devices.json   # Device list cache
               └─ miot-spec.json # MIoT spec cache
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

# Lint
golangci-lint run
```

## CI/CD

GitHub Actions builds binaries for Linux (amd64/arm64), macOS, and Windows on version tags (excluding `-alpha*`).

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Gin](https://github.com/gin-gonic/gin) - Web framework
- [go-openai](https://github.com/sashabaranov/go-openai) - OpenAI API client
- [pterm](https://github.com/pterm/pterm) - Terminal UI
- [req](https://github.com/imroc/req) - HTTP client
- [Edge TTS](https://github.com/microsoft/edge-tts) - Microsoft Edge TTS

---

**Disclaimer:** This project is for educational purposes only. Use at your own risk.