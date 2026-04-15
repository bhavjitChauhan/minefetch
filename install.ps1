# Minefetch installation script for Windows.

$ErrorActionPreference = "Stop"

$Url = "https://github.com/bhavjitChauhan/minefetch/releases/latest/download/minefetch_windows_amd64.exe"
$Install = "$HOME\AppData\Local\Minefetch"
$Exe = "$Install\minefetch.exe"

Set-PSDebug -Trace 1

New-Item -Path $Install -ItemType Directory -Force | Out-Null
Invoke-WebRequest -Uri $Url -OutFile $Exe

if (-not ($env:Path -split ';' -contains $Install)) {
    [Environment]::SetEnvironmentVariable('Path', "$env:Path;$Install", [EnvironmentVariableTarget]::User)
    $env:Path += ";$Install"
}

Set-PSDebug -Trace 0

if (Get-Command minefetch -ErrorAction SilentlyContinue) {
    Write-Host "Successfully installed Minefetch!" -ForegroundColor Green
    Write-Host "You can run it using " -NoNewline
    Write-Host " minefetch " -ForegroundColor Black -BackgroundColor White -NoNewline
    Write-Host " in your terminal."
} else {
    Write-Host "Something went wrong." -ForegroundColor Red
    exit 1
}
