<p align="center">
  <img alt="fixxl" width="560" src="assets/fixxl-preview.gif">
</p>

<p align="center">
  converts spreadsheet files into clean clones —
  <b>the source is never written.</b>
</p>

<p align="center">
  <a href="#install">install</a> ·
  <a href="#usage">usage</a> ·
  <a href="#what-it-checks">what it checks</a> ·
  <a href="./preview/fixxl.html">html preview</a>
</p>

## Why

Spreadsheets arrive as `.xlsx`, `.xls`, `.csv`, `.txt`, `.ods`, or worse.
People hand you a file that **your downstream system refuses** — a legacy
`.xls` when only `.xlsx` is accepted, a text paste that isn't a spreadsheet
at all, a file that "opens fine in Excel" but crashes the import.

`fixxl` reads any spreadsheet you point it at, converts what it can into a
modern `.xlsx`, and **refuses** the rest with a reason — never silently
skipping, never writing to the source file.

- **clones, never mutates** — output goes to `.fixxl-out/` next to the input
  (or `--out DIR`); the original never changes
- **structural assurance** — every conversion is read back and verified
  (row/column counts must match) or marked `refused`
- **legacy is honest** — `.xls`, `.xlsb`, `.ods` are recognized and refused
  with advice, not ignored
- **one binary** — fully offline, no runtime, cross-platform

## Install

```sh
# curl (macOS / linux)
curl -fsSL https://raw.githubusercontent.com/krishnasureshcpa/fixxl/main/scripts/install.sh | sh

# homebrew (once the formula lands)
brew install fixxl

# go (anywhere)
go install github.com/krishnasureshcpa/fixxl/cmd/fixxl@latest
```

Windows: `npx fixxl` is planned — or grab the
[release](https://github.com/krishnasureshcpa/fixxl/releases/latest) binary.

## Usage

```sh
fixxl ./invoices              # scan dir, convert everything convertible
fixxl ./invoices -o ./clean   # clone output elsewhere
fixxl . -p                    # plain text output, no TUI
fixxl demo                    # run a built-in sample batch, no files needed
```

Each file ends with one of:

| status        | meaning |
| ------------- | ------- |
| `structural ok` | converted, and the clone was read back with a matching grid |
| `intact ok`   | already a clean `.xlsx`; still cloned so the output is complete |
| `refused`     | not convertible (`legacy.xls`); a reason is shown |

`--out` clones are **noisy-is-clean**: every converted file lands there
even when it was already fine, so you can move the whole folder into your
downstream system in one go.

## What it checks

- opens every file with its real reader — `.xlsx`/`.xlsm` via
  `excelize`, `.csv`/`.txt` with robust separator detection, legacy and
  unknown with a refusal
- converts to a single normalized `.xlsx` (first sheet, headers honored)
- **read-back check**: the clone is reopened and cell-grid vs counts are
  compared to the source → `structural ok` or `refused`
- refuses `.xls`, `.xlsb`, `.ods`, XML, HTML with actionable advice
  ("re-save as .xlsx in the source app") — surfaced, never silently dropped

## Build & test

```sh
make build       # local binary at ./bin/fixxl
make test        # go test ./...
make dist        # cross-platform release binaries → dist/
```

CI: `.github/workflows/test.yml` (vet + test always, cross-build on PR) and
`.github/workflows/release.yml` (tag `v*` → binaries + SHA256SUMS).

## Examples

`examples/` holds a real multi-sheet workbook, a csv, a text table, and a
`legacy.xls` the tool politely refuses — regenerate with
`go run ./scripts/gen-examples`.

## License

MIT