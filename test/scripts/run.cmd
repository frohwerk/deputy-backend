@echo off

if "%1"=="" (
    docker compose up --build --force-recreate --renew-anon-volumes --exit-code-from go-test 
) else (
    docker compose up --build --force-recreate --renew-anon-volumes --exit-code-from go-test 
)

