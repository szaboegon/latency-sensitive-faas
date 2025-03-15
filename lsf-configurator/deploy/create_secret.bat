@echo off
set /p username=Enter your Docker registry username:
set /p password=Enter your Docker registry password:

kubectl create secret generic registry-secret ^
    --from-literal=REGISTRY_USER="%username%" ^
    --from-literal=REGISTRY_PASSWORD="%password%" ^
    --namespace=configurator

echo Secret created successfully.
pause