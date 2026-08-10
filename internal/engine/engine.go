// Package engine finds, converts, and clones spreadsheet files into a clean
// output directory. The source workbook is never modified.
package engine

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

// Job describes what happened to one file.
type Job struct {
	Name    string
	Sheets  int
	Rows    int64
	Cols    int
	Style   string // intact | structural | refused
	Verify  string // ok | refused
	Sec     string
	Refusal bool
	Launch  bool
	Advice  []string
	Path    string
	SizeIn  int64
	SizeOut int64
	Audit   []AuditLine
	Steps   []ResolveStep
}

// AuditLine is one inspiration-line in the inspector.
type AuditLine struct {
	Kind string
	Text string
}

// ResolveStep narrates cause -> action -> outcome.
type ResolveStep struct {
	Kind string
	Text string
}

// Options controls a batch.
type Options struct {
	OutDir string
}

func (o Options) out() string {
	if o.OutDir != "" {
		return o.OutDir
	}
	return ".fixxl-out"
}

// Discover returns spreadsheet files under root, sorted. Convertible formats
// are processed; legacy formats are surfaced so the run can refuse them with
// advice instead of silently ignoring them. Excel lock files (garbage like
// "~$book.xlsx"), hidden files, and clones already inside the out dir are
// skipped so a rerun never eats its own output.
func Discover(root string, o Options) ([]string, error) {
	var out []string
	err := filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if isNoise(p, info, o) {
			return nil
		}
		switch strings.ToLower(filepath.Ext(p)) {
		case ".xlsx", ".xlsm", ".csv", ".txt", ".xls", ".xlsb", ".ods", ".xml", ".html":
			out = append(out, p)
		}
		return nil
	})
	sort.Strings(out)
	return out, err
}

// isNoise reports files that must never be treated as spreadsheet input.
func isNoise(p string, info os.FileInfo, o Options) bool {
	b := filepath.Base(p)
	if strings.HasPrefix(b, "~$") || strings.HasPrefix(b, ".") {
		return true
	}
	if rel, err := filepath.Rel(o.out(), p); err == nil && rel != "." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && rel != ".." {
		return true
	}
	return false
}

// Process takes one source file and writes a clean clone into the clone
// directory. It never returns a hard error: anything that cannot be converted
// becomes a refusal job with the reason spelled out. start seeds the elapsed
// time reported in Sec.
func Process(src string, o Options, start time.Time) Job {
	info, err := os.Stat(src)
	if err != nil {
		return refused(filepath.Base(src), "cannot read source ("+err.Error()+")", secSince(start))
	}
	name := filepath.Base(src)
	ext := strings.ToLower(filepath.Ext(src))

	switch ext {
	case ".csv", ".txt":
		return convCSV(src, name, info.Size(), o, start)
	case ".xlsx", ".xlsm":
		return convWorkbook(src, name, info.Size(), o, start)
	default:
		return refused(name, "unsupported legacy format — re-save as .xlsx from Excel", secSince(start))
	}
}

// convWorkbook re-packs an xlsx/xlsm into a fresh clone. All sheets are kept;
// the reported row/column figures are the totals across every readable sheet,
// never just the first one.
func convWorkbook(src, name string, size int64, o Options, start time.Time) Job {
	f, err := excelize.OpenFile(src)
	if err != nil {
		return refused(name, "cannot open source ("+err.Error()+")", secSince(start))
	}
	defer f.Close()

	sheets := f.GetSheetList()
	out := cloneFile(name, ".xlsx", o)
	if err := os.MkdirAll(o.out(), 0o755); err != nil {
		return refused(name, "cannot create output dir ("+err.Error()+")", secSince(start))
	}
	if err := f.SaveAs(out); err != nil {
		return refused(name, "re-pack failed ("+err.Error()+")", secSince(start))
	}
	n, cols, warns := countWorkbook(f)
	so, soErr := fileSize(out)

	audit := []AuditLine{{Kind: "ok", Text: "workbook re-packed into a clean clone"}}
	if len(sheets) > 1 {
		audit = append(audit, AuditLine{
			Kind: "ok",
			Text: fmt.Sprintf("%d sheets kept · %d rows across all sheets", len(sheets), n),
		})
	}
	for _, w := range warns {
		audit = append(audit, AuditLine{Kind: "warn", Text: w})
	}
	if soErr != nil {
		audit = append(audit, AuditLine{Kind: "warn", Text: "could not measure clone size"})
	}
	return Job{
		Name: name, Sheets: len(sheets), Rows: n, Cols: cols,
		Style: "intact", Verify: "ok", Path: out, SizeIn: size, SizeOut: so,
		Sec:   secSince(start),
		Audit: audit,
	}
}

