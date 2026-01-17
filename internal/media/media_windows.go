//go:build windows

package media

import (
	"fmt"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32         = syscall.NewLazyDLL("user32.dll")
	procKeybdEvent = user32.NewProc("keybd_event")
)

const (
	VK_MEDIA_NEXT_TRACK = 0xB0
	VK_MEDIA_PREV_TRACK = 0xB1
	VK_MEDIA_PLAY_PAUSE = 0xB3
	KEYEVENTF_KEYUP     = 0x0002
)

func Control(action string) error {
	var vk int
	switch strings.ToLower(action) {
	case "play", "pause", "playpause":
		vk = VK_MEDIA_PLAY_PAUSE
	case "next":
		vk = VK_MEDIA_NEXT_TRACK
	case "prev", "previous":
		vk = VK_MEDIA_PREV_TRACK
	default:
		return fmt.Errorf("unknown action: %s", action)
	}

	// Press
	simulateKey(vk, false)
	// Release
	simulateKey(vk, true)

	return nil
}

func simulateKey(vk int, up bool) {
	var flags uintptr = 0
	if up {
		flags = KEYEVENTF_KEYUP
	}
	procKeybdEvent.Call(uintptr(vk), 0, flags, 0)
}

// GetStatus returns the current media status using GSMTC via syscalls
func GetStatus() (map[string]interface{}, error) {
	// Initialize COM for this thread
	// COINIT_MULTITHREADED = 0x0
	// We use CoInitializeEx to ensure COM is ready.
	// Note: In a real app, strict thread management is needed, but for a transient status check, likely ok.
	windows.CoInitializeEx(0, 0)
	defer windows.CoUninitialize()

	// 1. Get Activation Factory
	className := "Windows.Media.Control.GlobalSystemMediaTransportControlsSessionManager"
	hClassName, err := createHString(className)
	if err != nil {
		return nil, fmt.Errorf("failed to create HString: %v", err)
	}
	defer deleteHString(hClassName)

	var statics *IGlobalSystemMediaTransportControlsSessionManagerStatics
	hr, _, _ := procRoGetActivationFactory.Call(
		uintptr(hClassName),
		uintptr(unsafe.Pointer(&IID_IGlobalSystemMediaTransportControlsSessionManagerStatics)),
		uintptr(unsafe.Pointer(&statics)),
	)
	if hr != 0 {
		return nil, fmt.Errorf("RoGetActivationFactory failed: 0x%X", hr)
	}
	defer statics.Release()

	// 2. RequestAsync (Get Manager)
	var asyncOp *IAsyncOperation
	hr, _, _ = syscall.Syscall(statics.Vtbl.RequestAsync, 2,
		uintptr(unsafe.Pointer(statics)),
		uintptr(unsafe.Pointer(&asyncOp)), 0)
	if hr != 0 {
		return nil, fmt.Errorf("RequestAsync failed: 0x%X", hr)
	}
	defer asyncOp.Release()

	// Wait for async completion
	if err := awaitAsync(asyncOp); err != nil {
		return nil, err
	}

	// Get Results (Manager)
	var manager *IGlobalSystemMediaTransportControlsSessionManager
	hr, _, _ = syscall.Syscall(asyncOp.Vtbl.GetResults, 2,
		uintptr(unsafe.Pointer(asyncOp)),
		uintptr(unsafe.Pointer(&manager)), 0)
	if hr != 0 {
		return nil, fmt.Errorf("GetResults (Manager) failed: 0x%X", hr)
	}
	defer manager.Release()

	// 3. Get Current Session
	var session *IGlobalSystemMediaTransportControlsSession
	hr, _, _ = syscall.Syscall(manager.Vtbl.GetCurrentSession, 2,
		uintptr(unsafe.Pointer(manager)),
		uintptr(unsafe.Pointer(&session)), 0)

	// If session is null, nothing is playing
	if session == nil {
		return map[string]interface{}{"status": "stopped"}, nil
	}
	defer session.Release()

	// 4. Get Data (Title, Artist, Status)

	// A. Status (PlaybackInfo)
	var playbackInfo *IGlobalSystemMediaTransportControlsSessionPlaybackInfo
	hr, _, _ = syscall.Syscall(session.Vtbl.GetPlaybackInfo, 2,
		uintptr(unsafe.Pointer(session)),
		uintptr(unsafe.Pointer(&playbackInfo)), 0)
	if hr != 0 {
		return nil, fmt.Errorf("GetPlaybackInfo failed: 0x%X", hr)
	}
	defer playbackInfo.Release()

	// Note: PlaybackStatus VTable is complex to probe reliably.
	// If we have a valid session, we know media is active.
	// We'll determine exact status from metadata availability.
	_ = playbackInfo // Keep the object alive, but skip unreliable probing

	// B. Metadata (MediaProperties)
	var asyncOpProps *IAsyncOperation
	hr, _, _ = syscall.Syscall(session.Vtbl.TryGetMediaPropertiesAsync, 2,
		uintptr(unsafe.Pointer(session)),
		uintptr(unsafe.Pointer(&asyncOpProps)), 0)
	if hr != 0 {
		return nil, fmt.Errorf("TryGetMediaPropertiesAsync failed: 0x%X", hr)
	}
	defer asyncOpProps.Release()

	if err := awaitAsync(asyncOpProps); err != nil {
		return nil, err
	}

	var props *IGlobalSystemMediaTransportControlsSessionMediaProperties
	hr, _, _ = syscall.Syscall(asyncOpProps.Vtbl.GetResults, 2,
		uintptr(unsafe.Pointer(asyncOpProps)),
		uintptr(unsafe.Pointer(&props)), 0)
	if hr != 0 {
		return nil, fmt.Errorf("GetResults (Props) failed: 0x%X", hr)
	}
	defer props.Release()

	// Get Title
	var hTitle HSTRING
	syscall.Syscall(props.Vtbl.GetTitle, 2, uintptr(unsafe.Pointer(props)), uintptr(unsafe.Pointer(&hTitle)), 0)
	title := getStringFromHString(hTitle)
	deleteHString(hTitle)

	// Get Artist
	var hArtist HSTRING
	syscall.Syscall(props.Vtbl.GetArtist, 2, uintptr(unsafe.Pointer(props)), uintptr(unsafe.Pointer(&hArtist)), 0)
	artist := getStringFromHString(hArtist)
	deleteHString(hArtist)

	// Simple status: if we have metadata, media is active
	status := "active"
	if title == "" && artist == "" {
		status = "idle"
	}

	return map[string]interface{}{
		"status": status,
		"title":  title,
		"artist": artist,
	}, nil
}

