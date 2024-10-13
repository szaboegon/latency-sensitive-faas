@echo off

set OTELCOLLECTOR_CONFIG="otel\otelcollector.yaml"
set ES_CONFIG="elastic\elasticsearch.yaml"
set KIBANA_CONFIG="elastic\kibana.yaml"
set APM_SERVER_CONFIG="elastic\apmserver.yaml"
set PYTHON_INSTRUMENTATION="otel\python_instrumentation.yaml"

echo Creating observability namespace...
kubectl create namespace observability
if errorlevel 1 (
    echo Error creating namespace observability
    exit /b 1
)

echo Installing prerequisite: Cert Manager
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.6.3/cert-manager.yaml
if errorlevel 1 (
    echo Error installing Cert Manager
    exit /b 1
)


if errorlevel 1 (
    echo Error creating otel operator
    exit /b 1
)

echo Installing Otel Operator...
set MAX_RETRIES=50
set RETRY_COUNT=0
:RETRY_OTEL_OPERATOR
kubectl apply -f https://github.com/open-telemetry/opentelemetry-operator/releases/latest/download/opentelemetry-operator.yaml
if errorlevel 1 (
    set /a RETRY_COUNT+=1
    if %RETRY_COUNT% lss %MAX_RETRIES% (
        echo Retry %RETRY_COUNT% of %MAX_RETRIES% failed. Retrying...
        timeout /t 10 >nul
        goto RETRY_OTEL_OPERATOR
    ) else (
        echo Failed to deploy Jaeger instance after %MAX_RETRIES% retries
        exit /b 1
    )
)

set MAX_RETRIES=50
set RETRY_COUNT=0
:RETRY_OTEL_COLLECTOR
kubectl apply -n observability -f %OTELCOLLECTOR_CONFIG%
if errorlevel 1 (
    set /a RETRY_COUNT+=1
    if %RETRY_COUNT% lss %MAX_RETRIES% (
        echo Retry %RETRY_COUNT% of %MAX_RETRIES% failed. Retrying...
        timeout /t 10 >nul
        goto RETRY_OTEL_COLLECTOR
    ) else (
        echo Failed to deploy Jaeger instance after %MAX_RETRIES% retries
        exit /b 1
    )
)

kubectl apply -f %PYTHON_INSTRUMENTATION%
if errorlevel 1 (
    echo Error creating python instrumentation resource
    exit /b 1
)

helm install elastic-operator elastic/eck-operator -n observability
if errorlevel 1 (
    echo Error creating elastic operator
    exit /b 1
)

@REM set ES password to 'elastic', only for testing purposes
kubectl create secret generic -n observability elasticsearch-es-elastic-user --from-literal=elastic=elastic

kubectl apply -f %ES_CONFIG%
kubectl apply -f %KIBANA_CONFIG%
kubectl apply -f %APM_SERVER_CONFIG%

echo Port forwarding Kibana to local port 5601
set MAX_RETRIES=50
set RETRY_COUNT=0
:RETRY_PORT_FORWARD
kubectl port-forward service/kibana-kb-http 5601:5601 -n observability
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
