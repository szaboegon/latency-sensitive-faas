$cleanupScript = {
    Write-Host "`nStopping all kubectl port-forwards..."
    Get-CimInstance Win32_Process -Filter "Name = 'kubectl.exe'" |
        Where-Object { $_.CommandLine -like '*port-forward*' } |
        ForEach-Object { 
            Write-Host "Stopping PID $($_.ProcessId): $($_.CommandLine)"
            Stop-Process -Id $_.ProcessId -Force
        }
    Write-Host "All port-forwards stopped."
    exit
}

$null = Register-EngineEvent -SourceIdentifier Console_CancelKeyPress -Action $cleanupScript

Start-Process -NoNewWindow -FilePath "kubectl" -ArgumentList "port-forward", "svc/kibana-kb-http", "5601:5601", "-n", "observability"
Start-Process -NoNewWindow -FilePath "kubectl" -ArgumentList "port-forward", "svc/lsf-configurator", "8080:80", "-n", "configurator"
Start-Process -NoNewWindow -FilePath "kubectl" -ArgumentList "port-forward", "svc/sqlite-web", "8085:8080", "-n", "configurator"
Start-Process -NoNewWindow -FilePath "kubectl" -ArgumentList "port-forward", "svc/tekton-dashboard", "9097:9097", "-n", "tekton-pipelines"

Write-Host "Port-forwards started. Press Ctrl+C to stop them."

# Keep script alive until Ctrl+C
while ($true) {
    Start-Sleep -Seconds 1
}