@echo off
title TormentNexus Watchdog
cd /d "%~dp0"
echo ========================================
echo  TormentNexus Watchdog
echo  Started: %date% %time%
echo  CWD: %cd%
echo ========================================
echo.

:restart
echo [%date% %time%] Starting watchdog...
pythonw -u watchdog.py
echo [%date% %time%] Watchdog exited. Restarting in 5 seconds...
timeout /t 5 /nobreak >nul
goto restart
