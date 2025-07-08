# Mac-trap

A simple Go program for macOS that monitors for mouse and keyboard activity and locks the screen when input is detected. It's a one-shot program that exits after locking the screen.

## Features

- **One-shot Operation**: Run once, monitor until input is detected, then exit
- **Mouse & Keyboard Detection**: Monitors system idle time to detect any input activity
- **Automatic Screen Locking**: Uses `pmset displaysleepnow` to lock the screen
- **Simple Usage**: No configuration needed, just run and it works
- **No Special Permissions**: Uses system `ioreg` command, no accessibility permissions required

## Building

```bash
go build -o mac-trap main.go
```

## Usage

Simply run the program:
```bash
./mac-trap
```

The program will:
1. Start monitoring for activity immediately
2. Lock the screen when mouse movement or keyboard input is detected
3. Exit automatically after locking

To use it again, just run the program again.

## How It Works

1. The program starts and immediately begins monitoring system idle time
2. It uses the `ioreg -c IOHIDSystem` command to get the `HIDIdleTime` value
3. Every 500ms, it checks if the idle time has decreased (indicating user activity)
4. When activity is detected, it runs `pmset displaysleepnow` to lock the screen
5. The program then exits

## Requirements

- macOS (uses `ioreg` and `pmset` commands)
- Go 1.21 or later
- No special permissions required

## Notes

- The program runs silently and monitors both mouse and keyboard activity
- Press Ctrl+C to exit without locking the screen
- The program is designed to be run manually each time you want to enable the trap
- Works reliably without requiring accessibility permissions or special frameworks 