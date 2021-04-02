@echo off
start docker-compose up --build --force-recreate
echo Press any key to stop the container
pause >nul
docker-compose down
