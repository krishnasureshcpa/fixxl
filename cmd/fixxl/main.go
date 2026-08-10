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
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/krishnasureshcpa/fixxl/internal/engine"
	"github.com/krishnasureshcpa/fixxl/internal/ui"
)

func main() {
	// `fixxl demo` bypasses normal flag parsing so `-p` works after it.
	for _, a := range os.Args[1:] {
		if a == "demo" {
			plain := false
			for _, x := range os.Args[2:] {
				if x == "-p" || x == "-plain" || x == "--plain" || x == "--p" {
					plain = true
				}
			}
			demo(plain)
			return
		}
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
	abs, err := filepath.Abs(dir)
	if err != nil {
		abs = dir
	}

	if *plain {
		runPlain(abs, *out)
		return
	}

	model, err := ui.New(abs, *out)
	if err != nil {
		fmt.Fprintln(os.Stderr, "fixxl:", err)
		os.Exit(1)
	}
	if _, err := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "fixxl:", err)
		os.Exit(1)
	}
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
		j, err := engine.Process(f, opts, now())
		if err != nil {
			j = engine.Job{Name: filepath.Base(f), Refusal: true, Verify: "refused",
				Audit: []engine.AuditLine{{Kind: "err", Text: err.Error()}}}
		}
		mark := "✓"
		if j.Refusal {
			mark = "✕"
			refused++
		} else {
			ok++
		}
		fmt.Printf("  %s %-28s %5d rows  %-10s %s\n",
			mark, j.Name, j.Rows, j.Style, j.Verify)
	}
	fmt.Printf("\n%d converted · %d refused\n", ok, refused)
}

func now() time.Time { return time.Now() }

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
