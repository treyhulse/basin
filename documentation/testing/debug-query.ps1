# Test query parameter filtering with debug output

Write-Host "Testing query parameter filtering..." -ForegroundColor Green

# Step 1: Login and get token
Write-Host "Step 1: Login..." -ForegroundColor Yellow

$loginBody = @{
    email = "admin@example.com"
    password = "password"
} | ConvertTo-Json

try {
    $loginResponse = Invoke-RestMethod -Uri "http://localhost:8080/auth/login" -Method POST -Body $loginBody -ContentType "application/json"
    $token = $loginResponse.token
    Write-Host "Token obtained: $($token.Substring(0, 20))..." -ForegroundColor Cyan
    
    # Step 2: Query fields with collection_id filter
    Write-Host "Step 2: Querying fields with collection_id filter..." -ForegroundColor Yellow
    
    $headers = @{
        "Authorization" = "Bearer $token"
    }
    
    $url = "http://localhost:8080/items/fields?collection_id=5c059eeb-cd5d-4fb7-a35a-c9c485b2e024"
    Write-Host "URL: $url" -ForegroundColor Cyan
    
    $fieldResponse = Invoke-RestMethod -Uri $url -Method GET -Headers $headers
    
    Write-Host "Response received:" -ForegroundColor Green
    Write-Host "Count: $($fieldResponse.meta.count)" -ForegroundColor White
    Write-Host "Data:" -ForegroundColor White
    $fieldResponse.data | ForEach-Object {
        Write-Host "  - Field: $($_.name), Collection: $($_.collection_id)" -ForegroundColor Gray
    }
    
} catch {
    Write-Host "Error occurred:" -ForegroundColor Red
    Write-Host $_.Exception.Message -ForegroundColor Red
    
    if ($_.Exception.Response) {
        $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
        $responseBody = $reader.ReadToEnd()
        Write-Host "Response body: $responseBody" -ForegroundColor Red
    }
}

Write-Host "Test completed!" -ForegroundColor Green