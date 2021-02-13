@echo off
setlocal

set GOOS=linux
set GOARCH=amd64

echo Clean build directory...
del /Q build\*

echo Build %GOOS%-%GOARCH% binary...
go build -o build cmd\workshop\main.go

move build\main build\app

echo Build docker image...
docker build -t 172.30.1.1:5000/myproject/go-hello-world .
echo Push docker image...
docker push 172.30.1.1:5000/myproject/go-hello-world

echo Update pods...
kubectl delete deployment go-hello-world
kubectl apply -f deployments\go-hello-world.deployment.yaml

echo Done!

endlocal
