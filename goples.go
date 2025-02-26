package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/lxn/walk"
	"golang.org/x/sys/windows"
)

type WebhookPayload struct {
	Content string `json:"content"`
}

func main() {
	// Hide console window
	console := windows.NewLazySystemDLL("kernel32").NewProc("GetConsoleWindow")
	if console != nil {
		consoleWindow, _, _ := console.Call()
		if consoleWindow != 0 {
			user32 := windows.NewLazySystemDLL("user32")
			showWindow := user32.NewProc("ShowWindow")
			showWindow.Call(consoleWindow, 0) // SW_HIDE = 0
		}
	}

	// Define command line flags
	processName := flag.String("process", "", "Name of the process to monitor")
	webhookURL := flag.String("url", "", "URL to send webhook to")
	message := flag.String("message", "", "Message to send in the webhook")
	flag.Parse()

	if *processName == "" || *webhookURL == "" || *message == "" {
		// Create a simple GUI dialog for errors since we're hiding the console
		walk.MsgBox(nil, "Error", "Missing required parameters.\nUsage: process-monitor -process <name> -url <webhook_url> -message <message>", walk.MsgBoxIconError)
		return
	}

	// We need either a walk.MainWindow or a walk.Dialog for their message loop.
	// We will not make it visible, though.
	mw, err := walk.NewMainWindow()
	if err != nil {
		walk.MsgBox(nil, "Error", fmt.Sprintf("Failed to create main window: %v", err), walk.MsgBoxIconError)
		return
	}

	// Create system tray icon
	icon, err := walk.Resources.Icon("icon.ico")
	if err != nil {
		walk.MsgBox(nil, "Error", fmt.Sprintf("Failed to load icon: %v", err), walk.MsgBoxIconError)
		return
	}
	// Create the notify icon and make sure we clean it up on exit.
	ni, err := walk.NewNotifyIcon(mw)
	if err != nil {
		walk.MsgBox(nil, "Error", fmt.Sprintf("Failed to create notify icon: %v", err), walk.MsgBoxIconError)
		return
	}
	defer ni.Dispose()

	if err := ni.SetIcon(icon); err != nil {
		walk.MsgBox(nil, "Error", fmt.Sprintf("Failed to set icon: %v", err), walk.MsgBoxIconError)
		return
	}
	ni.SetToolTip(fmt.Sprintf("Goples: %s", *processName))

	// Add exit action to tray icon
	exitAction := walk.NewAction()
	if err := exitAction.SetText("Exit"); err != nil {
		walk.MsgBox(nil, "Error", fmt.Sprintf("Failed to set exit text: %v", err), walk.MsgBoxIconError)
		return
	}
	exitAction.Triggered().Attach(func() { walk.App().Exit(0) })
	if err := ni.ContextMenu().Actions().Add(exitAction); err != nil {
		walk.MsgBox(nil, "Error", fmt.Sprintf("Failed to add exit action: %v", err), walk.MsgBoxIconError)
		return
	}
	ni.SetVisible(true)

	// Set initial values
	var lastDetectedTime *time.Time

	// Start monitoring routine
	go func() {
		for {
			isRunning := isProcessRunning(*processName)

			if isRunning && lastDetectedTime == nil {
				// Process is running and we haven't detected it before (or it was restarted)
				now := time.Now()
				lastDetectedTime = &now

				// Send webhook notification
				err := sendWebhook(*webhookURL, *message)
				if err != nil {
					ni.ShowInfo("Error", fmt.Sprintf("Failed to send webhook: %v", err))
				}
			} else if !isRunning && lastDetectedTime != nil {
				// Process was running but now it's not
				lastDetectedTime = nil
			}

			// Wait for next check
			time.Sleep(60 * time.Second)
		}
	}()

	// Run the message loop
	mw.Run()
}

func isProcessRunning(processName string) bool {
	// Hide flashing console window
	cmd_path := "C:\\Windows\\system32\\cmd.exe"
	cmd_instance := exec.Command(cmd_path, "/c", "tasklist", "/FO", "CSV", "/NH")
	cmd_instance.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	output, err := cmd_instance.Output()
	if err != nil {
		return false
	}

	return strings.Contains(strings.ToLower(string(output)), strings.ToLower(processName))
}

func sendWebhook(url string, message string) error {
	payload := WebhookPayload{
		Content: message,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("JSON marshaling error: %v", err)
	}

	// Create a new request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	// Set the Content-Type header explicitly as in your curl command
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("wequest failed: %v", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned non-success status code: %d", resp.StatusCode)
	}

	return nil
}
