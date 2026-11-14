package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

const (
	DefaultBisyncFlags = "--fast-list --transfers 8 --checkers 16 --tpslimit 10 --drive-chunk-size 64M"
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

func LoadProfiles() ([]*ConnectorProfile, error) {
	f, err := os.Open(ProfilesPath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var profiles []*ConnectorProfile
	var current *ConnectorProfile

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			sectionName := strings.TrimSpace(line[1 : len(line)-1])
			current = &ConnectorProfile{
				Name:        sectionName,
				Remote:      sectionName + ":/",
				Mode:        "hybrid",
				Enabled:     true,
				BisyncFlags: DefaultBisyncFlags,
				MountFlags:  DefaultMountFlags,
			}
			profiles = append(profiles, current)
			continue
		}
		if current != nil && strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])

			switch key {
			case "remote":
				current.Remote = val
			case "mode":
				current.Mode = val
			case "enabled":
				if enabled, err := strconv.ParseBool(val); err == nil {
					current.Enabled = enabled
				}
			case "local_dir":
				current.LocalDir = ExpandTilde(val)
			case "virtual_mount":
				current.VirtualMount = ExpandTilde(val)
			case "bisync_flags":
				current.BisyncFlags = val
			case "mount_flags":
				current.MountFlags = val
			case "virtual_items":
				if val != "" {
					for _, item := range strings.Split(val, ",") {
						if trimmed := strings.TrimSpace(item); trimmed != "" {
							current.VirtualItems = append(current.VirtualItems, trimmed)
						}
					}
				}
			}
		}
	}
	return profiles, scanner.Err()
}

func SaveProfiles(profiles []*ConnectorProfile) error {
	_ = os.MkdirAll(filepath.Dir(ProfilesPath()), 0755)
	var sb strings.Builder

	for i, p := range profiles {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(fmt.Sprintf("[%s]\n", p.Name))
		sb.WriteString(fmt.Sprintf("remote = %s\n", p.Remote))
		sb.WriteString(fmt.Sprintf("mode = %s\n", p.Mode))
		sb.WriteString(fmt.Sprintf("enabled = %t\n", p.Enabled))
		sb.WriteString(fmt.Sprintf("local_dir = %s\n", p.LocalDir))
		sb.WriteString(fmt.Sprintf("virtual_mount = %s\n", p.VirtualMount))
		sb.WriteString(fmt.Sprintf("bisync_flags = %s\n", p.BisyncFlags))
		sb.WriteString(fmt.Sprintf("mount_flags = %s\n", p.MountFlags))
		sb.WriteString(fmt.Sprintf("virtual_items = %s\n", strings.Join(p.VirtualItems, ", ")))
	}
	return os.WriteFile(ProfilesPath(), []byte(sb.String()), 0644)
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
