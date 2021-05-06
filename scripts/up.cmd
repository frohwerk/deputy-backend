@echo off

if "%1"=="" goto noargs

setlocal

:: prefer using windows terminal (new-tabs) if availabe
where wt 1> nul 2> nul
if errorlevel 1 (
    echo Windows Terminal not on the path. Using fallback: cmd
    set term=start
    :: reset errorlevel
    ver > nul
) else (
    set term=wt nt -d . --title
)

if "%1"=="imgmatch" (
    go build -o bin/%1.exe ./cmd/%1
    if %errorlevel% neq 0 goto end
    %term% "%1" cmd /c bin\%1.exe --port 8092
) else if "%1"=="rthook" (
    go build -o bin/%1.exe ./cmd/%1
    if %errorlevel% neq 0 goto end
    %term% "%1" cmd /c bin\%1.exe --port 8082 --artifactory http://localhost:8091/libs-release-local
) else if "%1"=="server" (
    go build -o bin/%1.exe ./cmd/%1
    if %errorlevel% neq 0 goto end
    %term% "%1" cmd /c bin\%1.exe --port 8080
) else if "%1"=="mocktifactory" (
    %term% "%1" cmd /c ..\mocktifactory\bin\m.exe
) else if "%1"=="k8smon" (
    echo Building k8swatcher...
    go build -o bin/k8swatcher.exe ./cmd/k8swatcher
    if %errorlevel% neq 0 goto end
    echo Building k8smon...
    go build -o bin/%1.exe ./cmd/%1
    if %errorlevel% neq 0 goto end
    %term% "%1" cmd /c bin\%1.exe
) else if "%1"=="all" (
    call %~dp0up.cmd imgmatch
    call %~dp0up.cmd rthook
    call %~dp0up.cmd server
    call %~dp0up.cmd k8smon
    call %~dp0up.cmd mocktifactory
) else (
    go build -o bin/%1.exe ./cmd/%1
    if %errorlevel% neq 0 goto end
    %term% "%1" cmd /c bin\%1.exe
)

:end
endlocal
exit /b 0

:noargs
echo Please specify which component to start
goto end
