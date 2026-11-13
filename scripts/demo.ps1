# Bifrost SmartRoute - Interactive Hackathon Demonstration Script (PowerShell)
param (
    [string]$GatewayUrl = "http://localhost:8080"
)

function Write-Header($text) {
    Write-Host ""
    Write-Host "══════════════════════════════════════════════════════════════════════" -ForegroundColor Magenta
    Write-Host "  $text" -ForegroundColor Cyan
    Write-Host "══════════════════════════════════════════════════════════════════════" -ForegroundColor Magenta
}

function Write-Step($step, $title) {
    Write-Host ""
    Write-Host "[$step] $title" -ForegroundColor Yellow
    Write-Host ("-" * 60) -ForegroundColor DarkGray
}

Clear-Host
Write-Host "╔══════════════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║               BIFROST SMARTROUTE - LIVE DEMO TOUR                    ║" -ForegroundColor Cyan
Write-Host "║       Intelligent AI Model Routing & Cost Optimization System        ║" -ForegroundColor Cyan
Write-Host "╚══════════════════════════════════════════════════════════════════════╝" -ForegroundColor Cyan

# 1. Health & Metrics Check
Write-Step "1/5" "Inspecting Gateway Health & Historical Analytics..."
try {
    $metrics = Invoke-RestMethod -Uri "$GatewayUrl/api/smartroute/metrics" -UseBasicParsing -TimeoutSec 5
    Write-Host "  • Total Requests Processed: $($metrics.total_requests)" -ForegroundColor Green
    Write-Host "  • Cumulative Expenditure:   `$$([math]::Round($metrics.total_expenditure, 4))" -ForegroundColor Green
    Write-Host "  • Unoptimized Baseline:     `$$([math]::Round($metrics.baseline_cost, 4))" -ForegroundColor Yellow
    Write-Host "  • Net Dollar Savings:       `$$([math]::Round($metrics.estimated_savings, 4)) ($([math]::Round($metrics.savings_percent, 1))% SAVINGS)" -ForegroundColor Magenta
    Write-Host "  • Gateway Average Latency:  $([math]::Round($metrics.avg_response_time_ms, 1)) ms" -ForegroundColor Green
} catch {
    Write-Host "Error connecting to Bifrost SmartRoute at $GatewayUrl: $_" -ForegroundColor Red
    Write-Host "Please start the server first with: .\bifrost-smartroute.exe -port 8080" -ForegroundColor Red
    exit 1
}

# 2. Simple Query Routing
Write-Step "2/5" "Routing Simple Informational Query..."
$simplePrompt = "What is the capital of Australia?"
Write-Host "  Prompt: `"$simplePrompt`"" -ForegroundColor Gray
$body1 = @{ prompt = $simplePrompt } | ConvertTo-Json
$res1 = Invoke-RestMethod -Uri "$GatewayUrl/api/smartroute/route" -Method Post -Body $body1 -ContentType "application/json" -UseBasicParsing

Write-Host "  → Category Detected:  $($res1.decision.task_category.ToUpper()) (Complexity: $($res1.decision.complexity_score)/10)" -ForegroundColor Green
Write-Host "  → Selected Model:     $($res1.decision.selected_model) on $($res1.decision.selected_provider)" -ForegroundColor Cyan
Write-Host "  → Cost vs Baseline:   `$$($res1.decision.estimated_cost) vs `$$($res1.decision.baseline_cost) ($([math]::Round($res1.decision.savings_percent, 1))% Savings!)" -ForegroundColor Magenta
Write-Host "  → Justification:      $($res1.decision.explanation)" -ForegroundColor DarkCyan

# 3. Complex Coding Routing
Write-Step "3/5" "Routing Complex Engineering / Coding Prompt..."
$codePrompt = "Implement an asynchronous connection pool in Go with mutex striping and backoff retry logic"
Write-Host "  Prompt: `"$codePrompt`"" -ForegroundColor Gray
$body2 = @{ prompt = $codePrompt } | ConvertTo-Json
$res2 = Invoke-RestMethod -Uri "$GatewayUrl/api/smartroute/route" -Method Post -Body $body2 -ContentType "application/json" -UseBasicParsing

Write-Host "  → Category Detected:  $($res2.decision.task_category.ToUpper()) (Complexity: $($res2.decision.complexity_score)/10)" -ForegroundColor Green
Write-Host "  → Selected Model:     $($res2.decision.selected_model) on $($res2.decision.selected_provider)" -ForegroundColor Cyan
Write-Host "  → Fallback Chain:     $($res2.decision.fallback_chain.model_name -join ' -> ')" -ForegroundColor Yellow
Write-Host "  → Justification:      $($res2.decision.explanation)" -ForegroundColor DarkCyan

# 4. Deep Reasoning / Math Routing
Write-Step "4/5" "Routing Complex Multi-Step Reasoning Puzzle..."
$reasonPrompt = "Analyze the paradox of thrift under liquidity trap conditions and derive the IS-LM mathematical equilibrium"
Write-Host "  Prompt: `"$reasonPrompt`"" -ForegroundColor Gray
$body3 = @{ prompt = $reasonPrompt } | ConvertTo-Json
$res3 = Invoke-RestMethod -Uri "$GatewayUrl/api/smartroute/route" -Method Post -Body $body3 -ContentType "application/json" -UseBasicParsing

Write-Host "  → Category Detected:  $($res3.decision.task_category.ToUpper()) (Complexity: $($res3.decision.complexity_score)/10)" -ForegroundColor Green
Write-Host "  → Selected Model:     $($res3.decision.selected_model) on $($res3.decision.selected_provider)" -ForegroundColor Cyan
Write-Host "  → Justification:      $($res3.decision.explanation)" -ForegroundColor DarkCyan

# 5. Circuit Breaker Outage & Self-Healing Failover
Write-Step "5/5" "Simulating Provider Outage & Self-Healing Failover..."
Write-Host "  Simulating HTTP 503 Service Unavailable on 'anthropic'..." -ForegroundColor DarkYellow
$failBody = @{ provider = "anthropic"; error_code = 503 } | ConvertTo-Json
$null = Invoke-RestMethod -Uri "$GatewayUrl/api/smartroute/failover/simulate" -Method Post -Body $failBody -ContentType "application/json" -UseBasicParsing

$providers = Invoke-RestMethod -Uri "$GatewayUrl/api/smartroute/providers" -UseBasicParsing
$anthropicState = $providers.circuit_statuses.anthropic.state
Write-Host "  • Anthropic Circuit Breaker State: $($anthropicState.ToUpper())" -ForegroundColor Yellow

Write-Host "  Submitting request with primary provider under stress..." -ForegroundColor Gray
$failoverRes = Invoke-RestMethod -Uri "$GatewayUrl/api/smartroute/route" -Method Post -Body $body2 -ContentType "application/json" -UseBasicParsing
Write-Host "  → Failover Status: $($failoverRes.fallback_status)" -ForegroundColor Green
Write-Host "  → Active Routing:  $($failoverRes.decision.selected_model) (Backup chain ready)" -ForegroundColor Cyan

Write-Header "DEMONSTRATION COMPLETE!"
Write-Host "To explore the full interactive dashboard, open:" -ForegroundColor White
Write-Host "👉 http://localhost:8080" -ForegroundColor Cyan
Write-Host ""
