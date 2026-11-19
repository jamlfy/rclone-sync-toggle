package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/ini.v1"
)

const (
	DefaultBisyncFlags = "--fast-list --transfers 8 --checkers 16 --tpslimit 10 --drive-chunk-size 64M --drive-skip-dangling-shortcuts --drive-export-formats docx,xlsx,pptx --drive-import-formats docx,xlsx,pptx"
	DefaultMountFlags  = "--vfs-cache-mode full --vfs-cache-max-size 10G --vfs-cache-max-age 1h --dir-cache-time 96h --poll-interval 30s --tpslimit 10 --tpslimit-burst 10"
)

var GlobalIgnores = []string{"node_modules", ".git", "__pycache__", ".venv", "venv"}

type ConnectorProfile struct {
	Name         string
	Remote       string
	Mode         string
	Enabled      bool
	LocalDir     string
	VirtualMount string
	VirtualItems []string
	BisyncFlags  string
	MountFlags   string
	SyncMutex    sync.Mutex
}

func ProfilesPath() string {
	h, _ := os.UserHomeDir()
	return filepath.Join(h, ".config", "rclone", "sync-profiles.ini")
}

func RuntimeFiltersDir() string {
	h, _ := os.UserHomeDir()
	return filepath.Join(h, ".cache", "rclone-sync-filters")
}

func ExpandTilde(path string) string {
	if strings.HasPrefix(path, "~/") {
		h, _ := os.UserHomeDir()
		return filepath.Join(h, path[2:])
	}
	return path
}

func SanitizeBisyncFlags(flags string) string {
	fields := strings.Fields(flags)
	sanitized := make([]string, 0, len(fields))
	for _, field := range fields {
		if field != "--resync" {
			sanitized = append(sanitized, field)
		}
	}
	if len(sanitized) == 0 {
		return DefaultBisyncFlags
	}
	result := strings.Join(sanitized, " ")
	if !strings.Contains(result, "--drive-skip-dangling-shortcuts") {
		result += " --drive-skip-dangling-shortcuts"
	}
	if !strings.Contains(result, "--drive-export-formats") {
		result += " --drive-export-formats docx,xlsx,pptx"
	}
	if !strings.Contains(result, "--drive-import-formats") {
		result += " --drive-import-formats docx,xlsx,pptx"
	}
	return result
}

func LoadProfiles() ([]*ConnectorProfile, error) {
	cfg, err := ini.Load(ProfilesPath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var profiles []*ConnectorProfile
	for _, section := range cfg.Sections() {
		if section.Name() == ini.DefaultSection {
			continue
		}

		current := &ConnectorProfile{
			Name:        section.Name(),
			Remote:      section.Name() + ":/",
			Mode:        "hybrid",
			Enabled:     true,
			BisyncFlags: DefaultBisyncFlags,
			MountFlags:  DefaultMountFlags,
		}
		current.Remote = section.Key("remote").MustString(current.Remote)
		current.Mode = section.Key("mode").MustString(current.Mode)
		current.Enabled = section.Key("enabled").MustBool(true)
		current.LocalDir = ExpandTilde(section.Key("local_dir").MustString(""))
		current.VirtualMount = ExpandTilde(section.Key("virtual_mount").MustString(""))
		current.BisyncFlags = SanitizeBisyncFlags(section.Key("bisync_flags").MustString(DefaultBisyncFlags))
		current.MountFlags = section.Key("mount_flags").MustString(DefaultMountFlags)

		if virtualItems := strings.TrimSpace(section.Key("virtual_items").MustString("")); virtualItems != "" {
			for _, item := range strings.Split(virtualItems, ",") {
				if trimmed := strings.TrimSpace(item); trimmed != "" {
					current.VirtualItems = append(current.VirtualItems, trimmed)
				}
			}
		}
		profiles = append(profiles, current)
	}
	return profiles, nil
}

func SaveProfiles(profiles []*ConnectorProfile) error {
	_ = os.MkdirAll(filepath.Dir(ProfilesPath()), 0755)
	cfg := ini.Empty()

	for _, p := range profiles {
		p.BisyncFlags = SanitizeBisyncFlags(p.BisyncFlags)
		section, err := cfg.NewSection(p.Name)
		if err != nil {
			return err
		}
		section.Key("remote").SetValue(p.Remote)
		section.Key("mode").SetValue(p.Mode)
		section.Key("enabled").SetValue(fmt.Sprintf("%t", p.Enabled))
		section.Key("local_dir").SetValue(p.LocalDir)
		section.Key("virtual_mount").SetValue(p.VirtualMount)
		section.Key("bisync_flags").SetValue(p.BisyncFlags)
		section.Key("mount_flags").SetValue(p.MountFlags)
		if len(p.VirtualItems) > 0 {
			section.Key("virtual_items").SetValue(strings.Join(p.VirtualItems, ", "))
		}
	}
	if err := cfg.SaveTo(ProfilesPath()); err != nil {
		return err
	}
	return os.Chmod(ProfilesPath(), 0644)
}

func (p *ConnectorProfile) BuildFilterFile() (string, error) {
	_ = os.MkdirAll(RuntimeFiltersDir(), 0755)
	filterPath := filepath.Join(RuntimeFiltersDir(), p.Name+".filters")
	var sb strings.Builder

	for _, item := range p.VirtualItems {
		clean := "/" + strings.TrimPrefix(item, "/")
		sb.WriteString(fmt.Sprintf("- %s\n", clean))
		sb.WriteString(fmt.Sprintf("- %s/**\n", clean))
	}
	for _, g := range GlobalIgnores {
		sb.WriteString(fmt.Sprintf("- %s/**\n", g))
	}
	sb.WriteString("- .DS_Store\n- *.tmp\n- *~\n")

	err := os.WriteFile(filterPath, []byte(sb.String()), 0644)
	return filterPath, err
}

func (p *ConnectorProfile) IsPathVirtual(absPath string) bool {
	for _, item := range p.VirtualItems {
		cleanRel := strings.TrimPrefix(item, "/")
		exPath := filepath.Join(p.LocalDir, filepath.FromSlash(cleanRel))
		if absPath == exPath || strings.HasPrefix(absPath, exPath+string(os.PathSeparator)) {
			return true
		}
	}
	return false
}

func (p *ConnectorProfile) AddVirtualItem(relPath string) {
	for _, existing := range p.VirtualItems {
		if existing == relPath {
			return
		}
	}
	p.VirtualItems = append(p.VirtualItems, relPath)
}

func (p *ConnectorProfile) RemoveVirtualItem(relPath string) {
	var updated []string
	for _, existing := range p.VirtualItems {
		if existing != relPath {
			updated = append(updated, existing)
		}
	}

	p.VirtualItems = updated
}
