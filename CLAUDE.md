# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is ShowTa‘Ø, a personal/cloud storage solution built with Go + Vue3 that supports mounting various storage backends including local storage, Aliyun Drive, Baidu Netdisk, and 115 disk. It provides file browsing, sharing, and WebDAV access capabilities.

## Architecture

### Main Components
- **Backend**: Go/Gin web server with SQLite database
- **Frontend**: Vue3-based web interface (in `web/` directory)
- **Storage Engines**: Modular storage system supporting multiple backends
- **Authentication**: JWT-based user authentication with role-based access control
- **WebDAV**: Built-in WebDAV server implementation

### Key Directories
- `app/` - Main application code
  - `system/` - Core system components (models, logic, routes)
  - `storage/` - Storage engine interfaces and implementations
  - `internal/` - Internal utilities (JWT, signing, WebDAV)
- `web/` - Frontend Vue application
- `runtime/` - Runtime data (database, logs) - created at runtime

### Code Structure
1. **Main Entry**: `main.go` initializes config, database, and starts Gin server
2. **Routing**: `app/system/router/` defines API endpoints with different auth levels
3. **Models**: `app/system/model/` contains GORM models and DB operations
4. **Business Logic**: `app/system/logic/` implements core business logic
5. **Storage System**: `app/storage/` provides abstract storage interface with concrete implementations
6. **Configuration**: `app/system/conf/` handles INI config parsing and initialization

## Common Development Commands

### Building
```bash
# Basic build
go build -o showta .

# Build with version info
appVersion=1.0.1
ldflags="-X 'overlink.top/app/system/conf.AppVersion=$appVersion'"
go build -o showta -ldflags="$ldflags" .

# Full build with web assets (for release)
bash make.sh public
```

### Running
```bash
# Run development server
go run main.go

# Build and run
go build -o showta . && ./showta
```

### Docker
```bash
# Build Docker image
docker build -t showta .

# Run with volume mount for data persistence
docker run -d -p 8888:8888 -v ./runtime:/svc/runtime showta
```

### Dependencies
- Go modules managed via `go.mod`
- Gin for HTTP framework
- GORM with SQLite for database
- Resty for HTTP client operations

## Storage Engine System

The storage system is modular with a plugin-style architecture:
- Interface defined in `app/storage/storage.go`
- Implementations in `app/storage/engine/*/engine.go`
- Registration via init() functions and `logic.RegisterEngine()`

Supported engines:
- `native` - Local filesystem storage
- `alipan` - Alibaba Cloud Drive
- `baidunetdisk` - Baidu Netdisk
- `115disk` - 115 Disk
- `showta` - ShowTa remote instance

## Key APIs

### Authentication
- POST `/admin/login` - User login
- Various middleware for different auth levels

### File Operations
- GET `/fd/*path` - File proxy/download
- Various admin APIs under `/admin/` for user/storage/folder management

### WebDAV
- Full WebDAV implementation under `/` for compatible clients

## Configuration

Uses INI format config file (`config.ini`) with sections:
- `server` - Host, port, HTTPS settings
- `database` - SQLite path
- `log` - Logging configuration
- `secure` - JWT secret, token expiration

Auto-generated on first run if missing.