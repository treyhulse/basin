# Test Field Creation API
Write-Host "🧪 Testing Field Creation for blog_posts Collection..." -ForegroundColor Green

# Step 1: Login and get token
Write-Host "📝 Step 1: Login..." -ForegroundColor Yellow

$loginBody = @{
    email = "admin@example.com"
    password = "password"
} | ConvertTo-Json

try {
    $loginResponse = Invoke-RestMethod -Uri "http://localhost:8080/auth/login" -Method POST -Body $loginBody -ContentType "application/json"
    $token = $loginResponse.token
    Write-Host "Token: $($token.Substring(0, 20))..." -ForegroundColor Cyan
    
    # Step 2: Create a field
    Write-Host "📝 Step 2: Creating 'title' field..." -ForegroundColor Yellow
    
    $fieldBody = @{
        collection_id = "5c059eeb-cd5d-4fb7-a35a-c9c485b2e024"
        name = "title"
        display_name = "Title"
        type = "text"
        is_required = $true
        is_unique = $false
        sort_order = 1
    } | ConvertTo-Json
    
    $headers = @{
        "Authorization" = "Bearer $token"
        "Content-Type" = "application/json"
    }
    
    $fieldResponse = Invoke-RestMethod -Uri "http://localhost:8080/items/fields" -Method POST -Body $fieldBody -Headers $headers
    
    Write-Host "Field creation response:" -ForegroundColor Green
    $fieldResponse | ConvertTo-Json -Depth 10 | Write-Host
    
}
catch {
    Write-Host "❌ Error occurred:" -ForegroundColor Red
    Write-Host $_.Exception.Message -ForegroundColor Red
    
    if ($_.Exception.Response) {
        $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
        $responseBody = $reader.ReadToEnd()
        Write-Host "Response body: $responseBody" -ForegroundColor Red
    }
}

Write-Host "✅ Field creation test completed!" -ForegroundColor Green