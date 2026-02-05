HTTP Test 7: Error Case (key vacía)
powershellInvoke-WebRequest -Method POST -Uri "http://localhost:8080/kv" `
  -ContentType "application/json" `
  -Body '{"operation":"GET","key":""}' | Select-Object -ExpandProperty Content
Output esperado:
json{"success":false,"error":"key cannot be empty"}

🔌 FASE 6: Probar TCP Server
Método 1: Con PowerShell
powershell# Crear cliente TCP
$client = New-Object System.Net.Sockets.TcpClient("localhost", 9090)
$stream = $client.GetStream()
$writer = New-Object System.IO.StreamWriter($stream)
$reader = New-Object System.IO.StreamReader($stream)

# Enviar comando SET
$writer.WriteLine('{"operation":"SET","key":"tcp-test","value":"Hello TCP"}')
$writer.Flush()
$response = $reader.ReadLine()
Write-Host "Response: $response"

# Enviar comando GET
$writer.WriteLine('{"operation":"GET","key":"tcp-test"}')
$writer.Flush()
$response = $reader.ReadLine()
Write-Host "Response: $response"

# Cerrar
$client.Close()
Método 2: Con Netcat (si lo tienes instalado)
powershell# Descargar netcat de: https://eternallybored.org/misc/netcat/
# Luego:
echo '{"operation":"GET","key":"tcp-test"}' | nc localhost 9090

📡 FASE 7: Probar UDP Server
Con PowerShell
powershell# Crear cliente UDP
$client = New-Object System.Net.Sockets.UdpClient
$client.Connect("localhost", 9091)

# Enviar comando
$message = '{"operation":"SET","key":"udp-test","value":"Hello UDP"}'
$bytes = [System.Text.Encoding]::ASCII.GetBytes($message)
$client.Send($bytes, $bytes.Length)

# Recibir respuesta
$endpoint = New-Object System.Net.IPEndPoint([System.Net.IPAddress]::Any, 0)
$receivedBytes = $client.Receive([ref]$endpoint)
$response = [System.Text.Encoding]::ASCII.GetString($receivedBytes)
Write-Host "Response: $response"

# Cerrar
$client.Close()

🔄 FASE 8: Test de Integración (CRÍTICO)
Este prueba que los 3 protocolos comparten el mismo store.
Script Completo de Integración
Guarda esto como test_integration.ps1:
powershellWrite-Host "`n========================================" -ForegroundColor Cyan
Write-Host "  INTEGRATION TEST - Cross-Protocol" -ForegroundColor Cyan
Write-Host "========================================`n" -ForegroundColor Cyan

# 1. Escribir via HTTP
Write-Host "[1] Writing via HTTP..." -ForegroundColor Yellow
$result = Invoke-WebRequest -Method POST -Uri "http://localhost:8080/kv" `
  -ContentType "application/json" `
  -Body '{"operation":"SET","key":"integration","value":"Cross-Protocol Works!"}' | 
  Select-Object -ExpandProperty Content
Write-Host "    $result" -ForegroundColor Green

# 2. Leer via TCP
Write-Host "[2] Reading via TCP..." -ForegroundColor Yellow
$client = New-Object System.Net.Sockets.TcpClient("localhost", 9090)
$stream = $client.GetStream()
$writer = New-Object System.IO.StreamWriter($stream)
$reader = New-Object System.IO.StreamReader($stream)
$writer.WriteLine('{"operation":"GET","key":"integration"}')
$writer.Flush()
$response = $reader.ReadLine()
Write-Host "    $response" -ForegroundColor Green
$client.Close()

# 3. Verificar SIZE via HTTP
Write-Host "[3] Checking SIZE via HTTP..." -ForegroundColor Yellow
$result = Invoke-WebRequest -Method POST -Uri "http://localhost:8080/kv" `
  -ContentType "application/json" `
  -Body '{"operation":"SIZE"}' | 
  Select-Object -ExpandProperty Content
Write-Host "    $result" -ForegroundColor Green

# 4. Leer via UDP
Write-Host "[4] Reading via UDP..." -ForegroundColor Yellow
$udpClient = New-Object System.Net.Sockets.UdpClient
$udpClient.Connect("localhost", 9091)
$message = '{"operation":"GET","key":"integration"}'
$bytes = [System.Text.Encoding]::ASCII.GetBytes($message)
$udpClient.Send($bytes, $bytes.Length)
$endpoint = New-Object System.Net.IPEndPoint([System.Net.IPAddress]::Any, 0)
$receivedBytes = $udpClient.Receive([ref]$endpoint)
$response = [System.Text.Encoding]::ASCII.GetString($receivedBytes)
Write-Host "    $response" -ForegroundColor Green
$udpClient.Close()

Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "  SUCCESS! All protocols share data" -ForegroundColor Green
Write-Host "========================================`n" -ForegroundColor Cyan
Ejecutar:
powershell.\test_integration.ps1

📝 FASE 9: Script de Testing Completo
Guarda esto como test_all.ps1:
powershell#Requires -Version 5.0

$ErrorActionPreference = "Stop"

