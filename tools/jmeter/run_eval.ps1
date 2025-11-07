Remove-Item -Recurse -Force "dashboard-report" -ErrorAction SilentlyContinue
Remove-Item -Force "jmeter.log" -ErrorAction SilentlyContinue
Remove-Item -Force "results.csv" -ErrorAction SilentlyContinue

jmeter -n -t "eval.jmx" -l "results.csv" -e -o "dashboard-report"