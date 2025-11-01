@echo off
set NUM_NODES=3
set MEMORY_LIMIT="6g"
set CPUS=4

set SERVING_YAML_PATH="knative\serving.yaml"
set SCALE_CONFIG_PATH="knative\disable_scale_to_zero.yaml"
set APP_NAMESPACE_PATH="otel\application_namespace.yaml"
set METRICS_SERVER_PATH="metrics-server\metrics-server.yaml"
set STORAGECLASS_YAML_PATH="storageclass.yaml"
set ISTIO_ENVOY_FILTER_PATH="istio\envoy_filter.yaml"

echo %cd%
echo Starting minikube cluster with %NUM_NODES% nodes...
minikube start --nodes %NUM_NODES% -p knative --memory %MEMORY_LIMIT% --cpus %CPUS% --addons=ingress --driver=docker
if errorlevel 1 (
    echo Failed to start Minikube cluster.
    exit /b 1
)

echo Configuring addons for elasticsearch to work properly...
minikube addons disable storage-provisioner -p knative
if errorlevel 1 (
    echo Failed to disable storage-provisioner addon.
    exit /b 1
)

minikube addons disable default-storageclass -p knative
if errorlevel 1 (
    echo Failed to disable default-storageclass addon.
    exit /b 1
)

minikube addons enable volumesnapshots -p knative
if errorlevel 1 (
    echo Failed to enable volumesnapshots addon.
    exit /b 1
)

minikube addons enable csi-hostpath-driver -p knative
if errorlevel 1 (
    echo Failed to enable csi-hostpath-driver addon.
    exit /b 1
)

kubectl patch storageclass csi-hostpath-sc --patch-file storageclass_patch.json
if errorlevel 1 (
    echo Failed to patch storageclass csi-hostpath-sc.
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

istioctl verify-install
if errorlevel 1 (
    echo Failed to verify networking layer.
    exit /b 1
)

echo Configuring DNS...
kubectl apply -f https://github.com/knative/serving/releases/download/knative-v1.15.2/serving-default-domain.yaml
if errorlevel 1 (
    echo Failed to configure DNS.
    exit /b 1
)

kubectl apply -f %SCALE_CONFIG_PATH%
if errorlevel 1 (
    echo Failed to modify knative scale config.
    exit /b 1
)

kubectl apply -f %APP_NAMESPACE_PATH%
if errorlevel 1 (
    echo Failed to create namespace 'application'.
    exit /b 1
)

kubectl apply -f %METRICS_SERVER_PATH%
if errorlevel 1 (
    echo Failed to install metrics-server.
    exit /b 1
)

kubectl apply -f %STORAGECLASS_YAML_PATH%
if errorlevel 1 (
    echo Failed to install storage class.
    exit /b 1
)

echo Adding Istio envoy filter...
kubectl apply -f %ISTIO_ENVOY_FILTER_PATH%
if errorlevel 1 (
    echo Failed to apply Istio envoy filter.
    exit /b 1
)

echo Installation complete. Starting tunneling for minikube. Keep this console open for DNS services to work.
minikube tunnel -p knative
