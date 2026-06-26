Set-Location -LiteralPath $PSScriptRoot
$env:ADDR = if ($env:ADDR) { $env:ADDR } else { "0.0.0.0:8080" }
$env:TRUSTED_PROXIES = if ($env:TRUSTED_PROXIES) { $env:TRUSTED_PROXIES } else { "cloudflare" }
$env:POW_DIFFICULTY = if ($env:POW_DIFFICULTY) { $env:POW_DIFFICULTY } else { "12" }
.\bin\e2ee-chat-windows-amd64.exe