Write-Host "`n" -NoNewline
Write-Host "=========================================" -ForegroundColor Cyan
Write-Host "   COMPLETE TEST SUITE" -ForegroundColor Cyan
Write-Host "=========================================" -ForegroundColor Cyan
Write-Host "`n"

$passed = 0
$failed = 0

function Test-Step {
    param($Name, $ScriptBlock)
    
    Write-Host "Testing: $Name..." -ForegroundColor Yellow -NoNewline
    
    try {
        & $ScriptBlock
        Write-Host " PASS" -ForegroundColor Green
        $script:passed++
    } catch {
        Write-Host " FAIL" -ForegroundColor Red
        Write-Host "  Error: $_" -ForegroundColor Red
        $script:failed++
    }
}

# Test 1: Unit Tests
Test-Step "Unit Tests" {
    $result = go test ./... 2>&1
    if ($LASTEXITCODE -ne 0) { throw "Unit tests failed" }
}

# Test 2: Race Detector
Test-Step "Race Detector" {
    $result = go test -race ./internal/store 2>&1
    if ($LASTEXITCODE -ne 0) { throw "Race detector found issues" }
}

# Test 3: Benchmarks
Test-Step "Benchmarks" {
    $result = go test -bench=. -benchtime=1s ./internal/store 2>&1
    if ($LASTEXITCODE -ne 0) { throw "Benchmarks failed" }
}

# Test 4: Build
Test-Step "Build" {
    $result = go build -o bin/kvstore-test.exe ./cmd/kvstore 2>&1
    if ($LASTEXITCODE -ne 0) { throw "Build failed" }
    if (!(Test-Path bin/kvstore-test.exe)) { throw "Binary not created" }
}

# Test 5: Start Server
Write-Host "`nStarting server for integration tests..." -ForegroundColor Yellow
$serverProcess = Start-Process -FilePath ".\bin\kvstore-test.exe" -PassThru -WindowStyle Hidden
Start-Sleep -Seconds 3

if (!$serverProcess.HasExited) {
    Write-Host "Server started (PID: $($serverProcess.Id))" -ForegroundColor Green
    
    # Test 6: HTTP Health
    Test-Step "HTTP Health Check" {
        $result = curl http://localhost:8080/health 2>&1
        if ($result -notmatch "OK") { throw "Health check failed" }
    }
    
    # Test 7: HTTP SET
    Test-Step "HTTP SET" {
        $result = Invoke-WebRequest -Method POST -Uri "http://localhost:8080/kv" `
          -ContentType "application/json" `
          -Body '{"operation":"SET","key":"test","value":"works"}' |
          Select-Object -ExpandProperty Content
        if ($result -notmatch '"success":true') { throw "SET failed" }
    }
    
    # Test 8: HTTP GET
    Test-Step "HTTP GET" {
        $result = Invoke-WebRequest -Method POST -Uri "http://localhost:8080/kv" `
          -ContentType "application/json" `
          -Body '{"operation":"GET","key":"test"}' |
          Select-Object -ExpandProperty Content
        if ($result -notmatch '"value":"works"') { throw "GET failed" }
    }
    
    # Test 9: TCP Connection
    Test-Step "TCP Connection" {
        $client = New-Object System.Net.Sockets.TcpClient("localhost", 9090)
        $stream = $client.GetStream()
        $writer = New-Object System.IO.StreamWriter($stream)
        $reader = New-Object System.IO.StreamReader($stream)
        $writer.WriteLine('{"operation":"SIZE"}')
        $writer.Flush()
        $response = $reader.ReadLine()
        $client.Close()
        if ($response -notmatch '"success":true') { throw "TCP failed" }
    }
    
    # Test 10: UDP Connection
    Test-Step "UDP Connection" {
        $client = New-Object System.Net.Sockets.UdpClient
        $client.Connect("localhost", 9091)
        $message = '{"operation":"SIZE"}'
        $bytes = [System.Text.Encoding]::ASCII.GetBytes($message)
        $client.Send($bytes, $bytes.Length)
        $endpoint = New-Object System.Net.IPEndPoint([System.Net.IPAddress]::Any, 0)
        $receivedBytes = $client.Receive([ref]$endpoint)
        $response = [System.Text.Encoding]::ASCII.GetString($receivedBytes)
        $client.Close()
        if ($response -notmatch '"success":true') { throw "UDP failed" }
    }
    
    # Stop server
    Stop-Process -Id $serverProcess.Id -Force
    Write-Host "Server stopped" -ForegroundColor Yellow
} else {
    Write-Host "Server failed to start" -ForegroundColor Red
    $failed++
}

# Cleanup
Remove-Item bin/kvstore-test.exe -ErrorAction SilentlyContinue

Write-Host "`n" -NoNewline
Write-Host "=========================================" -ForegroundColor Cyan
Write-Host "   RESULTS" -ForegroundColor Cyan
Write-Host "=========================================" -ForegroundColor Cyan
Write-Host "Passed: $passed" -ForegroundColor Green
Write-Host "Failed: $failed" -ForegroundColor $(if ($failed -eq 0) { "Green" } else { "Red" })
Write-Host "=========================================" -ForegroundColor Cyan
Write-Host "`n"

if ($failed -gt 0) {
    exit 1
}
Ejecutar:
powershell.\test_all.ps1
