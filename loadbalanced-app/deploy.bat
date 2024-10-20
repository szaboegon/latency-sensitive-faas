@echo off

cd .\func-1
lsfunc deploy -v --build
cd .\func-2
lsfunc deploy -v --build
cd .\func-3
lsfunc deploy -v --build
cd .\func-4
lsfunc deploy -v --build
cd .\func-5
lsfunc deploy -v --build
