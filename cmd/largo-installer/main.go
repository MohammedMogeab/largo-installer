package main

import (
    "flag"
    "fmt"
    "os"

    tea "github.com/charmbracelet/bubbletea"

    "github.com/MohammedMogeab/largo-installer/internal/ui"
    "github.com/MohammedMogeab/largo-installer/internal/buildinfo"
)

const (
    // Default target module to install.
    defaultModule  = "github.com/MohammedMogeab/largo/cmd/largo"
    defaultVersion = "latest"
)

func main() {
    version := defaultVersion
    module := defaultModule
    noColor := false
    showVersion := false

    flag.StringVar(&version, "largo-version", version, "LarGo version to install (e.g. v0.1.0 or latest)")
    flag.StringVar(&module, "module", module, "Go module path for LarGo CLI")
    flag.BoolVar(&noColor, "no-color", false, "Disable colors in the UI")
    flag.BoolVar(&showVersion, "version", false, "Print installer version and exit")
    flag.Parse()

    if showVersion {
        fmt.Println(buildinfo.Version)
        return
    }

    m := ui.NewModel(version, module, ui.Options{NoColor: noColor})
    if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
        fmt.Fprintf(os.Stderr, "installer failed: %v\n", err)
        os.Exit(1)
    }
}
// this 