// convCSV lifts a flat CSV/TXT into a single-sheet workbook.
func convCSV(src, name string, size int64, o Options, start time.Time) Job {
	f, err := os.Open(src)
	if err != nil {
		return refused(name, "cannot open source ("+err.Error()+")", secSince(start))
	}
	defer f.Close()

	ef := excelize.NewFile()
	ef.SetActiveSheet(0)
	r := csv.NewReader(f)
	r.FieldsPerRecord = -1
	var n int64
	cols := 0
	row := 1
	for {
		rec, e := r.Read()
		if e == io.EOF {
			break
		}
		if e != nil {
			return refused(name, fmt.Sprintf("parse error near line %d (%s)", row, e), secSince(start))
		}
		n++
		if len(rec) > cols {
			cols = len(rec)
		}
		for c, v := range rec {
			cell, cerr := excelize.CoordinatesToCellName(c+1, row)
			if cerr != nil {
				return refused(name, "internal: bad cell coordinate", secSince(start))
			}
			ef.SetCellValue("Sheet1", cell, v)
		}
		row++
	}
	out := cloneFile(name, ".xlsx", o)
	if err := os.MkdirAll(o.out(), 0o755); err != nil {
		return refused(name, "cannot create output dir ("+err.Error()+")", secSince(start))
	}
	if err := ef.SaveAs(out); err != nil {
		return refused(name, "write clone failed ("+err.Error()+")", secSince(start))
	}
	so, soErr := fileSize(out)
	audit := []AuditLine{{Kind: "ok", Text: "CSV lifted into a single worksheet"}}
	if soErr != nil {
		audit = append(audit, AuditLine{Kind: "warn", Text: "could not measure clone size"})
	}
	return Job{
		Name: name, Sheets: 1, Rows: n, Cols: cols,
		Style: "structural", Verify: "ok", Path: out, SizeIn: size, SizeOut: so,
		Sec:   secSince(start),
		Audit: audit,
		Steps: []ResolveStep{
			{Kind: "cause", Text: "flat CSV has no workbook structure"},
			{Kind: "action", Text: "wrap rows into one worksheet"},
			{Kind: "outcome", Text: "now opens natively as xlsx"},
		},
	}
}

// refused fabricates a refusal job.
func refused(name, reason string, sec string) Job {
	return Job{
		Name: name, Style: "refused", Verify: "refused", Refusal: true,
		Sec:   sec,
		Audit: []AuditLine{{Kind: "err", Text: reason}},
		Advice: []string{
			"re-save the source as a .xlsx workbook from Excel",
			"then rerun fixxl on the new file",
		},
		Steps: []ResolveStep{
			{Kind: "cause", Text: "the file is not a readable spreadsheet in this build"},
			{Kind: "action", Text: "leave the original untouched"},
			{Kind: "outcome", Text: "nothing was converted, nothing was corrupted"},
		},
	}
}

// ---- small helpers ----

func secSince(t time.Time) string {
	return time.Since(t).Round(time.Millisecond).String()
}

func cloneFile(name, ext string, o Options) string {
	base := strings.TrimSuffix(name, filepath.Ext(name))
	return filepath.Join(o.out(), base+"_fixxl"+ext)
}

// countWorkbook totals rows and the widest row across every readable sheet.
// Sheets that cannot be streamed (e.g. chartsheets) yield a warning so the
// figure never silently claims to cover a sheet it missed.
func countWorkbook(f *excelize.File) (rows int64, cols int, warns []string) {
	for _, sheet := range f.GetSheetList() {
		n, c, err := countSheet(f, sheet)
		if err != nil {
			warns = append(warns, sheet+": "+err.Error())
			continue
		}
		rows += n
		if c > cols {
			cols = c
		}
	}
	return rows, cols, warns
}

// countSheet streams a sheet row by row so huge files are counted without
// loading the whole grid into memory.
func countSheet(f *excelize.File, sheet string) (int64, int, error) {
	r, err := f.Rows(sheet)
	if err != nil {
		return 0, 0, err
	}
	defer r.Close()
	var n int64
	cols := 0
	for r.Next() {
		cells, err := r.Columns()
		if err != nil {
			return 0, 0, err
		}
		n++
		if len(cells) > cols {
			cols = len(cells)
		}
	}
	return n, cols, r.Error()
}

func fileSize(p string) (int64, error) {
	fi, err := os.Stat(p)
	if err != nil {
		return 0, err
	}
	return fi.Size(), nil
}
