# Use cases

Real flows, distilled from the UAT runs on consumer spreadsheets
(`~/MasterBase/krishna-tools/fixxl/`).

## Batch "which of these old workbooks are still worth keeping?"

Point fixxl at a folder of archive workbooks. Files that convert are
structurally sound — the clone is proof. Files that refuse tell you why.

- Batch of 5 retail workbooks (~850 MB total, up to 213 MB each)
- clone output: `.fixxl-out/` (a fresh `--out` each pass)

## "Re-save as .xlsx so my tool can read it"

Other business tools only ingest `.xlsx`. Instead of opening Excel and
re-saving hundreds of files by hand, run fixxl over the folder. Every
supported workbook (`.xlsx`/`.xlsm`/`.csv`/`.txt`) lands in one output
folder as a clean `.xlsx` clone; the originals are never touched.

## "Twice the size, same data" — debloat a bloated workbook

Test result: `Retail_Inventory_Obsolescence_Prior.xlsx` — 213,746,838 bytes
in → 178,670,288 bytes out (~-16%), 672,087 rows re-verified as intact.

UAT numbers off the consumer's real retail inventory workbooks:

| file | rows | result |
| --- | --- | --- |
| Retail_Inventory_Obsolescence_Prior.xlsx | 672,087 | intact · ok |
| WeeklySalesOrderingMultistore.xlsx | 830,451 | intact · ok |

## "Validate my reports before I trust them"

That is the whole now-worth: a `.xlsx` that clones clean and re-verifies
row-for-row (`structural ok`) is one where every sheet can be re-read by a
fresh engine — a fair stand-in for "the file is not silently corrupt".

## "Ship a clean copy to the auditor, not the original"

Handing over financial workbooks per auditor request usually means
duplicating or flattening. fixxl emits a clean clone, one command, no
source-write. The output is auditor-re-read friendly.

## "It still runs on my old Windows box"

`npx fixxl demo` crosses the Windows gap without any binary move:

```sh
npx fixxl demo
```

(falls back to the GitHub windows binary + `get0` PowerShell shim.)

## "I don't want my files touched, but I need to know what's there"

`fixxl . -p` reads every workbook, prints a row-and-verify report, and
writes clones into a side folder. When the disk is full mid-batch it
refuses the rest individually with the reason (`no space left on device`).
Nothing is ever written next to the source.