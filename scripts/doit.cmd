 @echo off
 
 setlocal
 set GOOS=linux
 
 go build -o bin\linux ./cmd/oidc
 if %errorlevel% neq 0 goto end
 
 docker build -t 172.30.1.1:5000/myproject/oidc:latest -f ./build/oidc/Dockerfile ./bin/linux
 if %errorlevel% neq 0 goto end
 
 docker push 172.30.1.1:5000/myproject/oidc:latest
 if %errorlevel% neq 0 goto end
 
 oc delete deployment oidc
 if %errorlevel% neq 0 goto end
 
 oc apply -f deployments\minishift\02-oidc.yaml
 if %errorlevel% neq 0 goto end

:end
endlocal
 