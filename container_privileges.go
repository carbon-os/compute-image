//go:build windows

package compute_image

import (
	"fmt"

	"golang.org/x/sys/windows"
)

// enablePrivileges grants SeBackupPrivilege and SeRestorePrivilege to the
// current process. Required by the Win32 backup stream APIs used internally
// by ociwclayer when writing container layer files.
// Must be called as Administrator.
func enablePrivileges() error {
	privs := []string{"SeBackupPrivilege", "SeRestorePrivilege"}

	proc, err := windows.GetCurrentProcess()
	if err != nil {
		return err
	}
	var token windows.Token
	if err := windows.OpenProcessToken(proc, windows.TOKEN_ADJUST_PRIVILEGES|windows.TOKEN_QUERY, &token); err != nil {
		return fmt.Errorf("OpenProcessToken: %w", err)
	}
	defer token.Close()

	for _, name := range privs {
		var luid windows.LUID
		ptr, _ := windows.UTF16PtrFromString(name)
		if err := windows.LookupPrivilegeValue(nil, ptr, &luid); err != nil {
			return fmt.Errorf("LookupPrivilegeValue(%s): %w", name, err)
		}
		tp := windows.Tokenprivileges{
			PrivilegeCount: 1,
			Privileges: [1]windows.LUIDAndAttributes{
				{Luid: luid, Attributes: windows.SE_PRIVILEGE_ENABLED},
			},
		}
		if err := windows.AdjustTokenPrivileges(token, false, &tp, 0, nil, nil); err != nil {
			return fmt.Errorf("AdjustTokenPrivileges(%s): %w", name, err)
		}
	}
	return nil
}