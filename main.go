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

type InputTracker struct {
	lastIdleTime float64
	initialized  bool
}

func NewInputTracker() *InputTracker {
	return &InputTracker{
		initialized: false,
	}
}

func (it *InputTracker) generatePhotoFilename() string {
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("mac-trap_%s.jpg", timestamp)

	// Create photos directory if it doesn't exist
	photosDir := "mac-trap-photos"
	os.MkdirAll(photosDir, 0755)

	return filepath.Join(photosDir, filename)
}

func (it *InputTracker) takePhoto() error {
	filename := it.generatePhotoFilename()

	// Use imagesnap with 2 second warmup for better photo quality
	cmd := exec.Command("imagesnap", "-w", "2", filename)

	fmt.Printf("Tomando foto: %s\n", filename)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error al capturar foto: %v", err)
	}

	fmt.Printf("Foto guardada: %s\n", filename)
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
	// AppleScript modal dialog that appears but doesn't block the program
	script := `display dialog "🛡️ SISTEMA DE SEGURIDAD ACTIVADO

Esta computadora está siendo monitoreada para prevenir acceso no autorizado.

El sistema tomará fotos automáticamente si detecta actividad no autorizada." buttons {"Entendido"} default button "Entendido" with icon caution with title "Monitor de Seguridad"`

	// Run in background - program continues without waiting
	go exec.Command("osascript", "-e", script).Run()
}

func (it *InputTracker) handleDetection() {
	fmt.Println("🚨 ¡ACCESO NO AUTORIZADO DETECTADO!")

	// Start photo capture IMMEDIATELY in parallel
	photoDone := make(chan bool, 1)
	go func() {
		filename := it.generatePhotoFilename()
		fmt.Printf("📷 Capturando: %s\n", filename)
		cmd := exec.Command("imagesnap", filename) // No warmup delay
		err := cmd.Run()
		if err == nil {
			fmt.Printf("✅ FOTO GUARDADA: %s\n", filename)
		}
		photoDone <- true
	}()

	// LOCK SCREEN IMMEDIATELY - DON'T WAIT FOR PHOTO
	fmt.Println("🔒 BLOQUEANDO INMEDIATAMENTE...")
	it.lockScreen()

	// Wait for photo to complete (max 3 seconds)
	select {
	case <-photoDone:
		fmt.Println("📷 Foto completada")
	case <-time.After(3 * time.Second):
		fmt.Println("📷 Foto en proceso...")
	}

	fmt.Println("✅ Sistema activado. Saliendo...")
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
		return false // Silent error handling
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
	ticker := time.NewTicker(250 * time.Millisecond) // Check 4x per second for faster response
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if it.detectInput() {
				it.handleDetection()
			}
		}
	}
}

func checkImageSnapAvailability() {
	_, err := exec.LookPath("imagesnap")
	if err != nil {
		fmt.Println("⚠️  Cámara deshabilitada (instalar: brew install imagesnap)")
	} else {
		fmt.Println("📷 Cámara lista")
	}
}

func main() {
	fmt.Println("🛡️  MONITOR DE SEGURIDAD - Iniciando")
	checkImageSnapAvailability()

	tracker := NewInputTracker()

	// Show startup notification popup
	tracker.showStartupNotification()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\n🛑 Detenido.")
		os.Exit(0)
	}()

	// Start monitoring (this blocks until input is detected)
	fmt.Println("🔍 ACTIVO - Esperando actividad...")
	tracker.monitor()
}
