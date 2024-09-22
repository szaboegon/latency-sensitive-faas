@echo off
rem If using minikube, make sure that the minikube tunnel command is running, otherwise DNS services won't work

set JAEGER_INSTANCE_PATH="jaeger\allinone_instance.yaml"
set SERVING_TRACE_CONFIG_PATH="knative\serving_trace_config.yaml"
set EVENTING_TRACE_CONFIG_PATH="knative\eventing_trace_config.yaml"

echo Installing prerequisite: Cert Manager
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.6.3/cert-manager.yaml
if errorlevel 1 (
    echo Error installing Cert Manager
    exit /b 1
)

rem Cert if Cert Manager webhooks are ready with cmctl 
cmctl check api --wait=5m

echo Creating observability namespace and installing Jaeger operator...
rem Create namespace
kubectl create namespace observability
if errorlevel 1 (
    echo Error creating namespace observability
    exit /b 1
)

rem Install Jaeger operator CRDs
kubectl create -f jaeger\jaeger_operator.yaml -n observability
if errorlevel 1 (
    echo Error installing Jaeger operator
    exit /b 1
)

echo Jaeger operator installed successfully in the observability namespace.

echo Deploying Jaeger instance into cluster...
rem Retry if deploy fails, it means the operator is not initialized yet
set MAX_RETRIES=10
set RETRY_COUNT=0
:RETRY_JAEGER
kubectl apply -f %JAEGER_INSTANCE_PATH% -n observability
if errorlevel 1 (
    set /a RETRY_COUNT+=1
    if %RETRY_COUNT% lss %MAX_RETRIES% (
        echo Retry %RETRY_COUNT% of %MAX_RETRIES% failed. Retrying...
        timeout /t 10 >nul
        goto RETRY_JAEGER
    ) else (
        echo Failed to deploy Jaeger instance after %MAX_RETRIES% retries
        exit /b 1
    )
)

kubectl apply -f %SERVING_TRACE_CONFIG_PATH%
if errorlevel 1 (
    echo Failed to apply modifications to knative serving configmap
    exit /b 1
)

kubectl apply -f %EVENTING_TRACE_CONFIG_PATH%
if errorlevel 1 (
    echo Failed to apply modifications to knative eventing configmap
    exit /b 1
)

echo Port forwarding Jaeger UI to localhost port 16686
set MAX_RETRIES=10
set RETRY_COUNT=0
:RETRY_PORT_FORWARD
kubectl port-forward -n observability  deployments/my-jaeger 16686
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
