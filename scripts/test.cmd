@echo off

setlocal

pushd test
if %errorlevel% neq 0 goto end

docker compose up --build --quiet-pull --force-recreate --renew-anon-volumes --detach database
if %errorlevel% neq 0 goto back

if "%1"=="" (
    @echo on
    go clean -testcache
    go test .
    @echo off
) else (
    @echo on
    go clean -testcache
    go test . -run %1 -v
    @echo off
)

:cleanup
docker compose down database

:back
popd

:end
endlocal
exit /b 0
