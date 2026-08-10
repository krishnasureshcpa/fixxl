# fixxl install (Windows PowerShell) - downloads the release binary and
# verifies its SHA-256 before installing.
#
#   irm https://raw.githubusercontent.com/krishnasureshcpa/fixxl/main/scripts/install.ps1 | iex
#
# Overrides: $FIXXL_VERSION, $FIXXL_BASE, $prefix
$ErrorActionPreference = "Stop"

$repo = if ($env:FIXXL_REPO) { $env:FIXXL_REPO } else { "krishnasureshcpa/fixxl" }
$base = if ($env:FIXXL_BASE) { $env:FIXXL_BASE } else { "https://github.com/$repo/releases/download" }
$version = if ($env:FIXXL_VERSION) { $env:FIXXL_VERSION } else { "latest" }

if ($version -eq "latest") {
    $release = Invoke-RestMethod "https://api.github.com/repos/$repo/releases/latest"
    $version = $release.tag_name
}

$arch = switch ($env:PROCESSOR_ARCHITECTURE) {
    "AMD64" { "amd64" }
    "ARM64" { "arm64" }
    default { throw "fixxl: unsupported arch $env:PROCESSOR_ARCHITECTURE" }
}
$asset = "fixxl-windows-$arch.exe"
$url = "$base/$version/$asset"
$sumsUrl = "$base/$version/SHA256SUMS"

$tmp = New-Item -ItemType Directory (Join-Path $env:TEMP "fixxl-install-$(Get-Random)")

try {
    Write-Host "fixxl $version (windows/$arch)"
    Write-Host "  download  $url"
    Invoke-WebRequest -Uri $url -OutFile (Join-Path $tmp $asset)
    Invoke-WebRequest -Uri $sumsUrl -OutFile (Join-Path $tmp "SHA256SUMS")

    $wantHash = (Get-Content (Join-Path $tmp "SHA256SUMS") | Where-Object { $_ -like "*$asset" }) -split "\s+" | Select-Object -First 1
    $gotHash = (Get-FileHash (Join-Path $tmp $asset) -Algorithm SHA256).Hash.ToLower()
    if ($gotHash -ne $wantHash.ToLower()) {
        throw "fixxl: checksum mismatch - aborting"
    }

    $prefix = if ($env:FIXXL_PREFIX) { $env:FIXXL_PREFIX } else { Join-Path $HOME ".local\bin" }
    New-Item -ItemType Directory -Force $prefix | Out-Null
    $dest = Join-Path $prefix "fixxl.exe"
    Move-Item -Force (Join-Path $tmp $asset) $dest
    Write-Host "fixxl: installed $dest"
    Write-Host "  next: add $prefix to your PATH"
    Write-Host "  run:  fixxl demo"
} finally {
    Remove-Item -Recurse -Force $tmp
}