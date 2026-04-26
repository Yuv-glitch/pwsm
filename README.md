<img width="1388" height="400" alt="image" src="https://github.com/user-attachments/assets/a6b3540a-893a-475a-97fa-c4bd43d750de" />

# pwsm — Persistent Windows Service Manager

pwsm lets you run any executable or script as a persistent Windows service. Define your service in a YAML config file, and pwsm handles registration, lifecycle management, crash recovery, and logging — similar to how systemd works on Linux.

---

## Features

- Run any executable or script as a Windows service
- Auto-restart on crash with configurable delay
- Structured JSON error logging with child process stderr capture
- Simple YAML config per service
- Full CLI — init, install, uninstall, start, stop, status
- Services survive reboots (auto-start on boot)

---

## Requirements

- Windows 10/11
- [Go 1.21+](https://go.dev/dl) (to build from source)
- Elevated PowerShell (Run as Administrator) for all commands

---

## Installation

### Build from source

```powershell
git clone https://github.com/yourusername/pwsm
cd pwsm
go build -o pwsm.exe .
```

### Download prebuilt binary

Download the latest `pwsm.exe` from [Releases](https://github.com/yourusername/pwsm/releases) and place it somewhere on your PATH or run the binary from it's folder with an elevated powershell 

> **Note:** After downloading or building, you may need to unblock the binary:
> ```powershell
> Unblock-File -Path ".\pwsm.exe"
> ```

---

## Getting Started

All commands must be run from an **elevated PowerShell (Run as Administrator)**.

### 1. Initialise

Creates the required folders on your system:

```powershell
.\pwsm.exe init
```

This creates:
```
C:\ProgramData\pwsm\
  services\    ← drop your .yaml config files here
  logs\        ← JSON log files are written here
```

### 2. Create a config file

Create a `.yaml` file in `C:\ProgramData\pwsm\services\`. The filename is your service name.
Note: ProgramData is a hidden folder

Example — `C:\ProgramData\pwsm\services\myserver.yaml`:

```yaml
name: "myserver"
exec_path: "C:\\Python312\\python.exe"
path: "C:\\path\\to\\your\\project"
args:
  - "server.py"
restart: true
restart_delay: 5
error_logs: "C:\\ProgramData\\pwsm\\logs\\myserver.json"
```

### 3. Install and start

```powershell
.\pwsm.exe install myserver
.\pwsm.exe start myserver
```

---

## Commands

| Command | Description |
|---|---|
| `pwsm init` | Initialise folders under `C:\ProgramData\pwsm\` |
| `pwsm install <name>` | Register a service from its config file |
| `pwsm uninstall <name>` | Remove a registered service and delete its config file |
| `pwsm start <name>` | Start a registered service |
| `pwsm stop <name>` | Stop a running service |
| `pwsm status <name>` | Show current state and PID of a service |

---

## Config File Reference

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | yes | Display name of the service |
| `exec_path` | string | yes | Full path to the executable (e.g. `python.exe`, `node.exe`) |
| `path` | string | no | Working directory for the process |
| `args` | list | no | Arguments passed to the executable |
| `restart` | bool | yes | Auto-restart the process if it crashes |
| `restart_delay` | int | yes | Seconds to wait before restarting |
| `error_logs` | string | yes | Full path to the JSON log file |

---

## Examples

### Python HTTP server

```yaml
name: "pyserver"
exec_path: "C:\\Python312\\python.exe"
path: "C:\\projects\\myapp"
args:
  - "server.py"
restart: true
restart_delay: 5
error_logs: "C:\\ProgramData\\pwsm\\logs\\pyserver.json"
```

### Node.js app

```yaml
name: "nodeapp"
exec_path: "C:\\Program Files\\nodejs\\node.exe"
path: "C:\\projects\\nodeapp"
args:
  - "index.js"
restart: true
restart_delay: 3
error_logs: "C:\\ProgramData\\pwsm\\logs\\nodeapp.json"
```

### Ping (for testing)

```yaml
name: "pingtest"
exec_path: "C:\\Windows\\System32\\ping.exe"
path: ""
args:
  - "8.8.8.8"
  - "-t"
restart: false
restart_delay: 0
error_logs: "C:\\ProgramData\\pwsm\\logs\\pingtest.json"
```

---

## Logs

Each service writes structured JSON logs to the path defined in `error_logs`. Logs include:

- Service start and stop events
- Child process crash details and exit codes
- Restart attempts and failures
- Child process stderr output

Example log entry:
```json
{"time":"2026-04-26T10:00:00Z","level":"ERROR","msg":"child process exited with error","error":"exit status 1","exit_code":1}
```

---

## How It Works

pwsm registers itself with the Windows Service Control Manager (SCM) as a wrapper process. When SCM starts the service, pwsm reads the YAML config and launches your executable as a child process. It monitors the child process and handles restart logic, logging, and clean shutdown when SCM sends a stop signal.

Each service is an independent SCM entry pointing to the same `pwsm.exe` binary with a `--config <name>` flag that tells it which yaml file to load.

---

## Known Limitations

- All commands require elevated PowerShell (Administrator)
- Log files have no size limit in v1 — rotate manually if needed
- Windows only — no Linux/macOS support

---

## License

GPL-3.0
