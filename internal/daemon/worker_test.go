package daemon

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBisyncAlreadyRunningRemovesStaleLock(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	lockDir := filepath.Join(home, ".cache", "rclone", "bisync")
	if err := os.MkdirAll(lockDir, 0755); err != nil {
		t.Fatal(err)
	}
	lockPath := filepath.Join(lockDir, "stale.lck")
	if err := os.WriteFile(lockPath, []byte("999999999\n"), 0600); err != nil {
		t.Fatal(err)
	}

	if bisyncAlreadyRunning() {
		t.Fatal("stale lock should not report an active bisync")
	}
	if _, err := os.Stat(lockPath); !os.IsNotExist(err) {
		t.Fatal("stale lock should be removed")
	}
}
