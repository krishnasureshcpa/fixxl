// Package ui renders fixxl as an interactive terminal: a banner, a rotating
// tips rail, per-file progress, a cause->action->outcome narration, an
// assurance ledger, and a closing report.
package ui

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/krishnasureshcpa/fixxl/internal/engine"
)

type palette struct {
	bg, fg, dim, accent, ok, bad, border string
}

var ink = palette{
	bg: "#0c100d", fg: "#e6e8e3", dim: "#9aa5a0",
	accent: "#e2a94c", ok: "#86c06a", bad: "#e4706f", border: "#2a3a30",
}
var paper = palette{
	bg: "#f5f3ee", fg: "#16191b", dim: "#6c786f",
	accent: "#b0762c", ok: "#3e8350", bad: "#b34045", border: "#cfc6b2",
}

// Model is the interactive ferry state.
type Model struct {
	dir  string
	out  string
	t0   time.Time
	tips []string
	tip  int
	help bool

	files  []string
	jobs   []engine.Job
	errMsg string
	done   bool
	quit   bool

	runNo int

	ch        chan tea.Msg
	themeDark bool
	width     int
	height    int
}

// msgBuf bounds how many worker messages can queue ahead of the renderer.
const msgBuf = 16

// New builds a model for a directory.
func New(dir string, outDir string) tea.Model {
	m := Model{
		dir: dir, out: outDir, themeDark: true, t0: time.Now(),
		tips: []string{
			"Never overwrite the source - a clone is written out-of-place, the original never changes.",
			"Numbers stored as text sort wrong - fixxl coerces them back to numerics.",
			"XLSX is a zip: a 'corrupt' file often only needs the archive re-packed.",
			"An Excel-locked file still reads fine - fixxl clones the snapshot and moves on.",
		},
	}
	m.ch = make(chan tea.Msg, msgBuf)
	m.runNo = readRun()
	return m
}

// ---- messages ----

type filesFound struct{ files []string }
type workDone struct{ job engine.Job }
type noFiles struct{}
type wrong struct{ err error }
type endRun struct{}

func (m Model) Init() tea.Cmd {
	go m.worker()
	return m.nextRound()
}

// worker scans and converts every file, always ending with exactly one
// terminal message (endRun, wrong, or noFiles) and never closing the channel
// itself; defer closes it after the final send so nothing is dropped.
func (m *Model) worker() {
	defer close(m.ch)
	if err := os.MkdirAll(m.out, 0o755); err != nil {
		m.ch <- wrong{err}
		return
	}
	files, err := engine.Discover(m.dir, engine.Options{OutDir: m.out})
	if err != nil {
		m.ch <- wrong{err}
		return
	}
	if len(files) == 0 {
		m.ch <- noFiles{}
		return
	}
	opts := engine.Options{OutDir: m.out}
	m.ch <- filesFound{files}
	for _, f := range files {
		j := engine.Process(f, opts, time.Now())
		m.ch <- workDone{j}
	}
	m.ch <- endRun{}
}

func (m Model) nextRound() tea.Cmd {
	return func() tea.Msg {
		v, ok := <-m.ch
		if !ok {
			return nil
		}
		return v
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch v := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = v.Width, v.Height
		return m, nil
	case tea.KeyMsg:
		return m.onKey(v.String())
	case filesFound:
		m.files = v.files
		return m, m.nextRound()
	case workDone:
		m.jobs = append(m.jobs, v.job)
		m.tip = (m.tip + 1) % len(m.tips)
		return m, m.nextRound()
	case noFiles:
		m.errMsg = "no target spreadsheet files in " + m.dir
		m.done = true
		return m, nil
	case wrong:
		m.errMsg = v.err.Error()
		m.done = true
		return m, nil
	case endRun:
		m.done = true
		m.runNo++
		writeRun(m.runNo)
		return m, nil
	}
	// Channel closed without a terminal message: hold the last view rather
	// than quitting mid-frame out from under the user.
	if msg == nil {
		return m, nil
	}
	return m, nil
}

func (m Model) onKey(k string) (tea.Model, tea.Cmd) {
	switch k {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "r":
		if m.done || m.errMsg != "" {
			m.jobs = nil
			m.files = nil
			m.errMsg = ""
			m.done = false
			m.t0 = time.Now()
			m.ch = make(chan tea.Msg, msgBuf)
			go m.worker()
			return m, m.nextRound()
		}
	case "t":
		m.themeDark = !m.themeDark
	case "?":
		m.help = !m.help
	}
	return m, nil
}

func (m Model) p() palette {
	if m.themeDark {
		return ink
	}
	return paper
}

// ==== run persistence ====
// Run numbers are a nicety, not a contract: if the config home is missing or
// the file is unreadable we simply start at zero rather than crashing.

func runFile() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".fixxl-run"), nil
}

func readRun() int {
	f, err := runFile()
	if err != nil {
		return 0
	}
	b, err := os.ReadFile(f)
	if err != nil {
		return 0
	}
	n, err := strconv.Atoi(strings.TrimSpace(string(b)))
	if err != nil || n < 0 {
		return 0
	}
	return n
}

func writeRun(n int) {
	f, err := runFile()
	if err != nil {
		return
	}
	if err := os.WriteFile(f, []byte(strconv.Itoa(n)), 0o644); err != nil {
		return
	}
}

// ==== view ====

func (m Model) View() string {
	pr := m.p()
	if m.quit {
		return ""
	}
	body := m.frame(pr)
	return lipgloss.NewStyle().Background(lipgloss.Color(pr.bg)).Render(body + "\n")
}
