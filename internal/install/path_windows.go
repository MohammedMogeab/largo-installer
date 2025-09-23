//go:build windows

package install

import (
    "fmt"
    "io"
    "os"
    "path/filepath"
    "strings"
    "syscall"
    "unsafe"

    "golang.org/x/sys/windows"
    "golang.org/x/sys/windows/registry"
)

func ensurePathWindows(bin string, log io.Writer) error {
    // Informational: already in current PATH (session)?
    if onPath(bin, os.Getenv("Path")) {
        fmt.Fprintf(log, "✓ %s already on current PATH (session).\n", bin)
    }

    // Read current User PATH from registry
    userPath, err := getUserPath()
    if err != nil {
        // If we can't read but the session PATH works, don't hard fail
        if onPath(bin, os.Getenv("Path")) {
            fmt.Fprintf(log, "! Could not read User PATH, but session PATH works. Skipping persist.\n")
            return nil
        }
        return fmt.Errorf("failed to read User PATH: %w", err)
    }

    // Update if missing
    if !pathListContains(userPath, bin) {
        updated := appendPath(userPath, bin)
        if err := setUserPath(updated); err != nil {
            // If current session already has it, do not fail the whole step
            if onPath(bin, os.Getenv("Path")) {
                fmt.Fprintf(log, "! Could not persist PATH, but it's available in this session. Open a new terminal after install. (%v)\n", err)
                return nil
            }
            return fmt.Errorf("failed to set User PATH: %w", err)
        }
        fmt.Fprintf(log, "✓ Ensured %s in USER Path (open a NEW terminal)\n", bin)
    } else {
        fmt.Fprintf(log, "✓ User PATH already contains %s\n", bin)
    }
    return nil
}

// OS-selected entry point used from install.EnsurePath
func ensurePathOS(bin string, log io.Writer) error { return ensurePathWindows(bin, log) }

// ---- Registry helpers ----
func getUserPath() (string, error) {
    k, err := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.QUERY_VALUE)
    if err != nil {
        return "", err
    }
    defer k.Close()
    s, _, err := k.GetStringValue("Path")
    if err == registry.ErrNotExist {
        return "", nil
    }
    return s, err
}

func setUserPath(val string) error {
    k, err := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.SET_VALUE)
    if err != nil {
        return err
    }
    defer k.Close()
    // Use ExpandString so references like %USERPROFILE% are preserved if present
    if err := k.SetExpandStringValue("Path", val); err != nil {
        return err
    }
    // Notify the system so new processes see the change
    broadcastEnvChange()
    return nil
}

func broadcastEnvChange() {
    const HWND_BROADCAST = 0xFFFF
    const WM_SETTINGCHANGE = 0x1A
    user32 := windows.NewLazySystemDLL("user32.dll")
    proc := user32.NewProc("SendMessageTimeoutW")
    if proc.Find() != nil {
        return
    }
    // SMTO_ABORTIFHUNG (0x0002)
    proc.Call(uintptr(HWND_BROADCAST), uintptr(WM_SETTINGCHANGE), 0,
        uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("Environment"))),
        uintptr(0x0002), uintptr(5000), 0)
}

// ---- Utilities ----
func pathListContains(list, item string) bool {
    item = strings.TrimRight(strings.ToLower(item), `\\`)
    for _, p := range strings.Split(list, ";") {
        p = strings.TrimSpace(p)
        if p == "" {
            continue
        }
        if strings.TrimRight(strings.ToLower(p), `\\`) == item {
            return true
        }
    }
    return false
}

func appendPath(list, item string) string {
    list = strings.Trim(list, ";")
    if list == "" {
        return item
    }
    return list + ";" + item
}

func onPath(bin, PATH string) bool {
    parts := strings.Split(PATH, ";")
    for _, p := range parts {
        if sameDir(p, bin) {
            return true
        }
    }
    return false
}

func sameDir(a, b string) bool {
    ap, _ := filepath.Abs(a)
    bp, _ := filepath.Abs(b)
    return strings.EqualFold(ap, bp)
}
