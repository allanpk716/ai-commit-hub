package pushover

import (
	"os/exec"
	"runtime"

	"golang.org/x/sys/windows"
)

// Command creates a new exec.Cmd with hidden window on Windows
// This prevents console windows from popping up when running git commands
func Command(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)

	// On Windows, hide the console window to prevent popups
	if runtime.GOOS == "windows" {
		cmd.SysProcAttr = &windows.SysProcAttr{
			CreationFlags: 0x08000000, // CREATE_NO_WINDOW
		}
	}

	return cmd
}
