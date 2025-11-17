# Build Process Design

## Current State Analysis

The current build process has several issues:
1. Frontend assets are pulled from an external repository during build (`make.sh` calls `pullWeb` function)
2. The process uses shell scripts rather than standardized build tools
3. There's no clear separation between development and production builds
4. Multi-platform support is not well-defined

## Proposed Architecture

### Build Flow
1. **Development Mode**: Serve frontend separately (Vue dev server) with backend API proxy
2. **Production Build**:
   - Compile frontend to static assets
   - Embed assets in Go binary using `embed.FS`
   - Single binary deployment

### Directory Structure
```
project/
├── web/                 # Frontend source code
│   ├── package.json
│   ├── src/
│   └── dist/           # Compiled frontend (generated)
├── app/
│   └── web/
│       └── web.go      # Embed directive
├── Makefile            # Standardized build commands
└── go.mod
```

### Makefile Targets
- `make build`: Build for current platform
- `make build-all`: Build for all supported platforms
- `make build-linux`: Build for Linux
- `make build-windows`: Build for Windows
- `make build-darwin`: Build for macOS
- `make dev`: Start development environment
- `make clean`: Clean generated files

## Technical Considerations

### Frontend Build Integration
- Use `npm run build` or `pnpm build` to compile Vue frontend
- Copy `web/dist/` contents to `app/web/dist/`
- Go embed directive already exists in `app/web/web.go`

### Cross-compilation Support
Go supports easy cross-compilation:
- GOOS=linux GOARCH=amd64 go build
- GOOS=windows GOARCH=amd64 go build
- GOOS=darwin GOARCH=amd64 go build

### Asset Embedding
The existing `//go:embed dist` directive in `app/web/web.go` will automatically include the frontend assets.

## Implementation Steps
1. Create Makefile with standardized targets
2. Modify build process to compile frontend locally
3. Update documentation
4. Remove external dependency in build scripts
5. Test cross-platform builds