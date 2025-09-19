# Mac-trap Project Brief

## Overview
Mac-trap is a Go-based security monitoring tool for macOS that detects unauthorized input activity (mouse/keyboard) and responds with protective measures.

## Core Requirements
- **Input Detection**: Monitor system idle time to detect mouse and keyboard activity
- **Screen Locking**: Lock the screen when unauthorized activity is detected
- **Camera Capture**: Take photos when activity is detected for security evidence
- **One-shot Operation**: Run once, monitor until input detected, then exit
- **No Special Permissions**: Work without accessibility permissions where possible

## Current Implementation
- Uses `ioreg -c IOHIDSystem` to monitor HIDIdleTime
- Detects activity by checking if idle time decreases
- Currently locks screen using osascript with Ctrl+Cmd+Q
- Monitors every 1000ms (1 second intervals)
- Handles graceful shutdown with Ctrl+C

## Scope
This is a personal security tool designed for:
- Protecting an unattended Mac from unauthorized access
- Gathering evidence of attempted unauthorized use
- Simple, reliable operation without complex setup
