@echo off

set OTELCOLLECTOR_CONFIG="otel\otelcollector.yaml"
set ES_CONFIG="elastic\elasticsearch.yaml"
set KIBANA_CONFIG="elastic\kibana.yaml"
set APM_SERVER_CONFIG="elastic\apmserver.yaml"
set PYTHON_INSTRUMENTATION="otel\python_instrumentation.yaml"
set LSF_CONFIGURATOR_NAMESPACE_YAML_PATH="..\lsf-configurator\deploy\configurator-namespace.yaml"
set UPLOADS_PVC_YAML_PATH="..\lsf-configurator\deploy\uploads-pvc.yaml"
set CREATE_REGISTRY_SECRET_PATH=".\create_registry_secret.bat"
set LSF_CONFIGURATOR_YAML_PATH="..\lsf-configurator\deploy\lsf-configurator.yaml"
set REDIS_NAMESPACE_YAML_PATH="redis\redis-namespace.yaml"
set REDIS_YAML_PATH="redis\redis-master-replica.yaml"
set TEKTON_PIPELINES_YAML_PATH="tekton\pipelines.yaml"
set TEKTON_DASHBOARD_YAML_PATH="tekton\dashboard.yaml"
set TEKTON_BUILD_PIPELINE_YAML_PATH="tekton\function-build-pipeline.yaml"
set TEKTON_BUILDPACKS_YAML_PATH="tekton\buildpacks-phases.yaml"

echo Updating hellm repositories...
helm repo update

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
        echo Failed to deploy Otel operator instance after %MAX_RETRIES% retries
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
        echo Failed to deploy Otel collector instance after %MAX_RETRIES% retries
        exit /b 1
    )
)

echo Installing Elastic Operator, ES and Kibana...
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

echo Installing Configurator component...

kubectl apply -f %LSF_CONFIGURATOR_NAMESPACE_YAML_PATH%
if errorlevel 1 (
    echo Error creating configurator namespace
    exit /b 1
)

kubectl apply -f %UPLOADS_PVC_YAML_PATH%
if errorlevel 1 (
    echo Error creating uploads PVC
    exit /b 1
)

@REM needed only for private registries, commented out for local registry without auth
@REM call %CREATE_REGISTRY_SECRET_PATH%
@REM if errorlevel 1 (
@REM     echo Error creating registry secret for lsf-configurator
@REM     exit /b 1
@REM )

kubectl apply -f %LSF_CONFIGURATOR_YAML_PATH%
if errorlevel 1 (
    echo Error installing lsf-configurator
    exit /b 1
)

echo Installing Redis...
kubectl apply -f %REDIS_NAMESPACE_YAML_PATH%
if errorlevel 1 (
    echo Error creating Redis namespace
    exit /b 1
)
kubectl apply -f %REDIS_YAML_PATH%
if errorlevel 1 (
    echo Error creating Redis
    exit /b 1
)

echo Installing Tekton Pipelines...
kubectl apply -f %TEKTON_PIPELINES_YAML_PATH%
if errorlevel 1 (
    echo Error installing Tekton Pipelines
    exit /b 1
)
kubectl apply -f %TEKTON_DASHBOARD_YAML_PATH%
if errorlevel 1 (
    echo Error installing Tekton Dashboard
    exit /b 1
)
kubectl apply -f %TEKTON_BUILDPACKS_YAML_PATH% -n configurator
if errorlevel 1 (
    echo Error applying buildpacks phases task
    exit /b 1
)
kubectl apply -f %TEKTON_BUILD_PIPELINE_YAML_PATH%
if errorlevel 1 (
    echo Error applying Tekton function build pipeline
    exit /b 1
)


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
