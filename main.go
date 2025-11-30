package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func logWithTimestamp(format string, a ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	msg := fmt.Sprintf(format, a...)
	// \r moves to start of line, \033[K clears the line.
	// We print the log message and a newline.
	fmt.Printf("\r\033[K[%s] %s\n", timestamp, msg)
}

const WARNING_DURATION = 5.0 // 5 seconds warning before lock

type InputTracker struct {
	idleThreshold   float64
	lastIdleTime    float64
	initialized     bool
	warningActive   bool
	warningCmd      *exec.Cmd
	warningStart    time.Time
	safeSignalChan  chan bool
}

func NewInputTracker(idleThreshold float64) *InputTracker {
	return &InputTracker{
		idleThreshold: idleThreshold,
		initialized:   false,
		warningActive: false,
		safeSignalChan: make(chan bool, 1),
	}
}

func (it *InputTracker) generatePhotoFilename() string {
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("mac-trap_%s.jpg", timestamp)

	// Get the current working directory (project directory)
	workingDir, err := os.Getwd()
	if err != nil {
		// Fallback to current directory
		workingDir = "."
	}

	// Create photos directory in the project directory
	photosDir := filepath.Join(workingDir, "mac-trap-photos")
	os.MkdirAll(photosDir, 0755)

	return filepath.Join(photosDir, filename)
}

func (it *InputTracker) takePhoto() error {
	filename := it.generatePhotoFilename()

	// Use imagesnap without warmup for instant photo
	cmd := exec.Command("imagesnap", filename)

	logWithTimestamp("Taking photo: %s", filename)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error capturing photo: %v", err)
	}

	logWithTimestamp("Photo saved: %s", filename)
	return nil
}

func (it *InputTracker) lockScreen() error {
	cmd := exec.Command("osascript", "-e", `tell application "System Events" to keystroke "q" using {control down, command down}`)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to lock screen: %v", err)
	}
	return nil
}

func (it *InputTracker) startNotification() *exec.Cmd {
	// Compile swift notification if needed
	// We assume swiftc is available as checked
	compileCmd := exec.Command("swiftc", "notification.swift", "-o", "mac-trap-notification")
	if err := compileCmd.Run(); err != nil {
		logWithTimestamp("⚠️ Failed to compile notification app: %v", err)
		return nil
	}

	// Run the compiled app
	cmd := exec.Command("./mac-trap-notification")
	if err := cmd.Start(); err != nil {
		logWithTimestamp("⚠️ Failed to start notification app: %v", err)
		return nil
	}
	
	return cmd
}

func (it *InputTracker) startWarning() {
	// Compile swift warning if needed
	compileCmd := exec.Command("swiftc", "warning.swift", "-o", "mac-trap-warning")
	if err := compileCmd.Run(); err != nil {
		logWithTimestamp("⚠️ Failed to compile warning app: %v", err)
		return
	}

	cmd := exec.Command("./mac-trap-warning")
	
	// Create a pipe to read stdout from the warning app
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		logWithTimestamp("⚠️ Failed to create stdout pipe: %v", err)
		return
	}

	if err := cmd.Start(); err != nil {
		logWithTimestamp("⚠️ Failed to start warning app: %v", err)
		return
	}

	it.warningCmd = cmd
	it.warningActive = true
	it.warningStart = time.Now()
	
	// Drain channel just in case
	select {
	case <-it.safeSignalChan:
	default:
	}

	// Start a goroutine to listen for "SAFE" signal
	go func(r io.Reader) {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			text := scanner.Text()
			if strings.Contains(text, "SAFE") {
				it.safeSignalChan <- true
				break
			}
		}
	}(stdout)

	logWithTimestamp("⚠️ Warning started - 5 seconds to abort...")
}

func (it *InputTracker) stopWarning() {
	if it.warningCmd != nil && it.warningCmd.Process != nil {
		it.warningCmd.Process.Kill()
		it.warningCmd.Wait() // Clean up resources
	}
	it.warningCmd = nil
	it.warningActive = false
}

func (it *InputTracker) handleDetection() {
	logWithTimestamp("🚨 UNAUTHORIZED ACCESS DETECTED - LOCKING!")

	// Start photo capture IMMEDIATELY in parallel
	photoDone := make(chan bool, 1)
	go func() {
		filename := it.generatePhotoFilename()
		logWithTimestamp("📷 Capturing: %s", filename)
		cmd := exec.Command("imagesnap", filename) // No warmup delay
		err := cmd.Run()
		if err == nil {
			logWithTimestamp("✅ PHOTO SAVED: %s", filename)
		}
		photoDone <- true
	}()

	// LOCK SCREEN IMMEDIATELY - DON'T WAIT FOR PHOTO
	logWithTimestamp("🔒 LOCKING IMMEDIATELY...")
	it.lockScreen()

	// Wait for photo to complete (max 3 seconds)
	select {
	case <-photoDone:
		logWithTimestamp("📷 Photo completed")
	case <-time.After(3 * time.Second):
		logWithTimestamp("📷 Photo in progress...")
	}
	// Set grace period - don't lock again immediately after unlock
	// it.lastLockTime = time.Now()
	// it.inGracePeriod = true
	// logWithTimestamp("⏳ Grace period started (5 minutes)")
	
	// FIX: Reset state completely so we don't loop immediately upon unlock
	it.lastIdleTime = 0
	it.initialized = false
	logWithTimestamp("🔄 State reset - Waiting for new session...")
}

