// Command gen-examples produces the committed examples/ folder: a real
// multi-sheet workbook, a CSV, a text table, and a legacy .xls the tool
// will politely refuse.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/xuri/excelize/v2"
)

func main() {
	dir := "examples"
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		panic(err)
	}

	// sales.xlsx — two sheets, a few hundred rows so it feels real.
	f := excelize.NewFile()
	sheet := "Sales"
	f.SetSheetName("Sheet1", sheet)
	headers := []string{"region", "rep", "sku", "qty", "unit_price", "total"}
	for i, h := range headers {
		c, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, c, h)
	}
	regions := []string{"north", "south", "east", "west"}
	reps := []string{"ada", "lin", "marie", "jon", "priya"}
	skuBase := 40000
	for r := 0; r < 120; r++ {
		for i, v := range []any{
			regions[r%len(regions)],
			reps[r%len(reps)],
			skuBase + r,
			r%97 + 1,
			1.5 + float64(r%30)/4,
			nil, // filled below
		} {
			c, _ := excelize.CoordinatesToCellName(i+1, r+2)
			f.SetCellValue(sheet, c, v)
		}
		q := r%97 + 1
		pr := 1.5 + float64(r%30)/4
		total := float64(q) * pr
		c, _ := excelize.CoordinatesToCellName(6, r+2)
		f.SetCellValue(sheet, c, total)
	}
	idx, _ := f.NewSheet("Summary")
	f.SetActiveSheet(idx)
	f.SetCellValue("Summary", "A1", "total sales")
	totalCell, _ := excelize.CoordinatesToCellName(2, 1)
	f.SetCellValue("Summary", totalCell, "=SUM(Sales!F2:F121)")
	if err := f.SaveAs(filepath.Join(dir, "sales.xlsx")); err != nil {
		panic(err)
	}

	// sales.csv — plain tabular.
	csv := "name,qty,price\napples,5,1.20\npears,3,0.90\nmangos,12,2.10\n"
	if err := os.WriteFile(filepath.Join(dir, "sales.csv"), []byte(csv), 0o644); err != nil {
		panic(err)
	}

	// notes.txt — text table that gets lifted into a workbook.
	if err := os.WriteFile(filepath.Join(dir, "notes.txt"),
		[]byte("note,owner\nroadmap,ada\nbudget,lin\n"), 0o644); err != nil {
		panic(err)
	}

	// legacy.xls — OLE magic bytes; fixxl refuses with advice.
	if err := os.WriteFile(filepath.Join(dir, "legacy.xls"),
		[]byte("\xd0\xcf\x11\xe0\xa1\xb1\x1a\xe1 legacy workbook placeholder"), 0o644); err != nil {
		panic(err)
	}

	fmt.Println("examples written to", filepath.Join(dir))
}
