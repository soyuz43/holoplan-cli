# install.ps1
$ErrorActionPreference = 'Stop'

$projectPath = Split-Path -Parent $MyInvocation.MyCommand.Path
$binaryName = "holoplan"
$targetBinDir = "$env:USERPROFILE\bin"

# Make sure the target directory exists
if (!(Test-Path -Path $targetBinDir)) {
    New-Item -ItemType Directory -Path $targetBinDir | Out-Null
}

# Build from the src folder where main.go lives
Write-Host "🚧 Building from: $projectPath\src"
go build -o "$targetBinDir\$binaryName.exe" "$projectPath\src"

if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ $binaryName installed to $targetBinDir"
} else {
    Write-Host "❌ Build failed"
}
