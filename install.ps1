Set-PSDebug -Trace 1

$version = (Invoke-RestMethod -Uri 'https://api.github.com/repos/bhavjitChauhan/minefetch/releases/latest').tag_name.TrimStart('v')
if (-not $version) {
    Write-Host "Failed to get latest version" -ForegroundColor Red
    exit 1
}
$url = 'https://github.com/bhavjitChauhan/minefetch/releases/download/v$version/minefech_${version}_windows_amd64.exe'
$dir = "$HOME\AppData\Local\Minefetch"
$exe = "$dir\minefetch.exe"

if (-not (Test-Path -Path $dir)) {
    New-Item -ItemType Directory -Path $dir | Out-Null
}

Invoke-WebRequest -Uri $url -OutFile $exe

if (-not ($env:Path -split ';' -contains $dir)) {
    [Environment]::SetEnvironmentVariable('Path', "$env:Path;$dir", [EnvironmentVariableTarget]::User)
    $env:Path += ";$dir"
}

Set-PSDebug -Trace 0

if (Get-Command minefetch -ErrorAction SilentlyContinue) {
    Write-Host "Successfully installed Minefetch!" -ForegroundColor Green
    Write-Host "You can run it by executing " -NoNewline
    Write-Host "minefetch" -ForegroundColor Blue -NoNewline
    Write-Host " in your terminal."
} else {
    Write-Host "Something went wrong." -ForegroundColor Red
    exit 1
}
