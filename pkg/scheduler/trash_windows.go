package scheduler

import (
	"fmt"
	"path/filepath"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	FO_DELETE          = 3
	FOF_ALLOWUNDO      = 0x40
	FOF_NOCONFIRMATION = 0x10
	FOF_SILENT         = 0x04
)

type SHFILEOPSTRUCT struct {
	Hwnd                  windows.Handle
	Func                  uint32
	From                  *uint16
	To                    *uint16
	Flags                 uint16
	AnyOperationsAborted  int32
	NameMappings          uintptr
	ProgressTitle         *uint16
}

var (
	modshell32           = windows.NewLazySystemDLL("shell32.dll")
	procSHFileOperationW = modshell32.NewProc("SHFileOperationW")
)

func MoveToTrash(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	// Double null terminated string is required for SHFileOperation
	path16, err := windows.UTF16FromString(absPath)
	if err != nil {
		return err
	}
	// Append extra null character
	path16 = append(path16, 0)

	op := SHFILEOPSTRUCT{
		Func:  FO_DELETE,
		From:  &path16[0],
		Flags: FOF_ALLOWUNDO | FOF_NOCONFIRMATION | FOF_SILENT,
	}

	ret, _, _ := procSHFileOperationW.Call(uintptr(unsafe.Pointer(&op)))
	if ret != 0 {
		return fmt.Errorf("SHFileOperationW failed with code %d", ret)
	}
	return nil
}
