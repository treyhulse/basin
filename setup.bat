@echo off
echo Starting Go RBAC API Setup...
powershell -ExecutionPolicy Bypass -File "setup.ps1" %*
pause 