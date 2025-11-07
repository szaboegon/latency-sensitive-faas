Remove-Item -Recurse -Force "dashboard-report" -ErrorAction SilentlyContinue
Remove-Item -Force "jmeter.log" -ErrorAction SilentlyContinue
Remove-Item -Force "results.jtl" -ErrorAction SilentlyContinue
Remove-Item -Force "response_times.png" -ErrorAction SilentlyContinue
Remove-Item -Force "transactions.png" -ErrorAction SilentlyContinue

& "$env:JMETER_HOME\bin\jmeter.bat" -n -t "eval.jmx" -l "results.jtl" -e -o "dashboard-report"

& "$env:JMETER_HOME\bin\JMeterPluginsCMD.bat" `
 --generate-png "response_times.png" `
 --input-jtl "results.jtl" `
 --plugin-type "ResponseTimesOverTime" `
 --width 1920 --height 1080

& "$env:JMETER_HOME\bin\JMeterPluginsCMD.bat" `
 --generate-png "transactions.png" `
 --input-jtl "results.jtl" `
 --plugin-type "TransactionsPerSecond" `
 --width 1920 --height 1080 