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
		profiles, _ := config.LoadProfiles()

		// Si no hay perfiles aún, abrir directo el formulario de agregar
		if len(profiles) == 0 {
			newProf := promptAddMapping(profiles)
			if newProf == nil {
				return
			}
			profiles = append(profiles, newProf)
			_ = config.SaveProfiles(profiles)
			_ = os.MkdirAll(newProf.LocalDir, 0755)
			if newProf.Mode == "hybrid" {
				desktop.Notify("normal", "system-run", newProf.Name, i18n.T("ui.sync_new"))
				daemon.RunBisync(newProf, true)
			}
			_ = desktop.RestartService()
			desktop.Notify("normal", "emblem-default", i18n.T("ui.mapping_added"), fmt.Sprintf("%s -> %s", newProf.LocalDir, newProf.Remote))
			continue
		}

		args := []string{
			"--list",
			"--checklist",
			"--title=" + i18n.T("ui.manager_title"),
			"--text=" + i18n.T("ui.manager_text"),
			"--column=" + i18n.T("ui.active"),
			"--column=" + i18n.T("ui.id"),
			"--column=" + i18n.T("ui.local_folder"),
			"--column=" + i18n.T("ui.rclone_remote"),
			"--column=" + i18n.T("ui.mode"),
			"--hide-column=2",
			"--print-column=2",
			"--separator=|",
			"--extra-button=" + i18n.T("ui.add_synced_folder"),
			"--ok-label=" + i18n.T("ui.save_apply"),
			"--cancel-label=" + i18n.T("ui.close"),
			"--width=720",
			"--height=400",
		}

		profileMap := make(map[string]*config.ConnectorProfile)
		for _, p := range profiles {
			profileMap[p.Name] = p
			args = append(args, "TRUE", p.Name, p.LocalDir, p.Remote, p.Mode)
		}

		cmd := exec.Command("zenity", args...)
		outBytes, err := cmd.Output()
		outStr := strings.TrimSpace(string(outBytes))

		// Zenity devuelve exit code 1 al pulsar --extra-button, por eso se revisa primero outStr
		if outStr == i18n.T("ui.add_synced_folder") {
			if newProf := promptAddMapping(profiles); newProf != nil {
				profiles = append(profiles, newProf)
				_ = config.SaveProfiles(profiles)
				_ = os.MkdirAll(newProf.LocalDir, 0755)
				if newProf.Mode == "hybrid" {
					desktop.Notify("normal", "system-run", newProf.Name, i18n.T("ui.sync_new"))
					daemon.RunBisync(newProf, true)
				}
				_ = desktop.RestartService()
				desktop.Notify("normal", "emblem-default", i18n.T("ui.mapping_added"), fmt.Sprintf("%s -> %s", newProf.LocalDir, newProf.Remote))
			}
			continue
		}

		if err != nil {
			return
		}

		var keptProfiles []*config.ConnectorProfile
		if outStr != "" {
			for _, id := range strings.Split(outStr, "|") {
				if p, ok := profileMap[strings.TrimSpace(id)]; ok {
					keptProfiles = append(keptProfiles, p)
				}
			}
		}

		_ = config.SaveProfiles(keptProfiles)

		if len(keptProfiles) == 0 {
			_ = desktop.StopService()
			desktop.Notify("normal", "dialog-information", "Rclone Sync", i18n.T("ui.no_active_folders"))
		} else {
			_ = desktop.RestartService()
			desktop.Notify("normal", "emblem-default", "Rclone Sync", i18n.T("ui.applied"))
		}
		return
	}
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
