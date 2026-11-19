package daemon

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"rclone-sync-toggle/internal/config"
	"rclone-sync-toggle/internal/desktop"
	"rclone-sync-toggle/internal/i18n"
)

const (
	debounceTime = 3 * time.Second
	pollInterval = 3 * time.Minute
)

var changeRegex = regexp.MustCompile(`(Copied|Deleted|Moved)`)

func bisyncAlreadyRunning() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	lockDir := filepath.Join(home, ".cache", "rclone", "bisync")
	matches, err := filepath.Glob(filepath.Join(lockDir, "*.lck"))
	if err != nil {
		return false
	}
	running := false
	for _, lockPath := range matches {
		data, err := os.ReadFile(lockPath)
		if err != nil {
			continue
		}
		pidText := strings.TrimSpace(string(data))
		pid, err := strconv.Atoi(pidText)
		if err != nil || pid <= 0 {
			_ = os.Remove(lockPath)
			continue
		}
		cmdline, err := os.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid))
		if err != nil {
			_ = os.Remove(lockPath)
			continue
		}
		command := string(cmdline)
		if strings.Contains(command, "rclone") && strings.Contains(command, "bisync") {
			running = true
		} else {
			_ = os.Remove(lockPath)
		}
	}
	return running
}

func RunWorker(ctx context.Context, p *config.ConnectorProfile) {
	_ = os.MkdirAll(p.LocalDir, 0755)

	if p.Mode == "virtual" {
		log.Printf("[%s] %s", p.Name, fmt.Sprintf(i18n.Log("worker.virtual_mode"), p.LocalDir, p.Remote))
		superviseMount(ctx, p.Name, p.Remote, p.LocalDir, p.MountFlags)
		return
	}

	_ = os.MkdirAll(p.VirtualMount, 0755)
	go superviseMount(ctx, p.Name, p.Remote, p.VirtualMount, p.MountFlags)

	time.Sleep(2 * time.Second)
	ensureSymlinks(p)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("[%s] %s", p.Name, fmt.Sprintf(i18n.Log("worker.watcher_start_error"), err))
		return
	}
	defer watcher.Close()

	if err := addRecursive(watcher, p); err != nil {
		log.Printf("[%s] %s", p.Name, fmt.Sprintf(i18n.Log("worker.scan_error"), err))
	}

	pollTicker := time.NewTicker(pollInterval)
	defer pollTicker.Stop()

	var debounceTimer *time.Timer
	triggerChan := make(chan struct{}, 1)

	log.Printf("[%s] %s", p.Name, fmt.Sprintf(i18n.Log("worker.hybrid_mode"), p.LocalDir, p.Remote))
	go RunBisync(p, false)

	for {
		select {
		case <-ctx.Done():
			return

		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if shouldIgnore(event.Name, p) {
				continue
			}

			if event.Has(fsnotify.Create) {
				if info, err := os.Lstat(event.Name); err == nil && info.IsDir() && info.Mode()&os.ModeSymlink == 0 {
					_ = addRecursive(watcher, p)
				}
			}

			if debounceTimer != nil {
				debounceTimer.Stop()
			}
			debounceTimer = time.AfterFunc(debounceTime, func() {
				select {
				case triggerChan <- struct{}{}:
				default:
				}
			})

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Printf("[%s] %s", p.Name, fmt.Sprintf(i18n.Log("worker.watcher_error"), err))

		case <-triggerChan:
			go RunBisync(p, false)

		case <-pollTicker.C:
			go RunBisync(p, false)
		}
	}
}

