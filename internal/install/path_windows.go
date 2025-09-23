//go:build windows

package install

import (
    "bytes"
    "fmt"
    "io"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
)

func ensurePathWindows(bin string, log io.Writer) error {
    // Informational: already in current PATH (session)?
    if onPath(bin, os.Getenv("Path")) {
        fmt.Fprintf(log, "✓ %s already on current PATH (session).\n", bin)
    }

    // Try Windows PowerShell or PowerShell 7 (pwsh)
    psExe := ""
    if p, err := exec.LookPath("powershell"); err == nil { psExe = p } else if p, err := exec.LookPath("pwsh"); err == nil { psExe = p }

    // Append to User PATH via .NET API
    psScript := `[string]$bin="` + escapePS(bin) + `";
$user=[Environment]::GetEnvironmentVariable("Path","User");
if([string]::IsNullOrWhiteSpace($user)){$user=""}
$items=@($user.Split(";") | ForEach-Object { $_.TrimEnd("\\") } | Where-Object { $_ -ne "" })
if(-not ($items -contains $bin.TrimEnd("\\"))){
  $new=($user.TrimEnd(";")+";"+$bin).Trim(";")
  [Environment]::SetEnvironmentVariable("Path",$new,"User")
  "UPDATED"
} else { "UNCHANGED" }`

    if psExe != "" {
        cmd := exec.Command(psExe, "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", psScript)
        out, err := cmd.CombinedOutput()
        if err == nil && strings.Contains(strings.ToUpper(string(out)), "UPDATED") {
            fmt.Fprintf(log, "✓ Ensured %s in USER Path (open a NEW terminal)\n", bin)
            return nil
        }
        // fall back to registry
    }

    current, _ := readUserPathFromReg()
    if !pathListContains(current, bin) {
        updated := appendPath(current, bin)
        if err := writeUserPathWithReg(updated); err != nil {
            return fmt.Errorf("failed to set user PATH via reg.exe: %w", err)
        }
        fmt.Fprintf(log, "✓ Ensured %s in USER Path via registry (open a NEW terminal)\n", bin)
    } else {
        fmt.Fprintf(log, "✓ User PATH already contains %s\n", bin)
    }
    return nil
}

// OS-selected entry point used from install.EnsurePath
func ensurePathOS(bin string, log io.Writer) error { return ensurePathWindows(bin, log) }

func readUserPathFromReg() (string, error) {
    cmd := exec.Command("reg", "query", `HKCU\\Environment`, "/v", "Path")
    out, err := cmd.CombinedOutput()
    if err != nil {
        if bytes.Contains(out, []byte("ERROR:")) { return "", nil }
        return "", fmt.Errorf("reg query failed: %v - %s", err, string(out))
    }
    lines := strings.Split(string(out), "\n")
    for _, ln := range lines {
        if strings.Contains(ln, "REG_") {
            parts := strings.Fields(ln)
            if len(parts) >= 3 { return strings.Join(parts[2:], " "), nil }
        }
    }
    return "", nil
}

func writeUserPathWithReg(val string) error {
    cmd := exec.Command("reg", "add", `HKCU\\Environment`, "/v", "Path", "/t", "REG_EXPAND_SZ", "/d", val, "/f")
    out, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("reg add failed: %v - %s", err, string(out))
    }
    return nil
}

func escapePS(path string) string { return strings.ReplaceAll(path, `"`, "`\"") }

func pathListContains(list, item string) bool {
    item = strings.TrimRight(strings.ToLower(item), `\\`)
    for _, p := range strings.Split(list, ";") {
        p = strings.TrimSpace(p)
        if p == "" { continue }
        if strings.TrimRight(strings.ToLower(p), `\\`) == item { return true }
    }
    return false
}

func appendPath(list, item string) string {
    list = strings.Trim(list, ";")
    if list == "" { return item }
    return list + ";" + item
}

func onPath(bin, PATH string) bool {
    parts := strings.Split(PATH, ";")
    for _, p := range parts { if sameDir(p, bin) { return true } }
    return false
}

func sameDir(a, b string) bool {
    ap, _ := filepath.Abs(a)
    bp, _ := filepath.Abs(b)
    return strings.EqualFold(ap, bp)
}
