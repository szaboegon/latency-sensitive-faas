@echo off
set SERVING_FEATURES_CONFIG_PATH="..\kubernetes\knative\serving_features.yaml"
kubectl apply -f %SERVING_FEATURES_CONFIG_PATH%

cd .\func-1
lsfunc deploy -v 
cd ..\func-2
lsfunc deploy -v 
cd ..\func-3
lsfunc deploy -v 
cd ..\func-4
lsfunc deploy -v 
cd ..\func-5
lsfunc deploy -v 

@REM start cmd /c "cd .\func-1 && lsfunc deploy -v --build"
@REM start cmd /c "cd .\func-2 && lsfunc deploy -v --build"
@REM start cmd /c "cd .\func-3 && lsfunc deploy -v --build"
@REM start cmd /c "cd .\func-4 && lsfunc deploy -v --build"
@REM start cmd /c "cd .\func-5 && lsfunc deploy -v --build"
