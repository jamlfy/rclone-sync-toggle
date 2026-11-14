package ui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"rclone-sync-toggle/internal/config"
	"rclone-sync-toggle/internal/daemon"
	"rclone-sync-toggle/internal/desktop"
	"rclone-sync-toggle/internal/i18n"
)

func OpenConfigWindow() {
	for {
		profiles, err := config.LoadProfiles()
		if err != nil {
			desktop.Notify("critical", "dialog-error", "Rclone Sync", fmt.Sprintf(i18n.T("config.read_error"), err))
			return
		}

		if len(profiles) == 0 {
			if !addProfile(profiles) {
				return
			}
			continue
		}

		result := showConfigManager(profiles)
		if result.err != nil {
			desktop.Notify("critical", "dialog-error", "Rclone Sync", result.err.Error())
			return
		}
		if result.action == "cancel" {
			return
		}
		if result.action == "add" {
			addProfile(profiles)
			continue
		}
		if result.action == "edit" {
			var selected *config.ConnectorProfile
			for _, profile := range profiles {
				if profile.Name == result.selected {
					selected = profile
					break
				}
			}
			if selected == nil {
				continue
			}
			edited := editProfile(selected)
			if edited.err != nil {
				desktop.Notify("critical", "dialog-error", "Rclone Sync", edited.err.Error())
				return
			}
			if !edited.saved {
				continue
			}
			if err := applyProfileEdit(profiles, selected, edited.profile); err != nil {
				desktop.Notify("critical", "dialog-error", "Rclone Sync", err.Error())
				continue
			}
			continue
		}

		active := make(map[string]bool, len(result.active))
		for _, name := range result.active {
			active[name] = true
		}
		deleted := make(map[string]bool, len(result.deleted))
		for _, name := range result.deleted {
			deleted[name] = true
		}
		var keptProfiles []*config.ConnectorProfile
		for _, profile := range profiles {
			if deleted[profile.Name] {
				continue
			}
			profile.Enabled = active[profile.Name]
			keptProfiles = append(keptProfiles, profile)
		}
		if err := config.SaveProfiles(keptProfiles); err != nil {
			desktop.Notify("critical", "dialog-error", "Rclone Sync", fmt.Sprintf(i18n.T("ui.save_failed"), err))
			continue
		}

		activeCount := 0
		for _, profile := range keptProfiles {
			if profile.Enabled {
				activeCount++
			}
		}
		if activeCount == 0 {
			if err := desktop.StopService(); err != nil {
				desktop.Notify("critical", "dialog-error", "Rclone Sync", fmt.Sprintf(i18n.T("ui.stop_failed"), err))
				return
			}
			if len(result.deleted) > 0 {
				desktop.Notify("normal", "emblem-default", i18n.T("ui.mapping_removed"), fmt.Sprintf(i18n.T("ui.mapping_removed_details"), strings.Join(result.deleted, ", ")))
			}
			desktop.Notify("normal", "dialog-information", "Rclone Sync", i18n.T("ui.no_active_folders"))
			if len(keptProfiles) == 0 {
				return
			}
		} else {
			if err := desktop.RestartService(); err != nil {
				desktop.Notify("critical", "dialog-error", "Rclone Sync", fmt.Sprintf(i18n.T("ui.restart_failed"), err))
				return
			}
			if len(result.deleted) > 0 {
				desktop.Notify("normal", "emblem-default", i18n.T("ui.mapping_removed"), fmt.Sprintf(i18n.T("ui.mapping_removed_details"), strings.Join(result.deleted, ", ")))
			} else {
				desktop.Notify("normal", "emblem-default", "Rclone Sync", i18n.T("ui.applied"))
			}
		}
	}
}

