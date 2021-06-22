@echo off

setlocal

if "%1"=="" (
    echo Please specify an application name
    goto end
)

set "at=now"
if "%2" neq "" (
    set "at=%2"
)

set "source=e7ccea48-c007-4ff5-b2fb-74516e77da00"
if "%3" neq "" (
    set "source=%3"
)

set "target=c8f1d2d6-8305-48d6-a613-23cdb67b5a19"
if "%4" neq "" (
    set "target=%4"
)

go run ./cmd/workshop %1 %at% %source% %target%

:end
endlocal
exit /b 0

