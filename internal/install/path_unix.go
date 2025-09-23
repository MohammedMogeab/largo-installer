//go:build !windows

package install

import (
    "bufio"
    "errors"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "strings"
)

func ensurePathUnix(bin string, log io.Writer) error {
    // Informational: already in PATH?
    if onPathUnix(bin, os.Getenv("PATH")) {
        fmt.Fprintf(log, "✓ bin already on current PATH: %s\n", bin)
    }

    home, _ := os.UserHomeDir()
    if home == "" { return errors.New("cannot determine home directory for PATH update") }

    shell := filepath.Base(os.Getenv("SHELL"))
    var rcCandidates []string
    switch shell {
    case "zsh":
        rcCandidates = []string{filepath.Join(home, ".zshrc")}
    case "bash":
        rcCandidates = []string{filepath.Join(home, ".bashrc"), filepath.Join(home, ".bash_profile")}
    case "fish":
        rcCandidates = []string{filepath.Join(home, ".config", "fish", "config.fish")}
    default:
        rcCandidates = []string{filepath.Join(home, ".profile")}
    }

    exportLine := fmt.Sprintf(`export PATH="%s:$PATH"`, bin)
    added := false
    for _, rc := range rcCandidates { if fileContains(rc, bin) { added = true; break } }
    if !added {
        rc := rcCandidates[0]
        if err := os.MkdirAll(filepath.Dir(rc), 0o755); err != nil { return fmt.Errorf("prepare %s: %w", rc, err) }
        f, err := os.OpenFile(rc, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
        if err != nil { return fmt.Errorf("write %s: %w", rc, err) }
        defer f.Close()
        if _, err := f.WriteString("\n" + exportLine + "\n"); err != nil { return fmt.Errorf("append to %s: %w", rc, err) }
        fmt.Fprintf(log, "✓ Added to PATH in %s (open a NEW terminal): %s\n", rc, bin)
    }
    return nil
}

// OS-selected entry point used from install.EnsurePath
func ensurePathOS(bin string, log io.Writer) error { return ensurePathUnix(bin, log) }

func fileContains(path, needle string) bool {
    f, err := os.Open(path)
    if err != nil { return false }
    defer f.Close()
    sc := bufio.NewScanner(f)
    for sc.Scan() {
        if strings.Contains(sc.Text(), needle) { return true }
    }
    return false
}

func onPathUnix(bin, PATH string) bool {
    parts := filepath.SplitList(PATH)
    for _, p := range parts { if sameDirUnix(p, bin) { return true } }
    return false
}

func sameDirUnix(a, b string) bool {
    pa, _ := filepath.Abs(a)
    pb, _ := filepath.Abs(b)
    return pa == pb
}
