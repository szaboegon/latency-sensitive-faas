@echo off

docker build -t szaboegon/lsf-configurator -f ..\Dockerfile ..\
if errorlevel 1 (
    echo Failed to build image with minikube
    exit /b 1
)

docker push szaboegon/lsf-configurator
if errorlevel 1 (
    echo Failed to push image to repository
    exit /b 1
)

kubectl apply -f configurator-namespace.yaml

kubectl apply -f lsf-configurator.yaml
if errorlevel 1 (
    echo Failed to apply kubernetes manifest
    exit /b 1
)

kubectl rollout restart -n configurator deployment lsf-configurator-deployment
timeout /t 30 /nobreak > nul
kubectl port-forward service/lsf-configurator 8081:80 -n configurator 