package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/krishnasureshcpa/fixxl/internal/engine"
)

// ledger (assurance) + closing report + statusbar.

func (m Model) ledgerPanel(pr palette) string {
	var b strings.Builder
	b.WriteString(panelHead(pr, "assurance ledger", fmt.Sprintf("%d rows", len(m.jobs))))
	// header
	cols := []string{"R", "File", "Sh", "Rows", "Cols", "Style", "Verify", "Sec"}
	roww := []int{3, 22, 5, 10, 7, 10, 8, 6}
	hdr := make([]string, len(cols))
	for i, h := range cols {
		hdr[i] = lipgloss.NewStyle().Foreground(lipgloss.Color(pr.dim)).Render(pad(h, roww[i]))
	}
	b.WriteString("\n" + strings.Join(hdr, ""))
	for i, jb := range m.jobs {
		if i > 40 {
			b.WriteString("\n  …")
			break
		}
		g, col := glyph(jb, pr)
		cells := []string{
			g,
			truncate(jb.Name, roww[1]),
			fmt.Sprintf("%d", jb.Sheets),
			thou(jb.Rows),
			fmt.Sprintf("%d", jb.Cols),
			jb.Style,
			jb.Verify,
			jb.Sec,
		}
		line := make([]string, len(cells))
		line[0] = lipgloss.NewStyle().Foreground(lipgloss.Color(col)).Render(g)
		for i := 1; i < len(cells); i++ {
			line[i] = pad(cells[i], roww[i])
		}
		b.WriteString("\n" + strings.Join(line, ""))
	}
	return b.String()
}

func glyph(jb engine.Job, pr palette) (string, string) {
	if jb.Refusal {
		return lipgloss.NewStyle().Foreground(lipgloss.Color(pr.bad)).Render("✕"), pr.bad
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color(pr.ok)).Render("✓"), pr.ok
}

func (m Model) ledgerHead(pr palette, s string) string { return panelHead(pr, "assurance ledger", s) }

func (m Model) reportPanel(pr palette) string {
	if !m.done {
		return statusbar(m, pr)
	}
	var okFiles, rows int64
	for _, jb := range m.jobs {
		if !jb.Refusal {
			okFiles++
			rows += jb.Rows
		}
	}
	var b strings.Builder
	b.WriteString(panelHead(pr, "assurance report", "verified"))
	b.WriteString(fmt.Sprintf("\n%d files assured", okFiles))
	b.WriteString(fmt.Sprintf("   ·   %s rows processed", thou(rows)))
	b.WriteString(fmt.Sprintf("   ·   run #%d", m.runNo))
	b.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color(pr.dim)).Render("press r to run the next batch"))
	return b.String()
}

func statusbar(m Model, pr palette) string {
	var ok, ref int
	for _, jb := range m.jobs {
		if jb.Refusal {
			ref++
		} else {
			ok++
		}
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("ok %d   refused %d   |   queue %d/%d",
		ok, ref, len(m.jobs), len(m.files)))
	b.WriteString("   " + lipgloss.NewStyle().Foreground(lipgloss.Color(pr.dim)).Render(
		"q quit · t theme · r rerun · ? help"))
	return b.String()
}

func progressBarPct(pct float64) string { return progressBar(pct) }
