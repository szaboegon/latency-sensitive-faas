@echo off
set NUM_NODES=2
set SERVING_YAML_PATH="knative\knative_serving.yaml"

echo %cd%
echo Starting minikube cluster with %NUM_NODES% nodes...
minikube start --nodes %NUM_NODES% -p knative --addons=ingress

if errorlevel 1 (
    echo Failed to start Minikube cluster.
    exit /b 1
)

echo Installing knative operator...
kubectl apply -f https://github.com/knative/operator/releases/download/knative-v1.15.4/operator.yaml
if errorlevel 1 (
    echo Failed to install knative operator.
    exit /b 1
)

echo Installing knative serving component...
kubectl apply -f %SERVING_YAML_PATH%
if errorlevel 1 (
    echo Failed to install serving component.
    exit /b 1
)

echo Installing istio networking layer...
istioctl install -y
if errorlevel 1 (
    echo Failed to install istio.
    exit /b 1
)

kubectl apply -f https://github.com/knative/net-istio/releases/download/knative-v1.15.1/net-istio.yaml
if errorlevel 1 (
    echo Failed to apply networking layer.
    exit /b 1
)

echo Configuring DNS...
kubectl apply -f https://github.com/knative/serving/releases/download/knative-v1.15.2/serving-default-domain.yaml
if errorlevel 1 (
    echo Failed to configure DNS.
    exit /b 1
)

echo Installation complete. Starting tunneling for minikube. Keep this console open for DNS services to work.
minikube tunnel -p knative
