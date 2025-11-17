# Project Context

## Purpose
ShowTa云盘 is a personal/cloud storage solution that supports mounting various storage backends including local storage, Aliyun Drive, Baidu Netdisk, and 115 disk. It provides file browsing, sharing, and WebDAV access capabilities.

## Tech Stack
- Backend: Go, Gin framework
- Database: SQLite (via GORM)
- Frontend: Vue3, Element Plus
- Build System: Makefile (proposed), Go embed for asset embedding
- Containerization: Docker

## Project Conventions

### Code Style
- Go code follows standard Go conventions
- Error handling with descriptive error messages
- Structured logging with log levels

### Architecture Patterns
- Modular storage engine system with plugin-style architecture
- MVC-like organization with models, logic, and routes separation
- Embedded frontend assets for single-binary deployment

### Testing Strategy
- Unit tests for critical business logic
- Integration tests for API endpoints
- Manual testing for UI components

### Git Workflow
- Feature branches from main
- Pull requests for code review
- Semantic commit messages

## Domain Context
- Storage engines abstract different cloud storage providers
- Users have role-based access control
- Files and folders can have password protection
- WebDAV protocol support for third-party client access

## Important Constraints
- Single binary deployment preferred
- Minimal external dependencies
- Cross-platform compatibility (Windows, Linux, macOS)

## External Dependencies
- Go modules for backend dependencies
- NPM/PNPM for frontend dependencies
- Docker for containerization (optional)