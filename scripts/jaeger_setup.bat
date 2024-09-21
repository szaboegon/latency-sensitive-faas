@echo off
rem If using minikube, make sure that the minikube tunnel command is running, otherwise DNS services won't work

set JAEGER_INSTANCE_PATH="jaeger\allinone_instance.yaml"
set MAX_RETRIES=5
set RETRY_COUNT=0

echo Installing prerequisite: Cert Manager
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.6.3/cert-manager.yaml
if errorlevel 1 (
    echo Error installing Cert Manager
    exit /b 1
)

rem Cert if Cert Manager webhooks are ready with cmctl 
cmctl check api --wait=2m

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
kubectl apply -f %JAEGER_INSTANCE_PATH%
if errorlevel 1 (
    echo Failed to deploy Jaeger instance
    exit /b 1
)

exit /b 0
