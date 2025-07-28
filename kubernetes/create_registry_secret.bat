@echo off
set /p username=Enter your Docker registry username:
set /p password=Enter your Docker registry password:

kubectl create secret docker-registry registry-user-pass ^
    --docker-username="%username%" ^
    --docker-password="%password%" ^
    --docker-server="https://index.docker.io/v1/" ^
    --namespace=configurator

echo Secret created successfully.
pause