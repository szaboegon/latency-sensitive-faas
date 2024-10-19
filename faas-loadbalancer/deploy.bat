@echo off

docker build -t szaboegon/faas-loadbalancer -f ./deploy/Dockerfile .
if errorlevel 1 (
    echo Failed to build image with minikube
    exit /b 1
)

docker push szaboegon/faas-loadbalancer
if errorlevel 1 (
    echo Failed to push image to repository
    exit /b 1
)

kubectl apply -f .\deploy\faas-loadbalancer.yaml
if errorlevel 1 (
    echo Failed to apply kubernetes manifest
    exit /b 1
)

kubectl rollout restart -n loadbalancer daemonset faas-loadbalancer-daemonset