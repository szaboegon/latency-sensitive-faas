<#
.SYNOPSIS
Runs the JMeter test, extracts reconfiguration events from the Kubernetes controller logs,
and generates the combined RPS/Latency plot using the Python script.

.DESCRIPTION
1. Cleans up previous results.
2. Executes the JMeter test defined in eval.jmx.
3. Finds the single running lsf-confiugrator pod.
4. Extracts structured JSON events from the pod logs for a configurable time window.
5. Converts the events to CSV format.
6. Runs the Python plotting script (plot_jtl_metrics.py) with both results files.

.NOTES
Requires:
- kubectl (installed and configured)
- jq (for reliable JSON parsing of logs)
- JMETER_HOME environment variable to be set
#>

# --- Configuration ---
$LogTimeWindowMinutes = 60
$ControllerDeploymentName = "lsf-configurator"
$Namespace = "configurator"
$JtlFile = "results.csv"
$EventsFile = "reconfig_events.csv"
$JmeterTestFile = "eval.jmx"
$PythonPlotter = "plot_rps_latency.py" 
# ---------------------


Write-Host "--- 1. Cleanup Previous Results ---"
Remove-Item -Recurse -Force "dashboard-report" -ErrorAction SilentlyContinue
Remove-Item -Force "jmeter.log" -ErrorAction SilentlyContinue
Remove-Item -Force $JtlFile -ErrorAction SilentlyContinue
Remove-Item -Force $EventsFile -ErrorAction SilentlyContinue

Write-Host "--- 2. Running JMeter Test: $JmeterTestFile ---"
& "$env:JMETER_HOME\bin\jmeter.bat" -n -t $JmeterTestFile -o "dashboard-report"

Write-Host "--- 3. Finding Controller Pod for Deployment: $ControllerDeploymentName ---"

$ControllerPodName = kubectl get pods -n $Namespace -l app=$ControllerDeploymentName -o jsonpath='{.items[0].metadata.name}'

if (-not $ControllerPodName) {
    Write-Error "Could not find a running pod for deployment label 'app=$ControllerDeploymentName'. Check your deployment labels."
    exit 1
}

Write-Host "Found Controller Pod: $ControllerPodName"

Write-Host "--- 4. Extracting Structured Events (Last $LogTimeWindowMinutes min) ---"

$SinceDuration = "$LogTimeWindowMinutes" + "m"

# The pipeline below uses kubectl to fetch logs, filters for the JSON events,
# strips the log prefix, and uses jq to format the JSON objects into CSV lines.
try {
    $LogData = kubectl logs $ControllerPodName -n $Namespace --since=$SinceDuration 2>$null | 
                Select-String -Pattern "\[RECONFIG_EVENT\]" | 
                ForEach-Object { 
                    # Use -split to isolate the JSON part (index -1 after the marker)
                    ($_.ToString() -split '\[RECONFIG_EVENT\]', 2)[-1].Trim()
                }
    
    # Write the CSV header first
    "event_time,duration_ms,app_id" | Out-File $EventsFile -Encoding UTF8 -Force
    
    if ($LogData.Length -eq 0) {
        Write-Warning "No [RECONFIG_EVENT] logs found in the last $LogTimeWindowMinutes minutes."
    } else {
        $CsvContent = $LogData | jq -r '[.event_time, .duration_ms, .app_id] | @csv'
        $CsvContent | Out-File $EventsFile -Append -Encoding UTF8
        Write-Host "Successfully extracted $($LogData.Length) events to $EventsFile"
    }
} catch {
    Write-Error "An error occurred during log extraction. Check if 'jq' is installed and in your PATH."
    Write-Error $_.Exception.Message
    exit 1
}

Write-Host "--- 5. Running Python Plotting Script: $PythonPlotter ---"
python $PythonPlotter $JtlFile $EventsFile

Write-Host "--- Workflow Complete ---"