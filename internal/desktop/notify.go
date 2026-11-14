package desktop

import "os/exec"

func Notify(urgency, icon, title, body string) {
	_ = exec.Command("notify-send", "-u", urgency, "-i", icon, title, body).Run()
}

func RestartService() error {
	return exec.Command("systemctl", "--user", "restart", "rclone-sync-toggle.service").Run()
}

func StopService() error {
	return exec.Command("systemctl", "--user", "stop", "rclone-sync-toggle.service").Run()
}
