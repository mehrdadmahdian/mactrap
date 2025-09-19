# Office Security Monitor (Mac-trap)

A workplace security tool for macOS that monitors who uses your office MacBook when you step away from your desk. Features transparent consent dialogs and secure photo documentation.

## 🏢 Office Use Case
Perfect for:
- Monitoring who accesses your work computer during breaks
- Documenting unauthorized access attempts  
- Maintaining workspace security compliance
- Protecting sensitive work data

## ✅ Privacy & Consent Features
- **Consent Dialog**: Users see a clear security notice before access
- **Transparent Process**: Full disclosure of monitoring and photo capture
- **User Choice**: Users can decline and system locks immediately
- **Legitimate Purpose**: Clear office security context provided

## Installation & Usage

### Prerequisites
```bash
brew install imagesnap
```

### Build & Run
```bash
go build -o trap main.go
./trap
```

## How It Works

1. **🔍 Monitoring**: Detects when someone tries to use your MacBook
2. **⚠️ Consent Dialog**: Shows office security notice with consent options
3. **📷 Documentation**: Takes photo if user consents (for security records)
4. **📱 Notification**: Shows completion message and exits
5. **🔒 Protection**: Locks screen if user declines consent

## What Users See
When someone touches your computer, they get a clear dialog:

```
⚠️ OFFICE SECURITY NOTICE ⚠️

This computer is monitored for security purposes.

By continuing to use this device, you consent to:
• Photo capture for security monitoring
• Activity logging and evidence collection  
• Compliance with office security policies

This is an authorized security measure.

[Cancel] [I Accept & Continue]
```

## Privacy Compliance
- ✅ **Transparent**: Users know they're being monitored
- ✅ **Consensual**: Users must accept before proceeding
- ✅ **Legitimate**: Clear office security purpose
- ✅ **Local**: Photos stored locally, no network transmission
- ✅ **Contextual**: Appropriate for workplace environment

Perfect for office environments where security monitoring is standard practice.
