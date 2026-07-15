param(
    [string]$InstallDir = "$env:USERPROFILE\.local\bin"
)

$ErrorActionPreference = "Stop"

$RepoDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$WorkDir = Split-Path -Parent $RepoDir

Push-Location $RepoDir
try {
    go build -ldflags "-X 'main.defaultWorkDir=$WorkDir'" -o delbyapps.exe .
    New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
    Copy-Item -Force .\delbyapps.exe (Join-Path $InstallDir "delbyapps.exe")
} finally {
    Pop-Location
}

$UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
if (($UserPath -split ';') -notcontains $InstallDir) {
    [Environment]::SetEnvironmentVariable("Path", (($UserPath, $InstallDir | Where-Object { $_ }) -join ';'), "User")
    Write-Host "Added $InstallDir to your user PATH. Open a new terminal before running delbyapps."
}
