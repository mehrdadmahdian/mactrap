package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const IDLE_THRESHOLD = 60.0 // 10 seconds for testing
const GRACE_PERIOD = 300.0   // 5 minutes = 300 seconds

type InputTracker struct {
	lastIdleTime    float64
	initialized     bool
	lastLockTime    time.Time
	inGracePeriod   bool
}

func NewInputTracker() *InputTracker {
	return &InputTracker{
		initialized:   false,
		inGracePeriod: false,
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

	// Use imagesnap with 2 second warmup for better photo quality
	cmd := exec.Command("imagesnap", "-w", "2", filename)

	fmt.Printf("Taking photo: %s\n", filename)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error capturing photo: %v", err)
	}

	fmt.Printf("Photo saved: %s\n", filename)
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

func (it *InputTracker) showStartupNotification() {
	// AppleScript modal dialog - stays open until user clicks button
	script := `display dialog "🛡️ BEVEILIGINGSSYSTEEM GEACTIVEERD

Deze computer wordt gemonitord om ongeautoriseerde toegang te voorkomen.

Het systeem zal automatisch foto's maken als ongeautoriseerde activiteit wordt gedetecteerd." buttons {"Begrepen"} default button "Begrepen" with title "Beveiligingsmonitor"`

	// Run in background - program continues without waiting
	go exec.Command("osascript", "-e", script).Run()
}

func (it *InputTracker) handleDetection() {
	fmt.Println("🚨 UNAUTHORIZED ACCESS DETECTED - LOCKING!")

	// Start photo capture IMMEDIATELY in parallel
	photoDone := make(chan bool, 1)
	go func() {
		filename := it.generatePhotoFilename()
		fmt.Printf("📷 Capturing: %s\n", filename)
		cmd := exec.Command("imagesnap", filename) // No warmup delay
		err := cmd.Run()
		if err == nil {
			fmt.Printf("✅ PHOTO SAVED: %s\n", filename)
		}
		photoDone <- true
	}()

	// LOCK SCREEN IMMEDIATELY - DON'T WAIT FOR PHOTO
	fmt.Println("🔒 LOCKING IMMEDIATELY...")
	it.lockScreen()

	// Wait for photo to complete (max 3 seconds)
	select {
	case <-photoDone:
		fmt.Println("📷 Photo completed")
	case <-time.After(3 * time.Second):
		fmt.Println("📷 Photo in progress...")
	}

	// Set grace period - don't lock again immediately after unlock
	it.lastLockTime = time.Now()
	it.inGracePeriod = true
	fmt.Println("⏳ Grace period started (5 minutes)")
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

	// Check if we're in grace period (1 minute after unlock)
	if it.inGracePeriod {
		// If user is active (idle time decreased), exit grace period
		if idleTime < it.lastIdleTime {
			it.inGracePeriod = false
			fmt.Println("✅ Activity detected - grace period cancelled")
		}
		// If grace period has lasted 5 minutes, exit it
		if time.Since(it.lastLockTime) > GRACE_PERIOD*time.Second {
			it.inGracePeriod = false
			fmt.Println("✅ Grace period completed - monitoring...")
		}
		it.lastIdleTime = idleTime
		return false
	}

	// Check if idle time was exceeded and someone just became active
	// This means unauthorized access attempt!
	if it.lastIdleTime > IDLE_THRESHOLD && idleTime < it.lastIdleTime {
		fmt.Println("🚨 ACCESS DETECTED AFTER IDLE PERIOD!")
		return true
	}

	// Update last idle time
	it.lastIdleTime = idleTime

	// Don't auto-lock, just wait for someone to touch the computer
	// The check above will catch it
	return false
}

func (it *InputTracker) monitor() {
	ticker := time.NewTicker(5 * time.Second) // Check every 5 seconds
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

	if it.inGracePeriod {
		elapsed := time.Since(it.lastLockTime).Seconds()
		remaining := GRACE_PERIOD - elapsed
		if remaining > 0 {
			fmt.Printf("⏳ Grace period: %.0fs remaining | Idle: %.0fs\n", remaining, idleTime)
		} else {
			fmt.Printf("✅ Grace period completed | Idle: %.0fs\n", idleTime)
		}
	} else {
		if idleTime > IDLE_THRESHOLD {
			fmt.Printf("⚠️  System waiting | Idle: %.0fs | Waiting for activity to lock...\n", idleTime)
		} else {
			remaining := IDLE_THRESHOLD - idleTime
			fmt.Printf("🔍 Monitoring | Idle: %.0fs | Threshold in: %.0fs\n", idleTime, remaining)
		}
	}
}

func checkImageSnapAvailability() {
	_, err := exec.LookPath("imagesnap")
	if err != nil {
		fmt.Println("⚠️  Camera disabled (install: brew install imagesnap)")
	} else {
		fmt.Println("📷 Camera ready")
	}
}

func main() {
	// Wait 5 seconds before starting the application
	fmt.Println("🛡️  SECURITY MONITOR - Starting in 5 seconds...")

	fmt.Println("🛡️  SECURITY MONITOR - Starting")
	checkImageSnapAvailability()

	tracker := NewInputTracker()

	// Show startup notification popup
	tracker.showStartupNotification()
	time.Sleep(3 * time.Second)

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\n🛑 Stopped.")
		os.Exit(0)
	}()

	// Start monitoring (this blocks until input is detected)
	fmt.Println("🔍 ACTIVE - Waiting for activity...")
	tracker.monitor()
}
