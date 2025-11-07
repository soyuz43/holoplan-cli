$ErrorActionPreference = 'Stop'

$projectPath = Split-Path -Parent $MyInvocation.MyCommand.Path
$binaryName = "holoplan.exe"
$targetBinDir = "$env:USERPROFILE\bin"
$targetBinary = Join-Path $targetBinDir $binaryName

# Ensure target directory exists
if (-not (Test-Path $targetBinDir)) {
    Write-Host "Creating bin directory at $targetBinDir"
    New-Item -ItemType Directory -Path $targetBinDir | Out-Null
}

# Delete old binary if it exists
if (Test-Path $targetBinary) {
    Write-Host "🧹 Removing old $binaryName from $targetBinDir"
    Remove-Item $targetBinary -Force
}

# Build from src/
Write-Host "Building $binaryName from: $projectPath\src"
go build -o $targetBinary "$projectPath\src"

if ($LASTEXITCODE -eq 0) {
    Write-Host "$binaryName installed to $targetBinDir"
} else {
    Write-Error "Build failed. Check compilation errors."
    exit 1
}
