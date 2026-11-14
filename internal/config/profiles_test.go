package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProfileEnabledDefaultsOnAndPersists(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	configDir := filepath.Join(home, ".config", "rclone")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(configDir, "sync-profiles.ini")
	if err := os.WriteFile(configPath, []byte("[legacy]\nremote = drive:/\nlocal_dir = ~/Documents\n"), 0600); err != nil {
		t.Fatal(err)
	}

	profiles, err := LoadProfiles()
	if err != nil {
		t.Fatal(err)
	}
	if len(profiles) != 1 || !profiles[0].Enabled {
		t.Fatalf("legacy profile should be enabled by default, got %#v", profiles)
	}

	profiles[0].Enabled = false
	if err := SaveProfiles(profiles); err != nil {
		t.Fatal(err)
	}
	profiles, err = LoadProfiles()
	if err != nil {
		t.Fatal(err)
	}
	if profiles[0].Enabled {
		t.Fatal("disabled profile was saved as enabled")
	}
}
