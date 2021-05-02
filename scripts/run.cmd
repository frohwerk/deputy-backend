@echo off

setlocal
set SCRIPTPATH=%~dp0

call %SCRIPTPATH%build %*

bin\%1.exe

:$exit
endlocal
exit /b %errorlevel%
