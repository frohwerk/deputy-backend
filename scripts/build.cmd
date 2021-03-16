@echo off
go build -o build/deputy-k8s-importer.exe ./cmd/importer
go build -o build/deputy-api-server.exe ./cmd/server
