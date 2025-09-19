# System Patterns

## Architecture
```
InputTracker
├── Input Detection (ioreg monitoring)
├── Camera Capture (imagesnap integration)  
└── Screen Locking (osascript)
```

## Key Components

### InputTracker
- **Responsibility**: Core orchestration of monitoring and response
- **Pattern**: Stateful monitoring with initialization and continuous polling
- **Key Methods**:
  - `detectInput()`: Checks for activity by comparing idle times
  - `getSystemIdleTime()`: Queries system via ioreg command
  - `monitor()`: Main monitoring loop with ticker
  - `handleDetection()`: Orchestrates response (photo + lock)

### Camera Integration
- **Tool**: imagesnap command-line utility
- **Pattern**: External command execution with error handling
- **Responsibilities**:
  - Capture photo when activity detected
  - Save with timestamp for identification
  - Handle camera access permissions gracefully

### Screen Locking
- **Current**: osascript with Ctrl+Cmd+Q keystroke
- **Pattern**: AppleScript execution via command
- **Timing**: Executed after photo capture

## Data Flow
1. **Monitor Loop**: Continuous 1-second polling of system idle time
2. **Detection**: Compare current vs last idle time
3. **Response Sequence**: Photo capture → Screen lock → Exit
4. **State Management**: Track initialization and last idle time

## Error Handling
- Camera unavailable: Log error, proceed with locking
- Screen lock failure: Log error, still exit
- System command failures: Continue operation where possible
