package engine

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/xuri/excelize/v2"
)

func TestCSVBecomesWorkbook(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, ".fixxl-out")
	csv := filepath.Join(dir, "sales.csv")
	if err := os.WriteFile(csv, []byte("region,qty\nnorth,10\nsouth,20\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	o := Options{OutDir: outDir}
	j := Process(csv, o, time.Now())
	if j.Refusal {
		t.Fatalf("expected ok job, got refusal: %+v", j.Audit)
	}
	if j.Rows != 3 {
		t.Fatalf("rows = %d, want 3 (header + 2)", j.Rows)
	}
	if !fileExists(j.Path) {
		t.Fatalf("clone missing: %s", j.Path)
	}
	f, err := excelize.OpenFile(j.Path)
	if err != nil {
		t.Fatalf("clone not a valid xlsx: %v", err)
	}
	defer f.Close()
	got, _ := f.GetRows("Sheet1")
	if len(got) != 3 { // header + 2 data rows
		t.Fatalf("clone sheet rows = %d, want 3", len(got))
	}
}

func TestRefusedLegacyFormat(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "old.xls")
	if err := os.WriteFile(p, []byte("\xd0\xcf\x11\xe0 legacy bytes"), 0o644); err != nil {
		t.Fatal(err)
	}
	j := Process(p, Options{}, time.Now())
	if !j.Refusal {
		t.Fatal("expected legacy .xls to be refused")
	}
	if j.Verify != "refused" {
		t.Fatalf("Verify = %q, want refused", j.Verify)
	}
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}
