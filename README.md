# YouFLAC Core

Shared Go backend for [YouFLAC Mobile](https://github.com/kushiemoon-dev/YouFLAC-Mobile). Compiled as a shared library (.so/.a) and accessed via FFI JSON-RPC from Flutter.

## Build

### Prerequisites
- Go 1.25+
- Android NDK 29+ (for Android builds)
- Xcode + Command Line Tools (for iOS builds, macOS only)

### Android
```
make android
```

### iOS (macOS only)
```
make ios
```

### Host (for testing)
```
make host
```

## Testing
```
go test ./... -v
```
