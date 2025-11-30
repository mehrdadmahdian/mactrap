import Cocoa

class AppDelegate: NSObject, NSApplicationDelegate {
    var window: NSWindow!

    func applicationDidFinishLaunching(_ aNotification: Notification) {
        let width: CGFloat = 400
        let height: CGFloat = 150
        
        let screenRect = NSScreen.main?.frame ?? NSRect(x: 0, y: 0, width: 1000, height: 800)
        let x = (screenRect.width - width) / 2
        let y = (screenRect.height - height) / 2
        
        let rect = NSRect(x: x, y: y, width: width, height: height)
        
        // Create a floating utility window that stays on top
        window = NSWindow(
            contentRect: rect,
            styleMask: [.titled, .fullSizeContentView], 
            backing: .buffered,
            defer: false
        )
        
        window.title = "Beveiligingswaarschuwing"
        window.level = .floating // Always on top
        window.isMovableByWindowBackground = true
        
        // Set solid white background
        window.backgroundColor = .white
        
        // Remove standard buttons
        window.standardWindowButton(.closeButton)?.isHidden = true
        window.standardWindowButton(.miniaturizeButton)?.isHidden = true
        window.standardWindowButton(.zoomButton)?.isHidden = true
        
        // Create content view
        let content = NSView(frame: rect)
        window.contentView = content
        
        // Warning Label
        let label = NSTextField(labelWithString: "Het systeem maakt foto's bij ongeautoriseerde activiteit.")
        label.frame = NSRect(x: 20, y: 0, width: width - 40, height: height)
        label.alignment = .center
        // Normal font, size 14
        label.font = NSFont.systemFont(ofSize: 14, weight: .regular)
        // Black text
        label.textColor = .black
        
        content.addSubview(label)
        
        // Hidden Button (Bottom Left)
        // In Cocoa, (0,0) is bottom-left by default.
        let hiddenButton = NSButton(frame: NSRect(x: 0, y: 0, width: 50, height: 50))
        hiddenButton.title = ""
        hiddenButton.isTransparent = true
        hiddenButton.bezelStyle = .regularSquare
        hiddenButton.target = self
        hiddenButton.action = #selector(hiddenButtonClicked)
        
        // Add button to the content view
        content.addSubview(hiddenButton)
        
        window.makeKeyAndOrderFront(nil)
        NSApp.activate(ignoringOtherApps: true)
    }
    
    @objc func hiddenButtonClicked() {
        // Print SAFE signal to stdout
        if let data = "SAFE\n".data(using: .utf8) {
            FileHandle.standardOutput.write(data)
        }
        NSApp.terminate(nil)
    }
    
    func applicationShouldTerminateAfterLastWindowClosed(_ sender: NSApplication) -> Bool {
        return true
    }
}

let app = NSApplication.shared
let delegate = AppDelegate()
app.delegate = delegate
app.run()
