package ui

import (
    "bytes"
    "fmt"
    "io"
    "os"
    "strings"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/bubbles/spinner"
    "github.com/charmbracelet/lipgloss"

    "github.com/MohammedMogeab/largo-installer/internal/install"
)

type Options struct {
    NoColor bool
}

type stepStatus int

const (
    _ stepStatus = iota // pending
    _                   // running
    success
    failed
)

type step struct {
    name   string
    run    func(io.Writer) error
    status stepStatus
    err    error
    logBuf bytes.Buffer
}

type stepDoneMsg struct {
    index int
    err   error
    log   string
}

type model struct {
    version string
    module  string
    opts    Options
    steps   []step
    idx     int
    spin    spinner.Model
    done    bool
    width   int
    height  int
}

func NewModel(version, module string, opts Options) model {
    sp := spinner.New()
    sp.Spinner = spinner.Dot
    m := model{version: version, module: module, opts: opts, spin: sp}
    m.steps = []step{
        {name: "Check Go toolchain", run: func(w io.Writer) error { return install.EnsureGo(w) }},
        {name: "Prepare bin directory", run: func(w io.Writer) error {
            b := install.BinDir()
            if b == "" { return fmt.Errorf("cannot determine install bin dir (GOBIN/GOPATH)") }
            fmt.Fprintf(w, "Using bin: %s\n", b)
            return os.MkdirAll(b, 0o755)
        }},
        {name: "Ensure PATH contains bin", run: func(w io.Writer) error {
            b := install.BinDir()
            if b == "" { return fmt.Errorf("cannot determine install bin dir (GOBIN/GOPATH)") }
            return install.EnsurePath(b, w)
        }},
        {name: "Install LarGo (go install)", run: func(w io.Writer) error { return install.GoInstall(m.module, m.version, w) }},
        {name: "Verify 'largo' availability", run: func(w io.Writer) error { return install.VerifyLargo(w) }},
    }
    return m
}

func (m model) Init() tea.Cmd { return tea.Batch(m.spin.Tick, m.runCurrentStep()) }

func (m model) runCurrentStep() tea.Cmd {
    i := m.idx
    if i >= len(m.steps) { return nil }
    return func() tea.Msg {
        var buf bytes.Buffer
        err := m.steps[i].run(&buf)
        return stepDoneMsg{index: i, err: err, log: buf.String()}
    }
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "q", "esc", "enter", "ctrl+c":
            return m, tea.Quit
        }
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
    case stepDoneMsg:
        if msg.index < len(m.steps) {
            m.steps[msg.index].logBuf.WriteString(msg.log)
            if msg.err != nil {
                m.steps[msg.index].status = failed
                m.steps[msg.index].err = msg.err
                m.done = true
                return m, nil
            }
            m.steps[msg.index].status = success
            m.idx++
            if m.idx >= len(m.steps) {
                m.done = true
                return m, nil
            }
            return m, m.runCurrentStep()
        }
    case spinner.TickMsg:
        var cmd tea.Cmd
        m.spin, cmd = m.spin.Update(msg)
        return m, cmd
    }
    return m, nil
}

// ---------- Styles ----------
var (
    primary   = lipgloss.Color("63")
    successFg = lipgloss.Color("42")
    errorFg   = lipgloss.Color("196")
    dimFg     = lipgloss.Color("245")

    titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(primary)
    subtitleStyle = lipgloss.NewStyle().Foreground(dimFg)
    stepStyle     = lipgloss.NewStyle()
    okStyle       = lipgloss.NewStyle().Foreground(successFg)
    failStyle     = lipgloss.NewStyle().Foreground(errorFg)
    dimStyle      = lipgloss.NewStyle().Foreground(dimFg)

    boxStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1).BorderForeground(dimFg)
)

func (m model) View() string {
    w := m.width
    if w <= 0 { w = 80 }

    // Header with big framework name (simple ASCII, PowerShell safe)
    const nameArt = `
  _                        _____         
 | |                      / ____|        
 | |        __ _   _ __  | |  __    ___  
 | |       / _  | |  __| | | |_ |  / _ \ 
 | |____  | (_| | | |    | |__| | | (_) |
 |______|  \___|  |_|     \_____|  \___/ 
                                         `

    center := func(line string) string { return lipgloss.PlaceHorizontal(w, lipgloss.Center, line) }
    var nameB strings.Builder
    for _, ln := range strings.Split(strings.Trim(nameArt, "\n"), "\n") {
        nameB.WriteString(center(titleStyle.Render(ln)))
        nameB.WriteString("\n")
    }
    title := titleStyle.Render("LarGo Installer")
    sub := subtitleStyle.Render("Package: " + m.module + "  •  Version: " + m.version)
    header := nameB.String() + center(title) + "\n" + center(sub)

    // Steps list
    var stepsB strings.Builder
    for i, s := range m.steps {
        icon := "•"
        line := stepStyle
        switch s.status {
        case success:
            icon = "✔"
            line = okStyle
        case failed:
            icon = "✘"
            line = failStyle
        default:
            if i == m.idx && !m.done { icon = m.spin.View() }
        }
        stepsB.WriteString(line.Render(fmt.Sprintf(" %s %s", icon, s.name)))
        stepsB.WriteString("\n")
    }
    stepsBox := boxStyle.Render(stepsB.String())

    // Logs; fit within remaining height to keep header static
    h := m.height
    if h <= 0 { h = 30 }
    nameLines := strings.Count(strings.Trim(nameArt, "\n"), "\n") + 1
    headerHeight := nameLines + 2
    sepTop := 2
    stepHeight := len(m.steps) + 2
    sepBetween := 2
    footerHeight := 1
    remaining := h - (headerHeight + sepTop + stepHeight + sepBetween + footerHeight)
    if remaining < 3 { remaining = 3 }
    maxLogContent := remaining - 2
    if maxLogContent < 1 { maxLogContent = 1 }

    var logLines []string
    sel := m.idx
    if sel >= len(m.steps) { sel = len(m.steps) - 1 }
    if sel >= 0 && len(m.steps) > 0 {
        log := strings.TrimSpace(m.steps[sel].logBuf.String())
        if log != "" {
            logLines = strings.Split(log, "\n")
            if len(logLines) > maxLogContent {
                logLines = append([]string{"…"}, logLines[len(logLines)-maxLogContent:]...)
            }
        }
    }
    var logsB strings.Builder
    if len(logLines) == 0 { logsB.WriteString(dimStyle.Render("no output yet")) } else {
        for _, ln := range logLines { logsB.WriteString(dimStyle.Render(ln) + "\n") }
    }
    logsBox := boxStyle.Render(logsB.String())

    // Footer
    status := "Running… press q to quit"
    if m.done {
        failedAny := false
        for _, s := range m.steps { if s.status == failed { failedAny = true; break } }
        if failedAny { status = "Finished with errors • Press Enter or q to exit" } else { status = "All steps complete • Press Enter or q to exit" }
    }
    footer := dimStyle.Render(status + " • If PATH was updated, open a new terminal and run: largo version")

    body := stepsBox + "\n\n" + logsBox
    return header + "\n\n" + body + "\n\n" + footer
}
