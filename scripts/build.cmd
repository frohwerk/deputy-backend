@echo off

setlocal

if "%1"=="" (
    echo Specify one of these programs to build:
    rem dir /AD /B cmd
    pushd cmd
    for /D %%G in (*) do echo - %%G
    popd
    goto $exit
)

go build -o bin/%1.exe ./cmd/%1

:$exit
endlocal
exit /b %errorlevel%
