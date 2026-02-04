Write-Host "=== http.go ===" -ForegroundColor Cyan
Get-Content internal\server\http.go | Select-String -Pattern "listening" -Context 0,0

Write-Host "`n=== tcp.go ===" -ForegroundColor Yellow
Get-Content internal\server\tcp.go | Select-String -Pattern "listening" -Context 0,0

Write-Host "`n=== udp.go ===" -ForegroundColor Green
Get-Content internal\server\udp.go | Select-String -Pattern "listening" -Context 0,0