func applyProfileEdit(profiles []*config.ConnectorProfile, oldProfile, edited *config.ConnectorProfile) error {
	if err := os.MkdirAll(edited.LocalDir, 0755); err != nil {
		return err
	}

	updatedProfiles := make([]*config.ConnectorProfile, 0, len(profiles))
	for _, profile := range profiles {
		if profile.Name == oldProfile.Name {
			updatedProfiles = append(updatedProfiles, edited)
		} else {
			updatedProfiles = append(updatedProfiles, profile)
		}
	}
	if err := config.SaveProfiles(updatedProfiles); err != nil {
		return fmt.Errorf(i18n.T("ui.save_failed"), err)
	}

	activeCount := 0
	for _, profile := range updatedProfiles {
		if profile.Enabled {
			activeCount++
		}
	}
	if activeCount == 0 {
		if err := desktop.StopService(); err != nil {
			return fmt.Errorf(i18n.T("ui.stop_failed"), err)
		}
	} else {
		needsResync := edited.Enabled && edited.Mode == "hybrid" &&
			(oldProfile.Mode != "hybrid" || oldProfile.LocalDir != edited.LocalDir || oldProfile.Remote != edited.Remote)
		if needsResync {
			if err := desktop.StopService(); err != nil {
				return fmt.Errorf(i18n.T("ui.stop_failed"), err)
			}
			daemon.RunBisync(edited, true)
		}
		if err := desktop.RestartService(); err != nil {
			return fmt.Errorf(i18n.T("ui.restart_failed"), err)
		}
	}
	desktop.Notify("normal", "emblem-default", i18n.T("ui.mapping_updated"), edited.Name)
	return nil
}

func addProfile(profiles []*config.ConnectorProfile) bool {
	newProf := promptAddMapping(profiles)
	if newProf == nil {
		return false
	}

	if err := os.MkdirAll(newProf.LocalDir, 0755); err != nil {
		desktop.Notify("critical", "dialog-error", "Rclone Sync", err.Error())
		return false
	}
	updated := append(profiles, newProf)
	if err := config.SaveProfiles(updated); err != nil {
		desktop.Notify("critical", "dialog-error", "Rclone Sync", fmt.Sprintf(i18n.T("ui.save_failed"), err))
		return false
	}
	if newProf.Mode == "hybrid" {
		desktop.Notify("normal", "system-run", newProf.Name, i18n.T("ui.sync_new"))
		daemon.RunBisync(newProf, true)
	}
	if err := desktop.RestartService(); err != nil {
		desktop.Notify("critical", "dialog-error", "Rclone Sync", fmt.Sprintf(i18n.T("ui.restart_failed"), err))
		return false
	}
	desktop.Notify("normal", "emblem-default", i18n.T("ui.mapping_added"), fmt.Sprintf("%s -> %s", newProf.LocalDir, newProf.Remote))
	return true
}

