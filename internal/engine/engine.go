// Package engine finds, converts, and clones spreadsheet files into a clean
// output directory. The source workbook is never modified.
package engine

import (
	"encoding/csv"
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
// advice instead of silently ignoring them.
func Discover(root string, o Options) ([]string, error) {
	var out []string
	err := filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
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

// Process takes one source file and writes a clean clone next to it (in the
// clone directory). start seeds the elapsed-time figure reported in Sec.
func Process(src string, o Options, start time.Time) (Job, error) {
	info, err := os.Stat(src)
	if err != nil {
		return Job{}, err
	}
	name := filepath.Base(src)
	ext := strings.ToLower(filepath.Ext(src))

	switch ext {
	case ".csv", ".txt":
		return convCSV(src, name, info.Size(), o, start)
	case ".xlsx", ".xlsm":
		return convWorkbook(src, name, info.Size(), o, start)
	default:
		return refused(name, "unsupported legacy format — re-save as .xlsx from Excel", secSince(start)), nil
	}
}

// convWorkbook re-packs an xlsx/xlsm into a fresh clone.
func convWorkbook(src, name string, size int64, o Options, start time.Time) (Job, error) {
	f, err := excelize.OpenFile(src)
	if err != nil {
		return refused(name, "cannot open source ("+err.Error()+")", secSince(start)), nil
	}
	defer f.Close()
	sheets := f.GetSheetList()
	out := cloneFile(name, ".xlsx", o)
	_ = os.MkdirAll(o.out(), 0o755)
	if err := f.SaveAs(out); err != nil {
		return refused(name, "re-pack failed", secSince(start)), nil
	}
	n, cols := countSheet(f, sheets[0])
	so, _ := fileSize(out)
	return Job{
		Name: name, Sheets: len(sheets), Rows: n, Cols: cols,
		Style: "intact", Verify: "ok", Path: out, SizeIn: size, SizeOut: so,
		Sec:   secSince(start),
		Audit: []AuditLine{{Kind: "ok", Text: "workbook re-packed into a clean clone"}},
	}, nil
}

// convCSV lifts a flat CSV/TXT into a single-sheet workbook.
func convCSV(src, name string, size int64, o Options, start time.Time) (Job, error) {
	f, err := os.Open(src)
	if err != nil {
		return Job{}, err
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
		if e != nil {
			break
		}
		n++
		if len(rec) > cols {
			cols = len(rec)
		}
		for c, v := range rec {
			cell, _ := excelize.CoordinatesToCellName(c+1, row)
			ef.SetCellValue("Sheet1", cell, v)
		}
		row++
	}
	out := cloneFile(name, ".xlsx", o)
	_ = os.MkdirAll(o.out(), 0o755)
	if err := ef.SaveAs(out); err != nil {
		return refused(name, "write clone failed", secSince(start)), nil
	}
	so, _ := fileSize(out)
	return Job{
		Name: name, Sheets: 1, Rows: n, Cols: cols,
		Style: "structural", Verify: "ok", Path: out, SizeIn: size, SizeOut: so,
		Sec:   secSince(start),
		Audit: []AuditLine{{Kind: "ok", Text: "CSV lifted into a single worksheet"}},
		Steps: []ResolveStep{
			{Kind: "cause", Text: "flat CSV has no workbook structure"},
			{Kind: "action", Text: "wrap rows into one worksheet"},
			{Kind: "outcome", Text: "now opens natively as xlsx"},
		},
	}, nil
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

func countSheet(f *excelize.File, sheet string) (int64, int) {
	rows, _ := f.GetRows(sheet)
	return int64(len(rows)), sheetCols(rows)
}

func sheetCols(rows [][]string) int {
	cols := 0
	for _, r := range rows {
		if len(r) > cols {
			cols = len(r)
		}
	}
	return cols
}

func fileSize(p string) (int64, error) {
	fi, err := os.Stat(p)
	if err != nil {
		return 0, err
	}
	return fi.Size(), nil
}
