<#
.SYNOPSIS
    Run Kotoha evaluation tests and optionally post results to a running server.

.DESCRIPTION
    Executes search and agent eval tests, extracts MRR/NDCG metrics from test output,
    and can POST results to the /api/v1/eval/runs endpoint.

.PARAMETER ApiUrl
    Base URL of the running Kotoha server (e.g. http://localhost:8080).
    If omitted, results are only printed to stdout.

.PARAMETER JWTToken
    JWT access token for authentication. Required only if ApiUrl is specified.

.EXAMPLE
    .\scripts\run_eval.ps1
    # Run eval locally, print metrics

.EXAMPLE
    .\scripts\run_eval.ps1 -ApiUrl "http://localhost:8080" -JWTToken "eyJ..."
    # Run eval and POST results to server
#>

param(
    [string]$ApiUrl = "",
    [string]$JWTToken = ""
)

$ErrorActionPreference = "Stop"

Write-Host "[eval] Running evaluation tests..." -ForegroundColor Cyan

# Run eval tests
go test -v -run "Eval" ./internal/search/ ./internal/agent/ 2>&1 | Tee-Object -Variable testOutput

# Parse EVAL_RESULT line from test output
$evalLine = $testOutput | Select-String "EVAL_RESULT:" | Select-Object -Last 1

if (-not $evalLine) {
    Write-Host "[eval] No EVAL_RESULT found in test output" -ForegroundColor Yellow
    exit 0
}

# Extract JSON from line
$jsonStart = $evalLine.Line.IndexOf("{")
if ($jsonStart -lt 0) {
    Write-Host "[eval] Could not parse EVAL_RESULT JSON" -ForegroundColor Red
    exit 1
}

$metricsJson = $evalLine.Line.Substring($jsonStart)
$metrics = $metricsJson | ConvertFrom-Json

Write-Host ""
Write-Host "=== Evaluation Results ===" -ForegroundColor Green
Write-Host "MRR:       $($metrics.mrr)" -ForegroundColor White
Write-Host "NDCG@5:    $($metrics.ndcg_at_5)" -ForegroundColor White
Write-Host "NDCG@10:   $($metrics.ndcg_at_10)" -ForegroundColor White
Write-Host "Avg Recall: $($metrics.avg_recall)" -ForegroundColor White
Write-Host "Accuracy:  $($metrics.accuracy)% ($($metrics.passed)/$($metrics.total) passed)" -ForegroundColor White

# Optionally POST to API
if ($ApiUrl -ne "") {
    $body = @{
        category    = "search"
        total_tests = $metrics.total
        passed      = $metrics.passed
        failed      = $metrics.failed
        accuracy    = $metrics.accuracy
        metrics     = $metricsJson
    } | ConvertTo-Json

    $headers = @{
        "Content-Type"  = "application/json"
        "Authorization" = "Bearer $JWTToken"
    }

    Write-Host ""
    Write-Host "[eval] Posting results to $ApiUrl/api/v1/eval/runs..." -ForegroundColor Cyan

    try {
        $response = Invoke-RestMethod -Uri "$ApiUrl/api/v1/eval/runs" -Method Post -Body $body -Headers $headers
        Write-Host "[eval] Eval run saved (id=$($response.id))" -ForegroundColor Green
    }
    catch {
        Write-Host "[eval] Failed to POST results: $_" -ForegroundColor Red
    }
}