func promptAddMapping(existing []*config.ConnectorProfile) *config.ConnectorProfile {
	out, err := exec.Command("rclone", "listremotes").Output()
	if err != nil {
		_ = exec.Command("zenity", "--error", "--text="+i18n.T("ui.list_remotes_error")).Run()
		return nil
	}

	var remotes []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if name := strings.TrimSuffix(strings.TrimSpace(line), ":"); name != "" {
			remotes = append(remotes, name)
		}
	}
	if len(remotes) == 0 {
		_ = exec.Command("zenity", "--error", "--text="+i18n.T("ui.no_remotes")).Run()
		return nil
	}

	formArgs := []string{
		"--forms",
		"--title=" + i18n.T("ui.add_title"),
		"--text=" + i18n.T("ui.add_text"),
		"--add-entry=" + i18n.T("ui.local_path_hint"),
		"--add-combo=" + i18n.T("ui.remote"),
		"--combo-values=" + strings.Join(remotes, "|"),
		"--add-entry=" + i18n.T("ui.cloud_path"),
		"--add-combo=" + i18n.T("ui.mode"),
		"--combo-values=hybrid|virtual",
		"--add-entry=" + i18n.T("ui.bisync_flags"),
		"--add-entry=" + i18n.T("ui.mount_flags"),
		"--separator=|",
		"--width=620",
	}

	formOut, err := exec.Command("zenity", formArgs...).Output()
	if err != nil {
		return nil
	}

	fields := strings.Split(strings.TrimRight(string(formOut), "\r\n"), "|")
	if len(fields) < 6 {
		return nil
	}

	localPath := config.ExpandTilde(strings.TrimSpace(fields[0]))
	baseRemote := strings.TrimSpace(fields[1])
	subPath := strings.TrimSpace(fields[2])
	mode := strings.TrimSpace(fields[3])
	bisyncFlags := strings.TrimSpace(fields[4])
	mountFlags := strings.TrimSpace(fields[5])
	h, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	if baseRemote == "" {
		baseRemote = remotes[0]
	}
	if subPath == "" || subPath == "/" {
		subPath = "/"
	} else if !strings.HasPrefix(subPath, "/") {
		subPath = "/" + subPath
	}
	if localPath == "" {
		localPath = filepath.Join(h, baseRemote)
	}
	if mode != "virtual" {
		mode = "hybrid"
	}
	if bisyncFlags == "" {
		bisyncFlags = config.DefaultBisyncFlags
	}
	if mountFlags == "" {
		mountFlags = config.DefaultMountFlags
	}

	profName := strings.ReplaceAll(filepath.Base(localPath), " ", "_")
	for _, ex := range existing {
		if ex.Name == profName {
			profName = fmt.Sprintf("%s_%d", profName, len(existing)+1)
			break
		}
	}

	return &config.ConnectorProfile{
		Name:         profName,
		Remote:       baseRemote + ":" + subPath,
		Mode:         mode,
		Enabled:      true,
		LocalDir:     localPath,
		VirtualMount: filepath.Join(h, ".cache", strings.ToLower(profName)+"-virtual"),
		BisyncFlags:  bisyncFlags,
		MountFlags:   mountFlags,
	}
}

func ToggleItem(targetPath string, profiles []*config.ConnectorProfile) {
	absTarget, err := filepath.Abs(targetPath)
	if err != nil {
		return
	}

	var matchedProfile *config.ConnectorProfile
	for _, p := range profiles {
		if strings.HasPrefix(absTarget, p.LocalDir+string(os.PathSeparator)) {
			matchedProfile = p
			break
		}
	}

	if matchedProfile == nil {
		desktop.Notify("critical", "dialog-error", "Rclone Sync", i18n.T("ui.toggle_unmatched"))
		return
	}

	if matchedProfile.Mode == "virtual" {
		desktop.Notify("normal", "dialog-information", matchedProfile.Name, i18n.T("ui.already_virtual"))
		return
	}

	relPath, err := filepath.Rel(matchedProfile.LocalDir, absTarget)
	if err != nil || relPath == "." {
		return
	}
	rcloneRel := "/" + filepath.ToSlash(relPath)

	info, err := os.Lstat(absTarget)
	if err != nil {
		desktop.Notify("critical", "dialog-error", matchedProfile.Name, i18n.T("ui.item_missing"))
		return
	}

	isVirtual := info.Mode()&os.ModeSymlink != 0
	virtualTarget := filepath.Join(matchedProfile.VirtualMount, relPath)

	if isVirtual {
		matchedProfile.RemoveVirtualItem(rcloneRel)
		_ = os.Remove(absTarget)
		desktop.Notify("normal", "cloud-download", matchedProfile.Name, fmt.Sprintf(i18n.T("ui.downloading"), filepath.Base(absTarget)))
	} else {
		if _, err := os.Stat(virtualTarget); os.IsNotExist(err) {
			desktop.Notify("critical", "dialog-error", matchedProfile.Name, i18n.T("ui.cloud_item_missing"))
			return
		}
		matchedProfile.AddVirtualItem(rcloneRel)
		_ = os.RemoveAll(absTarget)
		_ = os.Symlink(virtualTarget, absTarget)
		desktop.Notify("normal", "cloud", matchedProfile.Name, fmt.Sprintf(i18n.T("ui.now_virtual"), filepath.Base(absTarget)))
	}

	_ = config.SaveProfiles(profiles)
	daemon.RunBisync(matchedProfile, true)
	_ = desktop.RestartService()
}
