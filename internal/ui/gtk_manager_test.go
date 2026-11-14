package ui

import (
	"runtime"
	"testing"

	"github.com/ebitengine/purego"
	"rclone-sync-toggle/internal/config"
)

func TestGTKRuntime(t *testing.T) {
	gtkOnce.Do(loadGTK)
	if gtkLoadError != nil {
		t.Fatalf("loadGTK() failed: %v", gtkLoadError)
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if gtk.gtkInitCheck(nil, nil) == 0 {
		t.Skip("no graphical display is available")
	}
	gtk.gTypeClassRef(gtk.gtkCellRendererGetType())
	for _, name := range []string{"gboolean", "gchararray", "GtkCellRendererMode"} {
		if got := gtk.gTypeFromName(name); got == 0 {
			t.Errorf("GTK type %q was not registered", name)
		}
	}
}

func TestConfigManagerRenders(t *testing.T) {
	gtkOnce.Do(loadGTK)
	if gtkLoadError != nil {
		t.Fatalf("loadGTK() failed: %v", gtkLoadError)
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if gtk.gtkInitCheck(nil, nil) == 0 {
		t.Skip("no graphical display is available")
	}

	profile := &config.ConnectorProfile{
		Name:    "example",
		Remote:  "Drive:/Documents",
		Mode:    "hybrid",
		Enabled: true,
	}
	timeout := purego.NewCallback(func(_ uintptr) int32 {
		gtk.gtkMainQuit()
		return 0
	})
	gtk.gtkTimeoutAdd(250, timeout, 0)
	result := showConfigManager([]*config.ConnectorProfile{profile})
	if result.err != nil {
		t.Fatalf("showConfigManager() failed: %v", result.err)
	}
	if result.action != "cancel" {
		t.Errorf("manager action = %q, want cancel after the smoke-test timeout", result.action)
	}
	if len(result.active) != 1 || result.active[0] != profile.Name {
		t.Errorf("manager returned active profiles %v, want [%s]", result.active, profile.Name)
	}
}

func TestProfileEditorRenders(t *testing.T) {
	gtkOnce.Do(loadGTK)
	if gtkLoadError != nil {
		t.Fatalf("loadGTK() failed: %v", gtkLoadError)
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if gtk.gtkInitCheck(nil, nil) == 0 {
		t.Skip("no graphical display is available")
	}

	timeout := purego.NewCallback(func(_ uintptr) int32 {
		gtk.gtkMainQuit()
		return 0
	})
	gtk.gtkTimeoutAdd(250, timeout, 0)
	result := editProfile(&config.ConnectorProfile{
		Name:        "example",
		Remote:      "Drive:/Documents",
		Mode:        "hybrid",
		Enabled:     true,
		LocalDir:    "/home/user/Documents",
		BisyncFlags: config.DefaultBisyncFlags,
		MountFlags:  config.DefaultMountFlags,
	})
	if result.err != nil {
		t.Fatalf("editProfile() failed: %v", result.err)
	}
	if result.saved {
		t.Error("profile editor saved without the Save button being clicked")
	}
}

func TestSplitRemote(t *testing.T) {
	tests := []struct {
		remote string
		name   string
		path   string
	}{
		{remote: "Drive:/Projects", name: "Drive", path: "/Projects"},
		{remote: "Drive:", name: "Drive", path: "/"},
		{remote: "/local/path", name: "/local/path", path: "/"},
	}
	for _, tt := range tests {
		name, path := splitRemote(tt.remote)
		if name != tt.name || path != tt.path {
			t.Errorf("splitRemote(%q) = (%q, %q), want (%q, %q)", tt.remote, name, path, tt.name, tt.path)
		}
	}
}
