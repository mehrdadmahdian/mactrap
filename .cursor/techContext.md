# Technical Context

## Technology Stack
- **Language**: Go 1.21+
- **Platform**: macOS only
- **Dependencies**: Standard library only
- **External Tools**: 
  - `ioreg` (system idle time monitoring)
  - `osascript` (screen locking)
  - `imagesnap` (camera capture - to be added)

## Development Setup
```bash
# Build
go build -o mac-trap main.go

# Run
./mac-trap
```

## System Commands Used

### Current Implementation
- `ioreg -c IOHIDSystem`: Get HIDIdleTime for input detection
- `osascript -e 'tell application "System Events"...'`: Screen locking

### Camera Integration (To Add)
- `imagesnap`: Command-line camera capture utility
  - Installation: `brew install imagesnap`
  - Usage: `imagesnap -w 2 filename.jpg` (2 second warmup)

## Technical Constraints
- **macOS Only**: Uses macOS-specific commands
- **Camera Permissions**: May require camera access permission
- **No Network**: Operates entirely locally
- **Single Shot**: Exits after one detection cycle

## Dependencies
```go
// Standard library only
import (
    "fmt"
    "log" 
    "os"
    "os/exec"
    "os/signal"
    "strconv"
    "strings"
    "syscall"
    "time"
)
```

## Build & Distribution
- Single binary output
- No external configuration files
- Self-contained operation
