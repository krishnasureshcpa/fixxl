package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/krishnasureshcpa/fixxl/internal/engine"
)

const blockFull = "█"
const blockEmpty = "░"

func (m Model) frame(pr palette) string {
	var b strings.Builder
	b.WriteString(banner(pr))
	b.WriteString("\n" + tipLine(pr, m.tips, m.tip))
	b.WriteString("\n" + m.workPanel(pr))
	b.WriteString("\n" + m.inspectPanel(pr))
	b.WriteString("\n" + m.resolvePanel(pr))
	b.WriteString("\n" + m.ledgerPanel(pr))
	b.WriteString("\n" + m.reportPanel(pr))
	return b.String()
}

func banner(pr palette) string {
	s := lipgloss.NewStyle().
		Foreground(lipgloss.Color(pr.accent)).
		Bold(true)
	title := s.Render("fixxl")
	sub := lipgloss.NewStyle().Foreground(lipgloss.Color(pr.dim)).Render("scan · clone · assure")
	return title + "  " + sub
}

func tip(pr palette, tips []string, i int) string {
	idx := i % len(tips)
	label := lipgloss.NewStyle().Foreground(lipgloss.Color(pr.accent)).Render("ti")
	txt := lipgloss.NewStyle().Foreground(lipgloss.Color(pr.fg)).Render(tips[idx])
	cnt := lipgloss.NewStyle().Foreground(lipgloss.Color(pr.dim)).Render(fmt.Sprintf("%d/%d", idx+1, len(tips)))
	return label + "  " + txt + "  " + cnt
}

func (m Model) currentJob() *engine.Job {
	if len(m.jobs) == 0 {
		return nil
	}
	i := len(m.jobs) - 1
	return &m.jobs[i]
}

func (m Model) workPanel(pr palette) string {
	n := len(m.files)
	c := len(m.jobs)
	pct := 0.0
	if n > 0 {
		pct = float64(c) / float64(n) * 100
	}
	blocks := progressBar(pct)
	name := "…"
	if cb := m.currentJob(); cb != nil {
		name = cb.Name
	}
	hdr := panelHead(pr, "active work", statusOf(m))
	title := lipgloss.NewStyle().Foreground(lipgloss.Color(pr.fg)).Bold(true).Render(name)
	bar := lipgloss.NewStyle().Foreground(lipgloss.Color(pr.accent)).Render(blocks)
	pctS := lipgloss.NewStyle().Foreground(lipgloss.Color(pr.dim)).Render(fmt.Sprintf("%3.1f%%  %d/%d", pct, c, n))
	return hdr + "\n" + title + "\n" + bar + " " + pctS
}

func (m Model) inspectPanel(pr palette) string {
	var b strings.Builder
	b.WriteString(panelHead(pr, "inspect", ""))
	if cb := m.currentJob(); cb != nil {
		for _, a := range cb.Audit {
			g := map[string]string{"ok": "✓", "warn": "▲", "err": "✕"}[a.Kind]
			c := map[string]string{"ok": pr.ok, "warn": pr.accent, "err": pr.bad}[a.Kind]
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(c)).Render(g) + " " + a.Text + "\n")
		}
	}
	return strings.TrimSuffix(b.String(), "\n")
}

func (m Model) resolvePanel(pr palette) string {
	cb := m.currentJob()
	var b strings.Builder
	if cb == nil {
		return panelHead(pr, "resolve", "auto")
	}
	b.WriteString(panelHead(pr, "resolve", ""))
	if cb.Refusal {
		b.WriteString(colored(pr.bad, "refused", pr) + "\n")
		for _, a := range cb.Advice {
			b.WriteString("  • " + a + "\n")
		}
	} else {
		for _, s := range cb.Steps {
			kind := lipgloss.NewStyle().Foreground(lipgloss.Color(pr.accent)).Render(s.Kind)
			b.WriteString(kind + "  " + s.Text + "\n")
		}
	}
	return strings.TrimSuffix(b.String(), "\n")
}

func panelHead(pr palette, title, right string) string {
	h := lipgloss.NewStyle().Foreground(lipgloss.Color(pr.accent)).Render("▮ " + title)
	if right != "" {
		h += lipgloss.NewStyle().Foreground(lipgloss.Color(pr.dim)).Render("  " + right)
	}
	return h
}

func panelHeadCore(pr palette, title, right string) string {
	return panelHead(pr, title, right)
}

func progressBar(pct float64) string {
	const n = 26
	full := int(pct / 100 * n)
	return strings.Repeat(blockFull, full) + strings.Repeat(blockEmpty, n-full)
}

func statusOf(m Model) string {
	if m.done {
		return "done"
	}
	if len(m.jobs) == 0 {
		return "scanning…"
	}
	return "processing"
}
