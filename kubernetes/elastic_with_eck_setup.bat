@echo off

set OTELCOLLECTOR_HELMCHART_VALUES="helmcharts\otelcollector.values.yaml"
set ES_CONFIG="elastic\elasticsearch.yaml"
set KIBANA_CONFIG="elastic\kibana.yaml"
set APM_SERVER_CONFIG="elastic\apmserver.yaml"

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

helm install elastic-operator elastic/eck-operator -n observability

@REM set ES password to 'elastic', only for testing purposes
kubectl create secret generic -n observability elasticsearch-es-elastic-user --from-literal=elastic=elastic

kubectl apply -f %ES_CONFIG%
kubectl apply -f %KIBANA_CONFIG%
kubectl apply -f %APM_SERVER_CONFIG%

echo Port forwarding Kibana to local port 5601
set MAX_RETRIES=10
set RETRY_COUNT=0
:RETRY_PORT_FORWARD
kubectl port-forward service/kibana-kibana 5601:5601 -n observability
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
