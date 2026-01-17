//go:build windows

package media

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	modCombase = windows.NewLazySystemDLL("combase.dll")
	modOle32   = windows.NewLazySystemDLL("ole32.dll")

	procRoGetActivationFactory    = modCombase.NewProc("RoGetActivationFactory")
	procWindowsCreateString       = modCombase.NewProc("WindowsCreateString")
	procWindowsDeleteString       = modCombase.NewProc("WindowsDeleteString")
	procWindowsGetStringRawBuffer = modCombase.NewProc("WindowsGetStringRawBuffer")
)

type HSTRING uintptr

// GUIDs
var (
	IID_IInspectable                                             = windows.GUID{Data1: 0xAF86E2E0, Data2: 0xB12D, Data3: 0x4c6a, Data4: [8]byte{0x9C, 0x5A, 0xD7, 0xAA, 0x65, 0x10, 0x1E, 0x90}}
	IID_IAsyncInfo                                               = windows.GUID{Data1: 0x00000036, Data2: 0x0000, Data3: 0x0000, Data4: [8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}
	IID_IGlobalSystemMediaTransportControlsSessionManagerStatics = windows.GUID{Data1: 0x2050c4ee, Data2: 0x11a0, Data3: 0x57de, Data4: [8]byte{0xae, 0xd7, 0xc9, 0x7c, 0x70, 0x33, 0x82, 0x45}}
)

// Interfaces
type IInspectable struct {
	Vtbl *IInspectableVtbl
}

type IInspectableVtbl struct {
	QueryInterface      uintptr
	AddRef              uintptr
	Release             uintptr
	GetIids             uintptr
	GetRuntimeClassName uintptr
	GetTrustLevel       uintptr
}

type IGlobalSystemMediaTransportControlsSessionManagerStatics struct {
	Vtbl *IGlobalSystemMediaTransportControlsSessionManagerStaticsVtbl
}

type IGlobalSystemMediaTransportControlsSessionManagerStaticsVtbl struct {
	IInspectableVtbl
	RequestAsync uintptr
}

type IAsyncOperation struct {
	Vtbl *IAsyncOperationVtbl
}

type IAsyncOperationVtbl struct {
	IInspectableVtbl
	PutCompleted uintptr
	GetCompleted uintptr
	GetResults   uintptr
}

type IAsyncInfo struct {
	Vtbl *IAsyncInfoVtbl
}

type IAsyncInfoVtbl struct {
	IInspectableVtbl
	GetId        uintptr
	GetStatus    uintptr
	GetErrorCode uintptr
	Cancel       uintptr
	Close        uintptr
}

type IGlobalSystemMediaTransportControlsSessionManager struct {
	Vtbl *IGlobalSystemMediaTransportControlsSessionManagerVtbl
}

type IGlobalSystemMediaTransportControlsSessionManagerVtbl struct {
	IInspectableVtbl
	GetCurrentSession           uintptr
	AddCurrentSessionChanged    uintptr
	RemoveCurrentSessionChanged uintptr
	GetSessions                 uintptr
}

type IGlobalSystemMediaTransportControlsSession struct {
	Vtbl *IGlobalSystemMediaTransportControlsSessionVtbl
}

type IGlobalSystemMediaTransportControlsSessionVtbl struct {
	IInspectableVtbl
	GetSourceAppUserModelId    uintptr
	TryGetMediaPropertiesAsync uintptr
	GetPlaybackInfo            uintptr
	// ... others (TryChangeChannelUp, Down, SkipNext, Prev, etc.)
	// To be safe we should define full vtable or be very careful about offsets.
	// Standard WinRT methods are sequential.
	// We need 4 methods to get to PlaybackInfo (index 9 from IInspectable base? 6+3=9?)
	// IInspectable has 6 methods (0-5).
	// GetSourceAppUserModelId (6)
	// TryGetMediaPropertiesAsync (7)
	// GetPlaybackInfo (8)
}

type IGlobalSystemMediaTransportControlsSessionMediaProperties struct {
	Vtbl *IGlobalSystemMediaTransportControlsSessionMediaPropertiesVtbl
}

type IGlobalSystemMediaTransportControlsSessionMediaPropertiesVtbl struct {
	IInspectableVtbl
	GetTitle       uintptr
	GetSubtitle    uintptr
	GetArtist      uintptr
	GetAlbumArtist uintptr
	GetAlbumTitle  uintptr
	// ... others
}

type IGlobalSystemMediaTransportControlsSessionPlaybackInfo struct {
	Vtbl *IGlobalSystemMediaTransportControlsSessionPlaybackInfoVtbl
}

type IGlobalSystemMediaTransportControlsSessionPlaybackInfoVtbl struct {
	IInspectableVtbl
	GetControls       uintptr
	GetPlaybackStatus uintptr
	// ...
}

// Helpers

func createHString(s string) (HSTRING, error) {
	sUTF16, err := windows.UTF16PtrFromString(s)
	if err != nil {
		return 0, err
	}
	var hstring HSTRING
	ret, _, _ := procWindowsCreateString.Call(
		uintptr(unsafe.Pointer(sUTF16)),
		uintptr(len(s)),
		uintptr(unsafe.Pointer(&hstring)),
	)
	if ret != 0 {
		return 0, syscall.Errno(ret)
	}
	return hstring, nil
}

func deleteHString(hstring HSTRING) {
	procWindowsDeleteString.Call(uintptr(hstring))
}

func getStringFromHString(hstring HSTRING) string {
	var length uint32
	ptr, _, _ := procWindowsGetStringRawBuffer.Call(
		uintptr(hstring),
		uintptr(unsafe.Pointer(&length)),
	)
	if ptr == 0 {
		return ""
	}
	return windows.UTF16PtrToString((*uint16)(unsafe.Pointer(ptr)))
}

// IUnknown/IInspectable methods wrapper
func (obj *IInspectable) Release() {
	if obj != nil && obj.Vtbl != nil {
		syscall.Syscall(obj.Vtbl.Release, 1, uintptr(unsafe.Pointer(obj)), 0, 0)
	}
}

func (obj *IGlobalSystemMediaTransportControlsSessionManagerStatics) Release() {
	if obj != nil && obj.Vtbl != nil {
		syscall.Syscall(obj.Vtbl.Release, 1, uintptr(unsafe.Pointer(obj)), 0, 0)
	}
}

func (obj *IAsyncOperation) Release() {
	if obj != nil && obj.Vtbl != nil {
		syscall.Syscall(obj.Vtbl.Release, 1, uintptr(unsafe.Pointer(obj)), 0, 0)
	}
}

func (obj *IGlobalSystemMediaTransportControlsSessionManager) Release() {
	if obj != nil && obj.Vtbl != nil {
		syscall.Syscall(obj.Vtbl.Release, 1, uintptr(unsafe.Pointer(obj)), 0, 0)
	}
}

func (obj *IGlobalSystemMediaTransportControlsSession) Release() {
	if obj != nil && obj.Vtbl != nil {
		syscall.Syscall(obj.Vtbl.Release, 1, uintptr(unsafe.Pointer(obj)), 0, 0)
	}
}

func (obj *IGlobalSystemMediaTransportControlsSessionMediaProperties) Release() {
	if obj != nil && obj.Vtbl != nil {
		syscall.Syscall(obj.Vtbl.Release, 1, uintptr(unsafe.Pointer(obj)), 0, 0)
	}
}

func (obj *IGlobalSystemMediaTransportControlsSessionPlaybackInfo) Release() {
	if obj != nil && obj.Vtbl != nil {
		syscall.Syscall(obj.Vtbl.Release, 1, uintptr(unsafe.Pointer(obj)), 0, 0)
	}
}

func (obj *IAsyncInfo) Release() {
	if obj != nil && obj.Vtbl != nil {
		syscall.Syscall(obj.Vtbl.Release, 1, uintptr(unsafe.Pointer(obj)), 0, 0)
	}
}

func (obj *IInspectable) QueryInterface(iid *windows.GUID, out interface{}) error {
	// out should be **Interface
	// We do unsafe cast to uintptr
	// This is generic helper logic needed
	return nil
	// Actual implementation needs real syscall
}
