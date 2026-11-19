package i18n

import (
	"bufio"
	"embed"
	"fmt"
	"os"
	"strconv"
	"strings"
)

//go:embed locales/*.po
var poFiles embed.FS

var language = "es"
var translations = map[string]map[string]string{}

var logMessages = map[string]string{
	"config.read_error":          "Error reading configuration: %v",
	"config.empty":               "No folders are configured in %s. Run with --config to add one.",
	"daemon.started":             "Daemon started with %d active mapping(s).",
	"daemon.stopping":            "Stopping daemon and unmounting virtual drives...",
	"worker.virtual_mode":        "100%% virtual mode: %s -> %s",
	"worker.watcher_start_error": "Error starting watcher: %v",
	"worker.scan_error":          "Error scanning directories: %v",
	"worker.hybrid_mode":         "Hybrid mode active: %s <-> %s",
	"worker.watcher_error":       "Watcher error: %v",
	"worker.filter_error":        "Error generating filters: %v",
	"worker.bisync_error":        "Bisync error: %v\n%s",
	"worker.sync_failed_title":   " - Error",
	"worker.sync_failed":         "Synchronization failed. Check journalctl.",
	"worker.files_updated":       "Updated %d file(s).",
	"worker.synced":              " Synced",
	"worker.mounting":            "Mounting %s at %s",
	"worker.mount_stopped":       "Mount stopped (%v), retrying in 5s...",
}

func Init(preferred string) error {
	explicit := preferred != ""
	if preferred == "" {
		preferred = detectLanguage()
	}
	normalized := normalizeLanguage(preferred)
	if normalized == "" {
		if strings.HasPrefix(strings.ToLower(preferred), "es") {
			normalized = "es"
		} else if strings.HasPrefix(strings.ToLower(preferred), "en") {
			normalized = "en"
		} else if !explicit {
			normalized = "es"
		} else {
			return fmt.Errorf("unsupported language %q (use es or en)", preferred)
		}
	}

	messages, err := loadPO(normalized)
	if err != nil {
		return err
	}
	language = normalized
	translations[normalized] = messages
	return nil
}

func T(key string) string {
	if value, ok := translations[language][key]; ok {
		return value
	}
	return key
}

func Log(key string) string {
	if value, ok := logMessages[key]; ok {
		return value
	}
	return key
}

func detectLanguage() string {
	for _, key := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		if locale := os.Getenv(key); locale != "" {
			return locale
		}
	}
	return "es"
}

func normalizeLanguage(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if len(value) >= 2 {
		switch value[:2] {
		case "es", "en":
			return value[:2]
		}
	}
	return ""
}

func loadPO(locale string) (map[string]string, error) {
	path := "locales/" + locale + ".po"
	file, err := poFiles.Open(path)
	if err != nil {
		return nil, fmt.Errorf("load translation %s: %w", locale, err)
	}
	defer file.Close()

	result := make(map[string]string)
	scanner := bufio.NewScanner(file)
	var msgID, msgStr string
	var inID, inStr bool
	flush := func() {
		if msgID != "" {
			result[msgID] = msgStr
		}
		msgID, msgStr = "", ""
		inID, inStr = false, false
	}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			flush()
			continue
		}
		if strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "msgid ") {
			flush()
			msgID = parsePOString(strings.TrimPrefix(line, "msgid "))
			inID, inStr = true, false
			continue
		}
		if strings.HasPrefix(line, "msgstr ") {
			msgStr = parsePOString(strings.TrimPrefix(line, "msgstr "))
			inID, inStr = false, true
			continue
		}
		if strings.HasPrefix(line, "\"") {
			value := parsePOString(line)
			if inID {
				msgID += value
			} else if inStr {
				msgStr += value
			}
		}
	}
	flush()
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func parsePOString(value string) string {
	value = strings.TrimSpace(value)
	if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
		if decoded, err := strconv.Unquote(value); err == nil {
			return decoded
		}
	}
	return value
}
