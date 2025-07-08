package main

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

type InputTracker struct {
	lastIdleTime float64
	initialized  bool
}

func NewInputTracker() *InputTracker {
	return &InputTracker{
		initialized: false,
	}
}

func (it *InputTracker) lockScreenAndExit() {
	fmt.Println("Input detected! Locking screen and exiting...")	
	cmd := exec.Command("osascript", "-e", `tell application "System Events" to keystroke "q" using {control down, command down}`)

	err := cmd.Run()
	if err != nil {
		log.Printf("Error locking screen: %v", err)
	}
	os.Exit(0)
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

func (it *InputTracker) detectInput() bool {
	// Get current system idle time
	idleTime, err := it.getSystemIdleTime()
	if err != nil {
		log.Printf("Error getting system idle time: %v", err)
		return false
	}

	// Initialize on first run
	if !it.initialized {
		it.lastIdleTime = idleTime
		it.initialized = true
		return false
	}

	// If idle time decreased, it means there was user activity
	if idleTime < it.lastIdleTime {
		return true
	}

	// Update last idle time
	it.lastIdleTime = idleTime
	return false
}

func (it *InputTracker) monitor() {
	ticker := time.NewTicker(1000 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if it.detectInput() {
				it.lockScreenAndExit()
			}
		}
	}
}

func main() {
	fmt.Println("Mac-trap starting...")
	fmt.Println("Monitoring for mouse and keyboard activity. Screen will lock when input is detected.")
	fmt.Println("Press Ctrl+C to exit without locking.")

	tracker := NewInputTracker()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nExiting Mac-trap...")
		os.Exit(0)
	}()

	// Start monitoring (this blocks until input is detected)
	fmt.Println("Monitoring active. Move your mouse or press a key to lock the screen.")
	tracker.monitor()
}
