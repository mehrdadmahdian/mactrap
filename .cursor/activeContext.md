# Active Context

## Current Status
✅ **Office Security Monitor with Consent System** - Fully functional workplace security tool with transparent consent process.

## Recently Completed
- **Consent Dialog System**: Added AppleScript modal requiring user acceptance
- **Office Security Context**: Rebranded for legitimate workplace monitoring
- **Transparent Process**: Users fully informed before any monitoring occurs
- **Choice & Control**: Users can decline consent (system locks immediately)
- **Professional UI**: Clear office security notice with proper messaging
- **Completion Notifications**: Users notified when security documentation complete

## Implementation Details
- **Photo Method**: `takePhoto()` with timestamped filename generation
- **Detection Handler**: `handleDetection()` orchestrates photo→lock→exit sequence
- **Directory Creation**: Automatic `mac-trap-photos/` directory creation
- **User Feedback**: Clear console messages and emojis for better UX
- **Graceful Degradation**: Works even without camera, warns user appropriately

## Current Architecture
```
InputTracker
├── Input Detection (ioreg HIDIdleTime monitoring)
├── Consent Dialog (AppleScript modal with office security notice)
├── Photo Capture (imagesnap with 2s warmup - if consented)
├── Screen Locking (osascript Ctrl+Cmd+Q - if declined)
├── User Notifications (completion alerts)
└── File Management (timestamped security documentation)
```

## Workflow Sequence
1. **🔍 Detection**: Activity detected via HIDIdleTime monitoring
2. **⚠️ Consent**: AppleScript dialog with office security notice
3. **✅ Accept Path**: Photo → Notification → Exit
4. **❌ Decline Path**: Screen Lock → Exit
5. **📁 Documentation**: Timestamped photos in mac-trap-photos/

## Next Potential Enhancements
- Custom photo storage location configuration
- Different photo formats/quality options
- Multiple camera support
- Silent mode (no console output)
- Photo compression/cleanup options
