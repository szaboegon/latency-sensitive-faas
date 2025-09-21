@echo off

start cmd /c "cd .\imagegrab && llf deploy -v --build"
start cmd /c "cd .\resize && llf deploy -v --build"
start cmd /c "cd .\grayscale && llf deploy -v --build"
start cmd /c "cd .\objectdetect && llf deploy -v --build"
start cmd /c "cd .\cut && llf deploy -v --build"
start cmd /c "cd .\objectdetect2 && llf deploy -v --build"
start cmd /c "cd .\tag && llf deploy -v --build"

@REM cd .\imagegrab
@REM llf deploy -v --build
@REM cd ..\resize
@REM llf deploy -v --build
@REM cd ..\grayscale
@REM llf deploy -v --build
@REM cd ..\objectdetect
@REM llf deploy -v --build
@REM cd ..\cut
@REM llf deploy -v --build"
@REM cd ..\objectdetect2 
@REM llf deploy -v --build
@REM cd ..\tag 
@REM llf deploy -v --build