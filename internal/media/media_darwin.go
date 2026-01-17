//go:build darwin

package media

import (
	"fmt"
	"os/exec"
	"strings"
)

func Control(action string) error {
	var keyCode int
	switch strings.ToLower(action) {
	case "play", "pause", "playpause":
		keyCode = 100 // F8 (Play/Pause)
	case "next":
		keyCode = 101 // F9 (Next)
	case "prev", "previous":
		keyCode = 98 // F7 (Prev)
	default:
		return fmt.Errorf("unknown action: %s", action)
	}

	// tell application "System Events" to key code X
	script := fmt.Sprintf("tell application \"System Events\" to key code %d", keyCode)
	return exec.Command("osascript", "-e", script).Run()
}

func GetStatus() (map[string]interface{}, error) {
	// Update to try fetching from common players
	// Returns basic info if available

	script := `
	tell application "System Events"
		set spotifyRunning to (name of processes) contains "Spotify"
		set musicRunning to (name of processes) contains "Music"
	end tell

	if spotifyRunning then
		tell application "Spotify"
			return "Spotify" & "||" & (player state as string) & "||" & (name of current track) & "||" & (artist of current track)
		end tell
	else if musicRunning then
		tell application "Music"
			return "Music" & "||" & (player state as string) & "||" & (name of current track) & "||" & (artist of current track)
		end tell
	end if
	return "None"
	`

	out, err := exec.Command("osascript", "-e", script).Output()
	if err != nil {
		return nil, err
	}

	res := strings.TrimSpace(string(out))
	if res == "None" || res == "" {
		return map[string]interface{}{"status": "stopped"}, nil
	}

	parts := strings.Split(res, "||")
	if len(parts) >= 4 {
		return map[string]interface{}{
			"status": strings.ToLower(parts[1]), // playing/paused
			"source": parts[0],
			"title":  parts[2],
			"artist": parts[3],
		}, nil
	}

	return map[string]interface{}{"status": "unknown"}, nil
}
