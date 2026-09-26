#!/usr/bin/env pwsh
#Requires -Version 7

$ErrorActionPreference = "Stop"

# --- Detect OS and architecture ---
if ($IsWindows) {
    $goos   = "windows"
    $ext    = ".exe"
} elseif ($IsLinux) {
    $goos   = "linux"
    $ext    = ""
} elseif ($IsMacOS) {
    $goos   = "darwin"
    $ext    = ""
} else {
    Write-Error "Unsupported OS"
    exit 1
}

$arch = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture
$goarch = switch ($arch) {
    "X64"   { "amd64" }
    "Arm64" { "arm64" }
    default {
        Write-Error "Unsupported architecture: $arch"
        exit 1
    }
}

$binaryName  = "terraform-provider-tcm$ext"
$installPath = Join-Path $HOME ".terraform.d/plugins/symplr/tcm/0.1.0/${goos}_${goarch}"
$outputPath  = Join-Path $installPath $binaryName

Write-Host "Building for $goos/$goarch -> $outputPath"

# --- Build ---
$scriptDir = Split-Path -Parent $PSScriptRoot
Push-Location $scriptDir
try {
    $env:GOOS   = $goos
    $env:GOARCH = $goarch

    & go build -o $outputPath .
    if ($LASTEXITCODE -ne 0) {
        Write-Error "go build failed"
        exit $LASTEXITCODE
    }
} finally {
    Remove-Item Env:GOOS   -ErrorAction SilentlyContinue
    Remove-Item Env:GOARCH -ErrorAction SilentlyContinue
    Pop-Location
}

Write-Host "Done. Provider installed to: $outputPath"
Write-Host ""
Write-Host "Ensure your .terraformrc (or TF_CLI_CONFIG_FILE) contains:"
Write-Host ""
Write-Host '  provider_installation {'
Write-Host '    dev_overrides {'
Write-Host "      `"symplr/tcm`" = `"$installPath`""
Write-Host '    }'
Write-Host '    direct {}'
Write-Host '  }'
