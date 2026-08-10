package main

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/xuri/excelize/v2"

	"github.com/krishnasureshcpa/fixxl/internal/ui"
)

// demo materialises a tiny sample spreadsheet into a temp folder and runs a
// batch over it, so a first-time user sees the tool work with zero setup.
func demo(plain bool) {
	dir, err := os.MkdirTemp("", "fixxl-demo")
	if err != nil {
		fmt.Fprintln(os.Stderr, "fixxl:", err)
		os.Exit(1)
	}
	defer os.RemoveAll(dir)
	if err := writeDemoFiles(dir); err != nil {
		fmt.Fprintln(os.Stderr, "fixxl:", err)
		os.Exit(1)
	}
	out := filepath.Join(dir, ".fixxl-out")

	if plain {
		runPlain(dir, out)
		return
	}

	model := ui.New(dir, out)
	if _, err := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "fixxl:", err)
		os.Exit(1)
	}
}

// sheetName is the one worksheet the demo workbook writes.
const sheetName = "Sales"

// cell is a strict CoordinatesToCellName: the demo is a fixed script, so a bad
// coordinate is a programming bug and failing fast is the honest response.
func cell(col, row int) string {
	c, err := excelize.CoordinatesToCellName(col, row)
	if err != nil {
		panic(err)
	}
	return c
}

func writeDemoFiles(dir string) error {
	f := excelize.NewFile()
	f.SetActiveSheet(0)
	f.SetSheetName("Sheet1", sheetName)
	for i, h := range []string{"region", "rep", "sku", "qty", "total"} {
		f.SetCellValue(sheetName, cell(i+1, 1), h)
	}
	for r := 1; r <= 40; r++ {
		f.SetCellValue(sheetName, cell(1, r+1), []string{"north", "south", "east"}[r%3])
		f.SetCellValue(sheetName, cell(2, r+1), fmt.Sprintf("rep-%d", r))
		f.SetCellValue(sheetName, cell(5, r+1), float64(r)*2.5)
	}
	if err := f.SaveAs(filepath.Join(dir, "sales.xlsx")); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "sales.csv"),
		[]byte("name,qty\napples,5\npears,3\nmangos,12\n"), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "notes.txt"),
		[]byte("note,owner\nroadmap,ada\nbudget,lin\n"), 0o644); err != nil {
		return err
	}
	// A real .xls is a compound document whose first bytes are the OLE magic
	// D0 CF 11 E0 A1 B1 1A E1. Writing that magic into a stub makes the demo
	// refusal honest: the tool refuses by format sniffing, not by content.
	return os.WriteFile(filepath.Join(dir, "legacy.xls"),
		[]byte("\xd0\xcf\x11\xe0\xa1\xb1\x1a\xe1 placeholder"), 0o644)
}