// awaitAsync polls for IAsyncOperation completion
func awaitAsync(asyncOp *IAsyncOperation) error {
	var asyncInfo *IAsyncInfo
	// Manual QI for IAsyncInfo
	hr, _, _ := syscall.Syscall(asyncOp.Vtbl.QueryInterface, 3,
		uintptr(unsafe.Pointer(asyncOp)),
		uintptr(unsafe.Pointer(&IID_IAsyncInfo)),
		uintptr(unsafe.Pointer(&asyncInfo)))

	if hr != 0 {
		return fmt.Errorf("QI for IAsyncInfo failed: 0x%X", hr)
	}
	defer asyncInfo.Release()

	for {
		var status int32
		syscall.Syscall(asyncInfo.Vtbl.GetStatus, 2, uintptr(unsafe.Pointer(asyncInfo)), uintptr(unsafe.Pointer(&status)), 0)

		if status == 1 { // Completed
			return nil
		}
		if status == 2 || status == 3 { // Canceled or Error
			var errCode int32
			syscall.Syscall(asyncInfo.Vtbl.GetErrorCode, 2, uintptr(unsafe.Pointer(asyncInfo)), uintptr(unsafe.Pointer(&errCode)), 0)
			return fmt.Errorf("Async op failed with status %d, code 0x%X", status, errCode)
		}

		// Wait a bit
		// Using simple sleep to avoid burning CPU
		// windows.Sleep is in milliseconds
		time_Sleep(10)
	}
}

// time_Sleep uses syscall to sleep
func time_Sleep(ms uint32) {
	time.Sleep(time.Duration(ms) * time.Millisecond)
}
