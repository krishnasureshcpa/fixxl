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

	model, err := ui.New(dir, out)
	if err != nil {
		fmt.Fprintln(os.Stderr, "fixxl:", err)
		os.Exit(1)
	}
	if _, err := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "fixxl:", err)
		os.Exit(1)
	}
}

func writeDemoFiles(dir string) error {
	f := excelize.NewFile()
	f.SetActiveSheet(0)
	f.SetSheetName("Sheet1", "Sales")
	for i, h := range []string{"region", "rep", "sku", "qty", "total"} {
		c, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue("Sales", c, h)
	}
	for r := 1; r <= 40; r++ {
		c1, _ := excelize.CoordinatesToCellName(1, r+1)
		c2, _ := excelize.CoordinatesToCellName(2, r+1)
		c5, _ := excelize.CoordinatesToCellName(5, r+1)
		f.SetCellValue("Sales", c1, []string{"north", "south", "east"}[r%3])
		f.SetCellValue("Sales", c2, fmt.Sprintf("rep-%d", r))
		f.SetCellValue("Sales", c5, float64(r)*2.5)
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
	// legacy refusal makes the demo honest about limits.
	return os.WriteFile(filepath.Join(dir, "legacy.xls"),
		[]byte("\xd0\xcf\x11\xe0\xa1\xb1\x1a\xe1 placeholder"), 0o644)
}
