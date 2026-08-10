// Command fixxl scans a directory of spreadsheets and writes clean clone
// workbooks into an output folder, without ever modifying the source files.
//
//	fixxl [searchDIR] [-o outdir] [-p]
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/krishnasureshcpa/fixxl/internal/engine"
	"github.com/krishnasureshcpa/fixxl/internal/ui"
)

func main() {
	// `fixxl demo` is a subcommand: demos never touch real files.
	args := os.Args[1:]
	if len(args) > 0 && args[0] == "demo" {
		demo(demoPlain(args[1:]))
		return
	}

	out := flag.String("out", ".fixxl-out", "clone output directory")
	flag.StringVar(out, "o", ".fixxl-out", "alias for -out")
	plain := flag.Bool("plain", false, "plain text output instead of the TUI")
	flag.BoolVar(plain, "p", false, "alias for -plain")
	help := flag.Bool("help", false, "show help")
	flag.BoolVar(help, "h", false, "alias for -help")
	flag.Parse()

	if *help {
		usage()
		return
	}

	dir := "."
	if args := flag.Args(); len(args) > 0 {
		dir = args[0]
	}

	// Go's flag package stops parsing at the first non-flag token, so
	// `fixxl DIR -p -o DIR` leaves the tail as positional args. Re-scan the
	// tail so flags after the directory still take effect.
	for i := 0; i < len(flag.Args()); i++ {
		a := flag.Args()[i]
		switch {
		case a == "-p" || a == "-plain" || a == "--plain" || a == "--p":
			*plain = true
		case a == "-h" || a == "-help" || a == "--help" || a == "--h":
			*help = true
		case strings.HasPrefix(a, "-out=") || strings.HasPrefix(a, "-o=") ||
			strings.HasPrefix(a, "--out=") || strings.HasPrefix(a, "--o="):
			*out = strings.SplitN(a, "=", 2)[1]
		case a == "-o" || a == "-out" || a == "--out" || a == "--o":
			if i+1 < len(flag.Args()) {
				*out = flag.Args()[i+1]
				i++
			}
		}
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "fixxl:", err)
		os.Exit(1)
	}

	if *plain {
		runPlain(abs, *out)
		return
	}

	model := ui.New(abs, *out)
	if _, err := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "fixxl:", err)
		os.Exit(1)
	}
}

// demoPlain reports whether demo should render plain output.
func demoPlain(args []string) bool {
	for _, a := range args {
		if a == "-p" || a == "-plain" || a == "--plain" || a == "--p" {
			return true
		}
	}
	return false
}

func runPlain(dir, out string) {
	opts := engine.Options{OutDir: out}
	files, err := engine.Discover(dir, opts)
	if err != nil {
		fmt.Fprintln(os.Stderr, "fixxl:", err)
		os.Exit(1)
	}
	if len(files) == 0 {
		fmt.Println("no target spreadsheets found in", dir)
		return
	}
	fmt.Printf("fixxl · %d file(s) · clone → %s\n\n", len(files), out)
	ok, refused := 0, 0
	for _, f := range files {
		j := engine.Process(f, opts, now())
		if j.Refusal {
			refused++
		} else {
			ok++
		}
		mark := "✓"
		if j.Refusal {
			mark = "✕"
		}
		fmt.Printf("  %s %-32s %12s rows  %-10s %s\n",
			mark, j.Name, formatRows(j.Rows), j.Style, j.Verify)
		if j.Refusal {
			for _, a := range j.Audit {
				if a.Kind == "err" {
					fmt.Printf("      ↳ %s\n", a.Text)
				}
			}
		}
	}
	fmt.Printf("\n%d converted · %d refused\n", ok, refused)
}

func now() time.Time { return time.Now() }

// formatRows renders a row count with thousands separators for the HTML-free
// plain report.
func formatRows(n int64) string {
	sign := ""
	if n < 0 {
		sign = "-"
		n = -n
	}
	d := strconv.FormatInt(n, 10)
	if len(d) <= 3 {
		return sign + d
	}
	var b strings.Builder
	b.WriteString(sign)
	for i, r := range d {
		if i > 0 && (len(d)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(r)
	}
	return b.String()
}

func usage() {
	fmt.Print(`fixxl  scan · clone · repair

converts spreadsheet files into clean clones; the source is never written.

usage:
  fixxl [SEARCH_DIR] [flags]
  fixxl demo          run a built-in sample batch, no files needed

flags:
  -o, -out DIR    clone output directory        (default .fixxl-out)
  -p, -plain       plain text output, no TUI
  -h, -help        show this help

examples:
  fixxl ./invoices
  fixxl ./invoices -o ./cleaned
  fixxl . -p
`)
}
