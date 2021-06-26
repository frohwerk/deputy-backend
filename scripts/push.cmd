@echo off

setlocal

set GOOS=linux

if "%1"=="" (
    goto targets
)

if "%1"=="oidc" (
    go build -o bin/linux/oidc ./cmd/oidc
    if %errorlevel% neq 0 goto end
    docker build -t 172.30.1.1:5000/myproject/oidc:latest -f ./build/oidc/Dockerfile ./bin/linux
    if %errorlevel% neq 0 goto end
    docker push 172.30.1.1:5000/myproject/oidc:latest
    if %errorlevel% neq 0 goto end
    oc delete deployment oidc
    oc apply -f deployments/minishift/02-oidc.yaml
) else if "%1"=="k8smon" (
    go build -o bin/linux/k8smon ./cmd/k8smon
    go build -o bin/linux/k8swatcher ./cmd/k8swatcher
    if %errorlevel% neq 0 goto end
    docker build -t 172.30.1.1:5000/myproject/k8smon:latest -f ./build/k8smon/Dockerfile ./bin/linux
    if %errorlevel% neq 0 goto end
    docker push 172.30.1.1:5000/myproject/k8smon:latest
    if %errorlevel% neq 0 goto end
    oc delete deployment k8smon
    oc apply -f deployments/minishift/03-k8smon.yaml
) else if "%1"=="api-server" (
    echo Buiding ./cmd/server as bin/linux/api-server
    go build -o bin/linux/api-server ./cmd/server
    if %errorlevel% neq 0 goto end
    docker build -t 172.30.1.1:5000/myproject/api-server:latest -f ./build/api-server/Dockerfile ./bin/linux
    if %errorlevel% neq 0 goto end
    docker push 172.30.1.1:5000/myproject/api-server:latest
    if %errorlevel% neq 0 goto end
    oc delete deployment api-server
    oc apply -f deployments/minishift/03-api-server.yaml
) else if "%1"=="rthook" (
    go build -o bin/linux/rthook ./cmd/rthook
    if %errorlevel% neq 0 goto end
    docker build -t 172.30.1.1:5000/myproject/rthook:latest -f ./build/rthook/Dockerfile ./bin/linux
    if %errorlevel% neq 0 goto end
    docker push 172.30.1.1:5000/myproject/rthook:latest
    if %errorlevel% neq 0 goto end
    oc delete deployment rthook
    oc apply -f deployments/minishift/04-rthook.yaml
) else if "%1"=="imgmatch" (
    go build -o bin/linux/imgmatch ./cmd/imgmatch
    if %errorlevel% neq 0 goto end
    docker build -t 172.30.1.1:5000/myproject/imgmatch:latest -f ./build/imgmatch/Dockerfile ./bin/linux
    if %errorlevel% neq 0 goto end
    docker push 172.30.1.1:5000/myproject/imgmatch:latest
    if %errorlevel% neq 0 goto end
    oc delete deployment imgmatch
    oc apply -f deployments/minishift/04-imgmatch.yaml
) else if "%1"=="tasks-server" (
    go build -o bin/linux/tasks-server ./cmd/workshop
    if %errorlevel% neq 0 goto end
    docker build -t 172.30.1.1:5000/myproject/tasks-server:latest -f ./build/tasks-server/Dockerfile ./bin/linux
    if %errorlevel% neq 0 goto end
    docker push 172.30.1.1:5000/myproject/tasks-server:latest
    if %errorlevel% neq 0 goto end
    oc delete deployment tasks-server
    oc apply -f deployments/minishift/04-tasks-server.yaml
) else (
    goto targets
)

goto end

:targets
echo Please specify a build target:
echo - api-server
echo - imgmatch
echo - k8smon
echo - oidc
echo - rthook
echo - tasks-server

:end
endlocal
exit /b 0
