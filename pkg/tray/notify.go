package tray

import (
	"os/exec"
	"runtime"
)

// Notify dispatches an OS-native desktop notification banner.
func Notify(title, message string) {
	switch runtime.GOOS {
	case "linux":
		_ = exec.Command("notify-send", title, message, "-a", "AppLauncher3").Run()
	case "darwin":
		script := `display notification "` + message + `" with title "` + title + `"`
		_ = exec.Command("osascript", "-e", script).Run()
	case "windows":
		psScript := `[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null;` +
			`$template = [Windows.UI.Notifications.ToastNotificationManager]::GetTemplateContent([Windows.UI.Notifications.ToastTemplateType]::ToastText02);` +
			`$template.GetElementsByTagName('text')[0].AppendChild($template.CreateTextNode('` + title + `')) | Out-Null;` +
			`$template.GetElementsByTagName('text')[1].AppendChild($template.CreateTextNode('` + message + `')) | Out-Null;` +
			`$toast = [Windows.UI.Notifications.ToastNotification]::new($template);` +
			`[Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier('AppLauncher3').Show($toast);`
		_ = exec.Command("powershell", "-NoProfile", "-Command", psScript).Run()
	}
}
