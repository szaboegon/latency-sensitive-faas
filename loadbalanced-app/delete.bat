@echo off
start cmd /c "cd .\func-1 && rmdir .func /s && kn func delete func-1 -n application"
start cmd /c "cd .\func-2 && rmdir .func /s && kn func delete func-2 -n application"
start cmd /c "cd .\func-3 && rmdir .func /s && kn func delete func-3 -n application"
start cmd /c "cd .\func-4 && rmdir .func /s && kn func delete func-4 -n application"
start cmd /c "cd .\func-5 && rmdir .func /s && kn func delete func-5 -n application"
