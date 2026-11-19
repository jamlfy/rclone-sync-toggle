package config

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/ini.v1"
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

func TestSanitizeBisyncFlagsRemovesPersistentResync(t *testing.T) {
	got := SanitizeBisyncFlags("--fast-list --resync --drive-skip-dangling-shortcuts")
	want := "--fast-list --drive-skip-dangling-shortcuts --drive-export-formats docx,xlsx,pptx --drive-import-formats docx,xlsx,pptx"
	if got != want {
		t.Fatalf("sanitize result = %q, want %q", got, want)
	}
}

func TestSanitizeBisyncFlagsAddsDanglingShortcutProtection(t *testing.T) {
	got := SanitizeBisyncFlags("--fast-list --tpslimit 5")
	want := "--fast-list --tpslimit 5 --drive-skip-dangling-shortcuts --drive-export-formats docx,xlsx,pptx --drive-import-formats docx,xlsx,pptx"
	if got != want {
		t.Fatalf("sanitize result = %q, want %q", got, want)
	}
}

func TestMissingVirtualItemsIsSkipped(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	configDir := filepath.Join(home, ".config", "rclone")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(configDir, "sync-profiles.ini")
	if err := os.WriteFile(configPath, []byte("[drive]\nremote = drive:/\nmode = hybrid\nlocal_dir = ~/Documents\n"), 0600); err != nil {
		t.Fatal(err)
	}

	profiles, err := LoadProfiles()
	if err != nil {
		t.Fatal(err)
	}
	if len(profiles) < 1 {
		t.Fatalf("profiles = %d, want 1", len(profiles))
	}
	if len(profiles[0].VirtualItems) != 0 {
		t.Fatalf("VirtualItems = %v, want empty", profiles[0].VirtualItems)
	}
	if profiles[0].IsPathVirtual(filepath.Join(home, "Documents", "file.txt")) {
		t.Fatal("path should not be virtual without virtual_items")
	}

	filterPath, err := profiles[0].BuildFilterFile()
	if err != nil {
		t.Fatal(err)
	}
	filter, err := os.ReadFile(filterPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(filter) == 0 {
		t.Fatal("filter file should still contain global ignores")
	}
}

func TestSaveProfilesOmitsEmptyVirtualItems(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	profiles := []*ConnectorProfile{{
		Name:         "drive",
		Remote:       "drive:/",
		Mode:         "hybrid",
		Enabled:      true,
		LocalDir:     filepath.Join(home, "Documents"),
		VirtualMount: filepath.Join(home, ".cache", "drive"),
		BisyncFlags:  DefaultBisyncFlags,
		MountFlags:   DefaultMountFlags,
	}}
	if err := SaveProfiles(profiles); err != nil {
		t.Fatal(err)
	}

	cfg, err := ini.Load(ProfilesPath())
	if err != nil {
		t.Fatal(err)
	}
	section := cfg.Section("drive")
	if section.HasKey("virtual_items") {
		t.Fatal("empty virtual_items should not be persisted")
	}
}