func (it *InputTracker) getSystemIdleTime() (float64, error) {
	cmd := exec.Command("ioreg", "-c", "IOHIDSystem")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "HIDIdleTime") {
			parts := strings.Split(line, "=")
			if len(parts) > 1 {
				numStr := strings.TrimSpace(parts[1])
				numStr = strings.Fields(numStr)[0]
				idleTime, err := strconv.ParseFloat(numStr, 64)
				if err != nil {
					return 0, err
				}
				return idleTime / 1000000000.0, nil
			}
		}
	}

	return 0, fmt.Errorf("could not find HIDIdleTime")
}

func (it *InputTracker) shouldLock() bool {
	// Get current system idle time
	idleTime, err := it.getSystemIdleTime()
	if err != nil {
		return false // Silent error handling
	}

	// Initialize on first run
	if !it.initialized {
		it.lastIdleTime = idleTime
		it.initialized = true
		return false
	}

	// WARNING PHASE LOGIC
	// If warning is active, check for SAFE signal or timeout or activity
	if it.warningActive {
		// 1. Check if SAFE signal received (non-blocking check)
		select {
		case <-it.safeSignalChan:
			logWithTimestamp("✅ SAFE signal received - Resetting idle timer.")
			it.stopWarning()
			it.lastIdleTime = idleTime // Reset idle tracking
			return false
		default:
			// No signal yet
		}

		// 2. Check for activity (Unauthorized access attempt during warning!)
		// If idle time decreased significantly, it means user moved mouse/keyboard
		if idleTime < it.lastIdleTime && (it.lastIdleTime-idleTime) > 1.0 {
			// CRITICAL: The activity might be the user clicking the hidden button!
			// We must wait a brief moment to see if the SAFE signal arrives.
			logWithTimestamp("⚠️ Activity detected... checking for SAFE signal...")
			
			// Wait up to 1.5 seconds for the signal (reduced from 3000ms for snappiness)
			// This gives the user time to move the mouse to the button and click it.
			select {
			case <-it.safeSignalChan:
				logWithTimestamp("✅ SAFE signal received (after activity) - Resetting.")
				it.stopWarning()
				it.lastIdleTime = idleTime
				return false
			case <-time.After(1500 * time.Millisecond):
				// Timeout waiting for signal -> It was unauthorized activity!
				logWithTimestamp("🚨 Activity detected during warning WITHOUT safe signal!")
				it.stopWarning()
				return true // LOCK!
			}
		}
		
		// 3. Check for timeout (5 seconds elapsed)
		// User requested to KEEP the warning open indefinitely.
		// So we do NOT stop the warning on timeout.
		// We just let it run. If user is truly idle, nothing happens.
		// If user moves mouse (intruder or owner), the check above handles it.
		
		it.lastIdleTime = idleTime
		return false
	}

	// NORMAL MONITORING LOGIC

	// Check if we should start warning
	// Start warning 5 seconds before threshold
	timeUntilThreshold := it.idleThreshold - idleTime
	if timeUntilThreshold <= WARNING_DURATION && timeUntilThreshold > 0 && !it.warningActive {
		it.startWarning()
		it.lastIdleTime = idleTime
		return false
	}

	// Check if idle time was exceeded and someone just became active
	// This means unauthorized access attempt!
	if it.lastIdleTime > it.idleThreshold && idleTime < it.lastIdleTime {
		logWithTimestamp("🚨 ACCESS DETECTED AFTER IDLE PERIOD!")
		return true
	}

	// Update last idle time
	it.lastIdleTime = idleTime

	return false
}

func (it *InputTracker) monitor() {
	ticker := time.NewTicker(1 * time.Second) // Check every 1 second
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Display current status
			it.displayStatus()
			
			if it.shouldLock() {
				it.handleDetection()
				// Continue monitoring after lock (don't exit)
			}
		}
	}
}

func (it *InputTracker) displayStatus() {
	idleTime, err := it.getSystemIdleTime()
	if err != nil {
		return
	}

	// Use \r to overwrite the current line and \033[K to clear the rest of the line
	if it.warningActive {
		elapsed := time.Since(it.warningStart).Seconds()
		fmt.Printf("\r\033[K⚠️  WARNING ACTIVE: %.0fs elapsed | CLICK HIDDEN BUTTON!", elapsed)
	} else {
		if idleTime > it.idleThreshold {
			fmt.Printf("\r\033[K⚠️  System waiting | Idle: %.0fs | Waiting for activity to lock...", idleTime)
		} else {
			remaining := it.idleThreshold - idleTime
			fmt.Printf("\r\033[K🔍 Monitoring | Idle: %.0fs | Threshold in: %.0fs", idleTime, remaining)
		}
	}
}

func checkImageSnapAvailability() {
	_, err := exec.LookPath("imagesnap")
	if err != nil {
		logWithTimestamp("⚠️  Camera disabled (install: brew install imagesnap)")
	}
}

func main() {
	timeoutFlag := flag.Float64("timeout", 60.0, "Idle timeout in seconds")
	flag.Parse()

	// Wait 5 seconds before starting the application
	logWithTimestamp("🛡️  SECURITY MONITOR - Starting in 5 seconds...")
	time.Sleep(5 * time.Second)

	checkImageSnapAvailability()

	tracker := NewInputTracker(*timeoutFlag)

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		if tracker.warningCmd != nil && tracker.warningCmd.Process != nil {
			tracker.warningCmd.Process.Kill()
		}
		os.Exit(0)
	}()

	// Start monitoring (this blocks until input is detected)
	tracker.monitor()
}
