@echo off

start cmd /c "cd .\imagegrab && kn func deploy -v"
start cmd /c "cd .\resize && kn func deploy -v"
start cmd /c "cd .\grayscale && kn func deploy -v"
start cmd /c "cd .\objectdetect && kn func deploy -v"
start cmd /c "cd .\cut && kn func deploy -v"