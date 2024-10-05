@echo off

set OTELCOLLECTOR_HELMCHART_VALUES="helmcharts\otelcollector.values.yaml"
set ES_HELMCHART_VALUES="helmcharts\elasticsearch.values.yaml"
set KIBANA_HELMCHART_VALUES="helmcharts\kibana.values.yaml"
set APM_SERVER_HELMCHART_VALUES="helmcharts\apm-server.values.yaml"
set METRICBEAT_HELMCHART_VALUES="helmcharts\metricbeat.values.yaml"

echo Creating observability namespace...
kubectl create namespace observability
if errorlevel 1 (
    echo Error creating namespace observability
    exit /b 1
)

@REM helm install otel-collector open-telemetry/opentelemetry-collector -n observability --values %OTELCOLLECTOR_HELMCHART_VALUES%
@REM if errorlevel 1 (
@REM     echo Error installing otel collector from helm chart
@REM     exit /b 1
@REM )

echo Installing elasticsearch from helm chart...
helm install elasticsearch elastic/elasticsearch -n observability --values %ES_HELMCHART_VALUES%
if errorlevel 1 (
    echo Error installing elasticsearch
    exit /b 1
)

echo Installing kibana from helm chart...
helm install kibana elastic/kibana -n observability --values %KIBANA_HELMCHART_VALUES%
if errorlevel 1 (
    echo Error installing kibana
    exit /b 1
)

echo Installing apm-server from helm chart...
helm install apm-server elastic/apm-server -n observability --values %APM_SERVER_HELMCHART_VALUES%
if errorlevel 1 (
    echo Error installing apm-server
    exit /b 1
)

echo Installing metricbeat from helm chart...
helm install metricbeat elastic/metricbeat -n observability --values %METRICBEAT_HELMCHART_VALUES%
if errorlevel 1 (
    echo Error installing metricbeat
    exit /b 1
)

echo Port forwarding Kibana to local port 5601
set MAX_RETRIES=10
set RETRY_COUNT=0
:RETRY_PORT_FORWARD
kubectl port-forward service/kibana 5601:5601 -n observability
if errorlevel 1 (
    set /a RETRY_COUNT+=1
    if %RETRY_COUNT% lss %MAX_RETRIES% (
        echo Retry %RETRY_COUNT% of %MAX_RETRIES% failed. Retrying...
        timeout /t 10 >nul
        goto RETRY_PORT_FORWARD
    ) else (
        echo Failed to forward port after %MAX_RETRIES% retries
        exit /b 1
    )
)
