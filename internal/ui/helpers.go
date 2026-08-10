package ui

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// tipLine renders the current worksheet tip.
func tipLine(pr palette, tips []string, i int) string {
	i = ((i % len(tips)) + len(tips)) % len(tips)
	label := lipgloss.NewStyle().Foreground(lipgloss.Color(pr.accent)).Render("ti")
	txt := lipgloss.NewStyle().Foreground(lipgloss.Color(pr.fg)).Render(tips[i])
	cnt := lipgloss.NewStyle().Foreground(lipgloss.Color(pr.dim)).Render(counter(i, len(tips)))
	return label + "  " + txt + "  " + cnt
}

func counter(i, total int) string {
	return itoa(i+1) + "/" + itoa(total)
}

func itoa(n int) string {
	if n < 10 {
		return string(rune('0' + n))
	}
	return strconv.Itoa(n)
}

func pad(s string, n int) string {
	if len(s) >= n {
		return s[:n]
	}
	return s + strings.Repeat(" ", n-len(s))
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}

func thou(n int64) string {
	return strconv.FormatInt(n, 10)
}

func colored(hex, text string, pr palette) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(hex)).Render(text)
}
