package install

import (
    "bytes"
    "errors"
    "fmt"
    "io"
    "os"
    "os/exec"
    "path/filepath"
    "runtime"
    "strings"
)

// EnsureGo verifies that the Go toolchain is available and logs its version.
func EnsureGo(log io.Writer) error {
    if _, err := exec.LookPath("go"); err != nil {
        return errors.New("Go toolchain not found. Install from https://go.dev/dl and re-run this installer")
    }
    out, err := exec.Command("go", "version").CombinedOutput()
    if err != nil {
        return fmt.Errorf("unable to run 'go version': %v", err)
    }
    fmt.Fprintf(log, "✓ Found %s\n", string(bytes.TrimSpace(out)))
    return nil
}

// GoEnv returns `go env <key>` value.
func GoEnv(key string) (string, error) {
    out, err := exec.Command("go", "env", key).CombinedOutput()
    if err != nil {
        return "", fmt.Errorf("go env %s failed: %s", key, string(out))
    }
    return strings.TrimSpace(string(out)), nil
}

// BinDir picks GOBIN, GOPATH/bin, or $HOME/go/bin.
func BinDir() string {
    if gobin, _ := GoEnv("GOBIN"); gobin != "" { return gobin }
    if gopath, _ := GoEnv("GOPATH"); gopath != "" { return filepath.Join(gopath, "bin") }
    if home, _ := os.UserHomeDir(); home != "" { return filepath.Join(home, "go", "bin") }
    return ""
}

// EnsurePath delegates to the OS-specific implementation.
// We use per-OS files to avoid referencing undefined symbols on other platforms.
func EnsurePath(bin string, log io.Writer) error { return ensurePathOS(bin, log) }

// GoInstall runs `go install module@version` and streams output to log.
func GoInstall(module, version string, log io.Writer) error {
    ref := "@" + version
    fmt.Fprintf(log, "• Installing %s%s ...\n", module, ref)
    cmd := exec.Command("go", "install", module+ref)
    cmd.Stdout = log
    cmd.Stderr = log
    return cmd.Run()
}

// VerifyLargo runs `largo version` from PATH or direct candidate location.
func VerifyLargo(log io.Writer) error {
    fmt.Fprintln(log, "• Verifying 'largo' on PATH ...")
    if out, err := exec.Command("largo", "version").CombinedOutput(); err == nil {
        fmt.Fprintf(log, "%s\n", string(bytes.TrimSpace(out)))
        return nil
    }
    bin := BinDir()
    if bin == "" { return errors.New("cannot locate bin dir to verify largo") }
    candidate := filepath.Join(bin, exeName("largo"))
    if _, err := os.Stat(candidate); err != nil {
        return fmt.Errorf("largo not found at %s and not on PATH yet", candidate)
    }
    out, err := exec.Command(candidate, "version").CombinedOutput()
    if err != nil {
        return fmt.Errorf("failed running %s version: %v\n%s", candidate, err, string(out))
    }
    fmt.Fprintf(log, "%s\n", string(bytes.TrimSpace(out)))
    fmt.Fprintf(log, "• Found at %s\n", candidate)
    return nil
}

func exeName(name string) string {
    if runtime.GOOS == "windows" { return name + ".exe" }
    return name
}

