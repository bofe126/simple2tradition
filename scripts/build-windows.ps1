# build-windows.ps1
# Usage (from project root):
#   powershell -ExecutionPolicy Bypass -File scripts\build-windows.ps1
#
# Produces: build\bin\simple2tradition-amd64-installer.exe

$ErrorActionPreference = "Stop"
$ProjectRoot = Split-Path -Parent $PSScriptRoot
Set-Location $ProjectRoot

function Step($msg) {
    Write-Host ""
    Write-Host "==> $msg" -ForegroundColor Cyan
}

# Sanity checks
if (-not (Get-Command wails -ErrorAction SilentlyContinue)) {
    Write-Error "wails CLI not found. Install with: go install github.com/wailsapp/wails/v2/cmd/wails@latest"
    exit 1
}
if (-not (Get-Command makensis -ErrorAction SilentlyContinue)) {
    Write-Error "makensis (NSIS) not found. Install NSIS and add its Bin dir to PATH."
    exit 1
}

Step "Cleaning build\bin"
if (Test-Path "build\bin") {
    Remove-Item -Recurse -Force "build\bin"
}

Step "Building frontend (npm install + vite build)"
Push-Location "frontend"
try {
    if (-not (Test-Path "node_modules")) {
        npm install
        if ($LASTEXITCODE -ne 0) { throw "npm install failed" }
    }
    npm run build
    if ($LASTEXITCODE -ne 0) { throw "npm run build failed" }
} finally {
    Pop-Location
}

Step "Building Windows installer (wails build -nsis -platform windows/amd64)"
wails build -nsis -platform windows/amd64 -clean
if ($LASTEXITCODE -ne 0) { throw "wails build failed" }

Step "Done"
$Installer = Join-Path $ProjectRoot "build\bin\simple2tradition-amd64-installer.exe"
if (Test-Path $Installer) {
    $Size = [math]::Round((Get-Item $Installer).Length / 1MB, 2)
    Write-Host "Installer: $Installer ($Size MB)" -ForegroundColor Green
} else {
    Write-Warning "Installer not found in expected location. Check build\bin."
    exit 1
}
