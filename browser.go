package main

import (
	"log"
	"os/exec"
	"runtime"
	"time"
)

// openBrowser tries the platform-appropriate command to launch the user's
// default browser. Best-effort only — failures are logged but never fatal.
func openBrowser(url string) {
	// Small delay so the listener is ready by the time the browser hits it.
	time.Sleep(150 * time.Millisecond)
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	if err := cmd.Start(); err != nil {
		log.Printf("could not auto-open browser: %v (visit %s manually)", err, url)
	}
}
