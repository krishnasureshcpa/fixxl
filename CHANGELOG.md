# Changelog

All notable changes to fixxl are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and it adopts
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.1] - 2026-08-10

### Fixed

- `fixxl DIR -p -o OUT` — flags placed after the directory argument are now
  honored; previously they were treated as positional args and the flags-only
  paths (plain output, custom out dir) were silently ignored

## [0.1.0] - 2026-08-10

First release. A spreadsheet scanner that converts files into clean `.xlsx`
clones out-of-place and structurally re-verifies every conversion.

### Added

- Scan any directory; converts `.xlsx`/`.xlsm`/`.csv`/`.txt` into single or
  multi-sheet `.xlsx` clones, refuses `.xls`/`.xlsb`/`.ods`/XML/HTML with
  actionable advice
- Structural read-back assurance: every clone is reopened and its grid
  compared to the source before it is marked `ok`
- Multi-sheet workbooks are kept whole; reported rows are the total across
  every readable sheet, never just the first
- Sources are never modified — output goes to `.fixxl-out/` (or `--out`)
- Skips Excel `~$` lock files, hidden files, and its own previous clones so
  a rerun never eats its own output
- Interactive TUI (ink and paper themes) and a `--plain` report mode, plus a
  zero-setup `fixxl demo` batch
- Cross-platform release builds (darwin/linux/windows, amd64/arm64) with
  SHA256SUMS, a curl installer (`scripts/install.sh`), a PowerShell installer
  (`scripts/install.ps1`), and an `npx fixxl` shim
- Homebrew formula template and CI test/release pipelines