func RunBisync(p *config.ConnectorProfile, isResync bool) {
	if !p.SyncMutex.TryLock() {
		return
	}
	defer p.SyncMutex.Unlock()

	if bisyncAlreadyRunning() {
		log.Printf("[%s] Bisync is already running; skipping this cycle", p.Name)
		return
	}
	log.Printf("[%s] Starting bisync (resync=%t)", p.Name, isResync)

	filterPath, err := p.BuildFilterFile()
	if err != nil {
		log.Printf("[%s] %s", p.Name, fmt.Sprintf(i18n.Log("worker.filter_error"), err))
		return
	}

	args := []string{
		"bisync", p.LocalDir, p.Remote,
		"--filters-file", filterPath,
		"--log-level", "INFO",
	}
	if p.BisyncFlags != "" {
		args = append(args, strings.Fields(p.BisyncFlags)...)
	}
	if isResync {
		args = append(args, "--resync")
	}

	cmd := exec.Command("rclone", args...)
	output, err := cmd.CombinedOutput()
	outStr := string(output)

	if err != nil {
		log.Printf("[%s] %s", p.Name, fmt.Sprintf(i18n.Log("worker.bisync_error"), err, outStr))
		desktop.Notify("critical", "dialog-error", p.Name+i18n.T("worker.sync_failed_title"), i18n.T("worker.sync_failed"))
		if !isResync && strings.Contains(strings.ToLower(outStr), "must run --resync to recover") {
			log.Printf("[%s] Bisync requires resync: starting automatic recovery", p.Name)
			go RunBisync(p, true)
		}
		return
	}
	log.Printf("[%s] Bisync completed successfully", p.Name)

	if !isResync {
		matches := changeRegex.FindAllString(outStr, -1)
		if len(matches) > 0 {
			msg := fmt.Sprintf(i18n.Log("worker.files_updated"), len(matches))
			log.Printf("[%s] %s", p.Name, msg)
			desktop.Notify("normal", "cloud-upload", p.Name+i18n.T("worker.synced"), msg)
		}
	}
}

func superviseMount(ctx context.Context, name, remote, mountPoint, mountFlags string) {
	_ = unmount(mountPoint)

	for {
		select {
		case <-ctx.Done():
			_ = unmount(mountPoint)
			return
		default:
		}

		args := []string{"mount", remote, mountPoint}
		if mountFlags != "" {
			args = append(args, strings.Fields(mountFlags)...)
		}

		cmd := exec.CommandContext(ctx, "rclone", args...)
		log.Printf("[%s] %s", name, fmt.Sprintf(i18n.Log("worker.mounting"), remote, mountPoint))
		err := cmd.Run()
		if ctx.Err() != nil {
			_ = unmount(mountPoint)
			return
		}
		log.Printf("[%s] %s", name, fmt.Sprintf(i18n.Log("worker.mount_stopped"), err))
		time.Sleep(5 * time.Second)
	}
}

func unmount(mountPoint string) error {
	command := "fusermount"
	if _, err := exec.LookPath(command); err != nil {
		command = "fusermount3"
	}
	return exec.Command(command, "-u", "-z", mountPoint).Run()
}

func ensureSymlinks(p *config.ConnectorProfile) {
	for _, item := range p.VirtualItems {
		cleanRel := strings.TrimPrefix(item, "/")
		if cleanRel == "" {
			continue
		}
		localPath := filepath.Join(p.LocalDir, filepath.FromSlash(cleanRel))
		virtualPath := filepath.Join(p.VirtualMount, filepath.FromSlash(cleanRel))

		if _, err := os.Lstat(localPath); os.IsNotExist(err) {
			_ = os.MkdirAll(filepath.Dir(localPath), 0755)
			_ = os.Symlink(virtualPath, localPath)
		}
	}
}

func addRecursive(watcher *fsnotify.Watcher, p *config.ConnectorProfile) error {
	return filepath.WalkDir(p.LocalDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.Type()&os.ModeSymlink != 0 || p.IsPathVirtual(path) || isGlobalIgnoredDir(d.Name()) {
			return filepath.SkipDir
		}
		if d.IsDir() {
			return watcher.Add(path)
		}
		return nil
	})
}

func isGlobalIgnoredDir(name string) bool {
	for _, g := range config.GlobalIgnores {
		if name == g {
			return true
		}
	}
	return false
}

func shouldIgnore(path string, p *config.ConnectorProfile) bool {
	base := filepath.Base(path)
	if strings.HasSuffix(base, "~") || strings.HasSuffix(base, ".tmp") || base == ".DS_Store" || strings.Contains(path, ".rclone") {
		return true
	}
	for _, g := range config.GlobalIgnores {
		if strings.Contains(path, string(os.PathSeparator)+g+string(os.PathSeparator)) || base == g {
			return true
		}
	}
	return p.IsPathVirtual(path)
}
